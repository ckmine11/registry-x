package webhook

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/smtp"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"mime/multipart"
	"net/textproto"
)

type EventType string

const (
	EventScanCompleted    EventType = "SCAN_COMPLETED"
	EventCriticalVuln     EventType = "CRITICAL_VULN_FOUND"
	EventPolicyViolation  EventType = "POLICY_VIOLATION"
)

type Event struct {
	Type        EventType       `json:"type"`
	Action      string          `json:"action"`
	Repository  string          `json:"repository"`
	Tag         string          `json:"tag"`
	Digest      string          `json:"digest"`
	Timestamp   time.Time       `json:\"timestamp\"`
	User        string          `json:"user"`
	Data        json.RawMessage `json:"data,omitempty"`
	PDFReport   []byte          `json:"-"` // PDF report data (not serialized to JSON)
	PDFFilename string          `json:"-"` // PDF filename
}

type Webhook struct {
	ID        uuid.UUID   `json:"id"`
	URL       string      `json:"url"`
	Type      string      `json:"type"`   // slack, discord, generic, email
	Events    []string    `json:"events"` // Array of EventTypes
	Enabled   bool        `json:"enabled"`
	CreatedAt time.Time   `json:"created_at"`
}

type Service struct {
	DB *sql.DB
}

func NewService(db *sql.DB) *Service {
	return &Service{DB: db}
}

// CRUD Operations

func (s *Service) List(ctx context.Context) ([]Webhook, error) {
	rows, err := s.DB.QueryContext(ctx, "SELECT id, url, type, events, enabled, created_at FROM webhooks ORDER BY created_at DESC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var hooks []Webhook
	for rows.Next() {
		var w Webhook
		var events []string
		if err := rows.Scan(&w.ID, &w.URL, &w.Type, pq.Array(&events), &w.Enabled, &w.CreatedAt); err != nil {
			return nil, err
		}
		w.Events = events
		hooks = append(hooks, w)
	}
	return hooks, nil
}

func (s *Service) Create(ctx context.Context, url string, hookType string, events []string) (*Webhook, error) {
	w := Webhook{
		ID:        uuid.New(),
		URL:       url,
		Type:      hookType,
		Events:    events,
		Enabled:   true,
		CreatedAt: time.Now(),
	}

	_, err := s.DB.ExecContext(ctx, "INSERT INTO webhooks (id, url, type, events, enabled, created_at) VALUES ($1, $2, $3, $4, $5, $6)",
		w.ID, w.URL, w.Type, pq.Array(w.Events), w.Enabled, w.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &w, nil
}

func (s *Service) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := s.DB.ExecContext(ctx, "DELETE FROM webhooks WHERE id = $1", id)
	return err
}

func (s *Service) Test(ctx context.Context, id uuid.UUID) error {
	// Fetch specific webhook
	var w Webhook
	var events []string
	err := s.DB.QueryRowContext(ctx, "SELECT id, url, type, events FROM webhooks WHERE id = $1", id).Scan(&w.ID, &w.URL, &w.Type, pq.Array(&events))
	if err != nil {
		return err
	}

	testEvent := Event{
		Type:       EventScanCompleted,
		Action:     "TEST_NOTIFICATION",
		Repository: "registryx/test-repo",
		Tag:        "v1.0.0",
		Timestamp:  time.Now(),
		User:       "system",
	}
	
	return s.send(ctx, w, testEvent)
}

// Notification Logic

func (s *Service) Notify(ctx context.Context, event Event) error {
	// 1. Get enabled webhooks that subscribe to this event
	// Using Postgres array "cleanup" logic in Go for simplicity, or complex query.
	// For MVP, fetch enabled hooks and filter in memory.
	
	rows, err := s.DB.QueryContext(ctx, "SELECT id, url, type, events FROM webhooks WHERE enabled = true")
	if err != nil {
		return err
	}
	defer rows.Close()

	count := 0
	for rows.Next() {
		var w Webhook
		var events []string
		if err := rows.Scan(&w.ID, &w.URL, &w.Type, pq.Array(&events)); err != nil {
			continue
		}
		w.Events = events

		// Check if webhook subscribes to this event
		shouldSend := false
		for _, e := range w.Events {
			if e == string(event.Type) || e == "*" {
				shouldSend = true
				break
			}
		}

		if shouldSend {
			go func(hook Webhook) {
				if err := s.send(context.Background(), hook, event); err != nil {
					log.Printf("Failed to send webhook %s: %v", hook.ID, err)
				}
			}(w)
			count++
		}
	}
	
	return nil
}

func (s *Service) send(ctx context.Context, w Webhook, event Event) error {
	// Handle email notifications differently
	if w.Type == "email" {
		return s.sendEmail(ctx, w, event)
	}

	// HTTP-based webhooks (Slack, Discord, Generic)
	// If there is a PDF report, use multipart/form-data for Discord and Generic
	if len(event.PDFReport) > 0 && (w.Type == "discord" || w.Type == "generic") {
		return s.sendMultipart(ctx, w, event)
	}

	payloadBytes, err := s.formatPayload(w, event)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", w.URL, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("status %d", resp.StatusCode)
	}
	return nil
}

