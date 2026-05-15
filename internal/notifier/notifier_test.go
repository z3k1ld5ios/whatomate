package notifier_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/shridarpatil/whatomate/internal/notifier"
)

func TestNew_DefaultTimeout(t *testing.T) {
	n := notifier.New(notifier.Config{})
	if n == nil {
		t.Fatal("expected non-nil notifier")
	}
}

func TestSend_Success(t *testing.T) {
	var received notifier.Payload
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&received); err != nil {
			t.Errorf("decode body: %v", err)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	n := notifier.New(notifier.Config{Timeout: 5 * time.Second})
	p := notifier.Payload{Event: "test.event", Message: "hello"}

	if err := n.Send(server.URL, p); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if received.Event != "test.event" {
		t.Errorf("expected event %q, got %q", "test.event", received.Event)
	}
}

func TestSend_NonSuccessStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	n := notifier.New(notifier.Config{})
	err := n.Send(server.URL, notifier.Payload{Event: "fail"})
	if err == nil {
		t.Fatal("expected error for non-2xx status")
	}
}

func TestSend_InvalidURL(t *testing.T) {
	n := notifier.New(notifier.Config{})
	err := n.Send("http://127.0.0.1:0", notifier.Payload{Event: "fail"})
	if err == nil {
		t.Fatal("expected error for unreachable URL")
	}
}
