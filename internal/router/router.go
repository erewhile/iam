package router

import (
	"github.com/erewhile/iam/internal/database"
	"github.com/erewhile/iam/internal/middleware"
	"github.com/erewhile/iam/internal/wire"
	"github.com/erewhile/iam/pkg/response"
	"github.com/gin-gonic/gin"
)

func Init(e *gin.Engine) {
	client := database.GetDB()
	app := wire.InitApp(client)

	// JSON Web Key Set
	e.GET("/.well-known/jwks.json", app.Cert.JWKS)

	api := e.Group("/api/v1")

	// Public auth endpoints
	auth := api.Group("/auth")
	{
		auth.POST("/login", app.User.Login)
		auth.POST("/refresh", app.User.Refresh)
	}

	// Protected routes
	protected := api.Group("")
	protected.Use(middleware.Auth())

	users := protected.Group("/users")
	{
		users.GET("/me", app.User.Profile)
		users.POST("/logout", app.User.Logout)
	}

	e.NoRoute(func(c *gin.Context) {
		response.NotFound(c.Writer)
	})
}
