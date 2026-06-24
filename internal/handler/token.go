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

type TokenHandler struct {
	srv *service.TokenService
}

func NewTokenHandler(srv *service.TokenService) *TokenHandler {
	return &TokenHandler{srv: srv}
}

func (h *TokenHandler) List(c *gin.Context) {
	var params req.TokenList
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

func (h *TokenHandler) Info(c *gin.Context) {
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

func (h *TokenHandler) Revoke(c *gin.Context) {
	var params req.TokenRevokePathParams
	if err := c.ShouldBindUri(&params); err != nil {
		response.Fail(c.Writer, code.Parameter)
		return
	}

	ctx := c.Request.Context()
	if err := h.srv.Revoke(ctx, params); err != nil {
		response.Custom(c.Writer, http.StatusOK, err.Error())
		return
	}

	response.OK(c.Writer)
}
