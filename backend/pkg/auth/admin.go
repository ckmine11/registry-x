package auth

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
)

// ListUsers returns a list of all users/
func (s *Service) ListUsers(ctx context.Context) ([]User, error) {
	rows, err := s.DB.QueryContext(ctx, "SELECT id, username, email, role, created_at, updated_at FROM users ORDER BY created_at DESC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var u User
		if err := rows.Scan(&u.ID, &u.Username, &u.Email, &u.Role, &u.CreatedAt, &u.UpdatedAt); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, nil
}

// UpdateRole updates a user's role.
func (s *Service) UpdateRole(ctx context.Context, userID uuid.UUID, role string) error {
    if role != "admin" && role != "developer" && role != "readonly" {
        return errors.New("invalid role")
    }

	_, err := s.DB.ExecContext(ctx, "UPDATE users SET role=$1, updated_at=$2 WHERE id=$3", role, time.Now(), userID)
	return err
}

// DeleteUser removes a user and their namespace.
func (s *Service) DeleteUser(ctx context.Context, userID uuid.UUID) error {
	log.Printf("[DeleteUser] Attempting to delete user: %s", userID)
	
	// Check if user exists
	var exists bool
	err := s.DB.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM users WHERE id=$1)", userID).Scan(&exists)
	if err != nil {
		log.Printf("[DeleteUser] Error checking user existence: %v", err)
		return err
	}
	if !exists {
		log.Printf("[DeleteUser] User not found: %s", userID)
		return fmt.Errorf("user not found")
	}

	// Delete related records first to avoid foreign key constraints
	// Delete namespace owned by user
	_, err = s.DB.ExecContext(ctx, "DELETE FROM namespaces WHERE owner_id=$1", userID)
	if err != nil {
		log.Printf("[DeleteUser] Error deleting user namespace: %v", err)
		// Continue anyway, namespace might not exist
	}

	// Delete the user
	result, err := s.DB.ExecContext(ctx, "DELETE FROM users WHERE id=$1", userID)
	if err != nil {
		log.Printf("[DeleteUser] Error deleting user: %v", err)
		return err
	}

	rowsAffected, _ := result.RowsAffected()
	log.Printf("[DeleteUser] Successfully deleted user %s (rows affected: %d)", userID, rowsAffected)
	
	return nil
}

func (s *Service) InviteUser(ctx context.Context, email, role string) (*User, string, error) {
    // Check existence
	var exists bool
	err := s.DB.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM users WHERE email=$1)", email).Scan(&exists)
	if err != nil {
		return nil, "", err
	}
	if exists {
		return nil, "", errors.New("email already exists")
	}

    // Generate temp password / key
    tempPass := uuid.New().String()
    hash, _ := HashPassword(tempPass)
    
    // Generate username from email (part before @)
    username := email
    if idx :=  len(email); idx > 0 {
        // simple split
    }
    // Better: just use email as username or random
    // For simplicity, let's auto-generate a username: email_prefix + random
    // But RegistryX uses `username` for namespaces.
    
    // Let's ask caller to provide username or derive it safely.
    // We'll derive it.
    // foo@bar.com -> foo
    parts :=  time.Now().Unix()
    username = fmt.Sprintf("user%d", parts) 

    id := uuid.New()
    now := time.Now()
    
    _, err = s.DB.ExecContext(ctx, `
		INSERT INTO users (id, username, email, password_hash, role, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $6)`,
		id, username, email, hash, role, now)
	
    if err != nil {
        return nil, "", err
    }
    
    // Create Namespace
    _, _ = s.DB.ExecContext(ctx, `INSERT INTO namespaces (name, type, owner_id) VALUES ($1, 'user', $2)`, username, id)

    // Send Invitation Email
    if s.Email != nil {
        go func() {
            err := s.Email.SendInvitationEmail(email, username, tempPass)
            if err != nil {
                log.Printf("[Auth] Failed to send invitation email to %s: %v", email, err)
            }
        }()
    }

    return &User{
        ID: id,
        Username: username,
        Email: email,
        Role: role, 
        CreatedAt: now,
    }, tempPass, nil
}
