package server

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"pub-sub/config"
	"pub-sub/handlers"
	"pub-sub/logger"
	"pub-sub/middleware"
	"pub-sub/pubsub"
	"pub-sub/services"

	"github.com/gorilla/mux"
)

// Server represents the HTTP server
type Server struct {
	config     *config.Config
	logger     logger.Logger
	pubSub     *pubsub.PubSub
	httpServer *http.Server
	router     *mux.Router
}

// NewServer creates a new server instance
func NewServer(cfg *config.Config, log logger.Logger, pubSub *pubsub.PubSub) *Server {
	server := &Server{
		config: cfg,
		logger: log,
		pubSub: pubSub,
	}

	server.setupRouter()
	server.setupHTTPServer()

	return server
}

// setupRouter configures the router with all endpoints and middleware
func (s *Server) setupRouter() {
	s.router = mux.NewRouter()

	// Initialize services
	topicService := services.NewTopicService(s.pubSub, s.logger)
	messageService := services.NewMessageService(s.pubSub, s.logger)
	systemService := services.NewSystemService(s.pubSub, s.logger)

	// Initialize handlers
	wsHandler := handlers.NewWebSocketHandler(s.pubSub, s.config, s.logger)
	restHandler := handlers.NewRestHandler(topicService, messageService, systemService, s.logger)

	// WebSocket endpoint
	s.router.HandleFunc("/ws", wsHandler.HandleWebSocket)

	// REST API endpoints
	s.router.HandleFunc("/topics", restHandler.CreateTopic).Methods("POST")
	s.router.HandleFunc("/topics", restHandler.ListTopics).Methods("GET")
	s.router.HandleFunc("/topics/{name}", restHandler.GetTopic).Methods("GET")
	s.router.HandleFunc("/topics/{name}", restHandler.DeleteTopic).Methods("DELETE")
	s.router.HandleFunc("/publish", restHandler.PublishMessage).Methods("POST")
	s.router.HandleFunc("/stats", restHandler.GetStats).Methods("GET")
	s.router.HandleFunc("/health", restHandler.GetHealth).Methods("GET")

	// Add middleware
	s.router.Use(middleware.LoggingMiddleware(s.logger))
	s.router.Use(middleware.CORSMiddleware())
}

// setupHTTPServer configures the HTTP server
func (s *Server) setupHTTPServer() {
	s.httpServer = &http.Server{
		Addr:         fmt.Sprintf("%s:%s", s.config.Host, s.config.Port),
		Handler:      s.router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
}

// Start starts the server in a goroutine
func (s *Server) Start() error {
	s.logger.Infof("Starting Pub/Sub server on %s:%s", s.config.Host, s.config.Port)
	s.logger.Infof("Configuration: MaxMessagesPerTopic=%d, MaxPublishRate=%d",
		s.config.MaxMessagesPerTopic, s.config.MaxPublishRate)

	go func() {
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			s.logger.Fatalf("Server failed to start: %v", err)
		}
	}()

	return nil
}

// Shutdown gracefully shuts down the server
func (s *Server) Shutdown(ctx context.Context) error {
	s.logger.Info("Shutting down server...")

	if err := s.httpServer.Shutdown(ctx); err != nil {
		s.logger.Errorf("Server forced to shutdown: %v", err)
		return err
	}

	s.logger.Info("Server exited")
	return nil
}

// GetAddr returns the server address
func (s *Server) GetAddr() string {
	return s.httpServer.Addr
}
