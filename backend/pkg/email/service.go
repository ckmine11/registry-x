package email

import (
    "fmt"
    "net/smtp"
    "github.com/registryx/registryx/backend/pkg/config"
)

type Service struct {
    Config *config.Config
}

func NewService(cfg *config.Config) *Service {
    return &Service{Config: cfg}
}

func (s *Service) IsEnabled() bool {
    return s.Config.SMTPHost != "" && s.Config.SMTPPass != ""
}

func (s *Service) SendResetEmail(to, token string) error {
    if s.Config.SMTPHost == "" || s.Config.SMTPPass == "" {
        // Disabled
        fmt.Println("[Email] SMTP Host or Password not configured. Skipping email (Simulated).")
        return nil
    }

    auth := smtp.PlainAuth("", s.Config.SMTPUser, s.Config.SMTPPass, s.Config.SMTPHost)
    
    // Construct message
    subject := "Subject: Password Reset Request\n"
    mime := "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n"
    
    // Using localhost link for now
    // Ideally this should come from config.FrontendURL
    link := fmt.Sprintf("http://localhost:5173/reset-password?token=%s", token)
    
    body := fmt.Sprintf(`
    <html>
    <body>
        <h2>Password Reset</h2>
        <p>You requested a password reset for RegistryX.</p>
        <p>Click the link below to reset your password:</p>
        <p><a href="%s">Reset Password</a></p>
        <p>If you didn't request this, please ignore this email.</p>
    </body>
    </html>
    `, link)
    
    msg := []byte(subject + mime + body)
    
    addr := fmt.Sprintf("%s:%s", s.Config.SMTPHost, s.Config.SMTPPort)
    err := smtp.SendMail(addr, auth, s.Config.SMTPFrom, []string{to}, msg)
    if err != nil {
        return fmt.Errorf("failed to send email: %v", err)
    }
    
    fmt.Printf("[Email] Sent reset link to %s\n", to)
    return nil
}
