package req

import (
	"time"

	"github.com/erewhile/iam/internal/model"
	"github.com/google/uuid"
)

type TokenList struct {
	UserID int `form:"user_id,omitempty"`
	Pagination
}

type TokenCreate struct {
	UserID        int
	Jti           uuid.UUID
	SessionID     uuid.UUID
	ApplicationID *int
	Type          model.TokenType
	TokenHash     []byte
	IP            string
	UserAgent     string
	ExpiresAt     time.Time
}

type TokenRevokePathParams struct {
	TokenID int `uri:"id" binding:"required"`
}
