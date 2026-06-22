package token

import (
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/erewhile/iam/internal/consts"
	"github.com/erewhile/iam/pkg/aes"
	"github.com/erewhile/iam/pkg/utils"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type TokenType string

const (
	TokenTypeAccess  TokenType = "access"
	TokenTypeRefresh TokenType = "refresh"
)

var (
	signKey   *rsa.PrivateKey
	VerifyKey *rsa.PublicKey
)

type UserPayload struct {
	UserID   int       `json:"user_id"`
	UserUUID uuid.UUID `json:"user_uuid"`
}

type Claims struct {
	SessionID     uuid.UUID `json:"session_id"`
	TokenType     TokenType `json:"token_type"`
	EncryptedData string    `json:"encrypted_data"`
	jwt.RegisteredClaims
}

type TokenPair struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	SessionID    uuid.UUID `json:"session_id"`
	ExpiresIn    int64     `json:"expires_in"`
}

func init() {
	_, filename, _, _ := runtime.Caller(0)
	dir := filepath.Dir(filename)

	var certDir string
	for {
		target := filepath.Join(dir, "certs")
		if fi, err := os.Stat(target); err == nil && fi.IsDir() {
			certDir = target
			break
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			panic("cannot find 'certs' directory in any parent folders")
		}
		dir = parent
	}

	privBytes, err := os.ReadFile(filepath.Join(certDir, "iam_private.pem"))
	if err != nil {
		panic(fmt.Sprintf("failed to read private key: %v", err))
	}
	signKey, err = jwt.ParseRSAPrivateKeyFromPEM(privBytes)
	if err != nil {
		panic(fmt.Sprintf("failed to parse private key: %v", err))
	}

	pubBytes, err := os.ReadFile(filepath.Join(certDir, "iam_public.pem"))
	if err != nil {
		panic(fmt.Sprintf("failed to read public key: %v", err))
	}
	VerifyKey, err = jwt.ParseRSAPublicKeyFromPEM(pubBytes)
	if err != nil {
		panic(fmt.Sprintf("failed to parse public key: %v", err))
	}
}

func Validate(
	tokenString string,
	aad []byte,
	expectedType TokenType,
) (*Claims, *UserPayload, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		if tokenKid, ok := token.Header["kid"].(string); ok {
			if tokenKid != Kid() {
				return nil, fmt.Errorf("invalid token kid: %s", tokenKid)
			}
		}

		return VerifyKey, nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, nil, errors.New("token is expired")
		}
		return nil, nil, fmt.Errorf("invalid token: %w", err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, nil, errors.New("invalid token claims")
	}

	if expectedType != "" && claims.TokenType != expectedType {
		return nil, nil, fmt.Errorf("invalid token type: %s", claims.TokenType)
	}

	decryptedBytes, err := base64.StdEncoding.DecodeString(claims.EncryptedData)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to base64 decode ciphertext: %w", err)
	}

	decryptedPayload, err := aes.Decrypt(decryptedBytes, aad)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to decrypt payload: %w", err)
	}

	var payload UserPayload
	if err := json.Unmarshal(decryptedPayload, &payload); err != nil {
		return nil, nil, fmt.Errorf("failed to unmarshal decrypted payload: %w", err)
	}

	return claims, &payload, nil
}

func Kid() string {
	kid := os.Getenv(consts.EnvJwtKid)
	if kid == "" {
		kid = consts.DefaultJwtKid
	}
	return kid
}

func Generate(
	userID int,
	userUUID uuid.UUID,
	sessionID uuid.UUID,
	aad []byte,
) (*TokenPair, error) {

	if sessionID == uuid.Nil {
		var err error
		sessionID, err = uuid.NewRandom()
		if err != nil {
			return nil, fmt.Errorf("failed to generate session id: %w", err)
		}
	}

	now := utils.Now()

	accessJTI, _ := uuid.NewRandom()
	refreshJTI, _ := uuid.NewRandom()

	accessStr, err := generate(
		userID,
		userUUID,
		sessionID,
		aad,
		TokenTypeAccess,
		accessJTI,
		now.Add(consts.AccessTokenTTL),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshStr, err := generate(
		userID,
		userUUID,
		sessionID,
		aad,
		TokenTypeRefresh,
		refreshJTI,
		now.Add(consts.RefreshTokenTTL),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	return &TokenPair{
		AccessToken:  accessStr,
		RefreshToken: refreshStr,
		SessionID:    sessionID,
		ExpiresIn:    int64(consts.AccessTokenTTL.Seconds()),
	}, nil
}

func generate(
	userID int,
	userUUID uuid.UUID,
	sessionID uuid.UUID,
	aad []byte,
	tokenType TokenType,
	jti uuid.UUID,
	exp time.Time,
) (string, error) {

	payload := UserPayload{
		UserID:   userID,
		UserUUID: userUUID,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal payload: %w", err)
	}

	encryptedBytes, err := aes.Encrypt(payloadBytes, aad)
	if err != nil {
		return "", fmt.Errorf("failed to encrypt payload: %w", err)
	}

	encodedCiphertext := base64.StdEncoding.EncodeToString(encryptedBytes)

	claims := Claims{
		SessionID:     sessionID,
		TokenType:     tokenType,
		EncryptedData: encodedCiphertext,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        jti.String(),
			ExpiresAt: jwt.NewNumericDate(exp),
			IssuedAt:  jwt.NewNumericDate(utils.Now()),
			NotBefore: jwt.NewNumericDate(utils.Now()),
			Issuer:    "erewhile/iam",
			Subject:   fmt.Sprintf("%d", userID),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	token.Header["kid"] = Kid()

	return token.SignedString(signKey)
}
