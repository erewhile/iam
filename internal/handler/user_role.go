package handler

import "github.com/erewhile/iam/internal/service"

type UserRoleHandler struct {
	srv *service.UserRoleService
}

func NewUserRoleHandler(srv *service.UserRoleService) *UserRoleHandler {
	return &UserRoleHandler{srv}
}
