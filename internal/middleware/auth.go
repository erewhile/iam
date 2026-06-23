package middleware

import (
	"net/http"

	"github.com/erewhile/iam/config"
	"github.com/erewhile/iam/internal/cache/redis"
	"github.com/erewhile/iam/internal/consts"
	"github.com/erewhile/iam/internal/token"
	"github.com/erewhile/iam/pkg/response"
	"github.com/gin-gonic/gin"
)

func Auth() gin.HandlerFunc {
	return func(c *gin.Context) {
		accessToken, err := c.Cookie(config.Get().Token.AccessTokenCookieKey)
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

		tokenCache := redis.NewTokenCache()
		online, err := tokenCache.ExistsAccess(c.Request.Context(), claims.SessionID)
		if err != nil || !online {
			response.Custom(c.Writer, http.StatusUnauthorized, "session expired or logged out")
			c.Abort()
			return
		}

		c.Set(consts.MiddlewareUserID, payload.UserID)
		c.Set(consts.MiddlewareUserUUID, payload.UserUUID)
		c.Set(consts.MiddlewareSessionID, claims.SessionID)

		c.Next()
	}
}
