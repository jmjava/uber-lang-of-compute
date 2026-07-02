package events

import "context"

// Handler processes snapshot events from the bus.
type Handler func(ctx context.Context, evt SnapshotEvent) error

// Bus publishes and subscribes to KBL compute fabric events.
type Bus interface {
	Publish(ctx context.Context, evt SnapshotEvent) error
	Subscribe(ctx context.Context, handler Handler) error
	Close() error
}

// NoopBus discards all events.
type NoopBus struct{}

func NewNoopBus() *NoopBus { return &NoopBus{} }

func (n *NoopBus) Publish(_ context.Context, _ SnapshotEvent) error { return nil }
func (n *NoopBus) Subscribe(_ context.Context, _ Handler) error   { return nil }
func (n *NoopBus) Close() error                                     { return nil }
