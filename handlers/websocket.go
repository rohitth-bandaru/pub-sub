package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"pub-sub/config"
	"pub-sub/logger"
	"pub-sub/models"
	"pub-sub/pubsub"
	"pub-sub/utils"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// WebSocketHandler handles WebSocket connections for pub-sub operations
type WebSocketHandler struct {
	pubsub   *pubsub.PubSub              // Reference to the pub-sub system
	upgrader websocket.Upgrader          // WebSocket upgrader
	clients  map[string]*WebSocketClient // Map of client IDs to WebSocket clients
	mutex    sync.RWMutex                // Mutex for thread-safe client management
	logger   logger.Logger               // Logger instance
}

// WebSocketClient represents a connected WebSocket client
type WebSocketClient struct {
	ID          string                     // Unique client identifier
	Conn        *websocket.Conn            // WebSocket connection
	Topics      map[string]string          // Map of topic names to subscription IDs
	SendChan    chan *models.ServerMessage // Channel for sending messages
	Handler     *WebSocketHandler          // Reference to the handler
	mutex       sync.RWMutex               // Client-level mutex
	stopChan    chan struct{}              // Channel to stop message forwarding
	ConnectedAt time.Time                  // When the client connected
}

// NewWebSocketHandler creates a new WebSocket handler
func NewWebSocketHandler(pubsub *pubsub.PubSub, cfg *config.Config, log logger.Logger) *WebSocketHandler {
	return &WebSocketHandler{
		pubsub: pubsub,
		logger: log,
		upgrader: websocket.Upgrader{
			ReadBufferSize:  cfg.ReadBufferSize,
			WriteBufferSize: cfg.WriteBufferSize,
			CheckOrigin: func(r *http.Request) bool {
				// Allow all origins for development
				// In production, implement proper origin checking
				return true
			},
		},
		clients: make(map[string]*WebSocketClient),
	}
}

// HandleWebSocket handles WebSocket upgrade and client management
func (h *WebSocketHandler) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	// Upgrade HTTP connection to WebSocket
	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		h.logger.Errorf("WebSocket upgrade failed: %v", err)
		return
	}

	// Generate unique client ID
	clientID := generateClientID()

	// Create new WebSocket client
	client := &WebSocketClient{
		ID:          clientID,
		Conn:        conn,
		Topics:      make(map[string]string),
		SendChan:    make(chan *models.ServerMessage, 100), // Buffer for messages
		Handler:     h,
		stopChan:    make(chan struct{}),
		ConnectedAt: time.Now(),
	}

	// Register client
	h.mutex.Lock()
	h.clients[clientID] = client
	h.mutex.Unlock()

	h.logger.Infof("WebSocket client connected successfully: client_id=%s, remote_addr=%s, user_agent=%s", clientID, r.RemoteAddr, r.UserAgent())

	// Start client goroutines
	go client.readPump()
	go client.writePump()

	// Send welcome message
	client.sendSystemMessage("Connected to Pub/Sub system", "")
}

// readPump reads messages from the WebSocket connection
func (c *WebSocketClient) readPump() {
	defer func() {
		c.Handler.removeClient(c.ID)
		c.Conn.Close()
	}()

	// Set read deadline
	c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		// Read message from WebSocket
		_, messageBytes, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				c.Handler.logger.Errorf("WebSocket read error for client %s: %v", c.ID, err)
			}
			break
		}

		// Parse WebSocket message
		var clientMessage models.ClientMessage
		if err := json.Unmarshal(messageBytes, &clientMessage); err != nil {
			c.sendErrorMessage("Invalid message format", "BAD_REQUEST", err.Error(), "")
			continue
		}

		// Handle message based on type
		c.handleMessage(&clientMessage)
	}
}

// writePump writes messages to the WebSocket connection
func (c *WebSocketClient) writePump() {
	ticker := time.NewTicker(54 * time.Second) // Send ping every 54 seconds
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.SendChan:
			if !ok {
				// Channel closed
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			// Set write deadline
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))

			// Send message
			if err := c.Conn.WriteJSON(message); err != nil {
				c.Handler.logger.Errorf("WebSocket write error for client %s: %v", c.ID, err)
				return
			}

		case <-ticker.C:
			// Send ping
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// handleMessage processes incoming WebSocket messages
func (c *WebSocketClient) handleMessage(clientMessage *models.ClientMessage) {
	switch clientMessage.Type {
	case "publish":
		c.handlePublish(clientMessage)
	case "subscribe":
		c.handleSubscribe(clientMessage)
	case "unsubscribe":
		c.handleUnsubscribe(clientMessage)
	case "ping":
		c.handlePing(clientMessage)
	default:
		c.sendErrorMessage("Unknown message type", "BAD_REQUEST", "Unsupported message type: "+clientMessage.Type, clientMessage.RequestID)
	}
}

