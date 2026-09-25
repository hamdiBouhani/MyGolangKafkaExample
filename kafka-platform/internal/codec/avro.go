package codec

import (
	"fmt"

	"github.com/linkedin/goavro/v2"
)

// UserEvent is the domain type. Add fields here as the schema evolves.
type UserEvent struct {
	UserID    string
	EventType string
	Timestamp int64
}

// ToNative converts the domain type into the map goavro expects.
// Keep this in sync with the Avro schema.
func (e UserEvent) ToNative() map[string]interface{} {
	return map[string]interface{}{
		"user_id":    e.UserID,
		"event_type": e.EventType,
		"timestamp":  e.Timestamp,
	}
}

// AvroCodec wraps a goavro codec with a typed API.
type AvroCodec struct {
	codec *goavro.Codec
}

// NewAvroCodec compiles the given Avro schema.
func NewAvroCodec(schema string) (*AvroCodec, error) {
	c, err := goavro.NewCodec(schema)
	if err != nil {
		return nil, fmt.Errorf("compile avro schema: %w", err)
	}
	return &AvroCodec{codec: c}, nil
}

// Encode serializes a UserEvent into Avro binary.
func (a *AvroCodec) Encode(e UserEvent) ([]byte, error) {
	b, err := a.codec.BinaryFromNative(nil, e.ToNative())
	if err != nil {
		return nil, fmt.Errorf("avro encode: %w", err)
	}
	return b, nil
}
