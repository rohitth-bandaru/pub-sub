package handlers

import (
	"encoding/json"
	"net/http"
	"pub-sub/logger"
	"pub-sub/models"
	"pub-sub/services"
	"time"

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
		h.sendErrorResponse(w, http.StatusBadRequest, "Invalid request body", "INVALID_JSON")
		return
	}

	response, err := h.topicService.CreateTopic(request.Name)
	if err != nil {
		h.logger.Errorf("Failed to create topic: %v", err)
		statusCode := http.StatusInternalServerError
		if models.IsErrorType(err, models.ErrTopicExists) {
			statusCode = http.StatusConflict
		}
		h.sendErrorResponse(w, statusCode, err.Error(), "TOPIC_CREATION_FAILED")
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
		statusCode := http.StatusInternalServerError
		if models.IsErrorType(err, models.ErrTopicNotFound) {
			statusCode = http.StatusNotFound
		}
		h.sendErrorResponse(w, statusCode, err.Error(), "TOPIC_DELETION_FAILED")
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
		statusCode := http.StatusInternalServerError
		if models.IsErrorType(err, models.ErrTopicNotFound) {
			statusCode = http.StatusNotFound
		}
		h.sendErrorResponse(w, statusCode, err.Error(), "TOPIC_NOT_FOUND")
		return
	}

	h.sendJSONResponse(w, http.StatusOK, topic)
}

// GetStats handles GET /stats endpoint
func (h *RestHandler) GetStats(w http.ResponseWriter, r *http.Request) {
	response := h.systemService.GetStats()
	h.sendJSONResponse(w, http.StatusOK, response)
}

// GetTopicStats handles GET /stats/{topic} endpoint
func (h *RestHandler) GetTopicStats(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	topicName := vars["topic"]

	stats, err := h.systemService.GetTopicStats(topicName)
	if err != nil {
		h.logger.Errorf("Failed to get topic stats: %v", err)
		statusCode := http.StatusInternalServerError
		if models.IsErrorType(err, models.ErrTopicNotFound) {
			statusCode = http.StatusNotFound
		}
		h.sendErrorResponse(w, statusCode, err.Error(), "STATS_RETRIEVAL_FAILED")
		return
	}

	h.sendJSONResponse(w, http.StatusOK, stats)
}

// GetActiveClients handles GET /clients endpoint
func (h *RestHandler) GetActiveClients(w http.ResponseWriter, r *http.Request) {
	response := h.systemService.GetActiveClients()
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
		h.sendErrorResponse(w, http.StatusBadRequest, "Invalid request body", "INVALID_JSON")
		return
	}

	response, err := h.messageService.PublishMessage(request.Topic, request.Message)
	if err != nil {
		h.logger.Errorf("Failed to publish message: %v", err)
		statusCode := http.StatusInternalServerError
		if models.IsErrorType(err, models.ErrTopicNotFound) {
			statusCode = http.StatusNotFound
		} else if models.IsErrorType(err, models.ErrTopicRequired) || 
		          models.IsErrorType(err, models.ErrMessageRequired) || 
		          models.IsErrorType(err, models.ErrMessageIDRequired) {
			statusCode = http.StatusBadRequest
		}
		h.sendErrorResponse(w, statusCode, err.Error(), "MESSAGE_PUBLISH_FAILED")
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
		h.sendErrorResponse(w, http.StatusInternalServerError, "Internal server error", "JSON_ENCODING_FAILED")
	}
}

// sendErrorResponse sends a structured error response
func (h *RestHandler) sendErrorResponse(w http.ResponseWriter, statusCode int, message, code string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	errorResponse := struct {
		Error struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
		Timestamp string `json:"timestamp"`
	}{
		Timestamp: time.Now().Format(time.RFC3339),
	}
	errorResponse.Error.Code = code
	errorResponse.Error.Message = message

	if err := json.NewEncoder(w).Encode(errorResponse); err != nil {
		h.logger.Errorf("Failed to encode error response: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

