package handler

import (
	"net/http"

	"github.com/erewhile/iam/internal/dto/req"
	"github.com/erewhile/iam/internal/dto/resp"
	"github.com/erewhile/iam/internal/service"
	"github.com/erewhile/iam/pkg/response"
	"github.com/erewhile/iam/pkg/response/code"
	"github.com/gin-gonic/gin"
)

type RoleHandler struct {
	srv *service.RoleService
}

func NewRoleHandler(srv *service.RoleService) *RoleHandler {
	return &RoleHandler{srv: srv}
}

func (h *RoleHandler) List(c *gin.Context) {
	var params req.RoleList
	if err := c.ShouldBindQuery(&params); err != nil {
		response.Fail(c.Writer, code.Parameter)
		return
	}

	ctx := c.Request.Context()
	content, count, err := h.srv.List(ctx, params)
	if err != nil {
		response.Custom(c.Writer, http.StatusOK, err.Error())
		return
	}

	response.OkData(c.Writer, resp.List{
		Content: content,
		Count:   count,
	})
}

func (h *RoleHandler) Info(c *gin.Context) {
	var params req.InfoPathParams
	if err := c.ShouldBindUri(&params); err != nil {
		response.Fail(c.Writer, code.Parameter)
		return
	}

	ctx := c.Request.Context()
	info, err := h.srv.Info(ctx, params)
	if err != nil {
		response.Custom(c.Writer, http.StatusOK, err.Error())
		return
	}

	response.OkData(c.Writer, info)
}

func (h *RoleHandler) Create(c *gin.Context) {
	var body req.RoleCreate
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Fail(c.Writer, code.Parameter)
		return
	}

	ctx := c.Request.Context()
	if err := h.srv.Create(ctx, body); err != nil {
		response.Custom(c.Writer, http.StatusOK, err.Error())
		return
	}

	response.OK(c.Writer)
}

func (h *RoleHandler) Update(c *gin.Context) {
	var params req.RoleUpdatePathParams
	if err := c.ShouldBindUri(&params); err != nil {
		response.Fail(c.Writer, code.Parameter)
		return
	}

	var body req.RoleUpdate
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Fail(c.Writer, code.Parameter)
		return
	}

	ctx := c.Request.Context()
	if err := h.srv.Update(ctx, params, body); err != nil {
		response.Custom(c.Writer, http.StatusOK, err.Error())
		return
	}

	response.OK(c.Writer)
}

func (h *RoleHandler) Delete(c *gin.Context) {
	var params req.DeletePathParams
	if err := c.ShouldBindUri(&params); err != nil {
		response.Fail(c.Writer, code.Parameter)
		return
	}

	ctx := c.Request.Context()
	if err := h.srv.Delete(ctx, params); err != nil {
		response.Custom(c.Writer, http.StatusOK, err.Error())
		return
	}

	response.OK(c.Writer)
}
