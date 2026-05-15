package webhook

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewHandler(t *testing.T) {
	h := NewHandler("test-secret")
	if h == nil {
		t.Fatal("expected non-nil handler")
	}
	if h.Secret != "test-secret" {
		t.Errorf("expected secret 'test-secret', got '%s'", h.Secret)
	}
}

func TestServeHTTP_MethodNotAllowed(t *testing.T) {
	h := NewHandler("")
	req := httptest.NewRequest(http.MethodGet, "/webhook", nil)
	rw := httptest.NewRecorder()
	h.ServeHTTP(rw, req)
	if rw.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", rw.Code)
	}
}

func TestServeHTTP_InvalidJSON(t *testing.T) {
	h := NewHandler("")
	req := httptest.NewRequest(http.MethodPost, "/webhook", bytes.NewBufferString("not-json"))
	rw := httptest.NewRecorder()
	h.ServeHTTP(rw, req)
	if rw.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rw.Code)
	}
}

func TestServeHTTP_MissingFields(t *testing.T) {
	h := NewHandler("")
	body, _ := json.Marshal(Payload{Event: "message"})
	req := httptest.NewRequest(http.MethodPost, "/webhook", bytes.NewBuffer(body))
	rw := httptest.NewRecorder()
	h.ServeHTTP(rw, req)
	if rw.Code != http.StatusUnprocessableEntity {
		t.Errorf("expected 422, got %d", rw.Code)
	}
}

func TestServeHTTP_ValidPayload(t *testing.T) {
	h := NewHandler("secret")
	body, _ := json.Marshal(Payload{
		Event:   "message",
		ChatID:  "chat-123",
		Message: "Hello, World!",
	})
	req := httptest.NewRequest(http.MethodPost, "/webhook", bytes.NewBuffer(body))
	rw := httptest.NewRecorder()
	h.ServeHTTP(rw, req)
	if rw.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rw.Code)
	}
	var resp Response
	if err := json.NewDecoder(rw.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp.Status != "ok" {
		t.Errorf("expected status 'ok', got '%s'", resp.Status)
	}
}
