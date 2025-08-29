package services

import (
	"pub-sub/logger"
	"pub-sub/models"
	"pub-sub/pubsub"
)

// SystemService handles system-related operations
type SystemService struct {
	pubSub           *pubsub.PubSub
	logger           logger.Logger
	wsClientProvider models.WebSocketClientProvider
}

// NewSystemService creates a new system service
func NewSystemService(pubSub *pubsub.PubSub, log logger.Logger, wsProvider models.WebSocketClientProvider) *SystemService {
	return &SystemService{
		pubSub:           pubSub,
		logger:           log,
		wsClientProvider: wsProvider,
	}
}

// GetStats returns system statistics
func (s *SystemService) GetStats() *models.Stats {
	stats := s.pubSub.GetStats()

	s.logger.Debugf("Raw stats from pubsub: TotalTopics=%d, TotalMessages=%d, TotalSubscribers=%d",
		stats.TotalTopics, stats.TotalMessages, stats.TotalSubscribers)

	// Override ActiveConnections with actual WebSocket connection count
	if s.wsClientProvider != nil {
		activeClients := s.wsClientProvider.GetActiveClients()
		stats.ActiveConnections = len(activeClients)
		s.logger.Debugf("Updated ActiveConnections to %d based on WebSocket clients", stats.ActiveConnections)
	} else {
		s.logger.Warn("WebSocket client provider not available, using pubsub subscriber count for ActiveConnections")
	}

	s.logger.Debugf("Final stats: TotalTopics=%d, TotalMessages=%d, TotalSubscribers=%d, ActiveConnections=%d",
		stats.TotalTopics, stats.TotalMessages, stats.TotalSubscribers, stats.ActiveConnections)

	return &stats
}

// GetHealth returns system health status
func (s *SystemService) GetHealth() *models.Health {
	health := s.pubSub.GetHealth()
	s.logger.Debug("System health check completed")
	return &health
}

// GetTopicStats returns statistics for a specific topic
func (s *SystemService) GetTopicStats(topicName string) (*models.TopicStats, error) {
	stats, err := s.pubSub.GetTopicStats(topicName)
	if err != nil {
		s.logger.Errorf("Failed to get topic stats for %s: %v", topicName, err)
		return nil, err
	}
	s.logger.Debugf("Topic stats retrieved for %s", topicName)
	return stats, nil
}

// GetActiveClients returns information about all active WebSocket clients
func (s *SystemService) GetActiveClients() *models.ClientList {
	if s.wsClientProvider == nil {
		s.logger.Warn("WebSocket client provider not available")
		return &models.ClientList{
			Clients: []models.ClientInfo{},
			Total:   0,
		}
	}

	clients := s.wsClientProvider.GetActiveClients()
	s.logger.Debugf("Retrieved %d active WebSocket clients", len(clients))

	return &models.ClientList{
		Clients: clients,
		Total:   len(clients),
	}
}
