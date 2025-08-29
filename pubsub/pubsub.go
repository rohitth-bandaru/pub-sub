package pubsub

import (
	"errors"
	"pub-sub/config"
	"pub-sub/logger"
	"pub-sub/models"
	"sync"
	"time"
)

// PubSub represents the main pub-sub system
type PubSub struct {
	topics      map[string]*Topic      // Map of topic names to Topic instances
	subscribers map[string]*Subscriber // Map of subscriber IDs to Subscriber instances
	config      *config.Config         // System configuration
	mutex       sync.RWMutex           // Read-write mutex for thread safety
	startTime   time.Time              // System start time for uptime calculation
	logger      logger.Logger          // Logger instance
}

// Topic represents a topic with its messages and subscribers
type Topic struct {
	Name          string                 // Topic name
	Messages      []*models.Message      // Circular buffer of messages
	Subscribers   map[string]*Subscriber // Map of subscriber IDs to Subscriber instances
	MessageCount  int                    // Total messages published
	CreatedAt     time.Time              // When topic was created
	LastMessageAt time.Time              // When last message was published
	mutex         sync.RWMutex           // Topic-level mutex for thread safety
}

// Subscriber represents a WebSocket connection that can receive messages
type Subscriber struct {
	ID       string                     // Unique subscriber identifier
	Topics   map[string]bool            // Set of subscribed topics
	SendChan chan *models.ServerMessage // Channel to send messages to this subscriber
	conn     interface{}                // WebSocket connection (will be set by WebSocket handler)
	mutex    sync.RWMutex               // Subscriber-level mutex
}

// NewPubSub creates a new pub-sub system instance
func NewPubSub(cfg *config.Config, log logger.Logger) *PubSub {
	return &PubSub{
		topics:      make(map[string]*Topic),
		subscribers: make(map[string]*Subscriber),
		config:      cfg,
		startTime:   time.Now(),
		logger:      log,
	}
}

// CreateTopic creates a new topic if it doesn't exist
func (ps *PubSub) CreateTopic(name string) error {
	ps.mutex.Lock()
	defer ps.mutex.Unlock()

	// Check if topic already exists
	if _, exists := ps.topics[name]; exists {
		return errors.New("topic already exists")
	}

	// Create new topic with circular buffer for messages
	topic := &Topic{
		Name:        name,
		Messages:    make([]*models.Message, 0, ps.config.MaxMessagesPerTopic),
		Subscribers: make(map[string]*Subscriber),
		CreatedAt:   time.Now(),
	}

	ps.topics[name] = topic
	ps.logger.WithFields(logger.Fields{
		"topic":  name,
		"action": "create",
	}).Info("Topic created successfully")
	return nil
}

// DeleteTopic deletes a topic and notifies all subscribers
func (ps *PubSub) DeleteTopic(name string) error {
	ps.mutex.Lock()
	defer ps.mutex.Unlock()

	topic, exists := ps.topics[name]
	if !exists {
		return errors.New("topic does not exist")
	}

	// Notify all subscribers that topic is being deleted
	topic.mutex.Lock()
	for _, subscriber := range topic.Subscribers {
		// Send deletion notification
		select {
		case subscriber.SendChan <- &models.ServerMessage{
			Type:  "info",
			Topic: name,
			Msg:   "topic_deleted",
			TS:    time.Now().Format(time.RFC3339),
		}:
		default:
			// Channel is full, skip
		}

		// Remove topic from subscriber's topic list
		subscriber.mutex.Lock()
		delete(subscriber.Topics, name)
		subscriber.mutex.Unlock()
	}
	topic.mutex.Unlock()

	// Remove topic from all subscribers
	for _, subscriber := range ps.subscribers {
		subscriber.mutex.Lock()
		delete(subscriber.Topics, name)
		subscriber.mutex.Unlock()
	}

	// Delete the topic
	delete(ps.topics, name)
	ps.logger.WithFields(logger.Fields{
		"topic":                name,
		"action":               "delete",
		"subscribers_affected": len(topic.Subscribers),
	}).Info("Topic deleted successfully")
	return nil
}

