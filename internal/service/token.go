package service

import "github.com/erewhile/iam/internal/repository"

type TokenService struct {
	repo repository.TokenRepository
}

func NewTokenService(repo repository.TokenRepository) *TokenService {
	return &TokenService{repo}
}
