package services

import (
	"pub-sub/logger"
	"pub-sub/models"
	"pub-sub/pubsub"
)

// SystemService handles system-related operations
type SystemService struct {
	pubSub *pubsub.PubSub
	logger logger.Logger
}

// NewSystemService creates a new system service
func NewSystemService(pubSub *pubsub.PubSub, log logger.Logger) *SystemService {
	return &SystemService{
		pubSub: pubSub,
		logger: log,
	}
}

// GetStats returns system statistics
func (s *SystemService) GetStats() *models.Stats {
	stats := s.pubSub.GetStats()
	s.logger.Debug("System stats retrieved successfully")
	return &stats
}

// GetHealth returns system health status
func (s *SystemService) GetHealth() *models.Health {
	health := s.pubSub.GetHealth()
	s.logger.Debug("System health check completed")
	return &health
}
