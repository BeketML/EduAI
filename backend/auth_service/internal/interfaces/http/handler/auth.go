package handler

import (
	"auth_service/internal/domain"
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
)

// RegisterInput represents user registration payload
type RegisterInput struct {
	Username  string `json:"username" binding:"required" example:"john_doe"`
	Email     string `json:"email" binding:"required,email" example:"john@example.com"`
	Password  string `json:"password" binding:"required,min=6" example:"password123"`
	FirstName string `json:"first_name" binding:"required" example:"John"`
	LastName  string `json:"last_name" binding:"required" example:"Doe"`
}

// LoginInput represents user login payload
type LoginInput struct {
	Username string `json:"username" binding:"required" example:"john_doe"`
	Password string `json:"password" binding:"required" example:"password123"`
}

// @Summary Register new user
// @Description Create a new user account
// @Tags auth
// @Accept json
// @Produce json
// @Param input body RegisterInput true "Register input"
// @Success 201 {object} RegisterResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /auth/register [post]
func (h *Handler) signUp(c *gin.Context) {
	ctx := c.Request.Context()

	var input RegisterInput
	if err := c.ShouldBindJSON(&input); err != nil {
		NewErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	// Service returns uuid.UUID
	userID, err := h.service.Register(ctx, domain.User{
		Username:  input.Username,
		Password:  input.Password,
		Email:     input.Email,
		FirstName: input.FirstName,
		LastName:  input.LastName,
	})
	if err != nil {
		NewErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	// Convert uuid.UUID to string for DTO
	c.JSON(http.StatusCreated, RegisterResponse{
		UserID: userID.String(),
	})
}

// @Summary Login user
// @Description Authenticate user and return access and refresh tokens
// @Tags auth
// @Accept json
// @Produce json
// @Param input body LoginInput true "Login input"
// @Success 200 {object} LoginResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /auth/login [post]
func (h *Handler) signIn(c *gin.Context) {
	ctx := c.Request.Context()

	var input LoginInput
	if err := c.ShouldBindJSON(&input); err != nil {
		NewErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}
	at, rt, err := h.service.Login(ctx, input.Username, input.Password)
	if err != nil {
		NewErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, LoginResponse{
		AccessToken:  at,
		RefreshToken: rt,
	})
}

// RefreshInput represents refresh token payload
type RefreshInput struct {
	RefreshToken string `json:"refresh_token" binding:"required" example:"your_refresh_token"`
}

// MeResponse represents current user response
type MeResponse struct {
	ID        string `json:"id" example:"uuid"`
	Username  string `json:"username" example:"john_doe"`
	Email     string `json:"email" example:"john@example.com"`
	FirstName string `json:"first_name" example:"Aibar"`
	LastName  string `json:"last_name" example:"Tlekbay"`
}

// @Summary Refresh tokens
// @Description Get new access and refresh tokens
// @Tags auth
// @Accept json
// @Produce json
// @Param input body RefreshInput true "Refresh token"
// @Success 200 {object} LoginResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /auth/refresh [post]
func (h *Handler) refresh(c *gin.Context) {
	ctx := c.Request.Context()

	var input RefreshInput
	if err := c.ShouldBindJSON(&input); err != nil {
		NewErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	at, rt, err := h.service.Auth.Refresh(ctx, input.RefreshToken)
	if err != nil {
		NewErrorResponse(c, http.StatusUnauthorized, err.Error())
		return
	}

	c.JSON(http.StatusOK, LoginResponse{
		AccessToken:  at,
		RefreshToken: rt,
	})
}

// @Summary Logout user
// @Description Logout user by deleting refresh token
// @Tags auth
// @Security BearerAuth
// @Produce json
// @Success 200 {object} map[string]string
// @Failure 401 {object} ErrorResponse
// @Router /auth/logout [post]
func (h *Handler) logout(c *gin.Context) {
	ctx := c.Request.Context()

	token := c.GetHeader("Authorization")
	token = strings.TrimPrefix(token, "Bearer ")

	err := h.service.Auth.Logout(ctx, token)
	if err != nil {
		NewErrorResponse(c, http.StatusUnauthorized, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "logged out",
	})
}

// @Summary Get current user
// @Description Get user info from access token
// @Tags auth
// @Security BearerAuth
// @Produce json
// @Success 200 {object} MeResponse
// @Failure 401 {object} ErrorResponse
// @Router /auth/me [get]
func (h *Handler) me(c *gin.Context) {
	ctx := c.Request.Context()

	token := c.GetHeader("Authorization")
	token = strings.TrimPrefix(token, "Bearer ")

	user, err := h.service.Auth.Me(ctx, token)
	if err != nil {
		NewErrorResponse(c, http.StatusUnauthorized, err.Error())
		return
	}

	c.JSON(http.StatusOK, MeResponse{
		ID:        user.Id.String(),
		Username:  user.Username,
		Email:     user.Email,
		FirstName: user.FirstName,
		LastName:  user.LastName,
	})
}
