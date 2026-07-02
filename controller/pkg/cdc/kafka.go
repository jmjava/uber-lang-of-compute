package cdc

import (
	"context"
	"fmt"
	"time"

	"github.com/segmentio/kafka-go"

	"github.com/jmjava/uber-lang-of-compute/controller/pkg/store"
)

// KafkaConfig configures Kafka CDC publish/consume.
type KafkaConfig struct {
	Brokers []string
	Topic   string
	GroupID string
}

// KafkaPublisher publishes CDC envelopes to Kafka.
type KafkaPublisher struct {
	writer *kafka.Writer
	topic  string
}

func NewKafkaPublisher(cfg KafkaConfig) (*KafkaPublisher, error) {
	if len(cfg.Brokers) == 0 {
		return nil, fmt.Errorf("kafka brokers required")
	}
	if cfg.Topic == "" {
		cfg.Topic = DefaultTopic
	}
	return &KafkaPublisher{
		writer: &kafka.Writer{
			Addr:         kafka.TCP(cfg.Brokers...),
			Topic:        cfg.Topic,
			Balancer:     &kafka.LeastBytes{},
			RequiredAcks: kafka.RequireOne,
		},
		topic: cfg.Topic,
	}, nil
}

func (p *KafkaPublisher) Publish(ctx context.Context, snapshotID string, env Envelope) error {
	body, err := MarshalEnvelope(env)
	if err != nil {
		return err
	}
	return p.writer.WriteMessages(ctx, kafka.Message{
		Key:   []byte(snapshotID),
		Value: body,
		Time:  time.Now().UTC(),
	})
}

func (p *KafkaPublisher) PublishBatch(ctx context.Context, snapshotID string, envs []Envelope) error {
	for _, env := range envs {
		if err := p.Publish(ctx, snapshotID, env); err != nil {
			return err
		}
	}
	return nil
}

func (p *KafkaPublisher) Close() error {
	if p.writer == nil {
		return nil
	}
	return p.writer.Close()
}

// KafkaConsumer reads CDC envelopes for a snapshot from Kafka.
type KafkaConsumer struct {
	reader *kafka.Reader
}

func NewKafkaConsumer(cfg KafkaConfig) (*KafkaConsumer, error) {
	if len(cfg.Brokers) == 0 {
		return nil, fmt.Errorf("kafka brokers required")
	}
	if cfg.Topic == "" {
		cfg.Topic = DefaultTopic
	}
	if cfg.GroupID == "" {
		cfg.GroupID = "kbl-cdc-replica"
	}
	return &KafkaConsumer{
		reader: kafka.NewReader(kafka.ReaderConfig{
			Brokers:        cfg.Brokers,
			Topic:          cfg.Topic,
			GroupID:        cfg.GroupID,
			MinBytes:       1,
			MaxBytes:       10e6,
			MaxWait:        500 * time.Millisecond,
			CommitInterval: time.Second,
		}),
	}, nil
}

func (c *KafkaConsumer) Consume(ctx context.Context, snapshotID string) ([]Envelope, error) {
	var out []Envelope
	deadline, ok := ctx.Deadline()
	if !ok {
		deadline = time.Now().Add(5 * time.Second)
	}

	for time.Now().Before(deadline) {
		readCtx, cancel := context.WithTimeout(ctx, 500*time.Millisecond)
		msg, err := c.reader.ReadMessage(readCtx)
		cancel()
		if err != nil {
			if len(out) > 0 {
				break
			}
			continue
		}
		if string(msg.Key) != snapshotID && snapshotID != "" {
			continue
		}
		env, err := UnmarshalEnvelope(msg.Value)
		if err != nil {
			continue
		}
		if !matchesSnapshot(snapshotID, env) {
			continue
		}
		out = append(out, env)
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("no cdc events for snapshot %s", snapshotID)
	}
	return out, nil
}

func (c *KafkaConsumer) Close() error {
	if c.reader == nil {
		return nil
	}
	return c.reader.Close()
}

// SyncFromConsumer applies CDC events from a consumer to a target store.
func SyncFromConsumer(ctx context.Context, consumer Consumer, target store.Backend, snapshotID string, dominoChain []string) (SyncProgress, error) {
	envs, err := consumer.Consume(ctx, snapshotID)
	if err != nil {
		return SyncProgress{}, err
	}
	progress, err := ApplyAll(target, snapshotID, envs)
	if err != nil {
		return progress, err
	}
	if !progress.IsComplete(len(dominoChain)) {
		return progress, fmt.Errorf("incomplete cdc sync: snapshot=%v dominos=%d/%d",
			progress.SnapshotApplied, progress.DominoCount, len(dominoChain))
	}
	return progress, nil
}
