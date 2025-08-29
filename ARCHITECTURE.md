# Pub/Sub System Architecture

This document describes the refactored architecture of the Pub/Sub system, which has been designed with clean coding standards and modularity in mind.

## Project Structure

```
pub-sub/
├── config/          # Configuration management
├── handlers/        # HTTP request handlers
├── logger/          # Logging abstraction layer
├── middleware/      # HTTP middleware
├── models/          # Data models and structures
├── pubsub/          # Core pub/sub business logic
├── server/          # HTTP server management
├── services/        # Business logic services
├── utils/           # Utility functions
├── main.go          # Application entry point
└── README.md        # Project documentation
```

## Architecture Layers

### 1. Presentation Layer
- **`handlers/`**: HTTP request handlers for REST API and WebSocket endpoints
- **`middleware/`**: HTTP middleware for logging, CORS, etc.

### 2. Business Logic Layer
- **`services/`**: Business logic services that coordinate between handlers and core logic
  - `TopicService`: Manages topic operations
  - `MessageService`: Handles message publishing
  - `SystemService`: Provides system stats and health information

### 3. Core Logic Layer
- **`pubsub/`**: Core pub/sub system implementation
- **`models/`**: Data structures and models

### 4. Infrastructure Layer
- **`config/`**: Configuration management
- **`logger/`**: Logging abstraction (currently using logrus)
- **`server/`**: HTTP server setup and lifecycle management
- **`utils/`**: Utility functions

## Key Design Principles

### 1. Separation of Concerns
- Each package has a single responsibility
- Business logic is separated from HTTP handling
- Configuration is centralized

### 2. Dependency Injection
- Services are injected into handlers
- Logger is injected into all components
- Configuration is passed down the chain

### 3. Interface Abstraction
- Logger interface abstracts logging implementation
- Easy to switch logging libraries without code changes

### 4. Error Handling
- Consistent error handling across layers
- Proper HTTP status codes
- Structured logging for debugging

## Benefits of Refactoring

1. **Maintainability**: Code is easier to understand and modify
2. **Testability**: Each layer can be tested independently
3. **Scalability**: New features can be added without affecting existing code
4. **Flexibility**: Easy to swap implementations (e.g., different loggers)
5. **Readability**: Clear separation of responsibilities

## Usage Example

```go
// Initialize logger
log := logger.NewLogger(cfg.LogLevel, cfg.LogFormat)

// Initialize core system
pubSub := pubsub.NewPubSub(cfg, log)

// Initialize services
topicService := services.NewTopicService(pubSub, log)
messageService := services.NewMessageService(pubSub, log)
systemService := services.NewSystemService(pubSub, log)

// Initialize handlers with services
restHandler := handlers.NewRestHandler(topicService, messageService, systemService, log)

// Create server
server := server.NewServer(cfg, log, pubSub)
```

## Future Enhancements

1. **Database Layer**: Add persistent storage for topics and messages
2. **Authentication**: Implement user authentication and authorization
3. **Rate Limiting**: Add rate limiting middleware
4. **Metrics**: Integrate with monitoring systems
5. **Testing**: Add comprehensive unit and integration tests
