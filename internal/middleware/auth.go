package middleware

import (
	"fmt"
	"net/http"

	"github.com/erewhile/iam/cmd/flags"
	"github.com/erewhile/iam/config"
	"github.com/erewhile/iam/internal/cache/rds"
	"github.com/erewhile/iam/internal/consts"
	"github.com/erewhile/iam/internal/dto/req"
	"github.com/erewhile/iam/internal/logger"
	"github.com/erewhile/iam/internal/token"
	"github.com/erewhile/iam/pkg/response"
	"github.com/erewhile/iam/pkg/utils"
	"github.com/gin-gonic/gin"
)

func Auth() gin.HandlerFunc {
	tokenCache := rds.NewTokenCache()

	return func(c *gin.Context) {
		cookieUtil := utils.NewCookieUtil(!flags.Debug)
		accessToken, err := cookieUtil.Get(c.Request, config.Get().Token.AccessTokenCookieKey)
		if err != nil || accessToken == "" {
			response.Custom(c.Writer, http.StatusUnauthorized, "missing access token")
			c.Abort()
			return
		}

		claims, payload, err := token.Validate(accessToken, req.GetRequestMeta(c.Request), []byte(config.Get().Token.Aad), token.TokenTypeAccess)
		if err != nil {
			response.Custom(c.Writer, http.StatusUnauthorized, "invalid token")
			c.Abort()
			return
		}

		logger.Info(fmt.Sprintf("debug: payload.Roles = %+v, len=%d", payload.Roles, len(payload.Roles)))

		online, err := tokenCache.ExistsAccess(c.Request.Context(), claims.SessionID)
		if err != nil || !online {
			response.Custom(c.Writer, http.StatusUnauthorized, "session expired or logged out")
			c.Abort()
			return
		}

		c.Set(consts.MiddlewareUserID, payload.UserID)
		c.Set(consts.MiddlewareUserUUID, payload.UserUUID)
		c.Set(consts.MiddlewareSessionID, claims.SessionID)
		c.Set(consts.MiddlewareApplicationID, payload.ApplicationID)
		c.Set(consts.MiddlewareRoles, payload.Roles)

		c.Next()
	}
}

func RequireRoles(codes ...string) gin.HandlerFunc {
	allowed := make(map[string]struct{}, len(codes))
	for _, c := range codes {
		allowed[c] = struct{}{}
	}

	return func(c *gin.Context) {
		rolesVal, _ := c.Get(consts.MiddlewareRoles)
		roles, _ := rolesVal.([]string)

		for _, r := range roles {
			if _, ok := allowed[r]; ok {
				c.Next()
				return
			}
		}

		response.Custom(c.Writer, http.StatusForbidden, "insufficient permissions")
		c.Abort()
	}
}
