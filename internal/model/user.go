package model

import "github.com/google/uuid"

type UserStatus uint8

const (
	UserStatusPending UserStatus = iota
	UserStatusActive
	UserStatusDisabled
)

func (s UserStatus) String() string {
	switch s {
	case UserStatusPending:
		return "pending"
	case UserStatusActive:
		return "active"
	case UserStatusDisabled:
		return "disabled"
	default:
		return "unknown"
	}
}

type User struct {
	ID       int
	Username string
	UUID     uuid.UUID
	Email    string
	Status   UserStatus
}
