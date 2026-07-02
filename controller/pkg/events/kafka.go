package events

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/segmentio/kafka-go"
)

// KafkaBus publishes and consumes snapshot events via Kafka (Debezium-compatible topic).
type KafkaBus struct {
	writer   *kafka.Writer
	reader   *kafka.Reader
	mu       sync.RWMutex
	handlers []Handler
	closed   bool
}

// KafkaConfig configures the Kafka event bus.
type KafkaConfig struct {
	Brokers []string
	Topic   string
	GroupID string
}

// NewKafkaBus creates a Kafka-backed event bus.
func NewKafkaBus(cfg KafkaConfig) (*KafkaBus, error) {
	if len(cfg.Brokers) == 0 {
		return nil, fmt.Errorf("kafka brokers required")
	}
	if cfg.Topic == "" {
		cfg.Topic = "kbl.snapshot.events"
	}
	if cfg.GroupID == "" {
		cfg.GroupID = "kbl-multiverse"
	}

	writer := &kafka.Writer{
		Addr:         kafka.TCP(cfg.Brokers...),
		Topic:        cfg.Topic,
		Balancer:     &kafka.LeastBytes{},
		RequiredAcks: kafka.RequireOne,
	}

	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  cfg.Brokers,
		Topic:    cfg.Topic,
		GroupID:  cfg.GroupID,
		MinBytes: 1,
		MaxBytes: 10e6,
		MaxWait:  500 * time.Millisecond,
	})

	return &KafkaBus{writer: writer, reader: reader}, nil
}

func (k *KafkaBus) Publish(ctx context.Context, evt SnapshotEvent) error {
	k.mu.Lock()
	defer k.mu.Unlock()
	if k.closed {
		return context.Canceled
	}
	body, err := json.Marshal(evt)
	if err != nil {
		return err
	}
	key := []byte(evt.SnapshotID)
	if len(key) == 0 {
		key = []byte(evt.Workflow)
	}
	return k.writer.WriteMessages(ctx, kafka.Message{
		Key:   key,
		Value: body,
		Time:  time.Now().UTC(),
	})
}

func (k *KafkaBus) Subscribe(ctx context.Context, handler Handler) error {
	k.mu.Lock()
	k.handlers = append(k.handlers, handler)
	k.mu.Unlock()

	go k.consumeLoop(ctx)
	return nil
}

func (k *KafkaBus) consumeLoop(ctx context.Context) {
	for {
		msg, err := k.reader.ReadMessage(ctx)
		if err != nil {
			if ctx.Err() != nil || k.closed {
				return
			}
			continue
		}
		var evt SnapshotEvent
		if err := json.Unmarshal(msg.Value, &evt); err != nil {
			continue
		}
		k.mu.RLock()
		handlers := append([]Handler(nil), k.handlers...)
		k.mu.RUnlock()
		for _, h := range handlers {
			_ = h(ctx, evt)
		}
	}
}

func (k *KafkaBus) Close() error {
	k.mu.Lock()
	defer k.mu.Unlock()
	if k.closed {
		return nil
	}
	k.closed = true
	var err1, err2 error
	if k.writer != nil {
		err1 = k.writer.Close()
	}
	if k.reader != nil {
		err2 = k.reader.Close()
	}
	if err1 != nil {
		return err1
	}
	return err2
}

// Ping verifies Kafka connectivity by listing brokers (writer dial).
func Ping(ctx context.Context, brokers []string) error {
	if len(brokers) == 0 {
		return fmt.Errorf("no brokers")
	}
	conn, err := kafka.DialContext(ctx, "tcp", brokers[0])
	if err != nil {
		return err
	}
	defer conn.Close()
	_, err = conn.Brokers()
	return err
}
