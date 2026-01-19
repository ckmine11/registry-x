package audit

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"
	"github.com/google/uuid"
)

type Service struct {
	DB *sql.DB
}

func NewService(db *sql.DB) *Service {
	return &Service{DB: db}
}

type LogEntry struct {
	ID        uuid.UUID       `json:"id"`
	UserID    uuid.UUID       `json:"user_id"`
	Action    string          `json:"action"`
	Details   json.RawMessage `json:"details"`
	CreatedAt time.Time       `json:"created_at"`
}

// Log records an audit event. repoID can be nil.
func (s *Service) Log(ctx context.Context, userID uuid.UUID, action string, repoID *uuid.UUID, details map[string]interface{}) error {
	detailsJSON, _ := json.Marshal(details)
	
	_, err := s.DB.ExecContext(ctx, `
		INSERT INTO audit_logs (user_id, action, repository_id, details, created_at)
		VALUES ($1, $2, $3, $4, CURRENT_TIMESTAMP)`,
		userID, action, repoID, detailsJSON)
	return err
}

// GetUserLogs retrieves logs for a specific user.
func (s *Service) GetUserLogs(ctx context.Context, userID uuid.UUID, limit int) ([]LogEntry, error) {
	rows, err := s.DB.QueryContext(ctx, `
		SELECT id, user_id, action, details, created_at 
		FROM audit_logs 
		WHERE user_id = $1 
		ORDER BY created_at DESC 
		LIMIT $2`, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []LogEntry
	for rows.Next() {
		var l LogEntry
		if err := rows.Scan(&l.ID, &l.UserID, &l.Action, &l.Details, &l.CreatedAt); err != nil {
			continue
		}
		logs = append(logs, l)
	}
	return logs, nil
}
