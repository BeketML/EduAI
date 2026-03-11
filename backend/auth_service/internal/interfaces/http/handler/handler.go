package handler

import (
	"auth_service/internal/infrastructure/logger"
	"auth_service/internal/usecase"
	"github.com/gin-contrib/cors"

	"github.com/gin-gonic/gin"
	files "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

type Handler struct {
	service *usecase.Service
	log     *logger.SlogLogger
}

func NewHandler(service *usecase.Service, log *logger.SlogLogger) *Handler {
	return &Handler{service: service, log: log}
}

func (h *Handler) InitRouter() *gin.Engine {
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())

	r.GET("/swagger/*any", ginSwagger.WrapHandler(files.Handler))

	r.Use(cors.New(cors.Config{
		AllowOrigins: []string{"http://localhost:3000"},
		AllowMethods: []string{"GET", "POST", "PUT", "DELETE"},
		AllowHeaders: []string{"Authorization", "Content-Type"},
	}))

	api := r.Group("/api/v1")

	auth := api.Group("/auth")
	{
		// PUBLIC
		auth.POST("/register", h.signUp)
		auth.POST("/login", h.signIn)
		auth.POST("/refresh", h.refresh)

		// PROTECTED
		protected := auth.Group("/")
		protected.Use(h.userIdentity)
		{
			protected.POST("/logout", h.logout)
			protected.GET("/me", h.me)
		}
	}

	return r
}
