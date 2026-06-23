package totp

import (
	"context"
	"encoding/base32"
	"errors"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"
)

func seedFor(algo Algorithm) string {
	switch algo {
	case SHA1:
		return "12345678901234567890" // 20 bytes
	case SHA256:
		return "12345678901234567890123456789012" // 32 bytes
	case SHA512:
		return "1234567890123456789012345678901234567890123456789012345678901234" // 64 bytes
	default:
		panic("unknown algorithm")
	}
}

func base32Secret(raw string) string {
	return base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString([]byte(raw))
}

func TestGenerate_RFC6238Vectors(t *testing.T) {
	cases := []struct {
		unixTime int64
		want     string
		algo     Algorithm
	}{
		{59, "94287082", SHA1},
		{59, "46119246", SHA256},
		{59, "90693936", SHA512},
		{1111111109, "07081804", SHA1},
		{1111111109, "68084774", SHA256},
		{1111111109, "25091201", SHA512},
		{1111111111, "14050471", SHA1},
		{1111111111, "67062674", SHA256},
		{1111111111, "99943326", SHA512},
		{1234567890, "89005924", SHA1},
		{1234567890, "91819424", SHA256},
		{1234567890, "93441116", SHA512},
		{2000000000, "69279037", SHA1},
		{2000000000, "90698825", SHA256},
		{2000000000, "38618901", SHA512},
		{20000000000, "65353130", SHA1},
		{20000000000, "77737706", SHA256},
		{20000000000, "47863826", SHA512},
	}

	for _, tc := range cases {
		name := fmt.Sprintf("%s@%d", tc.algo, tc.unixTime)
		t.Run(name, func(t *testing.T) {
			secret := base32Secret(seedFor(tc.algo))
			cfg := Config{Algorithm: tc.algo, Digits: 8, Timestep: 30}
			got, err := Generate(secret, time.Unix(tc.unixTime, 0).UTC(), cfg)
			if err != nil {
				t.Fatalf("Generate() error = %v", err)
			}
			if got != tc.want {
				t.Errorf("Generate() = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		cfg     Config
		wantErr bool
	}{
		{"zero value gets sane defaults", Config{}, false},
		{"valid explicit", Config{Algorithm: SHA256, Digits: 8, Timestep: 60, Window: 2}, false},
		{"unsupported algorithm", Config{Algorithm: "MD5", Digits: 6, Timestep: 30}, true},
		{"digits too high", Config{Algorithm: SHA1, Digits: 9, Timestep: 30}, true},
		{"digits negative", Config{Algorithm: SHA1, Digits: -1, Timestep: 30}, true},
		{"zero timestep after defaults is fine", Config{}, false},
		{"negative timestep", Config{Timestep: -30}, true},
		{"negative window", Config{Window: -1}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := tt.cfg.withDefaults()
			err := cfg.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNew_RejectsInvalidConfig(t *testing.T) {
	_, err := New(Config{Algorithm: "bogus"}, nil)
	if err == nil {
		t.Fatal("expected error for invalid algorithm, got nil")
	}

	_, err = New(Config{Window: -1}, nil)
	if err == nil {
		t.Fatal("expected error for negative window, got nil")
	}

	v, err := New(Config{}, nil)
	if err != nil {
		t.Fatalf("expected zero-value config to be valid (defaulted), got error: %v", err)
	}
	if v == nil {
		t.Fatal("expected non-nil validator")
	}
}

func TestConfig_WithDefaults(t *testing.T) {
	cfg := Config{}.WithDefaults()
	if cfg.Algorithm != SHA1 {
		t.Errorf("default Algorithm = %v, want SHA1", cfg.Algorithm)
	}
	if cfg.Digits != Digits6 {
		t.Errorf("default Digits = %v, want %v", cfg.Digits, Digits6)
	}
	if cfg.Timestep != Timestep30s {
		t.Errorf("default Timestep = %v, want %v", cfg.Timestep, Timestep30s)
	}
}

func TestDecodeSecret(t *testing.T) {
	validSecret := base32Secret(seedFor(SHA1))

	tests := []struct {
		name       string
		secret     string
		minKeySize int
		wantErr    bool
	}{
		{"valid uppercase", validSecret, 0, false},
		{"valid lowercase normalized", strings.ToLower(validSecret), 0, false},
		{"valid with spaces and padding", insertSpaces(validSecret) + "====", 0, false},
		{"empty secret", "", 0, true},
		{"illegal character (digit 1)", "AAAA1AAA", 0, true},
		{"illegal character (digit 0)", "AAAA0AAA", 0, true},
		{"illegal character (symbol)", "AAAA-AAA", 0, true},
		{"too short for min key size", base32Secret("short"), 16, true},
		{"exactly min key size", base32Secret(strings.Repeat("x", 16)), 16, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := decodeSecret(tt.secret, tt.minKeySize)
			if (err != nil) != tt.wantErr {
				t.Errorf("decodeSecret(%q) error = %v, wantErr %v", tt.secret, err, tt.wantErr)
			}
		})
	}
}

func insertSpaces(s string) string {
	var b strings.Builder
	for i, r := range s {
		if i > 0 && i%4 == 0 {
			b.WriteByte(' ')
		}
		b.WriteRune(r)
	}
	return b.String()
}

func TestDecodeSecret_InvalidLengthRemainder(t *testing.T) {
	if _, err := decodeSecret("AAAAA", 0); err != nil {
		t.Errorf("expected len%%8==5 to be valid base32, got error: %v", err)
	}

	for _, s := range []string{"A", "AAA", "AAAAAA"} {
		if _, err := decodeSecret(s, 0); err == nil {
			t.Errorf("decodeSecret(%q): expected error for invalid base32 length, got nil", s)
		}
	}
}

func TestGenerateSecret(t *testing.T) {
	if _, err := GenerateSecret(8); err == nil {
		t.Error("expected error for byteLen < 16, got nil")
	}

	s1, err := GenerateSecret(20)
	if err != nil {
		t.Fatalf("GenerateSecret() error = %v", err)
	}
	s2, err := GenerateSecret(20)
	if err != nil {
		t.Fatalf("GenerateSecret() error = %v", err)
	}
	if s1 == s2 {
		t.Error("two generated secrets are identical; RNG may be broken")
	}

	key, err := decodeSecret(s1, 0)
	if err != nil {
		t.Fatalf("generated secret failed to decode: %v", err)
	}
	if len(key) != 20 {
		t.Errorf("decoded key length = %d, want 20", len(key))
	}
}

func TestProvisioningURI(t *testing.T) {
	secret := base32Secret(seedFor(SHA1))
	uri := ProvisioningURI("iam", "eren@xiazhi.net", secret, Config{Algorithm: SHA256, Digits: 8, Timestep: 60})

	wantPrefix := "otpauth://totp/iam:eren@xiazhi.net?"
	if !strings.HasPrefix(uri, wantPrefix) {
		t.Errorf("ProvisioningURI() = %q, want prefix %q", uri, wantPrefix)
	}
	for _, want := range []string{"secret=" + secret, "issuer=iam", "algorithm=SHA256", "digits=8", "period=60"} {
		if !strings.Contains(uri, want) {
			t.Errorf("ProvisioningURI() = %q, missing %q", uri, want)
		}
	}
}

type memStore struct {
	mu       sync.Mutex
	data     map[string]int64
	failNext bool
	failErr  error
}

func newMemStore() *memStore {
	return &memStore{data: make(map[string]int64)}
}

func (s *memStore) GetLastTimestamp(_ context.Context, key string) (int64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.failNext {
		s.failNext = false
		return 0, s.failErr
	}
	ts, ok := s.data[key]
	if !ok {
		return 0, ErrTimestampNotFound
	}
	return ts, nil
}

func (s *memStore) SetLastTimestamp(_ context.Context, key string, ts int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[key] = ts
	return nil
}

func TestValidateAt_CorrectCode(t *testing.T) {
	secret := base32Secret(seedFor(SHA1))
	v, err := New(Config{Digits: 8, Timestep: 30}, newMemStore())
	if err != nil {
		t.Fatal(err)
	}

	at := time.Unix(59, 0).UTC()
	ok, err := v.ValidateAt(context.Background(), "user1", secret, "94287082", at)
	if err != nil {
		t.Fatalf("ValidateAt() error = %v", err)
	}
	if !ok {
		t.Error("ValidateAt() = false, want true for known-good code")
	}
}

func TestValidateAt_WrongCode(t *testing.T) {
	secret := base32Secret(seedFor(SHA1))
	v, err := New(Config{Digits: 8, Timestep: 30}, newMemStore())
	if err != nil {
		t.Fatal(err)
	}

	at := time.Unix(59, 0).UTC()
	ok, err := v.ValidateAt(context.Background(), "user1", secret, "00000000", at)
	if err != nil {
		t.Fatalf("ValidateAt() error = %v", err)
	}
	if ok {
		t.Error("ValidateAt() = true, want false for wrong code")
	}
}

func TestValidateAt_CodeLengthMismatch(t *testing.T) {
	secret := base32Secret(seedFor(SHA1))
	v, err := New(Config{Digits: 6, Timestep: 30}, nil)
	if err != nil {
		t.Fatal(err)
	}
	ok, err := v.ValidateAt(context.Background(), "user1", secret, "12345", time.Now())
	if err != nil {
		t.Fatalf("ValidateAt() error = %v", err)
	}
	if ok {
		t.Error("ValidateAt() = true for wrong-length code, want false")
	}
}

func TestValidateAt_WindowTolerance(t *testing.T) {
	secret := base32Secret(seedFor(SHA1))
	v, err := New(Config{Digits: 6, Timestep: 30, Window: 1}, newMemStore())
	if err != nil {
		t.Fatal(err)
	}

	now := time.Unix(1_700_000_000, 0).UTC()
	codeOneStepAhead, err := Generate(secret, now.Add(30*time.Second), Config{Digits: 6, Timestep: 30})
	if err != nil {
		t.Fatal(err)
	}

	ok, err := v.ValidateAt(context.Background(), "user1", secret, codeOneStepAhead, now)
	if err != nil {
		t.Fatalf("ValidateAt() error = %v", err)
	}
	if !ok {
		t.Error("expected code from adjacent step to be accepted within Window=1")
	}

	codeTwoStepsAhead, err := Generate(secret, now.Add(60*time.Second), Config{Digits: 6, Timestep: 30})
	if err != nil {
		t.Fatal(err)
	}
	ok, err = v.ValidateAt(context.Background(), "user2", secret, codeTwoStepsAhead, now)
	if err != nil {
		t.Fatalf("ValidateAt() error = %v", err)
	}
	if ok {
		t.Error("expected code two steps away to be rejected when Window=1")
	}
}

func TestValidateAt_ReplayIsRejected(t *testing.T) {
	secret := base32Secret(seedFor(SHA1))
	v, err := New(Config{Digits: 6, Timestep: 30}, newMemStore())
	if err != nil {
		t.Fatal(err)
	}

	now := time.Unix(1_700_000_000, 0).UTC()
	code, err := Generate(secret, now, Config{Digits: 6, Timestep: 30})
	if err != nil {
		t.Fatal(err)
	}

	ok, err := v.ValidateAt(context.Background(), "user1", secret, code, now)
	if err != nil || !ok {
		t.Fatalf("first use should succeed: ok=%v err=%v", ok, err)
	}

	ok, err = v.ValidateAt(context.Background(), "user1", secret, code, now.Add(2*time.Second))
	if err != nil {
		t.Fatalf("ValidateAt() error = %v", err)
	}
	if ok {
		t.Error("replayed code was accepted a second time; replay protection failed")
	}
}

func TestValidateAt_StoreErrorFailsClosed(t *testing.T) {
	secret := base32Secret(seedFor(SHA1))
	store := newMemStore()
	store.failNext = true
	store.failErr = errors.New("boom: store unavailable")

	v, err := New(Config{Digits: 6, Timestep: 30}, store)
	if err != nil {
		t.Fatal(err)
	}

	code, err := Generate(secret, time.Now(), Config{Digits: 6, Timestep: 30})
	if err != nil {
		t.Fatal(err)
	}

	ok, err := v.Validate(context.Background(), "user1", secret, code)
	if err == nil {
		t.Fatal("expected error when store lookup fails, got nil (fail-open bug)")
	}
	if ok {
		t.Error("expected validation to fail closed on store error, got ok=true")
	}
}

func TestValidateAt_ConcurrentReplaySerialized(t *testing.T) {
	secret := base32Secret(seedFor(SHA1))
	v, err := New(Config{Digits: 6, Timestep: 30}, newMemStore())
	if err != nil {
		t.Fatal(err)
	}

	now := time.Now()
	code, err := Generate(secret, now, Config{Digits: 6, Timestep: 30})
	if err != nil {
		t.Fatal(err)
	}

	const n = 20
	var wg sync.WaitGroup
	var successCount int32
	var mu sync.Mutex

	for range n {
		wg.Go(func() {
			ok, err := v.ValidateAt(context.Background(), "shared-user", secret, code, now)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			if ok {
				mu.Lock()
				successCount++
				mu.Unlock()
			}
		})
	}
	wg.Wait()

	if successCount != 1 {
		t.Errorf("expected exactly 1 successful validation out of %d concurrent attempts, got %d", n, successCount)
	}
}

func TestValidateAt_NilStoreSkipsReplayProtection(t *testing.T) {
	secret := base32Secret(seedFor(SHA1))
	v, err := New(Config{Digits: 6, Timestep: 30}, nil)
	if err != nil {
		t.Fatal(err)
	}

	now := time.Now()
	code, err := Generate(secret, now, Config{Digits: 6, Timestep: 30})
	if err != nil {
		t.Fatal(err)
	}

	for i := range 2 {
		ok, err := v.ValidateAt(context.Background(), "user1", secret, code, now)
		if err != nil || !ok {
			t.Fatalf("attempt %d: expected ok=true err=nil, got ok=%v err=%v (replay protection is opt-in via Store)", i, ok, err)
		}
	}
}

func TestHashFunc_RejectsUnknownAlgorithm(t *testing.T) {
	if _, err := Algorithm("DOES_NOT_EXIST").hashFunc(); err == nil {
		t.Error("expected error for unknown algorithm, got nil")
	}
}

func TestMinKeySize_RejectsWeakSecret(t *testing.T) {
	weakSecret := base32Secret("short")
	_, err := Generate(weakSecret, time.Now(), Config{Algorithm: SHA1, Digits: 6, Timestep: 30})
	if err == nil {
		t.Error("expected error for secret shorter than hash output size, got nil")
	}
}

func TestMinKeySize_DisabledWhenNegative(t *testing.T) {
	weakSecret := base32Secret("short")
	_, err := Generate(weakSecret, time.Now(), Config{Algorithm: SHA1, Digits: 6, Timestep: 30, MinKeySize: -1})
	if err != nil {
		t.Errorf("expected no error with MinKeySize disabled, got: %v", err)
	}
}
