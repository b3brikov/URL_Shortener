package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (h *Handler) InitRoutes() *gin.Engine {
	router := gin.New()

	router.GET("/health", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	test := router.Group("/test")
	test.Use(h.Auth())
	test.POST("/shorten", h.TestShorten)
	router.GET("/:code", h.GoToOriginal)
	shorten := router.Group("/shorten")
	shorten.Use(h.Auth())
	shorten.POST("/new", h.CreateNewCode)
	auth := router.Group("/auth")
	auth.POST("/register", h.Register)
	auth.GET("/login", h.Login)
	return router
}
