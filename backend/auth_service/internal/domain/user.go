package domain

import (
	"github.com/google/uuid"
	"time"
)

// User represents an application user.
type User struct {
	Id           uuid.UUID `json:"id" db:"id"`
	Username     string    `json:"username" db:"username"`
	Email        string    `json:"email" db:"email"`
	FirstName    string    `json:"first_name" db:"first_name"`
	LastName     string    `json:"last_name" db:"last_name"`
	Password     string    `json:"-" db:"password_hash"` // hide in JSON
	RefreshToken *string   `json:"-" db:"refresh_token"` // nullable
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
}
