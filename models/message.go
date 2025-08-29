package models

import (
	"time"
)

// ClientMessage represents messages sent from client to server
type ClientMessage struct {
	Type      string   `json:"type"`       // subscribe, unsubscribe, publish, ping
	Topic     string   `json:"topic"`      // required for subscribe/unsubscribe/publish
	Message   *Message `json:"message"`    // required for publish
	ClientID  string   `json:"client_id"`  // required for subscribe/unsubscribe
	LastN     int      `json:"last_n"`     // optional: number of historical messages to replay
	RequestID string   `json:"request_id"` // optional: correlation id
}

// ServerMessage represents messages sent from server to client
type ServerMessage struct {
	Type      string   `json:"type"`       // ack, event, error, pong, info
	RequestID string   `json:"request_id"` // echoed if provided
	Topic     string   `json:"topic"`      // topic name
	Message   *Message `json:"message"`    // message data for events
	Error     *Error   `json:"error"`      // error details
	Status    string   `json:"status"`     // status for ack messages
	Msg       string   `json:"msg"`        // info message
	TS        string   `json:"ts"`         // server timestamp
}

// Message represents a message published to a topic
type Message struct {
	ID      string      `json:"id"`      // Message identifier (UUID)
	Payload interface{} `json:"payload"` // Message payload
}

// Error represents error details
type Error struct {
	Code    string `json:"code"`    // Error code
	Message string `json:"message"` // Error message
}

// Topic represents a topic in the pub-sub system
type Topic struct {
	Name          string    `json:"name"`            // Topic name
	Subscribers   int       `json:"subscribers"`     // Number of active subscribers
	MessageCount  int       `json:"messages"`        // Total messages published
	CreatedAt     time.Time `json:"created_at"`      // When topic was created
	LastMessageAt time.Time `json:"last_message_at"` // When last message was published
}

// Stats represents system statistics
type Stats struct {
	TotalTopics       int                   `json:"total_topics"`
	TotalMessages     int                   `json:"total_messages"`
	TotalSubscribers  int                   `json:"total_subscribers"`
	ActiveConnections int                   `json:"active_connections"`
	UptimeSeconds     int                   `json:"uptime_seconds"`
	Topics            map[string]TopicStats `json:"topics"`
	GeneratedAt       string                `json:"generated_at"`
}

// TopicStats represents statistics for a specific topic
type TopicStats struct {
	Name          string    `json:"name"`
	Messages      int       `json:"messages"`
	Subscribers   int       `json:"subscribers"`
	CreatedAt     time.Time `json:"created_at"`
	LastMessageAt time.Time `json:"last_message_at"`
}

// Health represents system health status
type Health struct {
	UptimeSec   int `json:"uptime_sec"`  // System uptime in seconds
	Topics      int `json:"topics"`      // Total number of topics
	Subscribers int `json:"subscribers"` // Total number of subscribers
}

// TopicList represents a list of topics
type TopicList struct {
	Topics []TopicInfo `json:"topics"`
}

// TopicInfo represents basic topic information
type TopicInfo struct {
	Name        string `json:"name"`
	Subscribers int    `json:"subscribers"`
}

// TopicResponse represents topic operation responses
type TopicResponse struct {
	Status string `json:"status"`
	Topic  string `json:"topic"`
}

// PublishResponse represents message publishing responses
type PublishResponse struct {
	Status string `json:"status"`
	Topic  string `json:"topic"`
}

// ClientInfo represents information about a WebSocket client
type ClientInfo struct {
	ID          string    `json:"id"`           // Unique client identifier
	RemoteAddr  string    `json:"remote_addr"`  // Client's remote address
	Topics      []string  `json:"topics"`       // List of subscribed topics
	ConnectedAt time.Time `json:"connected_at"` // When the client connected
	IsConnected bool      `json:"is_connected"` // Current connection status
}

// ClientList represents a list of WebSocket clients
type ClientList struct {
	Clients []ClientInfo `json:"clients"`
	Total   int          `json:"total"` // Total number of clients
}

// WebSocketClientProvider interface for getting WebSocket client information
type WebSocketClientProvider interface {
	GetActiveClients() []ClientInfo
}
