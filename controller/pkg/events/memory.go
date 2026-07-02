package events

import (
	"context"
	"sync"
)

// MemoryBus is an in-process event bus for tests and single-controller deployments.
type MemoryBus struct {
	mu       sync.RWMutex
	handlers []Handler
	closed   bool
}

func NewMemoryBus() *MemoryBus {
	return &MemoryBus{}
}

func (m *MemoryBus) Publish(ctx context.Context, evt SnapshotEvent) error {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.closed {
		return context.Canceled
	}
	for _, h := range m.handlers {
		if err := h(ctx, evt); err != nil {
			return err
		}
	}
	return nil
}

func (m *MemoryBus) Subscribe(_ context.Context, handler Handler) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.closed {
		return context.Canceled
	}
	m.handlers = append(m.handlers, handler)
	return nil
}

func (m *MemoryBus) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.closed = true
	m.handlers = nil
	return nil
}