// handlePublish handles publish messages
func (c *WebSocketClient) handlePublish(clientMessage *models.ClientMessage) {
	if clientMessage.Topic == "" {
		c.sendErrorMessage("Missing topic", "BAD_REQUEST", "Topic is required for publish", clientMessage.RequestID)
		return
	}

	if clientMessage.Message == nil {
		c.sendErrorMessage("Missing message", "BAD_REQUEST", "Message is required for publish", clientMessage.RequestID)
		return
	}

	if clientMessage.Message.ID == "" {
		c.sendErrorMessage("Missing message ID", "BAD_REQUEST", "Message ID is required for publish", clientMessage.RequestID)
		return
	}

	// Publish message to topic
	err := c.Handler.pubsub.PublishMessage(clientMessage.Topic, clientMessage.Message)
	if err != nil {
		errorCode := "INTERNAL"
		if err.Error() == "TOPIC_NOT_FOUND" {
			errorCode = "TOPIC_NOT_FOUND"
		}
		c.sendErrorMessage("Publish failed", errorCode, err.Error(), clientMessage.RequestID)
		return
	}

	// Send acknowledgment
	c.sendAcknowledgment(clientMessage.Topic, "ok", clientMessage.RequestID)
}

// handleSubscribe handles subscribe messages
func (c *WebSocketClient) handleSubscribe(clientMessage *models.ClientMessage) {
	if clientMessage.Topic == "" {
		c.sendErrorMessage("Missing topic", "BAD_REQUEST", "Topic is required for subscribe", clientMessage.RequestID)
		return
	}

	// Use provided client ID or fall back to generated WebSocket client ID
	subscriberID := clientMessage.ClientID
	if subscriberID == "" {
		subscriberID = c.ID
		c.Handler.logger.Debugf("Using generated client ID %s for subscription to topic %s", subscriberID, clientMessage.Topic)
	}

	// Subscribe to topic
	err := c.Handler.pubsub.Subscribe(subscriberID, clientMessage.Topic, clientMessage.LastN)
	if err != nil {
		errorCode := "INTERNAL"
		if err.Error() == "TOPIC_NOT_FOUND" {
			errorCode = "TOPIC_NOT_FOUND"
		}
		c.sendErrorMessage("Subscribe failed", errorCode, err.Error(), clientMessage.RequestID)
		return
	}

	// Add topic to client's topic list with subscription ID
	c.mutex.Lock()
	c.Topics[clientMessage.Topic] = subscriberID
	c.mutex.Unlock()

	// Start a goroutine to forward messages from pubsub to WebSocket client
	go c.forwardMessagesFromPubSub(clientMessage.Topic)

	// Send acknowledgment
	c.sendAcknowledgment(clientMessage.Topic, "ok", clientMessage.RequestID)
}

// forwardMessagesFromPubSub forwards messages from the pubsub system to the WebSocket client
func (c *WebSocketClient) forwardMessagesFromPubSub(topicName string) {
	// Get the subscription ID for this topic
	c.mutex.RLock()
	subscriberID, exists := c.Topics[topicName]
	c.mutex.RUnlock()

	if !exists {
		c.Handler.logger.Errorf("Topic %s not found in client's topic list", topicName)
		return
	}

	// Get the subscriber's message channel from pubsub system
	messageChan := c.Handler.pubsub.GetSubscriberChannel(subscriberID)
	if messageChan == nil {
		c.Handler.logger.Errorf("Subscriber %s not found in pubsub system", subscriberID)
		return
	}

	// Listen for messages on the subscriber's channel
	for {
		select {
		case message, ok := <-messageChan:
			if !ok {
				// Channel closed, stop forwarding
				return
			}

			// Only forward messages for the specific topic
			if message.Topic == topicName {
				// Send message to WebSocket client
				select {
				case c.SendChan <- message:
					// Message sent successfully
				default:
					// Channel is full, log warning
					c.Handler.logger.Warnf("WebSocket client %s channel full, dropping message", c.ID)
				}
			}
		case <-c.stopChan:
			// Stop forwarding
			return
		}
	}
}

// handleUnsubscribe handles unsubscribe messages
func (c *WebSocketClient) handleUnsubscribe(clientMessage *models.ClientMessage) {
	if clientMessage.Topic == "" {
		c.sendErrorMessage("Missing topic", "BAD_REQUEST", "Topic is required for unsubscribe", clientMessage.RequestID)
		return
	}

	// Use provided client ID or fall back to generated WebSocket client ID
	subscriberID := clientMessage.ClientID
	if subscriberID == "" {
		subscriberID = c.ID
		c.Handler.logger.Debugf("Using generated client ID %s for unsubscription from topic %s", subscriberID, clientMessage.Topic)
	}

	// Unsubscribe from topic
	err := c.Handler.pubsub.Unsubscribe(subscriberID, clientMessage.Topic)
	if err != nil {
		errorCode := "INTERNAL"
		if err.Error() == "TOPIC_NOT_FOUND" {
			errorCode = "TOPIC_NOT_FOUND"
		}
		c.sendErrorMessage("Unsubscribe failed", errorCode, err.Error(), clientMessage.RequestID)
		return
	}

	// Remove topic from client's topic list
	c.mutex.Lock()
	delete(c.Topics, clientMessage.Topic)
	c.mutex.Unlock()

	// Stop message forwarding for this topic
	close(c.stopChan)
	c.stopChan = make(chan struct{}) // Create new stop channel for future subscriptions

	// Send acknowledgment
	c.sendAcknowledgment(clientMessage.Topic, "ok", clientMessage.RequestID)
}

