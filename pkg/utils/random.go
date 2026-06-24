package utils

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
)

func RandomString(n int) (string, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func RandomCryptoToken(n int) (string, error) {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ01233456789-_"

	bytes := make([]byte, n)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", fmt.Errorf("failed to generate crypto tokens: %w", err)
	}

	for i, b := range bytes {
		bytes[i] = charset[b%64]
	}

	return string(bytes), nil
}
