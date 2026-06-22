package model

import (
	"time"

	"github.com/google/uuid"
)

type TokenType uint8

const (
	TokenTypeAccess TokenType = 0
	TokenTypeRefresh
)

type Token struct {
	ID        int
	UserID    int
	Jti       uuid.UUID
	SessionID uuid.UUID
	Type      TokenType
	TokenHash string
	IP        string
	UserAgent string
	ExpiresAt time.Time
	RevokedAt *time.Time
}
