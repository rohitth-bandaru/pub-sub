# WebSocket Protocol Specification

This document describes the exact WebSocket protocol implemented by the Pub/Sub system.

## Client → Server Messages

### Message Format
```json
{
  "type": "subscribe" | "unsubscribe" | "publish" | "ping",
  "topic": "orders",           // required for subscribe/unsubscribe/publish
  "message": {                 // required for publish
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "payload": "..."
  },
  "client_id": "s1",          // required for subscribe/unsubscribe
  "last_n": 0,                // optional: number of historical messages to replay
  "request_id": "uuid-optional" // optional: correlation id
}
```

### Examples

#### Subscribe
```json
{
  "type": "subscribe",
  "topic": "orders",
  "client_id": "s1",
  "last_n": 5,
  "request_id": "550e8400-e29b-41d4-a716-446655440000"
}
```

#### Unsubscribe
```json
{
  "type": "unsubscribe",
  "topic": "orders",
  "client_id": "s1",
  "request_id": "340e8400-e29b-41d4-a716-4466554480098"
}
```

#### Publish
```json
{
  "type": "publish",
  "topic": "orders",
  "message": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "payload": {
      "order_id": "ORD-123",
      "amount": "99.5",
      "currency": "USD"
    }
  },
  "request_id": "340e8400-e29b-41d4-a716-4466554480098"
}
```

#### Ping
```json
{
  "type": "ping",
  "request_id": "570t8400-e29b-41d4-a716-4466554412345"
}
```

## Server → Client Messages

### Message Format
```json
{
  "type": "ack" | "event" | "error" | "pong" | "info",
  "request_id": "uuid-optional", // echoed if provided
  "topic": "orders",
  "message": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "payload": "..."
  },
  "error": {
    "code": "BAD_REQUEST",
    "message": "..."
  },
  "status": "ok",              // for ack messages
  "msg": "...",                // for info messages
  "ts": "2025-08-25T10:00:00Z" // optional server timestamp
}
```

### Examples

#### Ack (confirms successful operations)
```json
{
  "type": "ack",
  "request_id": "550e8400-e29b-41d4-a716-446655440000",
  "topic": "orders",
  "status": "ok",
  "ts": "2025-08-25T10:00:00Z"
}
```

#### Event (published message delivered to subscriber)
```json
{
  "type": "event",
  "topic": "orders",
  "message": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "payload": {
      "order_id": "ORD-123",
      "amount": 99.5,
      "currency": "USD"
    }
  },
  "ts": "2025-08-25T10:01:00Z"
}
```

#### Error (validation or flow errors)
```json
{
  "type": "error",
  "request_id": "req-67890",
  "error": {
    "code": "BAD_REQUEST",
    "message": "message.id must be a valid UUID"
  },
  "ts": "2025-08-25T10:02:00Z"
}
```

#### Pong (response to client ping)
```json
{
  "type": "pong",
  "request_id": "ping-abc",
  "ts": "2025-08-25T10:03:00Z"
}
```

#### Info (server-initiated notices)

Heartbeat:
```json
{
  "type": "info",
  "msg": "ping",
  "ts": "2025-08-25T10:04:00Z"
}
```

Topic deleted:
```json
{
  "type": "info",
  "topic": "orders",
  "msg": "topic_deleted",
  "ts": "2025-08-25T10:05:00Z"
}
```

## Error Codes

- **BAD_REQUEST**: Invalid message format or missing required fields
- **TOPIC_NOT_FOUND**: Publish/subscribe to non-existent topic
- **SLOW_CONSUMER**: Subscriber queue overflow
- **UNAUTHORIZED**: Invalid/missing auth (if implemented)
- **INTERNAL**: Unexpected server error

## HTTP REST Endpoints

### POST /topics
**Request:**
```json
{
  "name": "orders"
}
```

**Response:**
- **201 Created** → `{ "status": "created", "topic": "orders" }`
- **409 Conflict** if already exists

### DELETE /topics/{name}
**Response:**
- **200 OK** → `{ "status": "deleted", "topic": "orders" }`
- **404** if not found

### GET /topics
**Response:**
```json
{
  "topics": [
    {
      "name": "orders",
      "subscribers": 3
    }
  ]
}
```

### GET /health
**Response:**
```json
{
  "uptime_sec": 123,
  "topics": 2,
  "subscribers": 4
}
```

### GET /stats
**Response:**
```json
{
  "topics": {
    "orders": {
      "messages": 42,
      "subscribers": 3
    }
  }
}
```

### POST /publish
**Request:**
```json
{
  "topic": "orders",
  "message": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "payload": {
      "order_id": "ORD-123",
      "amount": 99.5,
      "currency": "USD"
    }
  }
}
```

**Response:**
- **200 OK** → `{ "status": "published", "topic": "orders" }`
- **404** if topic not found

## Implementation Notes

- **Message Replay**: The `last_n` parameter in subscribe requests enables historical message replay
- **Backpressure Handling**: When subscriber queues overflow, the system sends `SLOW_CONSUMER` errors
- **Graceful Shutdown**: Server stops accepting new operations, flushes existing messages, and closes sockets cleanly
- **Concurrency Safety**: All operations are thread-safe using read-write mutexes
- **Circular Buffer**: Messages per topic are limited to prevent memory issues
