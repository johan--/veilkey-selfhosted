package crypto

import (
	"crypto/rand"
	"crypto/sha256"

	"golang.org/x/crypto/pbkdf2"
)

const (
	SaltSize      = 32
	KDFIterations = 600000
	KEKSize       = 32
)

func GenerateSalt() ([]byte, error) {
	salt := make([]byte, SaltSize)
	if _, err := rand.Read(salt); err != nil {
		return nil, err
	}
	return salt, nil
}

func DeriveKEK(password string, salt []byte) []byte {
	return pbkdf2.Key([]byte(password), salt, KDFIterations, KEKSize, sha256.New)
}
