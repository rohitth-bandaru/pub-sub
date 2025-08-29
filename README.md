# Go Pub/Sub System

A simplified in-memory Pub/Sub system built in Go with WebSocket support for real-time messaging and REST API for management operations.

## Features

- **Real-time messaging** via WebSocket connections
- **Topic management** via REST API (create, delete, list topics)
- **In-memory storage** with configurable message limits per topic
- **Thread-safe operations** with proper mutex handling
- **Health monitoring** and system statistics
- **Docker support** for easy deployment
- **Configurable settings** via environment variables

## Architecture

The system consists of several key components:

- **PubSub Core**: Manages topics, messages, and subscribers
- **WebSocket Handler**: Handles real-time connections and messaging
- **REST Handler**: Provides HTTP API for management operations
- **Configuration**: Environment-based configuration system

### Design Choices

- **In-memory storage**: No external databases for simplicity and performance
- **Circular buffer**: Messages per topic are limited to prevent memory issues
- **Thread-safe operations**: Uses read-write mutexes for concurrent access
- **Backpressure handling**: Skips messages when subscriber channels are full
- **Graceful shutdown**: Proper cleanup of resources on server shutdown

## Configuration

The system can be configured using environment variables:

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `8080` | Server port |
| `HOST` | `0.0.0.0` | Server host |
| `MAX_MESSAGES_PER_TOPIC` | `1000` | Maximum messages stored per topic |
| `MAX_PUBLISH_RATE` | `100` | Maximum publish rate per topic (messages/sec) |
| `WS_READ_BUFFER_SIZE` | `1024` | WebSocket read buffer size |
| `WS_WRITE_BUFFER_SIZE` | `1024` | WebSocket write buffer size |

## API Endpoints

### WebSocket Endpoint

- **`/ws`** - WebSocket connection for real-time pub/sub operations

#### WebSocket Message Types

- **`publish`**: Publish a message to a topic
- **`subscribe`**: Subscribe to a topic
- **`unsubscribe`**: Unsubscribe from a topic
- **`ping`**: Health check ping

#### WebSocket Message Format

```json
{
  "type": "publish|subscribe|unsubscribe|ping",
  "topic": "topic-name",
  "data": "message-data"
}
```

### REST API Endpoints

#### Topic Management

- **`POST /topics`** - Create a new topic
- **`GET /topics`** - List all topics
- **`GET /topics/{name}`** - Get topic details
- **`DELETE /topics/{name}`** - Delete a topic

#### Messaging

- **`POST /publish`** - Publish a message via REST API

#### System Information

- **`GET /health`** - System health status
- **`GET /stats`** - System statistics

## Usage Examples

### Starting the Server

```bash
# Using Go directly
go run main.go

# Using Docker
docker build -t pub-sub .
docker run -p 8080:8080 pub-sub

# With custom configuration
MAX_MESSAGES_PER_TOPIC=500 PORT=9000 go run main.go
```

### WebSocket Client Example (JavaScript)

```javascript
const ws = new WebSocket('ws://localhost:8080/ws');

// Subscribe to a topic
ws.send(JSON.stringify({
  type: 'subscribe',
  topic: 'news'
}));

// Publish a message
ws.send(JSON.stringify({
  type: 'publish',
  topic: 'news',
  data: 'Hello World!'
}));

// Handle incoming messages
ws.onmessage = function(event) {
  const message = JSON.parse(event.data);
  console.log('Received:', message);
};
```

### REST API Examples

```bash
# Create a topic
curl -X POST http://localhost:8080/topics \
  -H "Content-Type: application/json" \
  -d '{"name": "news"}'

# List all topics
curl http://localhost:8080/topics

# Publish a message
curl -X POST http://localhost:8080/publish \
  -H "Content-Type: application/json" \
  -d '{"topic": "news", "data": "Breaking news!"}'

# Get system stats
curl http://localhost:8080/stats

# Health check
curl http://localhost:8080/health
```

## Development

### Prerequisites

- Go 1.21 or later
- Docker (optional)

### Building

```bash
# Download dependencies
go mod tidy

# Build the application
go build -o pub-sub .

# Run tests (if any)
go test ./...
```

### Project Structure

```
.
├── config/          # Configuration management
├── handlers/        # HTTP and WebSocket handlers
├── models/          # Data models and structures
├── pubsub/          # Core pub-sub logic
├── main.go          # Main application entry point
├── go.mod           # Go module file
├── Dockerfile       # Docker configuration
└── README.md        # This file
```

## Performance Considerations

- **Memory usage**: Limited by `MAX_MESSAGES_PER_TOPIC` setting
- **Concurrent connections**: Each WebSocket connection runs in its own goroutine
- **Message delivery**: Non-blocking message delivery with backpressure handling
- **Topic operations**: O(1) topic lookup, O(n) subscriber notification

## Limitations

- **No persistence**: All data is lost on server restart
- **Single instance**: No clustering or replication support
- **Memory constraints**: Limited by available system memory
- **No authentication**: No built-in security features

## Future Enhancements

- **Message persistence** to disk
- **Authentication and authorization**
- **Message filtering and routing**
- **Metrics and monitoring**
- **Horizontal scaling support**
- **Message acknowledgment and delivery guarantees**

## License

This project is open source and available under the MIT License.
