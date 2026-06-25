package hash

import (
	"bytes"
	"testing"
)

func TestHashBlake2b256(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		data []byte
	}{
		{"empty", []byte("")},
		{"text", []byte("hello world")},
		{"binary", []byte{0x00, 0x01, 0x02}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			h1 := HashBlake2b256(tt.data)
			h2 := HashBlake2b256(tt.data)

			if len(h1) != 32 {
				t.Errorf("expected 32 bytes, got %d", len(h1))
			}

			if !bytes.Equal(h1, h2) {
				t.Error("hash should be deterministic")
			}
		})
	}
}

func TestHMACBlake2b256(t *testing.T) {
	t.Parallel()
	key := []byte("secret-key")

	tests := []struct {
		name string
		data []byte
	}{
		{"hello", []byte("hello")},
		{"empty", []byte("")},
		{"golang", []byte("golang hmac test")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			h1 := HMACBlake2b256(tt.data, key)
			h2 := HMACBlake2b256(tt.data, key)

			if !bytes.Equal(h1, h2) {
				t.Error("HMAC should be deterministic")
			}

			if len(h1) != 32 {
				t.Errorf("expected 32 bytes HMAC, got %d", len(h1))
			}
		})
	}
}

func TestVerifyHMACBlake2b256(t *testing.T) {
	t.Parallel()
	key := []byte("secret-key")
	data := []byte("hello world")

	sign := HMACBlake2b256Hex(data, key)

	if !VerifyHMACBlake2b256(data, key, sign) {
		t.Error("expected verification to pass")
	}

	if VerifyHMACBlake2b256([]byte("tampered"), key, sign) {
		t.Error("expected verification to fail")
	}

	if VerifyHMACBlake2b256(data, key, "invalid-hex-string") {
		t.Error("expected invalid hex to fail")
	}
}
