package config

import (
	"crypto/pbkdf2"
	"crypto/sha512"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
)

const pbkdfKeyLen = 64 // PBKDF2 output length, for set_password parity

func HashAdminKey(plaintext, salt string) (string, error) {
	const (
		pbkdfIterations   = 1000
		pbkdfFallbackSalt = "17c4f39053ac5a50d5797c665ad1f4e6"
	)
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
