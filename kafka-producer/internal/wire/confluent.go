package wire

import "encoding/binary"

// MagicByte is the Confluent Schema Registry magic byte (always 0).
const MagicByte byte = 0x0

// Encode wraps an Avro payload in Confluent wire format:
//
//	[ magic byte (1) ][ schema ID (4, big-endian) ][ avro payload ]
func Encode(schemaID int, payload []byte) []byte {
	out := make([]byte, 5+len(payload))
	out[0] = MagicByte
	binary.BigEndian.PutUint32(out[1:5], uint32(schemaID))
	copy(out[5:], payload)
	return out
}

// Decode extracts the schema ID and payload from a Confluent-framed message.
// Useful for consumers and tests.
func Decode(b []byte) (schemaID int, payload []byte, ok bool) {
	if len(b) < 5 || b[0] != MagicByte {
		return 0, nil, false
	}
	id := int(binary.BigEndian.Uint32(b[1:5]))
	return id, b[5:], true
}
