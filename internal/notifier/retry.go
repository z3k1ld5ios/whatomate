package notifier

import (
	"fmt"
	"time"
)

// RetryConfig defines retry behaviour for Send operations.
type RetryConfig struct {
	MaxAttempts int
	Delay       time.Duration
}

// DefaultRetryConfig provides sensible defaults: 3 attempts with a 500ms delay.
var DefaultRetryConfig = RetryConfig{
	MaxAttempts: 3,
	Delay:       500 * time.Millisecond,
}

// SendWithRetry calls Send up to MaxAttempts times, waiting Delay between each.
func (n *Notifier) SendWithRetry(destURL string, p Payload, cfg RetryConfig) error {
	if cfg.MaxAttempts <= 0 {
		cfg.MaxAttempts = 1
	}

	var lastErr error
	for attempt := 1; attempt <= cfg.MaxAttempts; attempt++ {
		if err := n.Send(destURL, p); err == nil {
			return nil
		} else {
			lastErr = err
		}
		if attempt < cfg.MaxAttempts && cfg.Delay > 0 {
			time.Sleep(cfg.Delay)
		}
	}
	return fmt.Errorf("notifier: all %d attempts failed: %w", cfg.MaxAttempts, lastErr)
}
