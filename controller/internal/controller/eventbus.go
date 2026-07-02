package controller

import (
	"strings"

	"github.com/jmjava/uber-lang-of-compute/controller/pkg/events"
)

// NewEventBus returns the shared event bus for controllers.
func NewEventBus(kafkaBrokers, kafkaTopic string) events.Bus {
	brokers := splitBrokers(kafkaBrokers)
	if len(brokers) > 0 {
		bus, err := events.NewKafkaBus(events.KafkaConfig{
			Brokers: brokers,
			Topic:   kafkaTopic,
			GroupID: "kbl-controller",
		})
		if err == nil {
			return bus
		}
	}
	return events.NewMemoryBus()
}

func splitBrokers(s string) []string {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}
