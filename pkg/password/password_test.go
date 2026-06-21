package password

import (
	"errors"
	"sync"
	"testing"
)

func TestHashAndValidate(t *testing.T) {
	password := "123456"

	hash, err := Hash(password)
	if err != nil {
		t.Fatalf("hash failed: %v", err)
	}

	if hash == "" {
		t.Fatal("hash is empty")
	}

	ok, err := Validate(password, hash)
	if err != nil {
		t.Fatalf("validate error: %v", err)
	}

	if !ok {
		t.Fatal("password should be valid but got false")
	}
}

func TestValidateWrongPassword(t *testing.T) {
	password := "123456"

	hash, err := Hash(password)
	if err != nil {
		t.Fatalf("hash failed: %v", err)
	}

	ok, err := Validate("wrong-password", hash)
	if err != nil {
		t.Fatalf("validate error: %v", err)
	}

	if ok {
		t.Fatal("wrong password should fail")
	}
}

func TestConcurrency(t *testing.T) {
	const goroutines = 20

	var wg sync.WaitGroup
	errCh := make(chan error, goroutines)

	for range goroutines {
		wg.Go(func() {
			hash, err := Hash("123456")
			if err != nil {
				errCh <- err
				return
			}

			ok, err := Validate("123456", hash)
			if err != nil {
				errCh <- err
				return
			}

			if !ok {
				errCh <- errors.New("password should be valid but got false")
				return
			}
		})
	}

	wg.Wait()
	close(errCh)

	for err := range errCh {
		t.Fatalf("concurrent test failed: %v", err)
	}
}
