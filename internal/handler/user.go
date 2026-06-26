package handler

import (
	"errors"
	"net/http"
	"net/url"
	"strings"
	"text/template"

	"github.com/erewhile/iam/cmd/flags"
	"github.com/erewhile/iam/config"
	"github.com/erewhile/iam/internal/consts"
	"github.com/erewhile/iam/internal/dto/req"
	"github.com/erewhile/iam/internal/dto/resp"
	"github.com/erewhile/iam/internal/logger"
	"github.com/erewhile/iam/internal/model"
	"github.com/erewhile/iam/internal/service"
	"github.com/erewhile/iam/pkg/response"
	"github.com/erewhile/iam/pkg/response/code"
	"github.com/erewhile/iam/pkg/utils"
	"github.com/erewhile/iam/templates"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type UserHandler struct {
	srv *service.UserService
}

func NewUserHandler(srv *service.UserService) *UserHandler {
	return &UserHandler{srv: srv}
}

var loginTpl = template.Must(template.ParseFS(templates.FS, "login.html"))

func (h *UserHandler) validateRedirect(redirect string) (string, error) {
	if redirect == "" {
		return "/", nil
	}

	inputURL, err := url.Parse(redirect)
	if err != nil {
		return "/", nil
	}

	if inputURL.IsAbs() || inputURL.Host != "" {
		return "", errors.New("prohibited external redirect")
	}

	if !strings.HasPrefix(redirect, "/") {
		return "/", nil
	}

	return redirect, nil
}

func (h *UserHandler) ShowLogin(c *gin.Context) {
	var params req.UserShowLogin
	if err := c.ShouldBindQuery(&params); err != nil {
		response.Fail(c.Writer, code.Parameter)
		return
	}

	finalRedirect, err := h.validateRedirect(params.Redirect)
	if err != nil {
		logger.Error("validate redirect failed", err)
		finalRedirect = ""
	}

	if params.Redirect != "" && finalRedirect == "" {
		logger.Warn("Redirect URL was blocked or invalid", "input", params.Redirect)
	}

	c.Header("Content-Type", "text/html; charset=utf-8")
	c.Writer.WriteHeader(http.StatusOK)

	err = loginTpl.Execute(c.Writer, gin.H{
		"Redirect":    finalRedirect,
		"LoginApiUrl": consts.AuthLoginPath,
	})
	if err != nil {
		logger.Error("render login template failed", err)
	}
}

func (h *UserHandler) Login(c *gin.Context) {
	var params req.UserLogin
	if err := c.ShouldBindJSON(&params); err != nil {
		response.Fail(c.Writer, code.Parameter)
		return
	}

	params.RequestMeta = req.GetRequestMeta(c.Request)
	ctx := c.Request.Context()

	tokenPair, sid, err := h.srv.Login(ctx, params)
	if err != nil {
		response.BadRequest(c.Writer, err.Error())
		return
	}

	cookieUtil := utils.NewCookieUtil(!flags.Debug)
	cookieUtil.Set(
		c.Writer,
		config.Get().Token.AccessTokenCookieKey,
		tokenPair.AccessToken,
		int(config.Get().Token.AccessTokenTTL.Seconds()),
	)
	cookieUtil.Set(
		c.Writer,
		config.Get().Token.RefreshTokenCookieKey,
		tokenPair.RefreshToken,
		int(config.Get().Token.RefreshTokenTTL.Seconds()),
	)
	cookieUtil.Set(
		c.Writer,
		config.Get().Session.CookieKey,
		sid,
		int(config.Get().Session.CookieTTL.Seconds()),
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

	rolesVal, _ := c.Get(consts.MiddlewareRoles)
	roles, _ := rolesVal.([]string)

	response.OkData(c.Writer, &resp.UserProfile{
		UserID:   userID,
		UserUUID: userUUID,
		Roles:    roles,
	})
}

func (h *UserHandler) Refresh(c *gin.Context) {
	cookieUtil := utils.NewCookieUtil(!flags.Debug)
	refreshToken, err := cookieUtil.Get(c.Request, config.Get().Token.RefreshTokenCookieKey)
	if err != nil || refreshToken == "" {
		response.Custom(c.Writer, http.StatusOK, "missing refresh token")
		return
	}

	body := req.UserRefresh{
		Token:       refreshToken,
		RequestMeta: req.GetRequestMeta(c.Request),
	}

	ctx := c.Request.Context()
	tokenPair, err := h.srv.Refresh(ctx, body)
	if err != nil {
		response.BadRequest(c.Writer, err.Error())
		return
	}

	cookieUtil.Set(
		c.Writer,
		config.Get().Token.AccessTokenCookieKey,
		tokenPair.AccessToken,
		int(config.Get().Token.AccessTokenTTL.Seconds()),
	)
	cookieUtil.Set(
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

	cookieUtil := utils.NewCookieUtil(!flags.Debug)
	iamSID, _ := cookieUtil.Get(c.Request, config.Get().Session.CookieKey)

	ctx := c.Request.Context()
	if err := h.srv.Logout(ctx, sessionID, iamSID); err != nil {
		response.BadRequest(c.Writer, err.Error())
		return
	}

	cookieUtil.Delete(c.Writer, config.Get().Token.AccessTokenCookieKey)
	cookieUtil.Delete(c.Writer, config.Get().Token.RefreshTokenCookieKey)
	cookieUtil.Delete(c.Writer, config.Get().Session.CookieKey)

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
		response.BadRequest(c.Writer, err.Error())
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
		response.BadRequest(c.Writer, err.Error())
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
		response.BadRequest(c.Writer, err.Error())
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
		response.BadRequest(c.Writer, err.Error())
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
		response.BadRequest(c.Writer, err.Error())
		return
	}

	response.OK(c.Writer)
}

func (h *UserHandler) UserStatuses(c *gin.Context) {
	statuses := model.AllUserStatuses()
	options := make([]resp.UserStatusOption, 0, len(statuses))
	for _, s := range statuses {
		options = append(options, resp.UserStatusOption{Value: s, Label: s.String()})
	}

	response.OkData(c.Writer, options)
}
