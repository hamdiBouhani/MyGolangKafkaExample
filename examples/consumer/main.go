package main

import (
	"context"
	"log"

	"github.com/IBM/sarama"
	"github.com/linkedin/goavro"
	"github.com/riferrei/srclient"
)

const (
	brokers           = "localhost:9092"
	topic             = "user-events"
	groupID           = "user-events-consumer"
	schemaRegistryURL = "http://localhost:8081"
)

func main() {
	client := srclient.CreateSchemaRegistryClient(schemaRegistryURL)

	config := sarama.NewConfig()
	config.Consumer.Return.Errors = true
	config.Version = sarama.V2_8_0_0

	consumer, err := sarama.NewConsumerGroup([]string{brokers}, groupID, config)
	if err != nil {
		log.Fatalf("failed to create consumer group: %v", err)
	}

	handler := &AvroHandler{sr: client}
	for {
		ctx := context.Background()
		err := consumer.Consume(ctx, []string{topic}, handler)

		if err != nil {
			log.Printf("consume error: %v", err)
		}
	}
}

type AvroHandler struct {
	sr *srclient.SchemaRegistryClient
}

func (h *AvroHandler) Setup(_ sarama.ConsumerGroupSession) error   { return nil }
func (h *AvroHandler) Cleanup(_ sarama.ConsumerGroupSession) error { return nil }

func (h *AvroHandler) ConsumeClaim(sess sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {

		// Extract schema ID from Confluent wire format
		schemaID := int(msg.Value[1])<<24 |
			int(msg.Value[2])<<16 |
			int(msg.Value[3])<<8 |
			int(msg.Value[4])

		schema, err := h.sr.GetSchema(schemaID)
		if err != nil {
			log.Printf("schema fetch error: %v", err)
			continue
		}

		codec, err := goavro.NewCodec(schema.Schema())
		if err != nil {
			log.Printf("codec error: %v", err)
			continue
		}

		native, _, err := codec.NativeFromBinary(msg.Value[5:])
		if err != nil {
			log.Printf("decode error: %v", err)
			continue
		}

		log.Printf("Received Avro event: %+v", native)

		sess.MarkMessage(msg, "")
	}
	return nil
}
