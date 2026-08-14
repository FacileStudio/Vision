package authcrypto

import (
	"crypto/rand"
	"encoding/base64"
)

// NewToken returns a random 256-bit token encoded as unpadded base64url,
// suitable as an opaque credential.
func NewToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(bytes), nil
}
