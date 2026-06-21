package token

import (
	"testing"
	"time"

	"github.com/erewhile/iam/pkg/aes"
	"github.com/google/uuid"
)

func TestEncryptedToken(t *testing.T) {
	aesKey := []byte("x4AbfHqKSgDUQ6WrAWk5KLpe4al95BXa")
	aes.Init(aesKey)

	userID := 621
	userUUID := uuid.New()
	myAAD := []byte("iam_token_v1")

	tokenStr, err := GenerateToken(userID, userUUID, uuid.Nil, myAAD, time.Minute*10)
	if err != nil {
		t.Fatalf("bummer! failed to mint token: %v", err)
	}

	t.Logf("here is your token: %s", tokenStr)

	claims, payload, err := ValidateAndDecryptToken(tokenStr, myAAD)
	if err != nil {
		t.Fatalf("nope! token validation went sideways: %v", err)
	}

	if payload.UserID != userID || payload.UserUUID != userUUID {
		t.Errorf("hold up! decrypted data is messed up")
	}

	t.Logf("nailed it! decrypted UserID: %d, SessionID: %s", payload.UserID, claims.SessionID)
}
