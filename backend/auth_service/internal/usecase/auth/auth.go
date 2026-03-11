package auth

import (
	"auth_service/internal/domain"
	"auth_service/internal/infrastructure/logger"
	"auth_service/internal/infrastructure/repository"
	"context"
	"errors"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type TokenManager interface {
	NewAccessToken(userID string) (string, error)
	NewRefreshToken(userID string) (string, error)
	ParseAccessToken(ctx context.Context, token string) (string, error)
	ParseRefreshToken(ctx context.Context, token string) (string, error)
}

var ErrInvalidRefreshToken = errors.New("invalid refresh token")

type ServiceAuth struct {
	repo   repository.Auth
	log    *logger.SlogLogger
	tokens TokenManager
}

func NewServiceAuth(repo repository.Auth, log *logger.SlogLogger, tokens TokenManager) *ServiceAuth {
	return &ServiceAuth{
		repo:   repo,
		log:    log,
		tokens: tokens,
	}
}

func (s *ServiceAuth) Register(ctx context.Context, user domain.User) (uuid.UUID, error) {
	hash, err := hashPassword(user.Password)
	if err != nil {
		s.log.Error(ctx, "service auth: hash password error", err.Error())
		return uuid.UUID{}, err
	}
	user.Password = hash
	return s.repo.CreateUser(ctx, user)
}

func (s *ServiceAuth) Login(ctx context.Context, username, password string) (string, string, error) {
	user, err := s.repo.GetUserByUsername(ctx, username)
	if err != nil {
		s.log.Error(ctx, "repo auth: get user error", err.Error())
		return "", "", err
	}

	if err := checkPassword(password, user.Password); err != nil {
		s.log.Error(ctx, "repo auth: check password error", err.Error())
		return "", "", err
	}

	// Generate Access Token
	access, err := s.tokens.NewAccessToken(user.Id.String())
	if err != nil {
		s.log.Error(ctx, "service auth: access token generation error", err.Error())
		return "", "", err
	}

	// Generate Refresh Token
	refresh, err := s.tokens.NewRefreshToken(user.Id.String())
	if err != nil {
		s.log.Error(ctx, "service auth: refresh token generation error", err.Error())
		return "", "", err
	}

	// 🔥 SAVE REFRESH TOKEN TO DB (REQUIRED FOR /refresh)
	err = s.repo.SaveRefreshToken(ctx, user.Id, refresh)
	if err != nil {
		s.log.Error(ctx, "service auth: save refresh token error", err.Error())
		return "", "", err
	}

	return access, refresh, nil
}

func (s *ServiceAuth) ParseAccessToken(ctx context.Context, token string) (uuid.UUID, error) {
	userIdStr, err := s.tokens.ParseAccessToken(ctx, token)
	if err != nil {
		s.log.Error(ctx, "parse token error", err.Error())
		return uuid.UUID{}, err
	}

	userID, err := uuid.Parse(userIdStr)
	if err != nil {
		return uuid.UUID{}, err
	}

	return userID, nil
}

func (s *ServiceAuth) ParseRefreshToken(ctx context.Context, token string) (string, error) {
	return s.tokens.ParseRefreshToken(ctx, token)
}
func (s *ServiceAuth) GenerateAccessToken(userId string) (string, error) {
	return s.tokens.NewAccessToken(userId)
}
func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

func checkPassword(password, hash string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}

func (s *ServiceAuth) Logout(ctx context.Context, accessToken string) error {
	userIdStr, err := s.tokens.ParseAccessToken(ctx, accessToken)
	if err != nil {
		return err
	}

	userID, err := uuid.Parse(userIdStr)
	if err != nil {
		return err
	}

	return s.repo.DeleteRefreshToken(ctx, userID)
}

func (s *ServiceAuth) Me(ctx context.Context, accessToken string) (*domain.User, error) {
	userIdStr, err := s.tokens.ParseAccessToken(ctx, accessToken)
	if err != nil {
		return nil, err
	}

	userID, err := uuid.Parse(userIdStr)
	if err != nil {
		return nil, err
	}

	user, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	return &user, nil
}
func (s *ServiceAuth) Refresh(ctx context.Context, refreshToken string) (string, string, error) {
	userIdStr, err := s.tokens.ParseRefreshToken(ctx, refreshToken)
	if err != nil {
		s.log.Error(ctx, "refresh: parse error", err.Error())
		return "", "", err
	}

	userID, err := uuid.Parse(userIdStr)
	if err != nil {
		return "", "", err
	}

	dbToken, err := s.repo.GetRefreshToken(ctx, userID)
	if err != nil {
		s.log.Error(ctx, "refresh: db token error", err.Error())
		return "", "", err
	}

	if dbToken != refreshToken {
		return "", "", ErrInvalidRefreshToken
	}

	newAccess, err := s.tokens.NewAccessToken(userID.String())
	if err != nil {
		return "", "", err
	}

	newRefresh, err := s.tokens.NewRefreshToken(userID.String())
	if err != nil {
		return "", "", err
	}

	err = s.repo.SaveRefreshToken(ctx, userID, newRefresh)
	if err != nil {
		return "", "", err
	}

	return newAccess, newRefresh, nil
}
