package model

type TokenType uint8

const (
	TokenTypeAccess TokenType = iota
	TokenTypeRefresh
)

func (t TokenType) String() string {
	switch t {
	case TokenTypeAccess:
		return "access"
	case TokenTypeRefresh:
		return "refresh"
	default:
		return "unknown"
	}
}

func (t TokenType) IsValid() bool {
	switch t {
	case TokenTypeAccess, TokenTypeRefresh:
		return true
	default:
		return false
	}
}
