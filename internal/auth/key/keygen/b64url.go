package keygen

import (
	"crypto/rand"
	"encoding/base64"
)

type usingB64url struct {
	bytes int
}

func NewUsingB64url(bytes int) usingB64url {
	return usingB64url{
		bytes: bytes,
	}
}

func (u usingB64url) Generate() string {
	keyBytes := make([]byte, u.bytes)
	rand.Read(keyBytes)
	return base64.RawURLEncoding.EncodeToString(keyBytes)
}