// PublishMessage publishes a message to a topic
func (ps *PubSub) PublishMessage(topicName string, message *models.Message) error {
	ps.mutex.RLock()
	topic, exists := ps.topics[topicName]
	ps.mutex.RUnlock()

	if !exists {
		return errors.New("TOPIC_NOT_FOUND")
	}

	// Add message to topic with circular buffer logic
	topic.mutex.Lock()

	// Add new message
	topic.Messages = append(topic.Messages, message)

	// Maintain circular buffer size
	if len(topic.Messages) > ps.config.MaxMessagesPerTopic {
		topic.Messages = topic.Messages[1:] // Remove oldest message
	}

	topic.MessageCount++
	topic.LastMessageAt = time.Now()
	topic.mutex.Unlock()

	// Notify all subscribers
	ps.notifySubscribers(topicName, message)

	ps.logger.WithFields(logger.Fields{
		"topic":             topicName,
		"message_id":        message.ID,
		"action":            "publish",
		"subscribers_count": len(topic.Subscribers),
	}).Info("Message published successfully")
	return nil
}

// Subscribe adds a subscriber to a topic
func (ps *PubSub) Subscribe(subscriberID, topicName string, lastN int) error {
	ps.mutex.RLock()
	topic, exists := ps.topics[topicName]
	ps.mutex.RUnlock()

	if !exists {
		return errors.New("TOPIC_NOT_FOUND")
	}

	// Get or create subscriber
	ps.mutex.Lock()
	subscriber, exists := ps.subscribers[subscriberID]
	if !exists {
		subscriber = &Subscriber{
			ID:       subscriberID,
			Topics:   make(map[string]bool),
			SendChan: make(chan *models.ServerMessage, 100), // Buffer for messages
		}
		ps.subscribers[subscriberID] = subscriber
	}
	ps.mutex.Unlock()

	// Add topic to subscriber
	subscriber.mutex.Lock()
	subscriber.Topics[topicName] = true
	subscriber.mutex.Unlock()

	// Add subscriber to topic
	topic.mutex.Lock()
	topic.Subscribers[subscriberID] = subscriber
	topic.mutex.Unlock()

	// Send historical messages if requested
	if lastN > 0 {
		ps.sendHistoricalMessages(subscriber, topic, lastN)
	}

	ps.logger.WithFields(logger.Fields{
		"subscriber_id":       subscriberID,
		"topic":               topicName,
		"action":              "subscribe",
		"historical_messages": lastN,
		"total_subscribers":   len(topic.Subscribers),
	}).Info("Subscriber subscribed successfully")
	return nil
}

// Unsubscribe removes a subscriber from a topic
func (ps *PubSub) Unsubscribe(subscriberID, topicName string) error {
	ps.mutex.RLock()
	topic, exists := ps.topics[topicName]
	ps.mutex.RUnlock()

	if !exists {
		return errors.New("TOPIC_NOT_FOUND")
	}

	// Remove subscriber from topic
	topic.mutex.Lock()
	delete(topic.Subscribers, subscriberID)
	topic.mutex.Unlock()

	// Remove topic from subscriber
	ps.mutex.RLock()
	subscriber, exists := ps.subscribers[subscriberID]
	ps.mutex.RUnlock()

	if exists {
		subscriber.mutex.Lock()
		delete(subscriber.Topics, topicName)
		subscriber.mutex.Unlock()
	}

	ps.logger.WithFields(logger.Fields{
		"subscriber_id":         subscriberID,
		"topic":                 topicName,
		"action":                "unsubscribe",
		"remaining_subscribers": len(topic.Subscribers),
	}).Info("Subscriber unsubscribed successfully")
	return nil
}

// GetTopics returns a list of all topics
func (ps *PubSub) GetTopics() []models.TopicInfo {
	ps.mutex.RLock()
	defer ps.mutex.RUnlock()

	topics := make([]models.TopicInfo, 0, len(ps.topics))
	for _, topic := range ps.topics {
		topic.mutex.RLock()
		topics = append(topics, models.TopicInfo{
			Name:        topic.Name,
			Subscribers: len(topic.Subscribers),
		})
		topic.mutex.RUnlock()
	}

	return topics
}

// GetStats returns system statistics
func (ps *PubSub) GetStats() models.Stats {
	ps.mutex.RLock()
	defer ps.mutex.RUnlock()

	stats := models.Stats{}

	// Find orders topic if it exists
	if ordersTopic, exists := ps.topics["orders"]; exists {
		ordersTopic.mutex.RLock()
		stats.Topics.Orders.Messages = ordersTopic.MessageCount
		stats.Topics.Orders.Subscribers = len(ordersTopic.Subscribers)
		ordersTopic.mutex.RUnlock()
	}

	return stats
}

