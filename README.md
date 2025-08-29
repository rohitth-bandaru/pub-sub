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

The system can be configured using environment variables. Configuration is loaded only once when the system starts from either a `.env` file or system environment variables.

## Logging

The system uses [logrus](https://github.com/sirupsen/logrus) for comprehensive, structured logging with multiple output formats and log levels.

### Log Levels

- **`DEBUG`**: Detailed debugging information including function calls and file locations
- **`INFO`**: General operational information about system events
- **`WARN`**: Warning messages for potentially harmful situations
- **`ERROR`**: Error messages for failed operations
- **`FATAL`**: Critical errors that cause the system to exit

### Log Formats

- **`text`**: Human-readable text format with colors and timestamps
- **`json`**: Structured JSON format for log aggregation systems

### Structured Logging

All log messages include structured fields for better debugging and monitoring:

```go
logrus.WithFields(logrus.Fields{
    "component": "websocket",
    "action": "connect",
    "client_id": "client-123",
    "remote_addr": "192.168.1.100",
    "user_agent": "Mozilla/5.0...",
}).Info("WebSocket client connected successfully")
```

### Logging Examples

#### WebSocket Operations
```
INFO[2025-08-29T10:52:28+05:30] WebSocket client connected successfully  client_id=client-123 action=connect remote_addr=192.168.1.100 user_agent=Mozilla/5.0...
```

#### Pub/Sub Operations
```
INFO[2025-08-29T10:52:29+05:30] Topic created successfully  topic=orders action=create
INFO[2025-08-29T10:52:30+05:30] Message published successfully  topic=orders message_id=msg-456 action=publish subscribers_count=3
INFO[2025-08-29T10:52:31+05:30] Subscriber subscribed successfully  subscriber_id=sub-789 topic=orders action=subscribe historical_messages=5 total_subscribers=4
```

#### Error Handling
```
WARN[2025-08-29T10:52:32+05:30] Subscriber disconnected due to channel overflow  subscriber_id=sub-789 topic=orders action=disconnect reason=channel_overflow
ERROR[2025-08-29T10:52:33+05:30] WebSocket upgrade failed  error=invalid upgrade request remote_addr=192.168.1.100 user_agent=curl/7.68.0
```

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `8080` | Server port |
| `HOST` | `0.0.0.0` | Server host |
| `MAX_MESSAGES_PER_TOPIC` | `1000` | Maximum messages stored per topic |
| `MAX_PUBLISH_RATE` | `100` | Maximum publish rate per topic (messages/sec) |
| `WS_READ_BUFFER_SIZE` | `1024` | WebSocket read buffer size |
| `WS_WRITE_BUFFER_SIZE` | `1024` | WebSocket write buffer size |
| `LOG_LEVEL` | `info` | Logging level (debug, info, warn, error, fatal) |
| `LOG_FORMAT` | `text` | Logging format (text, json) |

### Configuration Files

- **`env`** - Contains the actual configuration values (included in Docker builds)
- **`env.example`** - Template showing all available configuration options

### Usage

```bash
# Use default configuration
./pub-sub

# Override specific values
PORT=9000 MAX_MESSAGES_PER_TOPIC=500 ./pub-sub

# Use custom .env file
cp env.example .env
# Edit .env file with your values
./pub-sub
```

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

### Debugging

The system provides comprehensive logging for debugging purposes:

#### Enable Debug Logging

```bash
# Set debug level in .env file
LOG_LEVEL=debug

# Or override at runtime
LOG_LEVEL=debug ./pub-sub
```

#### Debug Information Available

- **Function calls**: File names, line numbers, and function names for debug-level logs
- **Structured data**: All operations include relevant context (IDs, counts, timestamps)
- **Performance metrics**: Request durations, subscriber counts, message counts
- **Error details**: Full error context with stack traces for debugging

#### Log Analysis

```bash
# Filter logs by component
grep "component=websocket" logs.txt

# Filter logs by action
grep "action=connect" logs.txt

# Filter logs by client ID
grep "client_id=client-123" logs.txt

# JSON format for log aggregation
LOG_FORMAT=json ./pub-sub | jq '.component, .action, .client_id'
```

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
