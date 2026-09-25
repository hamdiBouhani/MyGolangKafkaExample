package schema

import (
	"fmt"
	"time"

	"github.com/riferrei/srclient"
)

// Registry is a thin, testable wrapper around srclient.
type Registry struct {
	client  *srclient.SchemaRegistryClient
	timeout time.Duration
}

func NewRegistry(url string, timeout time.Duration) *Registry {
	return &Registry{
		client:  srclient.CreateSchemaRegistryClient(url),
		timeout: timeout,
	}
}

// Register ensures the schema exists under subject and returns its ID.
// Idempotent: if the schema is already registered, the existing ID is returned.
func (r *Registry) Register(subject, schema string) (int, error) {
	// srclient uses an internal http.Client; we set a timeout defensively.
	r.client.SetTimeout(r.timeout)

	s, err := r.client.CreateSchema(subject, schema, srclient.Avro)
	if err != nil {
		return 0, fmt.Errorf("register schema %q: %w", subject, err)
	}
	return s.ID(), nil
}
