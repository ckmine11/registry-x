package webhook

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type Event struct {
	Action     string    `json:"action"`
	Repository string    `json:"repository"`
	Tag        string    `json:"tag"`
	Digest     string    `json:"digest"`
	Timestamp  time.Time `json:"timestamp"`
	User       string    `json:"user"`
}

type Service struct {
	WebhookURL string
}

func NewService(url string) *Service {
	return &Service{WebhookURL: url}
}

func (s *Service) Notify(ctx context.Context, event Event) error {
	if s.WebhookURL == "" {
		return nil
	}

	payload, err := json.Marshal(event)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", s.WebhookURL, bytes.NewBuffer(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	// Fire and forget-ish, but check status
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send webhook: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("webhook endpoint returned status: %d", resp.StatusCode)
	}

	return nil
}
