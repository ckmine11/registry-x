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
    if !s.IsEnabled() {
        // Disabled
        fmt.Println("[Email] SMTP Host or Password not configured. Skipping email (Simulated).")
        return nil
    }

    auth := smtp.PlainAuth("", s.Config.SMTPUser, s.Config.SMTPPass, s.Config.SMTPHost)
    
    // Construct message
    subject := "Subject: Password Reset Request\n"
    mime := "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n"
    
    // Use configured frontend URL
    link := fmt.Sprintf("%s/reset-password?token=%s", s.Config.FrontendURL, token)
    
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

func (s *Service) SendInvitationEmail(to, username, tempPass string) error {
    if !s.IsEnabled() {
        fmt.Println("[Email] SMTP Not Configured. Invitation Email skipped (Simulated).")
        return nil
    }

    auth := smtp.PlainAuth("", s.Config.SMTPUser, s.Config.SMTPPass, s.Config.SMTPHost)

    subject := fmt.Sprintf("Subject: Welcome to RegistryX - Your Account is Ready\n")
    mime := "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n"

    body := fmt.Sprintf(`
    <html>
    <body style="font-family: Arial, sans-serif;">
        <div style="max-width: 600px; margin: 0 auto; padding: 20px; border: 1px solid #eee; border-radius: 10px;">
            <h2 style="color: #4f46e5;">Welcome to RegistryX!</h2>
            <p>Hello,</p>
            <p>An administrator has created a RegistryX account for you. Here are your temporary login credentials:</p>
            
            <div style="background: #f9fafb; padding: 15px; border-radius: 5px; margin: 20px 0;">
                <p><strong>Username:</strong> %s</p>
                <p><strong>Temporary Password:</strong> %s</p>
            </div>

            <p>Please log in and change your password as soon as possible:</p>
            <p><a href="%s/login" style="background: #4f46e5; color: white; padding: 10px 20px; text-decoration: none; border-radius: 5px;">Login to RegistryX</a></p>

            <p style="color: #6b7280; font-size: 12px; margin-top: 30px;">
                If you were not expecting this invitation, please ignore this email.
            </p>
        </div>
    </body>
    </html>
    `, username, tempPass, s.Config.FrontendURL)

    msg := []byte(subject + mime + body)
    addr := fmt.Sprintf("%s:%s", s.Config.SMTPHost, s.Config.SMTPPort)
    
    err := smtp.SendMail(addr, auth, s.Config.SMTPFrom, []string{to}, msg)
    if err != nil {
        return fmt.Errorf("failed to send invitation email: %v", err)
    }

    fmt.Printf("[Email] Sent invitation to %s\n", to)
    return nil
}
