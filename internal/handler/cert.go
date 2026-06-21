package handler

import (
	"encoding/base64"
	"math/big"
	"net/http"

	"github.com/erewhile/iam/internal/token"
	"github.com/gin-gonic/gin"
)

type JWKKey struct {
	Kty string `json:"kty"`
	Alg string `json:"alg"`
	Use string `json:"use"`
	Kid string `json:"kid"`
	N   string `json:"n"`
	E   string `json:"e"`
}

type JWKSResponse struct {
	Keys []JWKKey `json:"keys"`
}

type CertHandler struct{}

func NewCertHandler() *CertHandler {
	return &CertHandler{}
}

func (h *CertHandler) JWKS(c *gin.Context) {
	pubKey := token.VerifyKey
	if pubKey == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "public key not initialized"})
		return
	}

	nStr := base64.RawURLEncoding.EncodeToString(pubKey.N.Bytes())
	eBytes := big.NewInt(int64(pubKey.E)).Bytes()
	eStr := base64.RawURLEncoding.EncodeToString(eBytes)
	kid := token.Kid()

	jwk := JWKKey{
		Kty: "RSA",
		Alg: "RS256",
		Use: "sig",
		Kid: kid,
		N:   nStr,
		E:   eStr,
	}

	c.JSON(http.StatusOK, JWKSResponse{
		Keys: []JWKKey{jwk},
	})
}
