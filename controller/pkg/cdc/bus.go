package cdc

import (
	"context"
	"encoding/json"
	"sync"
)

// Publisher publishes Debezium CDC envelopes.
type Publisher interface {
	Publish(ctx context.Context, snapshotID string, env Envelope) error
	PublishBatch(ctx context.Context, snapshotID string, envs []Envelope) error
	Close() error
}

// Consumer reads CDC envelopes for a snapshot ID.
type Consumer interface {
	Consume(ctx context.Context, snapshotID string) ([]Envelope, error)
	Close() error
}

// MemoryBus is an in-process CDC bus for dev and tests.
type MemoryBus struct {
	mu     sync.RWMutex
	events map[string][]Envelope
	closed bool
}

func NewMemoryBus() *MemoryBus {
	return &MemoryBus{events: make(map[string][]Envelope)}
}

var defaultMemory = NewMemoryBus()

// DefaultMemory returns the shared in-process CDC bus.
func DefaultMemory() *MemoryBus {
	return defaultMemory
}

func (m *MemoryBus) Publish(_ context.Context, snapshotID string, env Envelope) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.closed {
		return context.Canceled
	}
	m.events[snapshotID] = append(m.events[snapshotID], env)
	return nil
}

func (m *MemoryBus) PublishBatch(ctx context.Context, snapshotID string, envs []Envelope) error {
	for _, env := range envs {
		if err := m.Publish(ctx, snapshotID, env); err != nil {
			return err
		}
	}
	return nil
}

func (m *MemoryBus) Consume(_ context.Context, snapshotID string) ([]Envelope, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.closed {
		return nil, context.Canceled
	}
	out := append([]Envelope(nil), m.events[snapshotID]...)
	return out, nil
}

func (m *MemoryBus) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.closed = true
	m.events = nil
	return nil
}

// Reset clears buffered events (for tests).
func (m *MemoryBus) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.events = make(map[string][]Envelope)
	m.closed = false
}

// NoopPublisher discards CDC events.
type NoopPublisher struct{}

func (NoopPublisher) Publish(context.Context, string, Envelope) error        { return nil }
func (NoopPublisher) PublishBatch(context.Context, string, []Envelope) error { return nil }
func (NoopPublisher) Close() error                                           { return nil }

// MemoryPublisher wraps MemoryBus as a Publisher.
type MemoryPublisher struct{ Bus *MemoryBus }

func (p MemoryPublisher) Publish(ctx context.Context, snapshotID string, env Envelope) error {
	if p.Bus == nil {
		return nil
	}
	return p.Bus.Publish(ctx, snapshotID, env)
}

func (p MemoryPublisher) PublishBatch(ctx context.Context, snapshotID string, envs []Envelope) error {
	if p.Bus == nil {
		return nil
	}
	return p.Bus.PublishBatch(ctx, snapshotID, envs)
}

func (p MemoryPublisher) Close() error { return nil }

// MemoryConsumer wraps MemoryBus as a Consumer.
type MemoryConsumer struct{ Bus *MemoryBus }

func (c MemoryConsumer) Consume(ctx context.Context, snapshotID string) ([]Envelope, error) {
	if c.Bus == nil {
		return nil, nil
	}
	return c.Bus.Consume(ctx, snapshotID)
}

func (c MemoryConsumer) Close() error { return nil }

// MarshalEnvelope serializes an envelope to JSON bytes.
func MarshalEnvelope(env Envelope) ([]byte, error) {
	return json.Marshal(env)
}

// UnmarshalEnvelope parses JSON bytes into an envelope.
func UnmarshalEnvelope(data []byte) (Envelope, error) {
	var env Envelope
	err := json.Unmarshal(data, &env)
	return env, err
}
