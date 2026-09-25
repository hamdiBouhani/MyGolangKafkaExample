package producer

import (
	"fmt"
	"os"
	"time"

	"github.com/IBM/sarama"

	"github.com/hamdiBouhani/MyGolangKafkaExample/kafka-producer/internal/config"
)

func readSchema(path string) (string, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("read schema file %q: %w", path, err)
	}
	return string(b), nil
}

func buildSaramaConfig(cfg *config.Config) (*sarama.Config, error) {
	c := sarama.NewConfig()
	c.ClientID = cfg.ClientID
	c.Version = sarama.V3_5_0_0
	c.Producer.Return.Successes = true
	c.Producer.Return.Errors = true
	c.Producer.Retry.Max = cfg.RetryMax
	c.Producer.Retry.Backoff = cfg.RetryBackoff
	c.Producer.Flush.Frequency = cfg.FlushFrequency
	c.Producer.Flush.Messages = cfg.FlushMessages
	c.Producer.Flush.Bytes = cfg.FlushBytes
	c.Producer.Timeout = 10 * time.Second

	switch cfg.Acks {
	case "all":
		c.Producer.RequiredAcks = sarama.WaitForAll
	case "local":
		c.Producer.RequiredAcks = sarama.WaitForLocal
	case "none":
		c.Producer.RequiredAcks = sarama.NoResponse
	default:
		return nil, fmt.Errorf("unknown acks value %q", cfg.Acks)
	}

	if err := c.Validate(); err != nil {
		return nil, fmt.Errorf("sarama config: %w", err)
	}
	return c, nil
}