// GetHealth returns system health status
func (ps *PubSub) GetHealth() models.Health {
	ps.mutex.RLock()
	defer ps.mutex.RUnlock()

	totalSubscribers := 0
	for _, subscriber := range ps.subscribers {
		subscriber.mutex.RLock()
		totalSubscribers += len(subscriber.Topics)
		subscriber.mutex.RUnlock()
	}

	return models.Health{
		UptimeSec:   int(time.Since(ps.startTime).Seconds()),
		Topics:      len(ps.topics),
		Subscribers: totalSubscribers,
	}
}

// RemoveSubscriber removes a subscriber from all topics and the system
func (ps *PubSub) RemoveSubscriber(subscriberID string) {
	ps.mutex.Lock()
	defer ps.mutex.Unlock()

	subscriber, exists := ps.subscribers[subscriberID]
	if !exists {
		return
	}

	// Remove subscriber from all topics
	subscriber.mutex.RLock()
	topics := make([]string, 0, len(subscriber.Topics))
	for topicName := range subscriber.Topics {
		topics = append(topics, topicName)
	}
	subscriber.mutex.RUnlock()

	for _, topicName := range topics {
		if topic, exists := ps.topics[topicName]; exists {
			topic.mutex.Lock()
			delete(topic.Subscribers, subscriberID)
			topic.mutex.Unlock()
		}
	}

	// Close subscriber's message channel
	close(subscriber.SendChan)

	// Remove subscriber from system
	delete(ps.subscribers, subscriberID)

	ps.logger.WithFields(logger.Fields{
		"subscriber_id":         subscriberID,
		"action":                "remove",
		"topics_subscribed":     len(subscriber.Topics),
		"remaining_subscribers": len(ps.subscribers),
	}).Info("Subscriber removed successfully")
}

// notifySubscribers sends a message to all subscribers of a topic
func (ps *PubSub) notifySubscribers(topicName string, message *models.Message) {
	ps.mutex.RLock()
	topic, exists := ps.topics[topicName]
	ps.mutex.RUnlock()

	if !exists {
		return
	}

	topic.mutex.RLock()
	subscribers := make([]*Subscriber, 0, len(topic.Subscribers))
	for _, subscriber := range topic.Subscribers {
		subscribers = append(subscribers, subscriber)
	}
	topic.mutex.RUnlock()

	// Send message to all subscribers
	for _, subscriber := range subscribers {
		serverMessage := &models.ServerMessage{
			Type:    "event",
			Topic:   topicName,
			Message: message,
			TS:      time.Now().Format(time.RFC3339),
		}

		select {
		case subscriber.SendChan <- serverMessage:
			// Message sent successfully
		default:
			// Channel is full, send SLOW_CONSUMER error
			errorMessage := &models.ServerMessage{
				Type: "error",
				Error: &models.Error{
					Code:    "SLOW_CONSUMER",
					Message: "Subscriber queue overflow",
				},
				TS: time.Now().Format(time.RFC3339),
			}

			select {
			case subscriber.SendChan <- errorMessage:
				// Error sent successfully
			default:
				// Even error channel is full, disconnect subscriber
				ps.logger.WithFields(logger.Fields{
					"subscriber_id": subscriber.ID,
					"topic":         topicName,
					"action":        "disconnect",
					"reason":        "channel_overflow",
				}).Warn("Subscriber disconnected due to channel overflow")
				go ps.RemoveSubscriber(subscriber.ID)
			}
		}
	}
}

// sendHistoricalMessages sends the last N messages to a subscriber
func (ps *PubSub) sendHistoricalMessages(subscriber *Subscriber, topic *Topic, lastN int) {
	topic.mutex.RLock()
	messages := make([]*models.Message, len(topic.Messages))
	copy(messages, topic.Messages)
	topic.mutex.RUnlock()

	// Send last N messages in reverse order (newest first)
	start := len(messages) - lastN
	if start < 0 {
		start = 0
	}

	for i := len(messages) - 1; i >= start; i-- {
		serverMessage := &models.ServerMessage{
			Type:    "event",
			Topic:   topic.Name,
			Message: messages[i],
			TS:      time.Now().Format(time.RFC3339),
		}

		select {
		case subscriber.SendChan <- serverMessage:
			// Historical message sent successfully
		default:
			// Channel is full, stop sending historical messages
			ps.logger.WithFields(logger.Fields{
				"subscriber_id": subscriber.ID,
				"topic":         topic.Name,
				"action":        "historical_replay_stopped",
				"reason":        "channel_full",
				"messages_sent": i,
			}).Warn("Historical message replay stopped due to full channel")
			return
		}
	}
}
