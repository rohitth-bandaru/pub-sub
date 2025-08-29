package models

import "errors"

// Custom error types for better error handling
var (
	ErrTopicNotFound    = errors.New("TOPIC_NOT_FOUND")
	ErrTopicExists      = errors.New("TOPIC_EXISTS")
	ErrInvalidRequest   = errors.New("INVALID_REQUEST")
	ErrMessageRequired  = errors.New("MESSAGE_REQUIRED")
	ErrTopicRequired    = errors.New("TOPIC_REQUIRED")
	ErrMessageIDRequired = errors.New("MESSAGE_ID_REQUIRED")
	ErrSubscriberNotFound = errors.New("SUBSCRIBER_NOT_FOUND")
	ErrChannelOverflow   = errors.New("CHANNEL_OVERFLOW")
	ErrSlowConsumer      = errors.New("SLOW_CONSUMER")
)

// IsErrorType checks if an error is of a specific type
func IsErrorType(err error, target error) bool {
	return errors.Is(err, target)
}
