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

	oauthPublic := api.Group("/oauth")
	{
		oauthPublic.POST("/token", app.OAuth.ExchangeToken)
	}

	// Protected routes
	protected := api.Group("")
	protected.Use(middleware.Auth())

	protectedAuth := protected.Group("/auth")
	{
		protectedAuth.POST("/logout", app.User.Logout)
	}

	oauthProtected := protected.Group("/oauth")
	{
		oauthProtected.GET("/authorize", app.OAuth.Authorize)
	}

	users := protected.Group("/users")
	{
		users.GET("/me", app.User.Profile)

		users.GET("", app.User.List)
		users.GET("/:id", app.User.Info)
		users.POST("", app.User.Create)
		users.PUT("/:id", app.User.Update)
		users.DELETE("/:id", app.User.Delete)

		users.GET("/:id/roles", app.UserRole.Roles)
		users.PUT("/:id/roles", app.UserRole.Assign)
	}

	tokens := protected.Group("/tokens")
	{
		tokens.GET("", app.Token.List)
		tokens.GET("/:id", app.Token.Info)
		tokens.DELETE("/:id", app.Token.Revoke)
	}

	roles := protected.Group("/roles")
	{
		roles.GET("", app.Role.List)
		roles.GET("/:id", app.Role.Info)
		roles.POST("", app.Role.Create)
		roles.PUT("/:id", app.Role.Update)
		roles.DELETE("/:id", app.Role.Delete)
	}

	e.NoRoute(func(c *gin.Context) {
		response.NotFound(c.Writer)
	})
}
