package cdc_test

import (
	"context"
	"fmt"
	"net"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/jmjava/uber-lang-of-compute/controller/pkg/cdc"
	"github.com/jmjava/uber-lang-of-compute/controller/pkg/store"
)

// TestKafkaCDCRoundTrip publishes engine CDC envelopes to a real broker and
// applies them on a replica store. This is the Kafka bus, not Debezium capture.
//
// Skip when no broker is reachable and KAFKA_BROKERS is unset (CI / unit tests).
// Fail when KAFKA_BROKERS is set but the broker is down (research-kafka.sh).
func TestKafkaCDCRoundTrip(t *testing.T) {
	brokers := kafkaBrokersOrSkip(t)
	topic := fmt.Sprintf("kbl.cdc.test.%d", time.Now().UnixNano())
	group := fmt.Sprintf("kbl-cdc-test-%d", time.Now().UnixNano())
	const snapshotID = "snap-kafka-cdc-1"
	chain := []string{"load"}

	source, err := store.OpenSQLite(t.TempDir() + "/source.db")
	if err != nil {
		t.Fatal(err)
	}
	defer source.Close()
	target, err := store.OpenSQLite(t.TempDir() + "/target.db")
	if err != nil {
		t.Fatal(err)
	}
	defer target.Close()

	if err := source.SaveSnapshot(snapshotID, "2025-04-15", `{"v":1}`, true); err != nil {
		t.Fatal(err)
	}
	if err := source.SaveResult(snapshotID, "load", "in", "out", `{"ok":true}`, false, "", ""); err != nil {
		t.Fatal(err)
	}
	envs, err := cdc.ExportFromStore(source, snapshotID, chain)
	if err != nil {
		t.Fatal(err)
	}
	if len(envs) != 2 {
		t.Fatalf("expected snapshot + result envelopes, got %d", len(envs))
	}

	pub, err := cdc.NewKafkaPublisher(cdc.KafkaConfig{Brokers: brokers, Topic: topic})
	if err != nil {
		t.Fatal(err)
	}
	pubCtx, pubCancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer pubCancel()
	if err := pub.PublishBatch(pubCtx, snapshotID, envs); err != nil {
		t.Fatalf("publish: %v", err)
	}
	if err := pub.Close(); err != nil {
		t.Fatalf("publisher close: %v", err)
	}

	cons, err := cdc.NewKafkaConsumer(cdc.KafkaConfig{
		Brokers: brokers,
		Topic:   topic,
		GroupID: group,
	})
	if err != nil {
		t.Fatal(err)
	}
	defer cons.Close()

	consCtx, consCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer consCancel()
	progress, err := cdc.SyncFromConsumer(consCtx, cons, target, snapshotID, chain)
	if err != nil {
		t.Fatalf("sync from kafka: %v", err)
	}
	if !progress.IsComplete(len(chain)) {
		t.Fatalf("incomplete kafka cdc sync: %+v", progress)
	}

	_, data, sealed, err := target.GetSnapshot(snapshotID)
	if err != nil || data != `{"v":1}` || !sealed {
		t.Fatalf("replica snapshot missing: err=%v data=%q sealed=%v", err, data, sealed)
	}
	_, _, output, err := target.GetLatestResult(snapshotID, "load")
	if err != nil || output != `{"ok":true}` {
		t.Fatalf("replica result missing: err=%v output=%q", err, output)
	}
}

func kafkaBrokersOrSkip(t *testing.T) []string {
	t.Helper()
	raw, set := os.LookupEnv("KAFKA_BROKERS")
	if !set || strings.TrimSpace(raw) == "" {
		raw = "127.0.0.1:19092"
		if !kafkaReachable(raw) {
			t.Skipf("no kafka broker at %s (set KAFKA_BROKERS to require one)", raw)
		}
		return []string{raw}
	}
	brokers := splitBrokers(raw)
	if !kafkaReachable(brokers[0]) {
		t.Fatalf("KAFKA_BROKERS=%s is not reachable", raw)
	}
	return brokers
}

func splitBrokers(raw string) []string {
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

func kafkaReachable(addr string) bool {
	conn, err := net.DialTimeout("tcp", addr, 2*time.Second)
	if err != nil {
		return false
	}
	_ = conn.Close()
	return true
}
