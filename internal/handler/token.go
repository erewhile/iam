package handler

import "github.com/erewhile/iam/internal/service"

type TokenHandler struct {
	srv *service.TokenService
}

func NewTokenHandler(srv *service.TokenService) *TokenHandler {
	return &TokenHandler{srv}
}
