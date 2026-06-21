package router

import (
	"github.com/erewhile/iam/internal/wire"
	"github.com/erewhile/iam/pkg/response"
	"github.com/gin-gonic/gin"
)

func Init(e *gin.Engine) {
	app := wire.InitApp()

	// JSON Web Key Set
	e.GET("/.well-known/jwks.json", app.Cert.JWKS)

	e.NoRoute(func(c *gin.Context) {
		response.NotFound(c.Writer)
	})
}
