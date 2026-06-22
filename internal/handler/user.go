package handler

import (
	"net/http"

	"github.com/erewhile/iam/config"
	"github.com/erewhile/iam/internal/consts"
	"github.com/erewhile/iam/internal/dto/req"
	"github.com/erewhile/iam/internal/dto/resp"
	"github.com/erewhile/iam/internal/service"
	"github.com/erewhile/iam/pkg/response"
	"github.com/erewhile/iam/pkg/response/code"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
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

	tokenPair, err := h.srv.Login(ctx, param)
	if err != nil {
		response.Custom(c.Writer, http.StatusOK, err.Error())
		return
	}

	setCookie(
		c.Writer,
		config.Get().Token.AccessTokenCookieKey,
		tokenPair.AccessToken,
		int(config.Get().Token.AccessTokenTTL.Seconds()),
	)
	setCookie(
		c.Writer,
		config.Get().Token.RefreshTokenCookieKey,
		tokenPair.RefreshToken,
		int(config.Get().Token.RefreshTokenTTL.Seconds()),
	)

	response.OK(c.Writer)
}

func (h *UserHandler) Profile(c *gin.Context) {
	userID := c.GetInt(consts.MiddlewareUserID)

	uuidVal, exists := c.Get(consts.MiddlewareUserUUID)
	if !exists {
		response.Custom(c.Writer, http.StatusOK, "missing uuid")
		return
	}

	userUUID, ok := uuidVal.(uuid.UUID)
	if !ok {
		response.Custom(c.Writer, http.StatusOK, "invalid uuid type")
		return
	}

	response.OkData(c.Writer, &resp.UserProfile{
		UserID:   userID,
		UserUUID: userUUID,
	})
}

func (h *UserHandler) Refresh(c *gin.Context) {}

func (h *UserHandler) Logout(c *gin.Context) {}
