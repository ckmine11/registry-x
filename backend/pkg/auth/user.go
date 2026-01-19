package auth

import (
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID              uuid.UUID `json:"id"`
	Username        string    `json:"username"`
	Email           string    `json:"email"`
	PasswordHash    string    `json:"-"`
	Role            string    `json:"role"` // 'admin' or 'user'
	RecoveryKeyHash string    `json:"-"` // Stored hash
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type CreateUserRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type AuthResponse struct {
	Token       string `json:"token"`
	User        User   `json:"user"`
	RecoveryKey string `json:"recovery_key,omitempty"` // Only returned on creation
}

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
