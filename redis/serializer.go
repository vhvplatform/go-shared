package redis

import (
	"encoding/json"
	"fmt"
)

// Serializer defines the interface for cache value serialization
type Serializer interface {
	Serialize(v interface{}) ([]byte, error)
	Deserialize(data []byte, v interface{}) error
}

// JSONSerializer implements Serializer using JSON encoding
type JSONSerializer struct{}

// NewJSONSerializer creates a new JSON serializer
func NewJSONSerializer() *JSONSerializer {
	return &JSONSerializer{}
}

// Serialize converts a value to JSON bytes
func (s *JSONSerializer) Serialize(v interface{}) ([]byte, error) {
	// Handle string values directly
	if str, ok := v.(string); ok {
		return []byte(str), nil
	}

	// Handle byte slices directly
	if bytes, ok := v.([]byte); ok {
		return bytes, nil
	}

	// Marshal other types to JSON
	data, err := json.Marshal(v)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize value: %w", err)
	}
	return data, nil
}

// Deserialize converts JSON bytes back to a value
func (s *JSONSerializer) Deserialize(data []byte, v interface{}) error {
	// Handle string pointers directly
	if strPtr, ok := v.(*string); ok {
		*strPtr = string(data)
		return nil
	}

	// Handle byte slice pointers directly
	if bytesPtr, ok := v.(*[]byte); ok {
		*bytesPtr = data
		return nil
	}

	// Unmarshal JSON for other types
	if err := json.Unmarshal(data, v); err != nil {
		return fmt.Errorf("failed to deserialize value: %w", err)
	}
	return nil
}

// StringSerializer implements Serializer for simple string values
type StringSerializer struct{}

// NewStringSerializer creates a new string serializer
func NewStringSerializer() *StringSerializer {
	return &StringSerializer{}
}

// Serialize converts a value to bytes (expects string or []byte)
func (s *StringSerializer) Serialize(v interface{}) ([]byte, error) {
	switch val := v.(type) {
	case string:
		return []byte(val), nil
	case []byte:
		return val, nil
	case fmt.Stringer:
		return []byte(val.String()), nil
	default:
		return nil, fmt.Errorf("StringSerializer only supports string, []byte, or fmt.Stringer types")
	}
}

// Deserialize converts bytes to a string
func (s *StringSerializer) Deserialize(data []byte, v interface{}) error {
	switch ptr := v.(type) {
	case *string:
		*ptr = string(data)
		return nil
	case *[]byte:
		*ptr = data
		return nil
	default:
		return fmt.Errorf("StringSerializer only supports *string or *[]byte destination types")
	}
}
