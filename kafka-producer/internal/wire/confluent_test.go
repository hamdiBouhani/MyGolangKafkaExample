package wire

import (
	"bytes"
	"testing"
)

func TestEncodeDecodeRoundTrip(t *testing.T) {
	payload := []byte{0xDE, 0xAD, 0xBE, 0xEF}
	id := 42

	framed := Encode(id, payload)
	if len(framed) != 5+len(payload) {
		t.Fatalf("expected len %d, got %d", 5+len(payload), len(framed))
	}
	if framed[0] != MagicByte {
		t.Fatalf("expected magic byte 0, got %d", framed[0])
	}

	gotID, gotPayload, ok := Decode(framed)
	if !ok {
		t.Fatal("decode returned !ok")
	}
	if gotID != id {
		t.Fatalf("expected id %d, got %d", id, gotID)
	}
	if !bytes.Equal(gotPayload, payload) {
		t.Fatalf("payload mismatch: %x vs %x", gotPayload, payload)
	}
}

func TestDecodeRejectsBadInput(t *testing.T) {
	cases := [][]byte{
		nil,
		{},
		{1, 2, 3},       // too short
		{1, 0, 0, 0, 1}, // wrong magic byte
	}
	for i, c := range cases {
		if _, _, ok := Decode(c); ok {
			t.Fatalf("case %d: expected !ok", i)
		}
	}
}
