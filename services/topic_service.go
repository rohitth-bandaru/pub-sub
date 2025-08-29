package services

import (
	"pub-sub/logger"
	"pub-sub/models"
	"pub-sub/pubsub"
	"time"
)

// TopicService handles topic-related business logic
type TopicService struct {
	pubSub *pubsub.PubSub
	logger logger.Logger
}

// NewTopicService creates a new topic service
func NewTopicService(pubSub *pubsub.PubSub, log logger.Logger) *TopicService {
	return &TopicService{
		pubSub: pubSub,
		logger: log,
	}
}

// CreateTopic creates a new topic
func (s *TopicService) CreateTopic(name string) (*models.TopicResponse, error) {
	if name == "" {
		return nil, models.ErrTopicRequired
	}

	if err := s.pubSub.CreateTopic(name); err != nil {
		s.logger.Errorf("Failed to create topic %s: %v", name, err)
		return nil, err
	}

	s.logger.Infof("Topic %s created successfully", name)
	return &models.TopicResponse{
		Status: "created",
		Topic:  name,
	}, nil
}

// DeleteTopic deletes a topic
func (s *TopicService) DeleteTopic(name string) (*models.TopicResponse, error) {
	if name == "" {
		return nil, models.ErrTopicRequired
	}

	if err := s.pubSub.DeleteTopic(name); err != nil {
		s.logger.Errorf("Failed to delete topic %s: %v", name, err)
		return nil, err
	}

	s.logger.Infof("Topic %s deleted successfully", name)
	return &models.TopicResponse{
		Status: "deleted",
		Topic:  name,
	}, nil
}

// ListTopics returns a list of all topics
func (s *TopicService) ListTopics() *models.TopicList {
	topics := s.pubSub.GetTopics()
	return &models.TopicList{
		Topics: topics,
	}
}

// GetTopic returns a specific topic
func (s *TopicService) GetTopic(name string) (*models.Topic, error) {
	if name == "" {
		return nil, models.ErrTopicRequired
	}

	topics := s.pubSub.GetTopics()
	for _, topic := range topics {
		if topic.Name == name {
			return &models.Topic{
				Name:          topic.Name,
				Subscribers:   topic.Subscribers,
				MessageCount:  0, // This would need to be implemented in pubsub
				CreatedAt:     time.Time{}, // This would need to be implemented in pubsub
				LastMessageAt: time.Time{}, // This would need to be implemented in pubsub
			}, nil
		}
	}

	return nil, models.ErrTopicNotFound
}
