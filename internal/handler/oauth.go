package handler

import (
	"fmt"
	"net/http"

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
	userService *service.UserService
}

func NewOAuthHandler(userService *service.UserService) *OAuthHandler {
	return &OAuthHandler{userService}
}

func (h *OAuthHandler) Authorize(c *gin.Context) {
	var params req.Authorize
	if err := c.ShouldBindQuery(&params); err != nil {
		response.Fail(c.Writer, code.Parameter)
		return
	}

	userID := c.GetInt(consts.MiddlewareUserID)

	uuidVal, exists := c.Get(consts.MiddlewareUserUUID)
	if !exists {
		response.Custom(c.Writer, http.StatusOK, "missing uuid")
		c.Abort()
		return
	}
	userUUID, ok := uuidVal.(uuid.UUID)
	if !ok {
		response.Custom(c.Writer, http.StatusOK, "invalid uuid type")
		c.Abort()
		return
	}

	sessionIDVal, exists := c.Get(consts.MiddlewareSessionID)
	if !exists {
		response.Custom(c.Writer, http.StatusOK, "missing session id")
		c.Abort()
		return
	}
	sessionID, ok := sessionIDVal.(uuid.UUID)
	if !ok {
		response.Custom(c.Writer, http.StatusOK, "invalid session id type")
		c.Abort()
		return
	}

	code, err := utils.GenerateRandomString(32)
	if err != nil {
		logger.Error("failed to generate oauth code", err)
		response.InternalServer(c.Writer)
		return
	}

	tokenCache := rds.NewTokenCache()
	err = tokenCache.SetCode(c.Request.Context(), code, rds.OAuthCodePayload{
		UserID:    userID,
		UserUUID:  userUUID,
		SessionID: sessionID,
		ClientID:  params.ClientID,
	})
	if err != nil {
		logger.Error("failed to save code to cache", err)
		response.InternalServer(c.Writer)
		return
	}

	targetURL := fmt.Sprintf("%s?code=%s", params.RedirectURI, code)
	c.Redirect(http.StatusFound, targetURL)
}

func (h *OAuthHandler) ExchangeToken(c *gin.Context) {
	var params req.ExchangeToken
	if err := c.ShouldBindJSON(&params); err != nil {
		response.Fail(c.Writer, code.Parameter)
		return
	}

	ctx := c.Request.Context()
	tokenCache := rds.NewTokenCache()

	payload, err := tokenCache.GetAndDelCode(ctx, params.Code)
	if err != nil {
		response.BadRequest(c.Writer, "invalid or expired code")
		return
	}

	if payload.ClientID != params.ClientID {
		response.BadRequest(c.Writer, "client mismatch")
		return
	}

	mockMeta := req.GetRequestMeta(c.Request)
	tokenPair, err := h.userService.LoginWithOAuthCode(ctx, payload, mockMeta)
	if err != nil {
		response.Custom(c.Writer, http.StatusInternalServerError, "failed to issue token")
		return
	}

	response.OkData(c.Writer, tokenPair)
}
