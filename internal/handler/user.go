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
	var params req.UserLogin
	if err := c.ShouldBindJSON(&params); err != nil {
		response.Fail(c.Writer, code.Parameter)
		return
	}

	params.RequestMeta = req.GetRequestMeta(c.Request)
	ctx := c.Request.Context()

	tokenPair, err := h.srv.Login(ctx, params)
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

func (h *UserHandler) Refresh(c *gin.Context) {
	refreshToken, err := getCookie(c.Request, config.Get().Token.RefreshTokenCookieKey)
	if err != nil || refreshToken == "" {
		response.Custom(c.Writer, http.StatusOK, "missing refresh token")
		return
	}

	param := req.UserRefresh{
		Token:       refreshToken,
		RequestMeta: req.GetRequestMeta(c.Request),
	}

	ctx := c.Request.Context()
	tokenPair, err := h.srv.Refresh(ctx, param)
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

func (h *UserHandler) Logout(c *gin.Context) {
	sessionIDVal, exists := c.Get(consts.MiddlewareSessionID)
	if !exists {
		response.Custom(c.Writer, http.StatusOK, "missing session")
		return
	}

	sessionID, ok := sessionIDVal.(uuid.UUID)
	if !ok {
		response.Custom(c.Writer, http.StatusOK, "invalid session type")
		return
	}

	ctx := c.Request.Context()
	if err := h.srv.Logout(ctx, sessionID); err != nil {
		response.Custom(c.Writer, http.StatusOK, err.Error())
		return
	}

	deleteCookie(c.Writer, config.Get().Token.AccessTokenCookieKey)
	deleteCookie(c.Writer, config.Get().Token.RefreshTokenCookieKey)

	response.OK(c.Writer)
}

func (h *UserHandler) List(c *gin.Context) {
	var params req.UserList
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

func (h *UserHandler) Info(c *gin.Context) {
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

func (h *UserHandler) Create(c *gin.Context) {
	var params req.UserCreate
	if err := c.ShouldBindJSON(&params); err != nil {
		response.Fail(c.Writer, code.Parameter)
		return
	}

	ctx := c.Request.Context()
	if err := h.srv.Create(ctx, params); err != nil {
		response.Custom(c.Writer, http.StatusOK, err.Error())
		return
	}

	response.OK(c.Writer)
}

func (h *UserHandler) Update(c *gin.Context) {
	var params req.UserUpdatePathParams
	if err := c.ShouldBindUri(&params); err != nil {
		response.Fail(c.Writer, code.Parameter)
		return
	}

	var body req.UserUpdate
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

func (h *UserHandler) Delete(c *gin.Context) {
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
