package handler

import (
	"net/http"

	"github.com/erewhile/iam/internal/dto/req"
	"github.com/erewhile/iam/internal/service"
	"github.com/erewhile/iam/pkg/response"
	"github.com/erewhile/iam/pkg/response/code"
	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	srv *service.UserService
}

func NewUserHandler(srv *service.UserService) *UserHandler {
	return &UserHandler{srv}
}

func (h *UserHandler) Login(c *gin.Context) {
	var param req.UserLogin
	if err := c.ShouldBindJSON(&param); err != nil {
		response.Fail(c.Writer, code.Parameter)
		return
	}

	param.RequestMeta = req.GetRequestMeta(c.Request)
	ctx := c.Request.Context()
	if err := h.srv.Login(ctx, param); err != nil {
		response.Custom(c.Writer, http.StatusOK, err.Error())
		return
	}

	response.OK(c.Writer)
}

func (h *UserHandler) Profile(c *gin.Context) {}

func (h *UserHandler) Refresh(c *gin.Context) {}

func (h *UserHandler) Logout(c *gin.Context) {}
