package model

const UserSystemID = 1

type UserStatus uint8

const (
	UserStatusPending UserStatus = iota + 1
	UserStatusActive
	UserStatusDisabled
)

const (
	UserSystem   = true
	UserStandard = false
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

func AllUserStatuses() []UserStatus {
	return []UserStatus{UserStatusPending, UserStatusActive, UserStatusDisabled}
}