func (s *Service) sendMultipart(ctx context.Context, w Webhook, event Event) error {
	var b bytes.Buffer
	writer := multipart.NewWriter(&b)

	// Add payload
	payloadBytes, err := s.formatPayload(w, event)
	if err != nil {
		return err
	}

	if w.Type == "discord" {
		_ = writer.WriteField("payload_json", string(payloadBytes))
	} else {
		_ = writer.WriteField("event", string(payloadBytes))
	}

	// Add file
	h := make(textproto.MIMEHeader)
	h.Set("Content-Disposition", fmt.Sprintf(`form-data; name="file"; filename="%s"`, event.PDFFilename))
	h.Set("Content-Type", "application/pdf")
	part, err := writer.CreatePart(h)
	if err != nil {
		return err
	}
	_, _ = part.Write(event.PDFReport)

	writer.Close()

	req, err := http.NewRequestWithContext(ctx, "POST", w.URL, &b)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("status %d", resp.StatusCode)
	}
	return nil
}

func (s *Service) formatPayload(w Webhook, event Event) ([]byte, error) {
	if w.Type == "slack" {
		return s.formatSlack(event)
	} else if w.Type == "discord" {
		return s.formatDiscord(event)
	}
	// Default Generic
	return json.Marshal(event)
}

func (s *Service) formatSlack(event Event) ([]byte, error) {
	color := "#36a64f" // Green
	if event.Type == EventCriticalVuln || event.Type == EventPolicyViolation {
		color = "#ff0000" // Red
	}

	msg := fmt.Sprintf("*RegistryX Event: %s*\nRepository: `%s:%s`\nUser: %s", event.Type, event.Repository, event.Tag, event.User)

	payload := map[string]interface{}{
		"text": msg, // Fallback
		"attachments": []map[string]interface{}{
			{
				"color": color,
				"fields": []map[string]interface{}{
					{"title": "Event", "value": event.Type, "short": true},
					{"title": "Repository", "value": event.Repository, "short": true},
					{"title": "Tag", "value": event.Tag, "short": true},
					{"title": "User", "value": event.User, "short": true},
				},
				"footer": "RegistryX Security",
				"ts":     event.Timestamp.Unix(),
			},
		},
	}
	return json.Marshal(payload)
}

func (s *Service) formatDiscord(event Event) ([]byte, error) {
	// Discord Webhook format
	color := 3583567 // Green
	if event.Type == EventCriticalVuln || event.Type == EventPolicyViolation {
		color = 16711680 // Red
	}

	payload := map[string]interface{}{
		"embeds": []map[string]interface{}{
			{
				"title":       fmt.Sprintf("RegistryX Event: %s", event.Type),
				"description": fmt.Sprintf("Repository: `%s:%s`", event.Repository, event.Tag),
				"color":       color,
				"fields": []map[string]interface{}{
					{"name": "Action", "value": event.Action, "inline": true},
					{"name": "User", "value": event.User, "inline": true},
				},
				"timestamp": event.Timestamp.Format(time.RFC3339),
			},
		},
	}
	return json.Marshal(payload)
}

