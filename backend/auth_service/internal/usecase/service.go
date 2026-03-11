package usecase

import (
	"auth_service/internal/domain"
	"auth_service/internal/infrastructure/logger"
	"auth_service/internal/infrastructure/repository"
	"auth_service/internal/usecase/auth"
	"context"
	"github.com/google/uuid"
)

type Auth interface {
	Register(ctx context.Context, user domain.User) (uuid.UUID, error)
	Login(ctx context.Context, username, password string) (string, string, error)
	ParseRefreshToken(ctx context.Context, tokenR string) (string, error)
	ParseAccessToken(ctx context.Context, token string) (uuid.UUID, error)
	GenerateAccessToken(userId string) (string, error)
	Refresh(ctx context.Context, refreshToken string) (string, string, error)
	Logout(ctx context.Context, accessToken string) error
	Me(ctx context.Context, accessToken string) (*domain.User, error)
}

type Service struct {
	Auth
}

func NewService(rep *repository.Repository, log *logger.SlogLogger, tokens auth.TokenManager) *Service {
	return &Service{
		Auth: auth.NewServiceAuth(rep, log, tokens),
	}
}
