package handlers

import (
	"encoding/json"
	"net/http"
	"pub-sub/logger"
	"pub-sub/models"
	"pub-sub/services"

	"github.com/gorilla/mux"
)

// RestHandler handles HTTP REST API endpoints
type RestHandler struct {
	topicService   *services.TopicService
	messageService *services.MessageService
	systemService  *services.SystemService
	logger         logger.Logger
}

// NewRestHandler creates a new REST handler
func NewRestHandler(topicService *services.TopicService, messageService *services.MessageService, systemService *services.SystemService, log logger.Logger) *RestHandler {
	return &RestHandler{
		topicService:   topicService,
		messageService: messageService,
		systemService:  systemService,
		logger:         log,
	}
}

// CreateTopic handles POST /topics endpoint
func (h *RestHandler) CreateTopic(w http.ResponseWriter, r *http.Request) {
	var request struct {
		Name string `json:"name"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		h.logger.Warnf("Invalid request body: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	response, err := h.topicService.CreateTopic(request.Name)
	if err != nil {
		h.logger.Errorf("Failed to create topic: %v", err)
		http.Error(w, err.Error(), http.StatusConflict)
		return
	}

	h.sendJSONResponse(w, http.StatusCreated, response)
}

// DeleteTopic handles DELETE /topics/{name} endpoint
func (h *RestHandler) DeleteTopic(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	topicName := vars["name"]

	response, err := h.topicService.DeleteTopic(topicName)
	if err != nil {
		h.logger.Errorf("Failed to delete topic: %v", err)
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	h.sendJSONResponse(w, http.StatusOK, response)
}

// ListTopics handles GET /topics endpoint
func (h *RestHandler) ListTopics(w http.ResponseWriter, r *http.Request) {
	response := h.topicService.ListTopics()
	h.sendJSONResponse(w, http.StatusOK, response)
}

// GetTopic handles GET /topics/{name} endpoint
func (h *RestHandler) GetTopic(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	topicName := vars["name"]

	topic, err := h.topicService.GetTopic(topicName)
	if err != nil {
		h.logger.Errorf("Failed to get topic: %v", err)
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	h.sendJSONResponse(w, http.StatusOK, topic)
}

// GetStats handles GET /stats endpoint
func (h *RestHandler) GetStats(w http.ResponseWriter, r *http.Request) {
	response := h.systemService.GetStats()
	h.sendJSONResponse(w, http.StatusOK, response)
}

// GetHealth handles GET /health endpoint
func (h *RestHandler) GetHealth(w http.ResponseWriter, r *http.Request) {
	response := h.systemService.GetHealth()
	h.sendJSONResponse(w, http.StatusOK, response)
}

// PublishMessage handles POST /publish endpoint
func (h *RestHandler) PublishMessage(w http.ResponseWriter, r *http.Request) {
	var request struct {
		Topic   string          `json:"topic"`
		Message *models.Message `json:"message"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		h.logger.Warnf("Invalid request body: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	response, err := h.messageService.PublishMessage(request.Topic, request.Message)
	if err != nil {
		h.logger.Errorf("Failed to publish message: %v", err)
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	h.sendJSONResponse(w, http.StatusOK, response)
}

// sendJSONResponse sends a JSON response with proper headers
func (h *RestHandler) sendJSONResponse(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		h.logger.Errorf("Failed to encode JSON response: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}
