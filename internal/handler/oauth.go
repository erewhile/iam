package handler

import (
	"errors"
	"net/http"
	"net/url"

	"github.com/erewhile/iam/cmd/flags"
	"github.com/erewhile/iam/config"
	"github.com/erewhile/iam/internal/cache/rds"
	"github.com/erewhile/iam/internal/consts"
	"github.com/erewhile/iam/internal/dto/req"
	"github.com/erewhile/iam/internal/logger"
	"github.com/erewhile/iam/internal/service"
	"github.com/erewhile/iam/pkg/response"
	"github.com/erewhile/iam/pkg/response/code"
	"github.com/erewhile/iam/pkg/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type OAuthHandler struct {
	userService  *service.UserService
	oauthService *service.OAuthService
	cache        rds.TokenCache
}

func NewOAuthHandler(
	userService *service.UserService,
	oauthService *service.OAuthService,
	tokenCache rds.TokenCache,
) *OAuthHandler {
	return &OAuthHandler{
		userService:  userService,
		oauthService: oauthService,
		cache:        tokenCache,
	}
}

func (h *OAuthHandler) Authorize(c *gin.Context) {
	var params req.Authorize
	if err := c.ShouldBindQuery(&params); err != nil {
		response.Fail(c.Writer, code.Parameter)
		return
	}

	ctx := c.Request.Context()

	if _, err := h.oauthService.ValidateAuthorize(ctx, params.ClientID, params.RedirectURI); err != nil {
		switch {
		case errors.Is(err, service.ErrClientNotFound):
			response.BadRequest(c.Writer, "invalid client_id")
		case errors.Is(err, service.ErrRedirectURIInvalid):
			response.BadRequest(c.Writer, "redirect_uri not registered")
		default:
			logger.Error("validate authorize failed", err)
			response.InternalServer(c.Writer)
		}
		return
	}

	cookieUtil := utils.NewCookieUtil(!flags.Debug)
	sid, _ := cookieUtil.Get(c.Request, config.Get().Session.CookieKey)
	userID, userUUID, ok := h.userService.CheckSession(ctx, sid)
	if !ok {
		v := url.Values{}
		v.Set("redirect", c.Request.URL.RequestURI())
		loginURL := consts.AuthLoginPath + "?" + v.Encode()
		c.Redirect(http.StatusFound, loginURL)
		return
	}
	authCode, err := utils.RandomString(32)
	if err != nil {
		logger.Error("failed to generate oauth code", err)
		response.InternalServer(c.Writer)
		return
	}

	appSessionID := uuid.New()

	if err := h.cache.SetCode(ctx, authCode, rds.OAuthCodePayload{
		UserID:    userID,
		UserUUID:  userUUID,
		SessionID: appSessionID,
		ClientID:  params.ClientID,
	}); err != nil {
		logger.Error("failed to save code to cache", err)
		response.InternalServer(c.Writer)
		return
	}

	target, err := url.Parse(params.RedirectURI)
	if err != nil {
		logger.Error("invalid redirect_uri", err)
		response.InternalServer(c.Writer)
		return
	}
	q := target.Query()
	q.Set("code", authCode)
	target.RawQuery = q.Encode()

	c.Redirect(http.StatusFound, target.String())
}

func (h *OAuthHandler) ExchangeToken(c *gin.Context) {
	var params req.ExchangeToken
	if err := c.ShouldBindJSON(&params); err != nil {
		response.Fail(c.Writer, code.Parameter)
		return
	}

	ctx := c.Request.Context()

	if _, err := h.oauthService.ValidateClient(ctx, params.ClientID, params.ClientSecret); err != nil {
		switch {
		case errors.Is(err, service.ErrClientNotFound), errors.Is(err, service.ErrClientSecretWrong):
			response.BadRequest(c.Writer, "invalid client credentials")
		default:
			logger.Error("validate client failed", err)
			response.InternalServer(c.Writer)
		}
		return
	}

	payload, err := h.cache.GetAndDelCode(ctx, params.Code)
	if err != nil {
		response.BadRequest(c.Writer, "invalid or expired code")
		return
	}

	if payload.ClientID != params.ClientID {
		response.BadRequest(c.Writer, "client mismatch")
		return
	}

	meta := req.GetRequestMeta(c.Request)
	tokenPair, err := h.userService.LoginWithOAuthCode(ctx, payload, meta)
	if err != nil {
		response.Custom(c.Writer, http.StatusInternalServerError, "failed to issue token")
		return
	}

	response.OkData(c.Writer, tokenPair)
}
