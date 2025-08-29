package pubsub

import (
	"pub-sub/config"
	"pub-sub/logger"
	"pub-sub/models"
	"testing"
)

// MockLogger implements logger.Logger for testing
type MockLogger struct{}

func (m *MockLogger) Debug(args ...interface{})                 {}
func (m *MockLogger) Info(args ...interface{})                  {}
func (m *MockLogger) Warn(args ...interface{})                  {}
func (m *MockLogger) Error(args ...interface{})                 {}
func (m *MockLogger) Fatal(args ...interface{})                 {}
func (m *MockLogger) Debugf(format string, args ...interface{}) {}
func (m *MockLogger) Infof(format string, args ...interface{})  {}
func (m *MockLogger) Warnf(format string, args ...interface{})  {}
func (m *MockLogger) Errorf(format string, args ...interface{}) {}
func (m *MockLogger) Fatalf(format string, args ...interface{}) {}
func (m *MockLogger) WithField(key string, value interface{}) logger.Logger {
	return m
}
func (m *MockLogger) WithFields(fields logger.Fields) logger.Logger {
	return m
}
func (m *MockLogger) WithError(err error) logger.Logger {
	return m
}

func TestNewPubSub(t *testing.T) {
	cfg := &config.Config{
		MaxMessagesPerTopic: 100,
		MaxPublishRate:      50,
	}
	mockLogger := &MockLogger{}

	ps := NewPubSub(cfg, mockLogger)

	if ps == nil {
		t.Fatal("NewPubSub returned nil")
	}

	if ps.config != cfg {
		t.Error("Config not properly set")
	}

	if len(ps.topics) != 0 {
		t.Error("Topics map should be empty initially")
	}

	if len(ps.subscribers) != 0 {
		t.Error("Subscribers map should be empty initially")
	}
}

func TestCreateTopic(t *testing.T) {
	cfg := &config.Config{
		MaxMessagesPerTopic: 100,
		MaxPublishRate:      50,
	}
	mockLogger := &MockLogger{}

	ps := NewPubSub(cfg, mockLogger)

	// Test creating a topic
	err := ps.CreateTopic("test-topic")
	if err != nil {
		t.Errorf("Failed to create topic: %v", err)
	}

	// Test creating duplicate topic
	err = ps.CreateTopic("test-topic")
	if err == nil {
		t.Error("Should not allow duplicate topic names")
	}

	// Verify topic was created
	if len(ps.topics) != 1 {
		t.Error("Topic count should be 1")
	}

	if _, exists := ps.topics["test-topic"]; !exists {
		t.Error("Topic should exist in topics map")
	}
}

func TestDeleteTopic(t *testing.T) {
	cfg := &config.Config{
		MaxMessagesPerTopic: 100,
		MaxPublishRate:      50,
	}
	mockLogger := &MockLogger{}

	ps := NewPubSub(cfg, mockLogger)

	// Create a topic first
	ps.CreateTopic("test-topic")

	// Test deleting existing topic
	err := ps.DeleteTopic("test-topic")
	if err != nil {
		t.Errorf("Failed to delete topic: %v", err)
	}

	// Test deleting non-existent topic
	err = ps.DeleteTopic("non-existent")
	if err == nil {
		t.Error("Should not allow deleting non-existent topic")
	}

	// Verify topic was deleted
	if len(ps.topics) != 0 {
		t.Error("Topic count should be 0")
	}
}

func TestPublishMessage(t *testing.T) {
	cfg := &config.Config{
		MaxMessagesPerTopic: 100,
		MaxPublishRate:      50,
	}
	mockLogger := &MockLogger{}

	ps := NewPubSub(cfg, mockLogger)

	// Create a topic first
	ps.CreateTopic("test-topic")

	// Test publishing message
	message := &models.Message{
		ID:      "test-message-1",
		Payload: "Hello World",
	}

	err := ps.PublishMessage("test-topic", message)
	if err != nil {
		t.Errorf("Failed to publish message: %v", err)
	}

	// Test publishing to non-existent topic
	err = ps.PublishMessage("non-existent", message)
	if err == nil {
		t.Error("Should not allow publishing to non-existent topic")
	}

	// Verify message was published
	topic := ps.topics["test-topic"]
	if topic.MessageCount != 1 {
		t.Error("Message count should be 1")
	}

	if len(topic.Messages) != 1 {
		t.Error("Messages slice should have 1 message")
	}
}

func TestSubscribeUnsubscribe(t *testing.T) {
	cfg := &config.Config{
		MaxMessagesPerTopic: 100,
		MaxPublishRate:      50,
	}
	mockLogger := &MockLogger{}

	ps := NewPubSub(cfg, mockLogger)

	// Create a topic first
	ps.CreateTopic("test-topic")

	// Test subscribing
	err := ps.Subscribe("subscriber-1", "test-topic", 0)
	if err != nil {
		t.Errorf("Failed to subscribe: %v", err)
	}

	// Test subscribing to non-existent topic
	err = ps.Subscribe("subscriber-1", "non-existent", 0)
	if err == nil {
		t.Error("Should not allow subscribing to non-existent topic")
	}

	// Test unsubscribing
	err = ps.Unsubscribe("subscriber-1", "test-topic")
	if err != nil {
		t.Errorf("Failed to unsubscribe: %v", err)
	}

	// Verify subscriber was removed
	topic := ps.topics["test-topic"]
	if len(topic.Subscribers) != 0 {
		t.Error("Subscriber count should be 0")
	}
}

func TestGetTopics(t *testing.T) {
	cfg := &config.Config{
		MaxMessagesPerTopic: 100,
		MaxPublishRate:      50,
	}
	mockLogger := &MockLogger{}

	ps := NewPubSub(cfg, mockLogger)

	// Create some topics
	ps.CreateTopic("topic-1")
	ps.CreateTopic("topic-2")

	// Get topics
	topics := ps.GetTopics()

	if len(topics) != 2 {
		t.Errorf("Expected 2 topics, got %d", len(topics))
	}
}

func TestGetStats(t *testing.T) {
	cfg := &config.Config{
		MaxMessagesPerTopic: 100,
		MaxPublishRate:      50,
	}
	mockLogger := &MockLogger{}

	ps := NewPubSub(cfg, mockLogger)

	// Create orders topic and publish a message
	ps.CreateTopic("orders")
	message := &models.Message{
		ID:      "order-1",
		Payload: map[string]interface{}{"order_id": "ORD-123"},
	}
	ps.PublishMessage("orders", message)

	// Get stats
	stats := ps.GetStats()

	if stats.Topics.Orders.Messages != 1 {
		t.Errorf("Expected 1 message, got %d", stats.Topics.Orders.Messages)
	}

	if stats.Topics.Orders.Subscribers != 0 {
		t.Errorf("Expected 0 subscribers, got %d", stats.Topics.Orders.Subscribers)
	}
}

func TestGetHealth(t *testing.T) {
	cfg := &config.Config{
		MaxMessagesPerTopic: 100,
		MaxPublishRate:      50,
	}
	mockLogger := &MockLogger{}

	ps := NewPubSub(cfg, mockLogger)

	// Get health
	health := ps.GetHealth()

	if health.Topics != 0 {
		t.Errorf("Expected 0 topics, got %d", health.Topics)
	}

	if health.Subscribers != 0 {
		t.Errorf("Expected 0 subscribers, got %d", health.Subscribers)
	}

	if health.UptimeSec < 0 {
		t.Error("Uptime should be non-negative")
	}
}
