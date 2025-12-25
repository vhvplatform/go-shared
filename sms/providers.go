package sms

import (
	"fmt"
)

// Placeholder implementations for different providers
// TODO: Implement actual provider logic

func newTwilioClient(config Config) (Client, error) {
	return nil, fmt.Errorf("Twilio client not yet implemented")
}

func newAWSSNSClient(config Config) (Client, error) {
	return nil, fmt.Errorf("AWS SNS client not yet implemented")
}

func newNexmoClient(config Config) (Client, error) {
	return nil, fmt.Errorf("Nexmo client not yet implemented")
}

func newMessageBirdClient(config Config) (Client, error) {
	return nil, fmt.Errorf("MessageBird client not yet implemented")
}
