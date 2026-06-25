package middleware

import (
	"net/http"

	"github.com/erewhile/iam/cmd/flags"
	"github.com/erewhile/iam/config"
	"github.com/erewhile/iam/internal/cache/rds"
	"github.com/erewhile/iam/internal/consts"
	"github.com/erewhile/iam/internal/token"
	"github.com/erewhile/iam/pkg/response"
	"github.com/erewhile/iam/pkg/utils"
	"github.com/gin-gonic/gin"
)

func Auth() gin.HandlerFunc {
	tokenCache := rds.NewTokenCache() // 只在中间件初始化时创建一次

	return func(c *gin.Context) {
		cookieUtil := utils.NewCookieUtil(!flags.Debug)
		accessToken, err := cookieUtil.Get(c.Request, config.Get().Token.AccessTokenCookieKey)
		if err != nil || accessToken == "" {
			response.Custom(c.Writer, http.StatusUnauthorized, "missing access token")
			c.Abort()
			return
		}

		claims, payload, err := token.Validate(accessToken, []byte(config.Get().Token.Aad), token.TokenTypeAccess)
		if err != nil {
			response.Custom(c.Writer, http.StatusUnauthorized, "invalid token")
			c.Abort()
			return
		}

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
