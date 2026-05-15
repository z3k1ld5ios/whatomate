// Package notifier provides functionality for dispatching webhook payloads
// to external HTTP destinations.
//
// Basic usage:
//
//	n := notifier.New(notifier.Config{Timeout: 10 * time.Second})
//	err := n.Send("https://example.com/hook", notifier.Payload{
//		Event:   "order.created",
//		Message: "A new order has been placed.",
//		Meta:    map[string]string{"order_id": "42"},
//	})
//
// For resilient delivery use SendWithRetry:
//
//	err := n.SendWithRetry(destURL, payload, notifier.RetryConfig{
//		MaxAttempts: 3,
//		Delay:       2 * time.Second,
//	})
package notifier
