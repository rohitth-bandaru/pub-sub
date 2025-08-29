# Quick Start Guide

Get the Pub/Sub system up and running in minutes!

## Prerequisites

- Go 1.21 or later
- Docker (optional)

## Quick Start

### 1. Clone and Build

```bash
# Navigate to the project directory
cd pub-sub

# Install dependencies
go mod tidy

# Build the application
go build -o pub-sub .
```

### 2. Run the Server

```bash
# Start the server
./pub-sub
```

The server will start on `localhost:8080` with default configuration.

### 3. Test the System

#### Option A: Use the Demo Script
```bash
# In another terminal, run the demo
./examples/demo.sh
```

#### Option B: Use the Test Client
1. Open `test-client.html` in your browser
2. Click "Connect" to establish WebSocket connection
3. Create topics, subscribe, and publish messages

#### Option C: Use curl Commands
```bash
# Create a topic
curl -X POST http://localhost:8080/topics \
  -H "Content-Type: application/json" \
  -d '{"name": "orders"}'

# Publish a message
curl -X POST http://localhost:8080/publish \
  -H "Content-Type: application/json" \
  -d '{
    "topic": "orders",
    "message": {
      "id": "order-123",
      "payload": {"order_id": "ORD-123", "amount": 99.50}
    }
  }'

# List topics
curl http://localhost:8080/topics

# Check health
curl http://localhost:8080/health
```

## Configuration

The system loads configuration from environment variables or a `.env` file. Configuration is loaded only once when the system starts.

### Option 1: Environment Variables

```bash
# Set custom port and message limits
export PORT=9000
export MAX_MESSAGES_PER_TOPIC=500
export MAX_PUBLISH_RATE=50

# Run with custom config
./pub-sub
```

### Option 2: .env File

```bash
# Copy the example file
cp env.example .env

# Edit .env file with your values
nano .env

# Run the system
./pub-sub
```

### Option 3: Inline Override

```bash
# Override specific values inline
PORT=9000 MAX_MESSAGES_PER_TOPIC=500 ./pub-sub
```

## Docker

```bash
# Build Docker image
docker build -t pub-sub .

# Run container
docker run -p 8080:8080 pub-sub
```

## WebSocket Testing

Use the test client or connect directly:

```javascript
const ws = new WebSocket('ws://localhost:8080/ws');

// Subscribe to a topic
ws.send(JSON.stringify({
  type: 'subscribe',
  topic: 'orders',
  client_id: 'test-client',
  last_n: 5,
  request_id: 'req-123'
}));

// Publish a message
ws.send(JSON.stringify({
  type: 'publish',
  topic: 'orders',
  message: {
    id: 'msg-456',
    payload: { text: 'Hello World!' }
  },
  request_id: 'req-456'
}));
```

## What's Next?

- Read the [README.md](README.md) for detailed documentation
- Check [PROTOCOL.md](PROTOCOL.md) for WebSocket protocol details
- Run tests with `go test ./...`
- Explore the code structure in the `pubsub/`, `handlers/`, and `models/` directories

## Troubleshooting

- **Port already in use**: Change the port with `PORT=9000 ./pub-sub`
- **WebSocket connection fails**: Ensure the server is running and check the URL
- **Build errors**: Run `go mod tidy` to resolve dependencies
- **Permission denied**: Make sure the binary is executable with `chmod +x pub-sub`

## Support

If you encounter issues:
1. Check the server logs for error messages
2. Verify your WebSocket client implementation matches the protocol
3. Ensure all required fields are present in your messages
4. Check that topics exist before publishing or subscribing
