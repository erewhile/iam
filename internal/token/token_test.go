package token

import (
	"testing"

	"github.com/erewhile/iam/config"
	"github.com/erewhile/iam/pkg/aes"
	"github.com/google/uuid"
)

func TestEncryptedToken(t *testing.T) {
	config.Init()
	aes.Init([]byte(config.Get().Aes.Key))

	userID := 621
	userUUID := uuid.New()
	myAAD := []byte("iam_token_v1")

	tokenPair, err := Generate(UserPayload{UserID: userID, UserUUID: userUUID}, uuid.Nil, myAAD)
	if err != nil {
		t.Fatalf("bummer! failed to mint token: %v", err)
	}

	t.Logf("here is your access token: %s", tokenPair.AccessToken)

	claims, payload, err := Validate(tokenPair.AccessToken, myAAD, TokenTypeAccess)
	if err != nil {
		t.Fatalf("nope! token validation went sideways: %v", err)
	}

	if payload.UserID != userID || payload.UserUUID != userUUID {
		t.Errorf("hold up! decrypted data is messed up")
	}

	t.Logf("nailed it! decrypted UserID: %d, SessionID: %s", payload.UserID, claims.SessionID)
}
