package handler

import (
	"github.com/erewhile/iam/internal/service"
	"github.com/gin-gonic/gin"
)

type TokenHandler struct {
	srv *service.TokenService
}

func NewTokenHandler(srv *service.TokenService) *TokenHandler {
	return &TokenHandler{srv}
}

func (h *TokenHandler) List(c *gin.Context) {}

func (h *TokenHandler) Info(c *gin.Context) {}

func (h *TokenHandler) Revoke(c *gin.Context) {}
