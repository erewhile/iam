package resp

import "github.com/google/uuid"

type UserProfile struct {
	UserID   int       `json:"user_id"`
	UserUUID uuid.UUID `json:"user_uuid"`
}
