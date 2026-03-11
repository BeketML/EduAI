package repository

import (
	"auth_service/internal/domain"
	"auth_service/internal/infrastructure/logger"
	"auth_service/internal/infrastructure/postgres/user"
	"context"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type Auth interface {
	CreateUser(ctx context.Context, user domain.User) (uuid.UUID, error)
	GetUser(ctx context.Context, username, password string) (domain.User, error)
	GetUserByUsername(ctx context.Context, username string) (domain.User, error)

	SaveRefreshToken(ctx context.Context, id uuid.UUID, refresh string) error
	GetRefreshToken(ctx context.Context, id uuid.UUID) (string, error)
	DeleteRefreshToken(ctx context.Context, id uuid.UUID) error
	GetUserByID(ctx context.Context, id uuid.UUID) (domain.User, error)
}

type Repository struct {
	Auth
}

func NewRepository(db *sqlx.DB, log *logger.SlogLogger) *Repository {
	return &Repository{
		Auth: user.NewAuthRepository(db, log),
	}
}
