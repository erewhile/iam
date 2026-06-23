package totp

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"crypto/subtle"
	"encoding/base32"
	"encoding/binary"
	"errors"
	"fmt"
	"hash"
	"net/url"
	"strings"
	"sync"
	"time"
	"unicode"
)

type Algorithm string

const (
	SHA1   Algorithm = "SHA1"
	SHA256 Algorithm = "SHA256"
	SHA512 Algorithm = "SHA512"
)

func (a Algorithm) hashSize() int {
	switch a {
	case SHA256:
		return sha256.Size
	case SHA512:
		return sha512.Size
	case SHA1:
		return sha1.Size
	default:
		return 0
	}
}

func (a Algorithm) hashFunc() (func() hash.Hash, error) {
	switch a {
	case SHA1:
		return sha1.New, nil
	case SHA256:
		return sha256.New, nil
	case SHA512:
		return sha512.New, nil
	default:
		return nil, fmt.Errorf("totp: unsupported algorithm %q", a)
	}
}

const (
	Digits6 = 6
	Digits8 = 8

	Timestep30s int64 = 30
	Timestep60s int64 = 60

	maxDigits = 8
)

var pow10 = []uint32{1, 10, 100, 1000, 10000, 100000, 1000000, 10000000, 100000000}

type Config struct {
	Algorithm Algorithm
	Digits    int
	Timestep  int64
	Window    int

	MinKeySize int
}

func (cfg Config) WithDefaults() Config {
	return cfg.withDefaults()
}

func (cfg Config) withDefaults() Config {
	if cfg.Algorithm == "" {
		cfg.Algorithm = SHA1
	}
	if cfg.Digits == 0 {
		cfg.Digits = Digits6
	}
	if cfg.Timestep == 0 {
		cfg.Timestep = Timestep30s
	}
	if cfg.MinKeySize == 0 {
		cfg.MinKeySize = cfg.Algorithm.hashSize()
	}
	return cfg
}

func (cfg Config) Validate() error {
	if _, err := cfg.Algorithm.hashFunc(); err != nil {
		return err
	}
	if cfg.Digits < 1 || cfg.Digits > maxDigits {
		return fmt.Errorf("totp: digits must be between 1 and %d", maxDigits)
	}
	if cfg.Timestep <= 0 {
		return errors.New("totp: timestep must be positive")
	}
	if cfg.Window < 0 {
		return errors.New("totp: window must be >= 0")
	}
	return nil
}

var ErrTimestampNotFound = errors.New("totp: no timestamp recorded for key")

type Store interface {
	GetLastTimestamp(ctx context.Context, key string) (int64, error)
	SetLastTimestamp(ctx context.Context, key string, timestamp int64) error
}

type TOTP struct {
	config Config
	store  Store

	mu    sync.Mutex
	locks map[string]*sync.Mutex
}

func New(config Config, store Store) (*TOTP, error) {
	config = config.withDefaults()
	if err := config.Validate(); err != nil {
		return nil, err
	}
	return &TOTP{config: config, store: store, locks: make(map[string]*sync.Mutex)}, nil
}

func (v *TOTP) lockFor(key string) *sync.Mutex {
	v.mu.Lock()
	defer v.mu.Unlock()
	l, ok := v.locks[key]
	if !ok {
		l = &sync.Mutex{}
		v.locks[key] = l
	}
	return l
}

func (v *TOTP) Generate(secret string) (string, error) {
	return v.GenerateAt(secret, time.Now())
}

func (v *TOTP) GenerateAt(secret string, at time.Time) (string, error) {
	key, err := decodeSecret(secret, v.config.MinKeySize)
	if err != nil {
		return "", err
	}
	counter := uint64(at.Unix() / v.config.Timestep)
	return hotp(key, counter, v.config)
}

func (v *TOTP) Validate(ctx context.Context, key, secret, code string) (bool, error) {
	return v.ValidateAt(ctx, key, secret, code, time.Now())
}

