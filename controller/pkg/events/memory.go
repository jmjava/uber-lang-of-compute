package events

import (
	"context"
	"sync"
)

// MemoryBus is an in-process event bus for tests and single-controller deployments.
type MemoryBus struct {
	mu        sync.RWMutex
	handlers  []Handler
	seen      map[string]struct{}
	published []SnapshotEvent
	closed    bool
}

func NewMemoryBus() *MemoryBus {
	return &MemoryBus{seen: make(map[string]struct{})}
}

func (m *MemoryBus) Publish(ctx context.Context, evt SnapshotEvent) error {
	m.mu.Lock()
	if m.closed {
		m.mu.Unlock()
		return context.Canceled
	}
	if evt.EventID != "" {
		if _, ok := m.seen[evt.EventID]; ok {
			m.mu.Unlock()
			return nil
		}
		m.seen[evt.EventID] = struct{}{}
	}
	m.published = append(m.published, evt)
	handlers := append([]Handler(nil), m.handlers...)
	m.mu.Unlock()

	for _, h := range handlers {
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

func (m *MemoryBus) Published() []SnapshotEvent {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return append([]SnapshotEvent(nil), m.published...)
}

func (m *MemoryBus) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.closed = true
	m.handlers = nil
	return nil
}
