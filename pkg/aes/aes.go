package aes

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"sync"
)

var (
	global *AesGCM
	once   sync.Once
)

type AesGCM struct {
	gcm cipher.AEAD
}

func New(key []byte) (*AesGCM, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	return &AesGCM{gcm: gcm}, nil
}

func Init(key []byte) {
	once.Do(func() {
		instance, err := New(key)
		if err != nil {
			panic("aes init failed: " + err.Error())
		}
		global = instance
	})
}

func Global() *AesGCM {
	if global == nil {
		panic("aes: global instance not initialized, call Init() first")
	}
	return global
}

func (a *AesGCM) Encrypt(plaintext, aad []byte) ([]byte, error) {
	nonceSize := a.gcm.NonceSize()

	out := make([]byte, nonceSize+len(plaintext)+a.gcm.Overhead())

	nonce := out[:nonceSize]
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("failed to generate nonce: %w", err)
	}

	a.gcm.Seal(out[:nonceSize], nonce, plaintext, aad)

	return out, nil
}

func (a *AesGCM) Decrypt(data, aad []byte) ([]byte, error) {
	ns := a.gcm.NonceSize()
	if len(data) < ns {
		return nil, errors.New("ciphertext too short")
	}

	nonce := data[:ns]
	ciphertext := data[ns:]

	return a.gcm.Open(nil, nonce, ciphertext, aad)
}

func Encrypt(plaintext, aad []byte) ([]byte, error) {
	return Global().Encrypt(plaintext, aad)
}

func Decrypt(data, aad []byte) ([]byte, error) {
	return Global().Decrypt(data, aad)
}
