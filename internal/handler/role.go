package handler

import "github.com/erewhile/iam/internal/service"

type RoleHandler struct {
	srv *service.RoleService
}

func NewRoleHandler(srv *service.RoleService) *RoleHandler {
	return &RoleHandler{srv}
}
