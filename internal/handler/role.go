package handler

import (
	"github.com/erewhile/iam/internal/service"
	"github.com/gin-gonic/gin"
)

type RoleHandler struct {
	srv *service.RoleService
}

func NewRoleHandler(srv *service.RoleService) *RoleHandler {
	return &RoleHandler{srv}
}

func (h *RoleHandler) List(c *gin.Context) {}

func (h *RoleHandler) Info(c *gin.Context) {}

func (h *RoleHandler) Create(c *gin.Context) {}

func (h *RoleHandler) Update(c *gin.Context) {}

func (h *RoleHandler) Delete(c *gin.Context) {}
