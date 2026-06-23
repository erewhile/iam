package req

import (
	"time"

	"github.com/erewhile/iam/internal/model"
	"github.com/google/uuid"
)

type TokenCreate struct {
	UserID    int
	JTI       uuid.UUID
	SessionID uuid.UUID
	Type      model.TokenType
	TokenHash []byte
	IP        string
	UserAgent string
	ExpiresAt time.Time
}
