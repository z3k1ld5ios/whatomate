package notifier_test

import (
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/shridarpatil/whatomate/internal/notifier"
)

func TestSendWithRetry_SucceedsOnSecondAttempt(t *testing.T) {
	var calls int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if atomic.AddInt32(&calls, 1) < 2 {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	n := notifier.New(notifier.Config{})
	err := n.SendWithRetry(server.URL, notifier.Payload{Event: "retry"},
		notifier.RetryConfig{MaxAttempts: 3, Delay: time.Millisecond})
	if err != nil {
		t.Fatalf("expected success after retry, got: %v", err)
	}
	if atomic.LoadInt32(&calls) != 2 {
		t.Errorf("expected 2 calls, got %d", calls)
	}
}

func TestSendWithRetry_ExhaustsAttempts(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
	}))
	defer server.Close()

	n := notifier.New(notifier.Config{})
	err := n.SendWithRetry(server.URL, notifier.Payload{Event: "fail"},
		notifier.RetryConfig{MaxAttempts: 2, Delay: time.Millisecond})
	if err == nil {
		t.Fatal("expected error after exhausting retries")
	}
}

func TestSendWithRetry_ZeroAttempts(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	n := notifier.New(notifier.Config{})
	err := n.SendWithRetry(server.URL, notifier.Payload{Event: "zero"},
		notifier.RetryConfig{MaxAttempts: 0})
	if err == nil {
		t.Fatal("expected error with zero max attempts")
	}
}