func (s *Service) formatEmail(event Event) string {
	subject := fmt.Sprintf("[RegistryX] %s - %s:%s", event.Type, event.Repository, event.Tag)
	
	body := fmt.Sprintf(`RegistryX Notification

Event Type: %s
Repository: %s
Tag: %s
Digest: %s
User: %s
Timestamp: %s

Action: %s

---
This is an automated notification from RegistryX.
`,
		event.Type,
		event.Repository,
		event.Tag,
		event.Digest,
		event.User,
		event.Timestamp.Format(time.RFC3339),
		event.Action,
	)
	
	return fmt.Sprintf("Subject: %s\r\n\r\n%s", subject, body)
}

func (s *Service) sendEmail(ctx context.Context, w Webhook, event Event) error {
	// w.URL for email type should be the recipient email address
	to := w.URL
	
	// Get SMTP configuration from environment
	smtpHost := os.Getenv("SMTP_HOST")
	smtpPort := os.Getenv("SMTP_PORT")
	smtpUser := os.Getenv("SMTP_USER")
	smtpPass := os.Getenv("SMTP_PASS")
	smtpFrom := os.Getenv("SMTP_FROM")
	
	if smtpHost == "" || smtpPort == "" {
		return fmt.Errorf("SMTP not configured (SMTP_HOST, SMTP_PORT required)")
	}
	
	if smtpFrom == "" {
		smtpFrom = "noreply@registryx.local"
	}
	
	// Build MIME multipart message with PDF attachment
	var msg bytes.Buffer
	
	// Headers
	msg.WriteString(fmt.Sprintf("From: %s\r\n", smtpFrom))
	msg.WriteString(fmt.Sprintf("To: %s\r\n", to))
	msg.WriteString(fmt.Sprintf("Subject: RegistryX Scan Report - %s:%s\r\n", event.Repository, event.Tag))
	msg.WriteString("MIME-Version: 1.0\r\n")
	
	boundary := "===============boundary123456789=="
	msg.WriteString(fmt.Sprintf("Content-Type: multipart/mixed; boundary=\"%s\"\r\n", boundary))
	msg.WriteString("\r\n")
	
	// Text part
	msg.WriteString(fmt.Sprintf("--%s\r\n", boundary))
	msg.WriteString("Content-Type: text/plain; charset=utf-8\r\n")
	msg.WriteString("\r\n")
	msg.WriteString(s.formatEmail(event))
	msg.WriteString("\r\n\r\n")
	
	// PDF attachment (if present)
	if len(event.PDFReport) > 0 && event.PDFFilename != "" {
		msg.WriteString(fmt.Sprintf("--%s\r\n", boundary))
		msg.WriteString("Content-Type: application/pdf\r\n")
		msg.WriteString("Content-Transfer-Encoding: base64\r\n")
		msg.WriteString(fmt.Sprintf("Content-Disposition: attachment; filename=\"%s\"\r\n", event.PDFFilename))
		msg.WriteString("\r\n")
		
		// Encode PDF to base64
		encoded := base64.StdEncoding.EncodeToString(event.PDFReport)
		// Split into 76-character lines (MIME requirement)
		for i := 0; i < len(encoded); i += 76 {
			end := i + 76
			if end > len(encoded) {
				end = len(encoded)
			}
			msg.WriteString(encoded[i:end])
			msg.WriteString("\r\n")
		}
		msg.WriteString("\r\n")
	}
	
	// End boundary
	msg.WriteString(fmt.Sprintf("--%s--\r\n", boundary))
	
	// Setup authentication if credentials provided
	var auth smtp.Auth
	if smtpUser != "" && smtpPass != "" {
		auth = smtp.PlainAuth("", smtpUser, smtpPass, smtpHost)
	}
	
	addr := fmt.Sprintf("%s:%s", smtpHost, smtpPort)
	err := smtp.SendMail(addr, auth, smtpFrom, []string{to}, msg.Bytes())
	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}
	
	log.Printf("[Webhook] Email sent to %s for event %s (PDF attached: %v)", to, event.Type, len(event.PDFReport) > 0)
	return nil
}

