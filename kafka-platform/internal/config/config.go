package config

import (
	"fmt"
	"time"

	"github.com/caarlos0/env/v11"
)

// Config holds all runtime configuration for the producer.
// Values are loaded from environment variables (12-factor).
type Config struct {
	// Kafka
	Brokers        []string      `env:"KAFKA_BROKERS" envSeparator:"," envDefault:"localhost:9092"`
	Topic          string        `env:"KAFKA_TOPIC" envDefault:"user-events"`
	ClientID       string        `env:"KAFKA_CLIENT_ID" envDefault:"user-event-producer"`
	Acks           string        `env:"KAFKA_ACKS" envDefault:"all"` // all | local | none
	RetryMax       int           `env:"KAFKA_RETRY_MAX" envDefault:"5"`
	RetryBackoff   time.Duration `env:"KAFKA_RETRY_BACKOFF" envDefault:"250ms"`
	FlushFrequency time.Duration `env:"KAFKA_FLUSH_FREQUENCY" envDefault:"50ms"`
	FlushMessages  int           `env:"KAFKA_FLUSH_MESSAGES" envDefault:"100"`
	FlushBytes     int           `env:"KAFKA_FLUSH_BYTES" envDefault:"1048576"` // 1 MiB

	// Schema Registry
	SchemaRegistryURL string        `env:"SCHEMA_REGISTRY_URL" envDefault:"http://localhost:8081"`
	SchemaPath        string        `env:"SCHEMA_PATH" envDefault:"schemas/user_event.avsc"`
	SchemaSubject     string        `env:"SCHEMA_SUBJECT"` // defaults to <topic>-value if empty
	SchemaTimeout     time.Duration `env:"SCHEMA_TIMEOUT" envDefault:"10s"`

	// Observability
	LogLevel    string `env:"LOG_LEVEL" envDefault:"info"` // debug | info | warn | error
	MetricsAddr string `env:"METRICS_ADDR" envDefault:":9090"`

	// Runtime
	ShutdownTimeout time.Duration `env:"SHUTDOWN_TIMEOUT" envDefault:"15s"`
	NumMessages     int           `env:"NUM_MESSAGES" envDefault:"1"` // send N messages then exit; 0 = infinite
}

// Load parses env vars into Config and validates it.
func Load() (*Config, error) {
	cfg := &Config{}
	if err := env.Parse(cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	if cfg.SchemaSubject == "" {
		cfg.SchemaSubject = cfg.Topic + "-value"
	}
	return cfg, nil
}

func (c *Config) Validate() error {
	if len(c.Brokers) == 0 {
		return fmt.Errorf("KAFKA_BROKERS must not be empty")
	}
	if c.Topic == "" {
		return fmt.Errorf("KAFKA_TOPIC must not be empty")
	}
	switch c.Acks {
	case "all", "local", "none":
	default:
		return fmt.Errorf("KAFKA_ACKS must be one of all|local|none, got %q", c.Acks)
	}
	if c.SchemaRegistryURL == "" {
		return fmt.Errorf("SCHEMA_REGISTRY_URL must not be empty")
	}
	if c.SchemaPath == "" {
		return fmt.Errorf("SCHEMA_PATH must not be empty")
	}
	return nil
}
