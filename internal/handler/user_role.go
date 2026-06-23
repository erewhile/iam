package handler

import (
	"github.com/erewhile/iam/internal/dto/req"
	"github.com/erewhile/iam/internal/service"
	"github.com/erewhile/iam/pkg/response"
	"github.com/gin-gonic/gin"
)

type UserRoleHandler struct {
	srv *service.UserRoleService
}

func NewUserRoleHandler(srv *service.UserRoleService) *UserRoleHandler {
	return &UserRoleHandler{srv}
}

func (h *UserRoleHandler) Roles(c *gin.Context) {
	var params req.UserRoleRoles
	if err := c.ShouldBindUri(&params); err != nil {
		response.BadRequest(c.Writer, err.Error())
		return
	}

	roles, err := h.srv.Roles(c.Request.Context(), params)
	if err != nil {
		response.BadRequest(c.Writer, err.Error())
		return
	}

	response.OkData(c.Writer, roles)
}

func (h *UserRoleHandler) Assign(c *gin.Context) {
	var params req.UserRoleAssignPathParams
	if err := c.ShouldBindUri(&params); err != nil {
		response.BadRequest(c.Writer, err.Error())
		return
	}

	var body req.UserRoleAssign
	if err := c.ShouldBindJSON(&body); err != nil {
		response.BadRequest(c.Writer, err.Error())
		return
	}

	if err := h.srv.Assign(c.Request.Context(), params, body); err != nil {
		response.BadRequest(c.Writer, err.Error())
		return
	}

	response.OK(c.Writer)
}
