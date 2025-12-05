package email

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/bantuaku/backend/config"
	"github.com/bantuaku/backend/logger"
)

// MailjetService handles sending emails via Mailjet API
type MailjetService struct {
	apiKey    string
	apiSecret string
	baseURL   string
	client    *http.Client
	log       logger.Logger
}

// NewMailjetService creates a new Mailjet email service
func NewMailjetService(cfg *config.Config) *MailjetService {
	return &MailjetService{
		apiKey:    cfg.MailjetAPIKey,
		apiSecret: cfg.MailjetAPISecret,
		baseURL:   cfg.AppBaseURL,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		log: *logger.Default(),
	}
}

// MailjetSendRequest represents the Mailjet API send request
type MailjetSendRequest struct {
	Messages []MailjetMessage `json:"Messages"`
}

// MailjetMessage represents a single email message
type MailjetMessage struct {
	From     MailjetContact `json:"From"`
	To       []MailjetContact `json:"To"`
	Subject  string         `json:"Subject"`
	HTMLPart string         `json:"HTMLPart,omitempty"`
	TextPart string         `json:"TextPart,omitempty"`
}

// MailjetContact represents email contact
type MailjetContact struct {
	Email string `json:"Email"`
	Name  string `json:"Name,omitempty"`
}

// MailjetResponse represents Mailjet API response
type MailjetResponse struct {
	Messages []struct {
		Status string `json:"Status"`
		Errors []struct {
			ErrorIdentifier string `json:"ErrorIdentifier"`
			ErrorCode       string `json:"ErrorCode"`
			StatusCode      int    `json:"StatusCode"`
			ErrorMessage    string `json:"ErrorMessage"`
		} `json:"Errors,omitempty"`
	} `json:"Messages"`
}

// SendVerificationEmail sends an email verification email with OTP
func (s *MailjetService) SendVerificationEmail(toEmail, otpCode string) error {
	subject := "Verifikasi Email Anda - Bantuaku"
	htmlContent := generateVerificationEmailHTML(otpCode, toEmail)
	textContent := generateVerificationEmailText(otpCode, toEmail)

	return s.sendEmail(toEmail, subject, htmlContent, textContent)
}

// SendPasswordResetEmail sends a password reset email with token link
func (s *MailjetService) SendPasswordResetEmail(toEmail, resetToken string) error {
	subject := "Reset Password - Bantuaku"
	resetLink := fmt.Sprintf("%s/reset-password?token=%s", s.baseURL, resetToken)
	htmlContent := generatePasswordResetEmailHTML(resetLink, toEmail)
	textContent := generatePasswordResetEmailText(resetLink, toEmail)

	return s.sendEmail(toEmail, subject, htmlContent, textContent)
}

// sendEmail sends an email via Mailjet API
func (s *MailjetService) sendEmail(toEmail, subject, htmlContent, textContent string) error {
	if s.apiKey == "" || s.apiSecret == "" {
		s.log.Warn("Mailjet credentials not configured, skipping email send", "to", toEmail)
		return nil // Don't fail if email is not configured
	}

	message := MailjetMessage{
		From: MailjetContact{
			Email: "noreply@bantuaku.com",
			Name:  "Bantuaku",
		},
		To: []MailjetContact{
			{
				Email: toEmail,
			},
		},
		Subject:  subject,
		HTMLPart: htmlContent,
		TextPart: textContent,
	}

	request := MailjetSendRequest{
		Messages: []MailjetMessage{message},
	}

	jsonData, err := json.Marshal(request)
	if err != nil {
		return fmt.Errorf("failed to marshal email request: %w", err)
	}

	url := "https://api.mailjet.com/v3.1/send"
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Basic Auth with API key and secret
	auth := base64.StdEncoding.EncodeToString([]byte(s.apiKey + ":" + s.apiSecret))
	req.Header.Set("Authorization", "Basic "+auth)
	req.Header.Set("Content-Type", "application/json")

	// Retry logic for 5xx errors
	var resp *http.Response
	maxRetries := 3
	for i := 0; i < maxRetries; i++ {
		resp, err = s.client.Do(req)
		if err != nil {
			if i < maxRetries-1 {
				time.Sleep(time.Duration(i+1) * time.Second)
				continue
			}
			return fmt.Errorf("failed to send email request: %w", err)
		}
		defer resp.Body.Close()

		// If not 5xx error, break retry loop
		if resp.StatusCode < 500 {
			break
		}

		// Retry on 5xx errors
		if i < maxRetries-1 {
			s.log.Warn("Mailjet API returned 5xx error, retrying", "status", resp.StatusCode, "attempt", i+1)
			time.Sleep(time.Duration(i+1) * time.Second)
			continue
		}
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var mailjetResp MailjetResponse
		if err := json.Unmarshal(body, &mailjetResp); err == nil && len(mailjetResp.Messages) > 0 {
			if len(mailjetResp.Messages[0].Errors) > 0 {
				return fmt.Errorf("mailjet API error: %s (code: %s)", 
					mailjetResp.Messages[0].Errors[0].ErrorMessage,
					mailjetResp.Messages[0].Errors[0].ErrorCode)
			}
		}
		return fmt.Errorf("mailjet API returned status %d: %s", resp.StatusCode, string(body))
	}

	s.log.Info("Email sent successfully", "to", toEmail, "subject", subject)
	return nil
}

