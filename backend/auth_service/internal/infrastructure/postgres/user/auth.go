package user

import (
	"auth_service/internal/domain"
	"auth_service/internal/infrastructure/logger"
	"auth_service/internal/infrastructure/postgres"
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type Auth struct {
	db  *sqlx.DB
	log *logger.SlogLogger
}

func NewAuthRepository(db *sqlx.DB, log *logger.SlogLogger) *Auth {
	return &Auth{
		db:  db,
		log: log,
	}
}

func (r *Auth) CreateUser(ctx context.Context, user domain.User) (uuid.UUID, error) {
	var id uuid.UUID

	query := fmt.Sprintf(`
		INSERT INTO %s (
			username,
			email,
			first_name,
			last_name,
			password_hash
		)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`, postgres.Users)

	err := r.db.QueryRowContext(
		ctx,
		query,
		user.Username,
		user.Email,
		user.FirstName,
		user.LastName,
		user.Password,
	).Scan(&id)

	if err != nil {
		r.log.Error(ctx, "creating user error", err.Error())
		return uuid.UUID{}, err
	}

	r.log.Info(ctx, "creating user successfully")
	return id, nil
}

func (r *Auth) GetUser(ctx context.Context, username, password string) (domain.User, error) {
	var user domain.User
	r.log.Info(ctx, username, password)

	query := fmt.Sprintf("SELECT id FROM %s WHERE username=$1 AND password_hash=$2", postgres.Users)

	err := r.db.Get(&user, query, username, password)
	return user, err
}
func (r *Auth) GetUserByUsername(ctx context.Context, username string) (domain.User, error) {
	var user domain.User
	query := fmt.Sprintf("SELECT id, username, password_hash FROM %s WHERE username=$1", postgres.Users)
	err := r.db.QueryRowContext(ctx, query, username).Scan(&user.Id, &user.Username, &user.Password)
	if err != nil {
		r.log.Error(ctx, "postgres error", err.Error())
	}
	return user, err
}
func (r *Auth) SaveRefreshToken(ctx context.Context, userID uuid.UUID, token string) error {
	query := fmt.Sprintf(`
		UPDATE %s
		SET refresh_token = $1
		WHERE id = $2
	`, postgres.Users)

	_, err := r.db.ExecContext(ctx, query, token, userID)
	if err != nil {
		r.log.Error(ctx, "save refresh token error", err.Error())
		return err
	}

	return nil
}

func (r *Auth) GetRefreshToken(ctx context.Context, userID uuid.UUID) (string, error) {
	var token string

	query := fmt.Sprintf(`
		SELECT refresh_token
		FROM %s
		WHERE id = $1
	`, postgres.Users)

	err := r.db.QueryRowContext(ctx, query, userID).Scan(&token)
	if err != nil {
		r.log.Error(ctx, "get refresh token error", err.Error())
		return "", err
	}

	return token, nil
}

func (r *Auth) DeleteRefreshToken(ctx context.Context, userID uuid.UUID) error {
	query := fmt.Sprintf(`
		UPDATE %s
		SET refresh_token = NULL
		WHERE id = $1
	`, postgres.Users)

	_, err := r.db.ExecContext(ctx, query, userID)
	if err != nil {
		r.log.Error(ctx, "delete refresh token error", err.Error())
		return err
	}

	return nil
}

func (r *Auth) GetUserByID(ctx context.Context, userID uuid.UUID) (domain.User, error) {
	var user domain.User

	query := fmt.Sprintf(`
		SELECT id, username, email, last_name, first_name 
		FROM %s
		WHERE id = $1
	`, postgres.Users)

	err := r.db.QueryRowContext(ctx, query, userID).
		Scan(&user.Id, &user.Username, &user.Email, &user.LastName, &user.FirstName)

	if err != nil {
		r.log.Error(ctx, "get user by id error", err.Error())
		return domain.User{}, err
	}

	return user, nil
}
