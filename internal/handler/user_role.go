package handler

import (
	"github.com/erewhile/iam/internal/service"
	"github.com/gin-gonic/gin"
)

type UserRoleHandler struct {
	srv *service.UserRoleService
}

func NewUserRoleHandler(srv *service.UserRoleService) *UserRoleHandler {
	return &UserRoleHandler{srv}
}

func (h *UserRoleHandler) Assign(c *gin.Context) {}

func (h *UserRoleHandler) Roles(c *gin.Context) {}
