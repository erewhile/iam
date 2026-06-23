package model

type TokenType uint8

const (
	TokenTypeAccess TokenType = iota
	TokenTypeRefresh
)
