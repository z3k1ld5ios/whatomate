package notifier

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// Notifier sends webhook payloads to configured destinations.
type Notifier struct {
	client  *http.Client
	timeout time.Duration
}

// Config holds configuration for the Notifier.
type Config struct {
	Timeout time.Duration
}

// Payload represents the data sent to a destination URL.
type Payload struct {
	Event   string            `json:"event"`
	Message string            `json:"message"`
	Meta    map[string]string `json:"meta,omitempty"`
}

// New creates a new Notifier with the given config.
func New(cfg Config) *Notifier {
	if cfg.Timeout == 0 {
		cfg.Timeout = 10 * time.Second
	}
	return &Notifier{
		client:  &http.Client{Timeout: cfg.Timeout},
		timeout: cfg.Timeout,
	}
}

// Send dispatches a Payload to the given destination URL.
func (n *Notifier) Send(destURL string, p Payload) error {
	body, err := json.Marshal(p)
	if err != nil {
		return fmt.Errorf("notifier: marshal payload: %w", err)
	}

	resp, err := n.client.Post(destURL, "application/json", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("notifier: post to %s: %w", destURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("notifier: unexpected status %d from %s", resp.StatusCode, destURL)
	}
	return nil
}