func (v *TOTP) ValidateAt(ctx context.Context, key, secret, code string, at time.Time) (bool, error) {
	if len(code) != v.config.Digits {
		return false, nil
	}

	decodedKey, err := decodeSecret(secret, v.config.MinKeySize)
	if err != nil {
		return false, err
	}

	mu := v.lockFor(key)
	mu.Lock()
	defer mu.Unlock()

	var lastUsedTS int64
	if v.store != nil {
		ts, err := v.store.GetLastTimestamp(ctx, key)
		switch {
		case err == nil:
			lastUsedTS = ts
		case errors.Is(err, ErrTimestampNotFound):
			lastUsedTS = 0
		default:
			return false, fmt.Errorf("totp: replay store lookup failed: %w", err)
		}
	}

	codeBytes := []byte(code)

	for i := -v.config.Window; i <= v.config.Window; i++ {
		t := at.Add(time.Duration(i) * time.Duration(v.config.Timestep) * time.Second)
		currentStepTS := (t.Unix() / v.config.Timestep) * v.config.Timestep
		if v.store != nil && currentStepTS <= lastUsedTS {
			continue
		}

		counter := uint64(t.Unix() / v.config.Timestep)
		candidate, err := hotp(decodedKey, counter, v.config)
		if err != nil {
			return false, err
		}

		if subtle.ConstantTimeCompare([]byte(candidate), codeBytes) == 1 {
			if v.store != nil {
				if err := v.store.SetLastTimestamp(ctx, key, currentStepTS); err != nil {
					return false, fmt.Errorf("totp: replay store update failed: %w", err)
				}
			}
			return true, nil
		}
	}
	return false, nil
}

func Generate(secret string, t time.Time, config Config) (string, error) {
	config = config.withDefaults()
	if err := config.Validate(); err != nil {
		return "", err
	}
	key, err := decodeSecret(secret, config.MinKeySize)
	if err != nil {
		return "", err
	}
	counter := uint64(t.Unix() / config.Timestep)
	return hotp(key, counter, config)
}

func hotp(key []byte, counter uint64, config Config) (string, error) {
	hashFn, err := config.Algorithm.hashFunc()
	if err != nil {
		return "", err
	}

	var buf [8]byte
	binary.BigEndian.PutUint64(buf[:], counter)

	h := hmac.New(hashFn, key)
	h.Write(buf[:])
	sum := h.Sum(nil)

	offset := sum[len(sum)-1] & 0xf
	binCode := binary.BigEndian.Uint32(sum[offset:offset+4]) & 0x7fffffff
	code := binCode % pow10[config.Digits]
	return fmt.Sprintf("%0*d", config.Digits, code), nil
}

func GenerateSecret(byteLen int) (string, error) {
	if byteLen < 16 {
		return "", errors.New("totp: secret length must be at least 16 bytes (128 bits)")
	}
	raw := make([]byte, byteLen)
	if _, err := rand.Read(raw); err != nil {
		return "", fmt.Errorf("totp: failed to generate random secret: %w", err)
	}
	return base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(raw), nil
}

// See https://github.com/google/google-authenticator/wiki/Key-Uri-Format
func ProvisioningURI(issuer, accountName, secret string, config Config) string {
	config = config.withDefaults()
	label := accountName
	if issuer != "" {
		label = issuer + ":" + accountName
	}

	q := url.Values{}
	q.Set("secret", secret)
	if issuer != "" {
		q.Set("issuer", issuer)
	}
	q.Set("algorithm", string(config.Algorithm))
	q.Set("digits", fmt.Sprintf("%d", config.Digits))
	q.Set("period", fmt.Sprintf("%d", config.Timestep))

	u := url.URL{
		Scheme:   "otpauth",
		Host:     "totp",
		Path:     "/" + label,
		RawQuery: q.Encode(),
	}
	return u.String()
}

func decodeSecret(secret string, minKeySize int) ([]byte, error) {
	var invalid bool
	cleaned := strings.Map(func(r rune) rune {
		if unicode.IsSpace(r) || r == '=' {
			return -1
		}
		if 'a' <= r && r <= 'z' {
			r -= 32
		}
		if (r >= 'A' && r <= 'Z') || (r >= '2' && r <= '7') {
			return r
		}
		invalid = true
		return -1
	}, secret)

	if invalid {
		return nil, errors.New("totp: illegal base32 character in secret")
	}
	if cleaned == "" {
		return nil, errors.New("totp: empty secret")
	}

	if rem := len(cleaned) % 8; rem == 1 || rem == 3 || rem == 6 {
		return nil, errors.New("totp: invalid base32 secret length")
	}
	if mod := len(cleaned) % 8; mod != 0 {
		cleaned += strings.Repeat("=", 8-mod)
	}

	key, err := base32.StdEncoding.DecodeString(cleaned)
	if err != nil {
		return nil, fmt.Errorf("totp: failed to decode secret: %w", err)
	}

	if minKeySize > 0 && len(key) < minKeySize {
		return nil, fmt.Errorf("totp: secret too short, need >= %d bytes, got %d", minKeySize, len(key))
	}
	return key, nil
}
