package resp

import (
	"github.com/erewhile/iam/internal/model"
	"github.com/google/uuid"
)

type UserProfile struct {
	UUID     uuid.UUID `json:"uuid"`
	Email    string    `json:"email"`
	Username string    `json:"username"`
}

type UserListItem struct {
	ID           int              `json:"id"`
	Username     string           `json:"username"`
	UUID         uuid.UUID        `json:"uuid"`
	Email        string           `json:"email"`
	Status       model.UserStatus `json:"status"`
	StatusDetail string           `json:"status_detail"`
}

type UserInfo struct {
	ID           int              `json:"id"`
	Username     string           `json:"username"`
	UUID         uuid.UUID        `json:"uuid"`
	Email        string           `json:"email"`
	Status       model.UserStatus `json:"status"`
	StatusDetail string           `json:"status_detail"`
}

type UserStatusOption struct {
	Value model.UserStatus `json:"value"`
	Label string           `json:"label"`
}

type UserSelectOption struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
}
