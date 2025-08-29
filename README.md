# Go Pub/Sub System

A high-performance, race-condition-free pub/sub system built in Go with WebSocket support for real-time messaging and REST API for management operations.

## 🐳 Docker Quick Start

### Build and Run
```bash
# Build the Docker image
docker build -t pub-sub .

# Run the container
docker run -p 8080:8080 -e PORT=8080 pub-sub
```

### Environment Variables
```bash
docker run -p 8080:8080 \
  -e PORT=8080 \
  -e HOST=0.0.0.0 \
  -e LOG_LEVEL=info \
  -e MAX_MESSAGES_PER_TOPIC=500 \
  pub-sub
```

### Docker Compose
```yaml
version: '3.8'
services:
  pub-sub:
    build: .
    ports:
      - "8080:8080"
    environment:
      - PORT=8080
      - HOST=0.0.0.0
      - LOG_LEVEL=info
    restart: unless-stopped
```

Run with: `docker-compose up -d`

## 🚀 Features

- **Race-Condition Free**: Comprehensive mutex usage and proper goroutine management
- **Real-time Messaging**: WebSocket support for instant message delivery
- **REST API**: Full CRUD operations for topics and messages
- **Graceful Shutdown**: Proper resource cleanup and connection handling
- **Error Handling**: Structured error responses and comprehensive logging
- **High Performance**: Efficient in-memory storage with configurable limits

## 📡 API Endpoints

- `POST /topics` - Create topic
- `GET /topics` - List all topics
- `DELETE /topics/{name}` - Delete topic
- `POST /publish` - Publish message
- `GET /stats` - System statistics
- `GET /health` - Health check
- `GET /ws` - WebSocket endpoint

## 🔧 Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `8080` | Server port |
| `HOST` | `localhost` | Server host |
| `LOG_LEVEL` | `info` | Logging level |
| `MAX_MESSAGES_PER_TOPIC` | `100` | Max messages per topic |

## 🧪 Testing

```bash
# Run with race detector
go test -race ./...

# Build with race detector
go build -race .
```

## 📦 Dependencies

- Go 1.19+
- Gorilla WebSocket
- Gorilla Mux
- Structured logging
