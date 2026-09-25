package producer

import (
	"context"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/IBM/sarama"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"go.uber.org/zap"

	"github.com/hamdiBouhani/MyGolangKafkaExample/kafka-producer/internal/codec"
	"github.com/hamdiBouhani/MyGolangKafkaExample/kafka-producer/internal/config"
	"github.com/hamdiBouhani/MyGolangKafkaExample/kafka-producer/internal/schema"
	"github.com/hamdiBouhani/MyGolangKafkaExample/kafka-producer/internal/wire"
)

var (
	messagesSent = promauto.NewCounter(prometheus.CounterOpts{
		Name: "producer_messages_sent_total",
		Help: "Total number of successfully produced messages.",
	})
	messagesFailed = promauto.NewCounter(prometheus.CounterOpts{
		Name: "producer_messages_failed_total",
		Help: "Total number of failed produce attempts.",
	})
	sendLatency = promauto.NewHistogram(prometheus.HistogramOpts{
		Name:    "producer_send_duration_seconds",
		Help:    "Latency of producing a single message.",
		Buckets: prometheus.DefBuckets,
	})
)

// Service is the production-grade Kafka producer.
type Service struct {
	cfg      *config.Config
	log      *zap.Logger
	sarama   sarama.SyncProducer
	codec    *codec.AvroCodec
	registry *schema.Registry
	schemaID int
	closed   atomic.Bool
}

// New wires together config, logging, Schema Registry, Avro codec, and Sarama.
func New(cfg *config.Config, log *zap.Logger) (*Service, error) {
	// 1. Schema Registry
	reg := schema.NewRegistry(cfg.SchemaRegistryURL, cfg.SchemaTimeout)

	// 2. Avro codec
	schemaStr, err := readSchema(cfg.SchemaPath)
	if err != nil {
		return nil, err
	}
	avroCodec, err := codec.NewAvroCodec(schemaStr)
	if err != nil {
		return nil, err
	}

	// 3. Register schema, get ID
	schemaID, err := reg.Register(cfg.SchemaSubject, schemaStr)
	if err != nil {
		return nil, err
	}
	log.Info("schema registered",
		zap.String("subject", cfg.SchemaSubject),
		zap.Int("schema_id", schemaID),
	)

	// 4. Sarama producer
	saramaCfg, err := buildSaramaConfig(cfg)
	if err != nil {
		return nil, err
	}
	sp, err := sarama.NewSyncProducer(cfg.Brokers, saramaCfg)
	if err != nil {
		return nil, fmt.Errorf("create sarama producer: %w", err)
	}

	return &Service{
		cfg:      cfg,
		log:      log,
		sarama:   sp,
		codec:    avroCodec,
		registry: reg,
		schemaID: schemaID,
	}, nil
}

// ProduceUserEvent encodes and sends a single UserEvent.
func (s *Service) ProduceUserEvent(ctx context.Context, evt codec.UserEvent) error {
	if s.closed.Load() {
		return fmt.Errorf("producer is closed")
	}

	avroPayload, err := s.codec.Encode(evt)
	if err != nil {
		messagesFailed.Inc()
		return err
	}

	framed := wire.Encode(s.schemaID, avroPayload)

	msg := &sarama.ProducerMessage{
		Topic: s.cfg.Topic,
		Key:   sarama.StringEncoder(evt.UserID), // key by user_id for partition affinity
		Value: sarama.ByteEncoder(framed),
		Headers: []sarama.RecordHeader{
			{Key: []byte("content-type"), Value: []byte("application/avro")},
		},
	}

	start := time.Now()
	partition, offset, err := s.sarama.SendMessage(msg)
	sendLatency.Observe(time.Since(start).Seconds())

	if err != nil {
		messagesFailed.Inc()
		return fmt.Errorf("send message: %w", err)
	}

	messagesSent.Inc()
	s.log.Info("message produced",
		zap.String("topic", s.cfg.Topic),
		zap.Int32("partition", partition),
		zap.Int64("offset", offset),
		zap.String("user_id", evt.UserID),
		zap.String("event_type", evt.EventType),
	)
	return nil
}

// Close flushes and closes the producer. Safe to call multiple times.
func (s *Service) Close() error {
	if !s.closed.CompareAndSwap(false, true) {
		return nil
	}
	return s.sarama.Close()
}
