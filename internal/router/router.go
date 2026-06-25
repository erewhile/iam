package router

import (
	"github.com/erewhile/iam/internal/consts"
	"github.com/erewhile/iam/internal/database"
	"github.com/erewhile/iam/internal/middleware"
	"github.com/erewhile/iam/internal/model"
	"github.com/erewhile/iam/internal/wire"
	"github.com/erewhile/iam/pkg/response"
	"github.com/gin-gonic/gin"
)

func Init(e *gin.Engine) {
	client := database.GetDB()
	app := wire.InitApp(client)

	e.Use(middleware.CORS())

	// JSON Web Key Set
	e.GET("/.well-known/jwks.json", app.Cert.JWKS)

	api := e.Group(consts.APIBase)

	// Public auth endpoints
	auth := api.Group("/auth")
	{
		auth.GET("/login", app.User.ShowLogin)
		auth.POST("/login", app.User.Login)
		auth.POST("/refresh", app.User.Refresh)
	}

	oauth := api.Group("/oauth")
	{
		oauth.GET("/authorize", app.OAuth.Authorize)
		oauth.POST("/token", app.OAuth.ExchangeToken)
	}

	// Protected routes
	protected := api.Group("")
	protected.Use(middleware.Auth())

	protected.POST("/auth/logout", app.User.Logout)

	users := protected.Group("/users")
	{
		users.GET("/me", app.User.Profile)
	}

	admin := protected.Group("")
	admin.Use(middleware.RequireRoles(model.RoleSuperAdmin))
	{
		adminUsers := admin.Group("/users")
		adminUsers.GET("", app.User.List)
		adminUsers.GET("/:id", app.User.Info)
		adminUsers.POST("", app.User.Create)
		adminUsers.PUT("/:id", app.User.Update)
		adminUsers.DELETE("/:id", app.User.Delete)
		adminUsers.GET("/:id/roles", app.UserRole.Roles)
		adminUsers.PUT("/:id/roles", app.UserRole.Assign)

		admin.Group("/tokens").
			GET("", app.Token.List)
		admin.GET("/tokens/:id", app.Token.Info)
		admin.DELETE("/tokens/:id", app.Token.Revoke)

		adminRoles := admin.Group("/roles")
		adminRoles.GET("", app.Role.List)
		adminRoles.GET("/:id", app.Role.Info)
		adminRoles.POST("", app.Role.Create)
		adminRoles.PUT("/:id", app.Role.Update)
		adminRoles.DELETE("/:id", app.Role.Delete)

		adminApps := admin.Group("/applications")
		adminApps.GET("", app.Application.List)
		adminApps.GET("/:id", app.Application.Info)
		adminApps.POST("", app.Application.Create)
		adminApps.PUT("/:id", app.Application.Update)
		adminApps.PUT("/:id/secret", app.Application.UpdateSecret)
		adminApps.DELETE("/:id", app.Application.Delete)
	}

	e.NoRoute(func(c *gin.Context) {
		response.NotFound(c.Writer)
	})
}
