package middleware

import (
	"github.com/erewhile/iam/config"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func CORS() gin.HandlerFunc {
	config := cors.Config{
		AllowOrigins:     config.Get().CORS.AllowOrigins,
		AllowMethods:     config.Get().CORS.AllowMethods,
		AllowHeaders:     config.Get().CORS.AllowHeaders,
		AllowCredentials: config.Get().CORS.AllowCredentials,
		MaxAge:           config.Get().CORS.MaxAge,
	}

	return cors.New(config)
}
