package resp

import (
	"time"

	"github.com/google/uuid"
)

type TokenListItem struct {
	ID         int       `json:"id"`
	UserID     int       `json:"user_id"`
	Jti        uuid.UUID `json:"jti"`
	SessionID  uuid.UUID `json:"session_id"`
	TypeDetail string    `json:"type_detail"`
	IP         string    `json:"ip"`
	UserAgent  string    `json:"user_agent"`
	ExpiresAt  time.Time `json:"expires_at"`
}

type TokenInfo struct {
	ID         int       `json:"id"`
	UserID     int       `json:"user_id"`
	Jti        uuid.UUID `json:"jti"`
	SessionID  uuid.UUID `json:"session_id"`
	TypeDetail string    `json:"type_detail"`
	IP         string    `json:"ip"`
	UserAgent  string    `json:"user_agent"`
	ExpiresAt  time.Time `json:"expires_at"`
}
