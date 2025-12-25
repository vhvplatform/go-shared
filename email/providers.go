package email

import (
	"fmt"
)

// Placeholder implementations for different providers
// TODO: Implement actual provider logic

func newSMTPClient(config Config) (Client, error) {
	return nil, fmt.Errorf("SMTP client not yet implemented")
}

func newSendGridClient(config Config) (Client, error) {
	return nil, fmt.Errorf("SendGrid client not yet implemented")
}

func newAWSSESClient(config Config) (Client, error) {
	return nil, fmt.Errorf("AWS SES client not yet implemented")
}

func newMailgunClient(config Config) (Client, error) {
	return nil, fmt.Errorf("Mailgun client not yet implemented")
}
