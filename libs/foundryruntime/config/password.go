package config

import (
	"crypto/pbkdf2"
	"crypto/sha512"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
)

// PBKDF2 parameters for set_password parity.
const (
	pbkdfIterations   = 1000
	pbkdfKeyLen       = 64
	pbkdfFallbackSalt = "17c4f39053ac5a50d5797c665ad1f4e6"
)

// HashAdminKey hashes a plaintext admin key using PBKDF2-SHA512 with
// 1000 iterations and a 64-byte output, compatible with Foundry's
// set_password algorithm. When salt is empty, pbkdfFallbackSalt is used.
func HashAdminKey(plaintext, salt string) (string, error) {
	plaintext = strings.TrimSpace(plaintext)
	if plaintext == "" {
		return "", errors.New("config: empty admin key")
	}
	if salt == "" {
		salt = pbkdfFallbackSalt
	}
	key, err := pbkdf2.Key(sha512.New, plaintext, []byte(salt), pbkdfIterations, pbkdfKeyLen)
	if err != nil {
		return "", fmt.Errorf("config: pbkdf2: %w", err)
	}
	return hex.EncodeToString(key), nil
}
