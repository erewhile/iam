package model

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

func (s UserStatus) IsValid() bool {
	switch s {
	case UserStatusPending, UserStatusActive, UserStatusDisabled:
		return true
	default:
		return false
	}
}
