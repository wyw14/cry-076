package platform

import (
	"context"
	"sync"

	"github.com/wyw14/cry-076/internal/application"
)

type ResumeNotifier struct {
	mu       sync.Mutex
	Messages []application.Notification
}

func (n *ResumeNotifier) Notify(ctx context.Context, message application.Notification) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	n.mu.Lock()
	defer n.mu.Unlock()
	n.Messages = append(n.Messages, message)
	return nil
}

type Callback func(context.Context, map[string]any) error
type TemplateEventBus struct {
	mu       sync.RWMutex
	handlers map[string][]Callback
}

func NewTemplateEventBus() *TemplateEventBus {
	return &TemplateEventBus{handlers: map[string][]Callback{}}
}
func (c *TemplateEventBus) Subscribe(topic string, handler Callback) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.handlers[topic] = append(c.handlers[topic], handler)
}
func (c *TemplateEventBus) Publish(ctx context.Context, topic string, payload map[string]any) error {
	c.mu.RLock()
	handlers := append([]Callback(nil), c.handlers[topic]...)
	c.mu.RUnlock()
	for _, handler := range handlers {
		if err := handler(ctx, payload); err != nil {
			return err
		}
	}
	return nil
}
