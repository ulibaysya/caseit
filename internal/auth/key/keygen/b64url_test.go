package keygen

import (
	"testing"
)

func TestUsingB64urlGenerate(t *testing.T) {
	const bytes = 32

	key := NewUsingB64url(bytes).Generate()

	const encodedKeyLen = 43
	if keyLen := len(key); keyLen != encodedKeyLen {
		t.Fatalf("unexpected key length %v; expected %v", keyLen, encodedKeyLen)
	}
}
