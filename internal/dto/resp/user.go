package resp

import (
	"github.com/google/uuid"
)

type UserProfile struct {
	UserID   int       `json:"user_id"`
	UserUUID uuid.UUID `json:"user_uuid"`
	Roles    []string  `json:"roles,omitempty"`
}

type UserListItem struct {
	ID           int       `json:"id"`
	Username     string    `json:"username"`
	UUID         uuid.UUID `json:"uuid"`
	Email        string    `json:"email"`
	StatusDetail string    `json:"status_detail"`
}

type UserInfo struct {
	ID           int       `json:"id"`
	Username     string    `json:"username"`
	UUID         uuid.UUID `json:"uuid"`
	Email        string    `json:"email"`
	StatusDetail string    `json:"status_detail"`
}
