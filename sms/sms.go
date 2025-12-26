// Package sms provides an abstraction layer for sending SMS messages
// with support for multiple providers (Twilio, AWS SNS, Nexmo, etc.)
package sms

import (
	"context"
	"fmt"
	"time"
)

// Provider represents an SMS service provider
type Provider string

const (
	// ProviderTwilio uses Twilio API
	ProviderTwilio Provider = "twilio"
	// ProviderAWSSNS uses AWS Simple Notification Service
	ProviderAWSSNS Provider = "aws_sns"
	// ProviderNexmo uses Vonage/Nexmo API
	ProviderNexmo Provider = "nexmo"
	// ProviderMessageBird uses MessageBird API
	ProviderMessageBird Provider = "messagebird"
)

// Message represents an SMS message
type Message struct {
	From    string   // Sender phone number or ID
	To      []string // Recipient phone numbers
	Body    string   // Message text content
	Unicode bool     // Whether message contains Unicode characters
}

// SendResult contains the result of an SMS send operation
type SendResult struct {
	MessageID string    // Unique message identifier
	SentAt    time.Time // Time when SMS was sent
	Provider  Provider  // Provider used to send the SMS
	Status    Status    // Current message status
	Cost      float64   // Cost of sending (if available)
	Segments  int       // Number of SMS segments used
}

// Status represents the status of an SMS message
type Status string

const (
	// StatusQueued indicates message is queued for sending
	StatusQueued Status = "queued"
	// StatusSending indicates message is being sent
	StatusSending Status = "sending"
	// StatusSent indicates message was sent successfully
	StatusSent Status = "sent"
	// StatusDelivered indicates message was delivered to recipient
	StatusDelivered Status = "delivered"
	// StatusFailed indicates message failed to send
	StatusFailed Status = "failed"
	// StatusUndelivered indicates message could not be delivered
	StatusUndelivered Status = "undelivered"
)

// Client is the interface that all SMS providers must implement
type Client interface {
	// Send sends an SMS message
	Send(ctx context.Context, msg *Message) (*SendResult, error)

	// SendBulk sends multiple SMS messages in batch
	SendBulk(ctx context.Context, messages []*Message) ([]*SendResult, error)

	// GetStatus retrieves the status of a sent message
	GetStatus(ctx context.Context, messageID string) (*SendResult, error)

	// ValidatePhoneNumber checks if a phone number is valid
	ValidatePhoneNumber(phoneNumber string) error

	// Close closes the SMS client and releases resources
	Close() error
}

// Config contains configuration for SMS client
type Config struct {
	Provider Provider          // SMS provider to use
	From     string            // Default sender phone number or ID
	Options  map[string]string // Provider-specific options
}

// TwilioConfig contains Twilio-specific configuration
type TwilioConfig struct {
	AccountSID string // Twilio account SID
	AuthToken  string // Twilio auth token
	FromNumber string // Sender phone number
}

// AWSSNSConfig contains AWS SNS-specific configuration
type AWSSNSConfig struct {
	Region          string // AWS region
	AccessKeyID     string // AWS access key ID
	SecretAccessKey string // AWS secret access key
}

// NexmoConfig contains Nexmo/Vonage-specific configuration
type NexmoConfig struct {
	APIKey    string // Nexmo API key
	APISecret string // Nexmo API secret
	FromName  string // Sender name or number
}

// MessageBirdConfig contains MessageBird-specific configuration
type MessageBirdConfig struct {
	APIKey     string // MessageBird API key
	Originator string // Sender name or number
}

// NewClient creates a new SMS client based on the provider
func NewClient(config Config) (Client, error) {
	switch config.Provider {
	case ProviderTwilio:
		return newTwilioClient(config)
	case ProviderAWSSNS:
		return newAWSSNSClient(config)
	case ProviderNexmo:
		return newNexmoClient(config)
	case ProviderMessageBird:
		return newMessageBirdClient(config)
	default:
		return nil, fmt.Errorf("unsupported SMS provider: %s", config.Provider)
	}
}

// Validate checks if a message is valid
func (m *Message) Validate() error {
	if m.From == "" {
		return fmt.Errorf("from number is required")
	}
	if len(m.To) == 0 {
		return fmt.Errorf("at least one recipient is required")
	}
	if m.Body == "" {
		return fmt.Errorf("message body is required")
	}
	if len(m.Body) > 1600 {
		return fmt.Errorf("message body exceeds maximum length of 1600 characters")
	}
	return nil
}

// CalculateSegments calculates the number of SMS segments needed
func (m *Message) CalculateSegments() int {
	bodyLength := len(m.Body)

	// Standard SMS: 160 characters per segment
	// Unicode SMS: 70 characters per segment
	// Multi-part: 153/67 characters per segment (7/3 chars reserved for UDH)

	var singleSegmentLimit, multiSegmentLimit int
	if m.Unicode {
		singleSegmentLimit = 70
		multiSegmentLimit = 67
	} else {
		singleSegmentLimit = 160
		multiSegmentLimit = 153
	}

	if bodyLength <= singleSegmentLimit {
		return 1
	}

	segments := bodyLength / multiSegmentLimit
	if bodyLength%multiSegmentLimit > 0 {
		segments++
	}

	return segments
}

// EstimateCost estimates the cost of sending the message
// Note: This is a rough estimate. Actual cost depends on the provider and destination
func (m *Message) EstimateCost(costPerSegment float64) float64 {
	return float64(m.CalculateSegments()*len(m.To)) * costPerSegment
}
