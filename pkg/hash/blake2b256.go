package hash

import (
	"crypto/hmac"
	"encoding/base64"
	"encoding/hex"
	"hash"

	"golang.org/x/crypto/blake2b"
)

func HashBlake2b256(data []byte) []byte {
	sum := blake2b.Sum256(data)
	return sum[:]
}

func HashBlake2b256Hex(data []byte) string {
	return hex.EncodeToString(HashBlake2b256(data))
}

func HMACBlake2b256(data []byte, key []byte) []byte {
	mac := hmac.New(func() hash.Hash {
		h, _ := blake2b.New256(nil)
		return h
	}, key)

	mac.Write(data)
	return mac.Sum(nil)
}

func HMACBlake2b256Hex(data []byte, key []byte) string {
	return hex.EncodeToString(HMACBlake2b256(data, key))
}

func VerifyHMACBlake2b256(data []byte, key []byte, expectedHex string) bool {
	expected, err := hex.DecodeString(expectedHex)
	if err != nil {
		return false
	}

	actual := HMACBlake2b256(data, key)
	return hmac.Equal(actual, expected)
}

func HMACBlake2b256Base64(data []byte, key []byte) string {
	return base64.RawURLEncoding.EncodeToString(HMACBlake2b256(data, key))
}

func VerifyHMACBlake2b256Base64(data []byte, key []byte, expectedBase64 string) bool {
	expected, err := base64.RawURLEncoding.DecodeString(expectedBase64)
	if err != nil {
		return false
	}

	actual := HMACBlake2b256(data, key)
	return hmac.Equal(actual, expected)
}

func MACBlake2b256(data []byte, key []byte) ([]byte, error) {
	h, err := blake2b.New256(key)
	if err != nil {
		return nil, err
	}

	h.Write(data)
	return h.Sum(nil), nil
}

func VerifyMACBlake2b256(data []byte, key []byte, expectedHex string) bool {
	expected, err := hex.DecodeString(expectedHex)
	if err != nil {
		return false
	}

	actual, err := MACBlake2b256(data, key)
	if err != nil {
		return false
	}

	return hmac.Equal(actual, expected)
}

func VerifyMACBlake2b256Base64(data []byte, key []byte, expectedBase64 string) bool {
	expected, err := base64.RawURLEncoding.DecodeString(expectedBase64)
	if err != nil {
		return false
	}

	actual, err := MACBlake2b256(data, key)
	if err != nil {
		return false
	}

	return hmac.Equal(actual, expected)
}
