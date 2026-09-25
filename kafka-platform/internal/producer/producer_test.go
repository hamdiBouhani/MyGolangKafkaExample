package producer

import (
	"context"
	"testing"
	"time"

	"github.com/IBM/sarama"
	"github.com/IBM/sarama/mocks"
	"go.uber.org/zap"

	"github.com/hamdiBouhani/MyGolangKafkaExample/kafka-platform/internal/codec"
	"github.com/hamdiBouhani/MyGolangKafkaExample/kafka-platform/internal/config"
)

const testSchema = `{
  "type": "record",
  "name": "UserEvent",
  "namespace": "com.example",
  "fields": [
    {"name": "user_id", "type": "string"},
    {"name": "event_type", "type": "string"},
    {"name": "timestamp", "type": "long"}
  ]
}`

func TestService_ProduceUserEvent_EncodesConfluentWireFormat(t *testing.T) {
	mockProducer := mocks.NewSyncProducer(t, nil)
	mockProducer.ExpectSendMessageAndSucceed()

	avroCodec, err := codec.NewAvroCodec(testSchema)
	if err != nil {
		t.Fatalf("codec: %v", err)
	}

	svc := &Service{
		cfg: &config.Config{
			Topic: "user-events",
		},
		log:      zap.NewNop(),
		sarama:   mockProducer,
		codec:    avroCodec,
		schemaID: 7,
	}

	evt := codec.UserEvent{
		UserID:    "user-1",
		EventType: "LOGIN",
		Timestamp: time.Now().Unix(),
	}
	if err := svc.ProduceUserEvent(context.Background(), evt); err != nil {
		t.Fatalf("produce: %v", err)
	}
	if err := mockProducer.Close(); err != nil {
		t.Fatalf("close mock: %v", err)
	}
}

func TestService_ProduceUserEvent_PropagatesSendError(t *testing.T) {
	mockProducer := mocks.NewSyncProducer(t, nil)
	mockProducer.ExpectSendMessageAndFail(sarama.ErrOutOfBrokers)

	avroCodec, err := codec.NewAvroCodec(testSchema)
	if err != nil {
		t.Fatalf("codec: %v", err)
	}

	svc := &Service{
		cfg:      &config.Config{Topic: "user-events"},
		log:      zap.NewNop(),
		sarama:   mockProducer,
		codec:    avroCodec,
		schemaID: 1,
	}

	err = svc.ProduceUserEvent(context.Background(), codec.UserEvent{
		UserID: "u", EventType: "LOGIN", Timestamp: 1,
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
