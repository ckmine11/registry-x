package auth

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// var jwtKey removed - using s.JWTSecret

type Claims struct {
	UserID uuid.UUID `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

// RegisterUser creates a new user.
func (s *Service) RegisterUser(ctx context.Context, username, email, password string) (*User, string, error) {
    if len(password) < 8 {
        return nil, "", errors.New("password must be at least 8 characters")
    }
	// Check if user exists
	var exists bool
	err := s.DB.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM users WHERE username=$1 OR email=$2)", username, email).Scan(&exists)
	if err != nil {
		return nil, "", err
	}
	if exists {
		return nil, "", errors.New("username or email already exists")
	}

	// Hash password
	hash, err := HashPassword(password)
	if err != nil {
		return nil, "", err
	}

    // Generate Recovery Key (RX-XXXX-XXXX)
    keyRaw := uuid.New().String()
    recoveryKey := fmt.Sprintf("RX-%s-%s", keyRaw[0:4], keyRaw[4:8])
    recoveryHash, err := HashPassword(recoveryKey)
    if err != nil {
        return nil, "", err
    }

	// Start Transaction
    tx, err := s.DB.BeginTx(ctx, nil)
    if err != nil {
        return nil, "", err
    }
    defer tx.Rollback()

	// Insert User
	id := uuid.New()
	now := time.Now()
	_, err = tx.ExecContext(ctx, `
		INSERT INTO users (id, username, email, password_hash, recovery_key_hash, role, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, 'user', $6, $6)`,
		id, username, email, hash, recoveryHash, now)
	if err != nil {
		return nil, "", err
	}

	// Create Personal Namespace
	_, err = tx.ExecContext(ctx, `
		INSERT INTO namespaces (name, type, owner_id)
		VALUES ($1, 'user', $2)`, username, id)
	if err != nil {
		return nil, "", fmt.Errorf("failed to create namespace: %v", err)
	}

    if err := tx.Commit(); err != nil {
        return nil, "", err
    }

	return &User{
		ID:        id,
		Username:  username,
		Email:     email,
		Role:      "user",
		CreatedAt: now,
		UpdatedAt: now,
	}, recoveryKey, nil
}

// ResetPasswordWithKey resets password using request recovery key
func (s *Service) ResetPasswordWithKey(ctx context.Context, email, key, newPassword string) error {
    var userID uuid.UUID
    var storedHash sql.NullString // Handle nulls if existing users don't have keys
    
    // Get user and hash
    err := s.DB.QueryRowContext(ctx, "SELECT id, recovery_key_hash FROM users WHERE email=$1", email).Scan(&userID, &storedHash)
    if err != nil {
         return errors.New("invalid email or key")
    }
    
    if !storedHash.Valid || storedHash.String == "" {
        return errors.New("recovery not set up for this user")
    }
    
    // Verify Key
    if !CheckPasswordHash(key, storedHash.String) {
         return errors.New("invalid recovery key")
    }
    
    // Update Password
    return s.UpdatePassword(ctx, userID, newPassword)
}

// LoginUser authenticates a user and returns a JWT token.
func (s *Service) LoginUser(ctx context.Context, username, password string) (*User, string, error) {
	var user User
	err := s.DB.QueryRowContext(ctx, `
		SELECT id, username, email, password_hash, role, created_at, updated_at 
		FROM users WHERE username=$1`, username).Scan(
		&user.ID, &user.Username, &user.Email, &user.PasswordHash, &user.Role, &user.CreatedAt, &user.UpdatedAt)
	
	if err == sql.ErrNoRows {
		fmt.Printf("[Auth] Login failed: user '%s' not found\n", username)
		return nil, "", errors.New("invalid credentials")
	} else if err != nil {
		fmt.Printf("[Auth] Login DB error for '%s': %v\n", username, err)
		return nil, "", err
	}

	fmt.Printf("[Auth] Login attempt for '%s', hash length: %d\n", username, len(user.PasswordHash))
	if !CheckPasswordHash(password, user.PasswordHash) {
		fmt.Printf("[Auth] Login failed: password mismatch for '%s'\n", username)
		return nil, "", errors.New("invalid credentials")
	}
	fmt.Printf("[Auth] Login successful for '%s'\n", username)
	
	// Audit Log
	if s.Audit != nil {
		_ = s.Audit.Log(ctx, user.ID, "LOGIN", nil, map[string]interface{}{"method": "password"})
	}

	// Generate Token with Session ID (JTI)
	sessionID := uuid.New().String()
	expirationTime := time.Now().Add(24 * time.Hour)
	
	claims := &Claims{
		UserID: user.ID,
		Username: user.Username,
		Role: user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject: user.ID.String(),
			ID:      sessionID, // Set JTI for session tracking
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)
	tokenString, err := token.SignedString([]byte(s.JWTSecret))
	if err != nil {
		return nil, "", err
	}

	// Store Session in Redis
	if s.Redis != nil {
		sessionKey := "session:" + sessionID
		sessionData := map[string]interface{}{
			"user_id":  user.ID.String(),
			"username": user.Username,
			"role":     user.Role,
			"login_at": time.Now().Format(time.RFC3339),
		}
		
		err := s.Redis.HMSet(ctx, sessionKey, sessionData).Err()
		if err != nil {
			fmt.Printf("[Auth] Failed to store session in Redis: %v\n", err)
			return nil, "", fmt.Errorf("session initialization failed")
		}
		s.Redis.Expire(ctx, sessionKey, 24*time.Hour)
		fmt.Printf("[Auth] Created session %s for user %s\n", sessionID, user.Username)
	}

	return &user, tokenString, nil
}

// Logout invalidates a user session.
func (s *Service) Logout(ctx context.Context, sessionID string) error {
	if s.Redis == nil {
		return nil // Redis not enabled - nothing to do
	}

	fmt.Printf("[Auth] Logging out session %s\n", sessionID)
	
	// Delete from Redis
	err := s.Redis.Del(ctx, "session:"+sessionID).Err()
	if err != nil {
		return fmt.Errorf("failed to clear session: %w", err)
	}

	// Optional: Get user ID from session before deleting for audit log
	// But since we just deleted it, we'll keep it simple for now.

	return nil
}

type SessionInfo struct {
	ID        string            `json:"id"`
	UserID    string            `json:"user_id"`
	Username  string            `json:"username"`
	Role      string            `json:"role"`
	LoginAt   string            `json:"login_at"`
}

func (s *Service) ListSessions(ctx context.Context) ([]SessionInfo, error) {
	if s.Redis == nil {
		return nil, errors.New("redis session store not available")
	}

	keys, err := s.Redis.Keys(ctx, "session:*").Result()
	if err != nil {
		return nil, err
	}

	var sessions []SessionInfo
	for _, key := range keys {
		data, err := s.Redis.HGetAll(ctx, key).Result()
		if err != nil {
			continue
		}
		
		sid := strings.TrimPrefix(key, "session:")
		sessions = append(sessions, SessionInfo{
			ID:       sid,
			UserID:   data["user_id"],
			Username: data["username"],
			Role:     data["role"],
			LoginAt:  data["login_at"],
		})
	}

	return sessions, nil
}

func (s *Service) RevokeSession(ctx context.Context, sessionID string) error {
	if s.Redis == nil {
		return nil
	}
	return s.Redis.Del(ctx, "session:"+sessionID).Err()
}

// ValidateCredentials checks username and password and returns the User if valid.
func (s *Service) ValidateCredentials(ctx context.Context, username, password string) (*User, error) {
	var user User
	err := s.DB.QueryRowContext(ctx, `
		SELECT id, username, email, password_hash, role 
		FROM users WHERE username=$1`, username).Scan(
		&user.ID, &user.Username, &user.Email, &user.PasswordHash, &user.Role)
	
	if err == sql.ErrNoRows {
		return nil, errors.New("invalid credentials")
	} else if err != nil {
		return nil, err
	}

	if !CheckPasswordHash(password, user.PasswordHash) {
		return nil, errors.New("invalid credentials")
	}
	return &user, nil
}

// UpdatePassword updates the user's password.
func (s *Service) UpdatePassword(ctx context.Context, userID uuid.UUID, newPassword string) error {
	fmt.Printf("[Auth] UpdatePassword called for user %s, new password length: %d\n", userID, len(newPassword))
	hash, err := HashPassword(newPassword)
	if err != nil {
		fmt.Printf("[Auth] UpdatePassword hash generation failed: %v\n", err)
		return err
	}

	fmt.Printf("[Auth] UpdatePassword generated hash length: %d\n", len(hash))
	result, err := s.DB.ExecContext(ctx, "UPDATE users SET password_hash=$1, updated_at=$2 WHERE id=$3", hash, time.Now(), userID)
	if err != nil {
		fmt.Printf("[Auth] UpdatePassword DB error: %v\n", err)
		return err
	}
	rowsAffected, _ := result.RowsAffected()
	fmt.Printf("[Auth] UpdatePassword successful, rows affected: %d\n", rowsAffected)
	return nil
}

// RequestPasswordReset generates a reset token for the given email
func (s *Service) RequestPasswordReset(ctx context.Context, email string) (string, error) {
	var userID uuid.UUID
	err := s.DB.QueryRowContext(ctx, "SELECT id FROM users WHERE email=$1", email).Scan(&userID)
	if err == sql.ErrNoRows {
		// Return nil error to prevent email enumeration, but return empty token
		return "", nil
	} else if err != nil {
		return "", err
	}

	// Generate Token (simple UUID for now, in prod used cryptographically secure random string)
	token := uuid.New().String()
	expiresAt := time.Now().Add(1 * time.Hour) // 1 Hour expiry

	_, err = s.DB.ExecContext(ctx, `
		INSERT INTO password_resets (user_id, token, expires_at)
		VALUES ($1, $2, $3)`,
		userID, token, expiresAt)
	
	if err != nil {
		return "", err
	}

	// Send Email using Email Service
	if s.Email != nil {
		if err := s.Email.SendResetEmail(email, token); err != nil {
			fmt.Printf("Failed to send email: %v\n", err)
			// Return error so user knows it failed
			return "", err
		}
	}

	// Return token for debug/logging (Frontend won't see it if Email works?)
	// Actually frontend handlers uses this token for debug.
	return token, nil
}

// ResetPassword resets the password using the token
func (s *Service) ResetPassword(ctx context.Context, token, newPassword string) error {
	var userID uuid.UUID
	var expiresAt time.Time

	// Find valid token
	err := s.DB.QueryRowContext(ctx, `
		SELECT user_id, expires_at FROM password_resets 
		WHERE token=$1 AND expires_at > NOW()`, token).Scan(&userID, &expiresAt)
	
	if err == sql.ErrNoRows {
		return errors.New("invalid or expired token")
	} else if err != nil {
		return err
	}

	// Update Password
	if err := s.UpdatePassword(ctx, userID, newPassword); err != nil {
		return err
	}

	// Cleanup used token (or all tokens for this user)
	_, _ = s.DB.ExecContext(ctx, "DELETE FROM password_resets WHERE user_id=$1", userID)

	return nil
}
