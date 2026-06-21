package password

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"golang.org/x/crypto/argon2"
)

const (
	defaultVariant = "argon2id"
	defaultVersion = argon2.Version
)

type Config struct {
	Time, Memory       uint32
	Threads            uint8
	KeyLength, SaltLen uint32
	MaxConcurrency     int
	WaitTimeout        time.Duration
}

type hasher struct {
	config    *Config
	semaphore chan struct{}
}

func DefaultConfig() *Config {
	return &Config{
		Time: 3, Memory: 1 << 15, Threads: 2,
		KeyLength: 32, SaltLen: 16,
		MaxConcurrency: 8, WaitTimeout: 5 * time.Second,
	}
}

func New(cfg *Config) *hasher {
	if cfg == nil {
		cfg = DefaultConfig()
	}
	return &hasher{
		config:    cfg,
		semaphore: make(chan struct{}, cfg.MaxConcurrency),
	}
}

var defaultHasher = New(nil)

func Hash(password string) (string, error) { return defaultHasher.Hash(password) }

func Validate(password, encodedHash string) (bool, error) {
	return defaultHasher.Validate(password, encodedHash)
}

func (h *hasher) Hash(password string) (string, error) {
	if password == "" {
		return "", errors.New("password cannot be empty")
	}

	if err := h.acquire(); err != nil {
		return "", err
	}
	defer h.release()

	salt := make([]byte, h.config.SaltLen)
	if _, err := io.ReadFull(rand.Reader, salt); err != nil {
		return "", fmt.Errorf("failed to generate salt: %w", err)
	}

	hash := argon2.IDKey(
		[]byte(password),
		salt,
		h.config.Time,
		h.config.Memory,
		h.config.Threads,
		h.config.KeyLength,
	)

	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Hash := base64.RawStdEncoding.EncodeToString(hash)

	return fmt.Sprintf("$%s$v=%d$m=%d,t=%d,p=%d$%s$%s",
		defaultVariant, defaultVersion, h.config.Memory, h.config.Time, h.config.Threads, b64Salt, b64Hash), nil
}

func (h *hasher) Validate(password string, encodedHash string) (bool, error) {
	if password == "" || encodedHash == "" {
		return false, errors.New("credentials missing")
	}

	p := strings.Split(encodedHash, "$")
	if len(p) != 6 || p[1] != defaultVariant {
		return false, errors.New("invalid hash format or variant")
	}

	if !strings.HasPrefix(p[2], "v=") {
		return false, errors.New("invalid version format")
	}
	v, err := strconv.Atoi(p[2][2:])
	if err != nil || v != defaultVersion {
		return false, errors.New("unsupported or invalid version")
	}

	var m, t uint32
	var th uint8
	params := strings.Split(p[3], ",")
	if len(params) != 3 {
		return false, errors.New("invalid parameter format")
	}

	for _, param := range params {
		kv := strings.Split(param, "=")
		if len(kv) != 2 {
			return false, errors.New("invalid parameter kv")
		}
		val, err := strconv.ParseUint(kv[1], 10, 32)
		if err != nil {
			return false, errors.New("invalid parameter value")
		}
		switch kv[0] {
		case "m":
			m = uint32(val)
		case "t":
			t = uint32(val)
		case "p":
			th = uint8(val)
		default:
			return false, errors.New("unknown parameter")
		}
	}

	if m == 0 || t == 0 || th == 0 {
		return false, errors.New("invalid cryptographic parameters")
	}

	if err := h.acquire(); err != nil {
		return false, err
	}
	defer h.release()

	salt, err1 := base64.RawStdEncoding.DecodeString(p[4])
	actualHash, err2 := base64.RawStdEncoding.DecodeString(p[5])
	if err1 != nil || err2 != nil {
		return false, errors.New("decode failed")
	}

	if len(actualHash) < 16 || len(actualHash) > 128 {
		return false, errors.New("invalid actual hash length")
	}

	expectHash := argon2.IDKey([]byte(password), salt, t, m, th, uint32(len(actualHash)))
	return subtle.ConstantTimeCompare(expectHash, actualHash) == 1, nil
}

func (h *hasher) acquire() error {
	select {
	case h.semaphore <- struct{}{}:
		return nil
	default:
		timer := time.NewTimer(h.config.WaitTimeout)
		defer timer.Stop()

		select {
		case h.semaphore <- struct{}{}:
			return nil
		case <-timer.C:
			return errors.New("server too busy")
		}
	}
}

func (h *hasher) release() { <-h.semaphore }
