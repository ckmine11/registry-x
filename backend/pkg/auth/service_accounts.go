package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/registryx/registryx/backend/pkg/audit"
	"github.com/registryx/registryx/backend/pkg/email"
)

type ServiceAccount struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Prefix      string    `json:"-"`
	Status      string    `json:"status"`
	LastUsedAt  *time.Time `json:"lastUsed"`
	CreatedAt   time.Time `json:"created"`
}

type Service struct {
	DB        *sql.DB
	Email     *email.Service
	Audit     *audit.Service
	Redis     *redis.Client
	JWTSecret string
}

func NewService(db *sql.DB, email *email.Service, audit *audit.Service, redisClient *redis.Client, jwtSecret string) *Service {
	return &Service{DB: db, Email: email, Audit: audit, Redis: redisClient, JWTSecret: jwtSecret}
}

// Create generates a new service account and API Key.
// Returns the ServiceAccount object and the raw API Key (only time it's seen).
func (s *Service) Create(ctx context.Context, name, description string) (*ServiceAccount, string, error) {
	// 1. Generate Key
	rawKey, err := generateRandomString(32)
	if err != nil {
		return nil, "", err
	}
	apiKey := "rx_" + rawKey

	// 2. Hash Key
	hash := sha256.New()
	hash.Write([]byte(apiKey))
	keyHash := hex.EncodeToString(hash.Sum(nil))

	// 3. Insert
	id := uuid.New()
	now := time.Now()
	_, err = s.DB.ExecContext(ctx, `
		INSERT INTO service_accounts (id, name, description, api_key_hash, prefix, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, 'active', $6, $6)`,
		id, name, description, keyHash, "rx_"+rawKey[:4], now)
	if err != nil {
		return nil, "", fmt.Errorf("failed to insert service account: %w", err)
	}

	return &ServiceAccount{
		ID:          id,
		Name:        name,
		Description: description,
		Status:      "active",
		CreatedAt:   now,
	}, apiKey, nil
}

// List returns all service accounts.
func (s *Service) List(ctx context.Context) ([]ServiceAccount, error) {
	rows, err := s.DB.QueryContext(ctx, `
		SELECT id, name, description, status, last_used_at, created_at 
		FROM service_accounts ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var accounts []ServiceAccount
	for rows.Next() {
		var acc ServiceAccount
		var lastUsed sql.NullTime
		var desc sql.NullString
		if err := rows.Scan(&acc.ID, &acc.Name, &desc, &acc.Status, &lastUsed, &acc.CreatedAt); err != nil {
			return nil, err
		}
		if lastUsed.Valid {
			acc.LastUsedAt = &lastUsed.Time
		}
		if desc.Valid {
			acc.Description = desc.String
		}
		accounts = append(accounts, acc)
	}
	return accounts, nil
}

// Revoke changes status to revoked.
func (s *Service) Revoke(ctx context.Context, id uuid.UUID) error {
	_, err := s.DB.ExecContext(ctx, "UPDATE service_accounts SET status = 'revoked', updated_at = NOW() WHERE id = $1", id)
	return err
}

func generateRandomString(n int) (string, error) {
	bytes := make([]byte, n)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