// handlePing handles ping messages
func (c *WebSocketClient) handlePing(clientMessage *models.ClientMessage) {
	// Send pong response
	response := models.ServerMessage{
		Type:      "pong",
		RequestID: clientMessage.RequestID,
		TS:        time.Now().Format(time.RFC3339),
	}

	c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
	if err := c.Conn.WriteJSON(response); err != nil {
		c.Handler.logger.Errorf("Failed to send pong to client %s: %v", c.ID, err)
	}
}

// sendAcknowledgment sends an acknowledgment message to the client
func (c *WebSocketClient) sendAcknowledgment(topic, status, requestID string) {
	ackMessage := models.ServerMessage{
		Type:      "ack",
		RequestID: requestID,
		Topic:     topic,
		Status:    status,
		TS:        time.Now().Format(time.RFC3339),
	}

	select {
	case c.SendChan <- &ackMessage:
		// Acknowledgment sent successfully
	default:
		// Channel is full, log error
		c.Handler.logger.Warnf("Failed to send acknowledgment to client %s: channel full", c.ID)
	}
}

// sendSystemMessage sends a system message to the client
func (c *WebSocketClient) sendSystemMessage(message, topic string) {
	infoMessage := models.ServerMessage{
		Type: "info",
		Msg:  message,
		TS:   time.Now().Format(time.RFC3339),
	}

	if topic != "" {
		infoMessage.Topic = topic
	}

	select {
	case c.SendChan <- &infoMessage:
		// System message sent successfully
	default:
		// Channel is full, log error
		c.Handler.logger.Warnf("Failed to send system message to client %s: channel full", c.ID)
	}
}

// sendErrorMessage sends an error message to the client
func (c *WebSocketClient) sendErrorMessage(message, code, details, requestID string) {
	errorMessage := models.ServerMessage{
		Type:      "error",
		RequestID: requestID,
		Error: &models.Error{
			Code:    code,
			Message: details,
		},
		TS: time.Now().Format(time.RFC3339),
	}

	select {
	case c.SendChan <- &errorMessage:
		// Error message sent successfully
	default:
		// Channel is full, log error
		c.Handler.logger.Warnf("Failed to send error message to client %s: channel full", c.ID)
	}
}

// removeClient removes a client from the handler
func (h *WebSocketHandler) removeClient(clientID string) {
	h.mutex.Lock()
	client, exists := h.clients[clientID]
	if exists {
		topicsCount := len(client.Topics)
		delete(h.clients, clientID)
		h.mutex.Unlock()

		// Remove all subscriptions for this client
		for topicName, subscriberID := range client.Topics {
			h.logger.Debugf("Removing subscription %s from topic %s", subscriberID, topicName)
			h.pubsub.RemoveSubscriber(subscriberID)
		}

		h.logger.Infof("WebSocket client disconnected: client_id=%s, topics_subscribed=%d", clientID, topicsCount)
	} else {
		delete(h.clients, clientID)
		h.mutex.Unlock()
		// Remove client from pub-sub system as fallback
		h.pubsub.RemoveSubscriber(clientID)

		h.logger.Infof("WebSocket client disconnected: client_id=%s, topics_subscribed=0", clientID)
	}
}

// GetActiveClients returns details of all active WebSocket clients
func (h *WebSocketHandler) GetActiveClients() []models.ClientInfo {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	clients := make([]models.ClientInfo, 0, len(h.clients))
	for _, client := range h.clients {
		client.mutex.RLock()
		topics := make([]string, 0, len(client.Topics))
		for topicName := range client.Topics {
			topics = append(topics, topicName)
		}
		client.mutex.RUnlock()

		clientInfo := models.ClientInfo{
			ID:          client.ID,
			RemoteAddr:  client.Conn.RemoteAddr().String(),
			Topics:      topics,
			ConnectedAt: client.ConnectedAt,
			IsConnected: true,
		}
		clients = append(clients, clientInfo)
	}

	return clients
}

// generateClientID generates a unique client identifier
func generateClientID() string {
	return utils.GenerateClientID()
}

// Shutdown gracefully shuts down all WebSocket connections
func (h *WebSocketHandler) Shutdown(ctx context.Context) {
	h.logger.Info("Shutting down WebSocket handler...")

	// Get all clients and close their connections
	h.mutex.RLock()
	clients := make([]*WebSocketClient, 0, len(h.clients))
	for _, client := range h.clients {
		clients = append(clients, client)
	}
	h.mutex.RUnlock()

	// Close all client connections
	for _, client := range clients {
		// Stop message forwarding goroutines
		close(client.stopChan)

		// Close WebSocket connection
		client.Conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseGoingAway, "Server shutting down"))
		client.Conn.Close()
	}

	h.logger.Info("WebSocket handler shutdown complete")
}
