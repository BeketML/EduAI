package handler

import (
	"github.com/gin-gonic/gin"
	"log/slog"
)

// ErrorResponse represents an API error response
type ErrorResponse struct {
	Message string `json:"message" example:"internal server error"`
}

// StatusResponse represents a simple status response
type StatusResponse struct {
	Status string `json:"status" example:"ok"`
}

// RegisterResponse represents registration response
type RegisterResponse struct {
	UserID string `json:"user_id" example:"01234567-89ab-cdef-0123-456789abcdef"`
}

// LoginResponse represents login response
type LoginResponse struct {
	AccessToken  string `json:"access_token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	RefreshToken string `json:"refresh_token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
}

func NewErrorResponse(c *gin.Context, statusCode int, message string) {
	slog.Error(message)
	c.AbortWithStatusJSON(statusCode, ErrorResponse{Message: message})
}
