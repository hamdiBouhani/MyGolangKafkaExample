package main

import (
	"log"
	"os"
	"time"

	"github.com/IBM/sarama"
	"github.com/linkedin/goavro"
	"github.com/riferrei/srclient"
)

const (
	brokers           = "localhost:9092"
	topic             = "user-events"
	schemaRegistryURL = "http://localhost:8081"
)

func main() {
	// Load Avro schema
	schemaBytes, err := os.ReadFile("../schemas/user_event.avsc")
	if err != nil {
		log.Fatalf("failed to read schema: %v", err)
	}
	schema := string(schemaBytes)

	// Schema Registry client
	client := srclient.CreateSchemaRegistryClient(schemaRegistryURL)
	registeredSchema, err := client.CreateSchema(topic+"-value", schema, srclient.Avro)
	if err != nil {
		log.Fatalf("failed to register schema: %v", err)
	}

	// Avro codec
	codec, err := goavro.NewCodec(schema)
	if err != nil {
		log.Fatalf("failed to create codec: %v", err)
	}

	// Sarama producer
	config := sarama.NewConfig()
	config.Producer.Return.Successes = true

	producer, err := sarama.NewSyncProducer([]string{brokers}, config)
	if err != nil {
		log.Fatalf("failed to create producer: %v", err)
	}
	defer producer.Close()

	// Build Avro message
	native := map[string]interface{}{
		"user_id":    "user-123",
		"event_type": "LOGIN",
		"timestamp":  time.Now().Unix(),
	}

	binary, err := codec.BinaryFromNative(nil, native)
	if err != nil {
		log.Fatalf("failed to encode avro: %v", err)
	}

	id := registeredSchema.ID()

	// Confluent wire format: magic byte + schema ID + payload
	// Magic byte is 0, schema ID is 4 bytes, followed by the Avro payload
	// [ magic byte ][ schema ID (4 bytes) ][ avro payload ]
	// | 0 | schemaID byte1 | schemaID byte2 | schemaID byte3 | schemaID byte4 | avro payload... |
	wire := make([]byte, 5+len(binary))
	wire[0] = 0
	wire[1] = byte(id >> 24)
	wire[2] = byte(id >> 16)
	wire[3] = byte(id >> 8)
	wire[4] = byte(id)
	copy(wire[5:], binary)

	msg := &sarama.ProducerMessage{
		Topic: topic,
		Value: sarama.ByteEncoder(wire),
	}

	partition, offset, err := producer.SendMessage(msg)
	if err != nil {
		log.Fatalf("failed to send message: %v", err)
	}

	log.Printf("Message sent to partition %d at offset %d", partition, offset)
}
