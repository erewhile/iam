package utils

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"math/big"
)

func RandomString(n int) (string, error) {
	if n <= 0 {
		return "", fmt.Errorf("n must be positive")
	}
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

const (
	alphanumericCharset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	unambiguousCharset  = "abcdefghijkmnpqrstuvwxyzABCDEFGHJKLMNPQRSTUVWXYZ23456789"
)

func RandomAlphanumeric(length int) (string, error) {
	return randomFromCharset(length, alphanumericCharset)
}

func RandomUnambiguous(length int) (string, error) {
	return randomFromCharset(length, unambiguousCharset)
}

func randomFromCharset(length int, charset string) (string, error) {
	if length <= 0 {
		return "", fmt.Errorf("length must be positive")
	}
	charsetLen := big.NewInt(int64(len(charset)))
	result := make([]byte, length)
	for i := range result {
		idx, err := rand.Int(rand.Reader, charsetLen)
		if err != nil {
			return "", fmt.Errorf("failed to generate random char: %w", err)
		}
		result[i] = charset[idx.Int64()]
	}
	return string(result), nil
}
