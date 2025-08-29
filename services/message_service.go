package services

import (
	"errors"
	"pub-sub/logger"
	"pub-sub/models"
	"pub-sub/pubsub"
)

// MessageService handles message-related business logic
type MessageService struct {
	pubSub *pubsub.PubSub
	logger logger.Logger
}

// NewMessageService creates a new message service
func NewMessageService(pubSub *pubsub.PubSub, log logger.Logger) *MessageService {
	return &MessageService{
		pubSub: pubSub,
		logger: log,
	}
}

// PublishMessage publishes a message to a topic
func (s *MessageService) PublishMessage(topic string, message *models.Message) (*models.PublishResponse, error) {
	if topic == "" {
		return nil, errors.New("topic is required")
	}

	if message == nil {
		return nil, errors.New("message is required")
	}

	if message.ID == "" {
		return nil, errors.New("message ID is required")
	}

	if err := s.pubSub.PublishMessage(topic, message); err != nil {
		s.logger.Errorf("Failed to publish message to topic %s: %v", topic, err)
		return nil, err
	}

	s.logger.Infof("Message %s published to topic %s successfully", message.ID, topic)
	return &models.PublishResponse{
		Status: "published",
		Topic:  topic,
	}, nil
}
