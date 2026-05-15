package webhook

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
)

// Payload represents an incoming webhook payload
type Payload struct {
	Event   string                 `json:"event"`
	Data    map[string]interface{} `json:"data"`
	ChatID  string                 `json:"chat_id"`
	Message string                 `json:"message"`
}

// Response represents the webhook handler response
type Response struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

// Handler processes incoming webhook requests
type Handler struct {
	Secret string
}

// NewHandler creates a new webhook Handler
func NewHandler(secret string) *Handler {
	return &Handler{Secret: secret}
}

// ServeHTTP handles the incoming HTTP webhook request
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("webhook: failed to read body: %v", err)
		http.Error(w, "failed to read request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var payload Payload
	if err := json.Unmarshal(body, &payload); err != nil {
		log.Printf("webhook: failed to parse payload: %v", err)
		http.Error(w, "invalid JSON payload", http.StatusBadRequest)
		return
	}

	if payload.ChatID == "" || payload.Message == "" {
		http.Error(w, "chat_id and message are required", http.StatusUnprocessableEntity)
		return
	}

	log.Printf("webhook: received event=%s chat_id=%s", payload.Event, payload.ChatID)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(Response{
		Status:  "ok",
		Message: "webhook received",
	})
}
