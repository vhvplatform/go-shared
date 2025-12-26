// Package email provides an abstraction layer for sending emails
// with support for multiple providers (SMTP, SendGrid, AWS SES, etc.)
package email

import (
	"context"
	"fmt"
	"time"
)

// Provider represents an email service provider
type Provider string

const (
	// ProviderSMTP uses standard SMTP protocol
	ProviderSMTP Provider = "smtp"
	// ProviderSendGrid uses SendGrid API
	ProviderSendGrid Provider = "sendgrid"
	// ProviderAWSSES uses AWS Simple Email Service
	ProviderAWSSES Provider = "aws_ses"
	// ProviderMailgun uses Mailgun API
	ProviderMailgun Provider = "mailgun"
)

// Message represents an email message
type Message struct {
	From        string            // Sender email address
	To          []string          // Recipient email addresses
	CC          []string          // Carbon copy recipients
	BCC         []string          // Blind carbon copy recipients
	Subject     string            // Email subject
	Body        string            // Email body (plain text or HTML)
	HTML        bool              // Whether body is HTML
	Attachments []Attachment      // File attachments
	Headers     map[string]string // Custom email headers
	ReplyTo     string            // Reply-to address
	Priority    Priority          // Email priority
}

// Attachment represents an email attachment
type Attachment struct {
	Filename    string // Filename to display
	Content     []byte // File content
	ContentType string // MIME type
}

// Priority represents email priority level
type Priority int

const (
	// PriorityLow indicates low priority email
	PriorityLow Priority = iota
	// PriorityNormal indicates normal priority email
	PriorityNormal
	// PriorityHigh indicates high priority email
	PriorityHigh
)

// SendResult contains the result of an email send operation
type SendResult struct {
	MessageID string    // Unique message identifier
	SentAt    time.Time // Time when email was sent
	Provider  Provider  // Provider used to send the email
}

// Client is the interface that all email providers must implement
type Client interface {
	// Send sends an email message
	Send(ctx context.Context, msg *Message) (*SendResult, error)

	// SendBulk sends multiple emails in batch
	SendBulk(ctx context.Context, messages []*Message) ([]*SendResult, error)

	// ValidateAddress checks if an email address is valid
	ValidateAddress(email string) error

	// Close closes the email client and releases resources
	Close() error
}

// Config contains configuration for email client
type Config struct {
	Provider Provider          // Email provider to use
	From     string            // Default sender address
	Options  map[string]string // Provider-specific options
}

// SMTPConfig contains SMTP-specific configuration
type SMTPConfig struct {
	Host     string // SMTP server hostname
	Port     int    // SMTP server port
	Username string // SMTP username
	Password string // SMTP password
	UseTLS   bool   // Whether to use TLS
}

// SendGridConfig contains SendGrid-specific configuration
type SendGridConfig struct {
	APIKey string // SendGrid API key
}

// AWSSESConfig contains AWS SES-specific configuration
type AWSSESConfig struct {
	Region          string // AWS region
	AccessKeyID     string // AWS access key ID
	SecretAccessKey string // AWS secret access key
	ConfigSetName   string // Optional configuration set name
}

// MailgunConfig contains Mailgun-specific configuration
type MailgunConfig struct {
	Domain    string // Mailgun domain
	APIKey    string // Mailgun API key
	PublicKey string // Mailgun public key
	BaseURL   string // Mailgun API base URL (optional)
}

// NewClient creates a new email client based on the provider
func NewClient(config Config) (Client, error) {
	switch config.Provider {
	case ProviderSMTP:
		return newSMTPClient(config)
	case ProviderSendGrid:
		return newSendGridClient(config)
	case ProviderAWSSES:
		return newAWSSESClient(config)
	case ProviderMailgun:
		return newMailgunClient(config)
	default:
		return nil, fmt.Errorf("unsupported email provider: %s", config.Provider)
	}
}

// Validate checks if a message is valid
func (m *Message) Validate() error {
	if m.From == "" {
		return fmt.Errorf("from address is required")
	}
	if len(m.To) == 0 {
		return fmt.Errorf("at least one recipient is required")
	}
	if m.Subject == "" {
		return fmt.Errorf("subject is required")
	}
	if m.Body == "" {
		return fmt.Errorf("body is required")
	}
	return nil
}

// SetHTML sets the message body as HTML
func (m *Message) SetHTML(html string) {
	m.Body = html
	m.HTML = true
}

// SetText sets the message body as plain text
func (m *Message) SetText(text string) {
	m.Body = text
	m.HTML = false
}

// AddAttachment adds a file attachment to the message
func (m *Message) AddAttachment(filename string, content []byte, contentType string) {
	if m.Attachments == nil {
		m.Attachments = make([]Attachment, 0)
	}
	m.Attachments = append(m.Attachments, Attachment{
		Filename:    filename,
		Content:     content,
		ContentType: contentType,
	})
}

// AddHeader adds a custom header to the message
func (m *Message) AddHeader(key, value string) {
	if m.Headers == nil {
		m.Headers = make(map[string]string)
	}
	m.Headers[key] = value
}
