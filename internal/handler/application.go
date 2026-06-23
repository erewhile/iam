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

type ApplicationHandler struct {
	srv *service.ApplicationService
}

func NewApplicationHandler(srv *service.ApplicationService) *ApplicationHandler {
	return &ApplicationHandler{srv}
}

func (h *ApplicationHandler) List(c *gin.Context) {
	var params req.ApplicationList
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

func (h *ApplicationHandler) Info(c *gin.Context) {
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

func (h *ApplicationHandler) Create(c *gin.Context) {
	var body req.ApplicationCreate
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

func (h *ApplicationHandler) Update(c *gin.Context) {
	var params req.ApplicationUpdatePathParams
	if err := c.ShouldBindUri(&params); err != nil {
		response.Fail(c.Writer, code.Parameter)
		return
	}

	var body req.ApplicationUpdate
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

func (h *ApplicationHandler) Delete(c *gin.Context) {
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
