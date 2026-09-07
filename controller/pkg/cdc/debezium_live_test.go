package cdc_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/jmjava/uber-lang-of-compute/controller/pkg/cdc"
	"github.com/jmjava/uber-lang-of-compute/controller/pkg/store"
)

// TestDebeziumCapturesSealedSnapshot requires Postgres WAL + Debezium Connect
// publishing onto the Kafka bus. This is Phase 40: real connector capture, not
// engine-published envelopes.
//
// Skip when DEBEZIUM_PROOF is unset and the stack is down.
// Fail when DEBEZIUM_PROOF=1 and capture does not happen.
func TestDebeziumCapturesSealedSnapshot(t *testing.T) {
	proof := os.Getenv("DEBEZIUM_PROOF") == "1"
	dsn := getenv("POSTGRES_DSN", "postgres://kbl:kbl@127.0.0.1:15432/kbl?sslmode=disable")
	brokers := splitBrokers(getenv("KAFKA_BROKERS", "127.0.0.1:19092"))
	topic := getenv("DEBEZIUM_TOPIC", "kbl.public.snapshots")
	connectURL := getenv("DEBEZIUM_CONNECT", "http://127.0.0.1:8083")

	pg, err := store.OpenPostgres(dsn)
	if err != nil {
		if proof {
			t.Fatalf("postgres required for DEBEZIUM_PROOF: %v", err)
		}
		t.Skipf("postgres not reachable: %v", err)
	}
	defer pg.Close()

	if !kafkaReachable(brokers[0]) {
		if proof {
			t.Fatalf("kafka broker %s required for DEBEZIUM_PROOF", brokers[0])
		}
		t.Skipf("kafka broker %s not reachable", brokers[0])
	}

	if err := assertDebeziumConnector(connectURL); err != nil {
		if proof {
			t.Fatalf("debezium connect: %v", err)
		}
		t.Skipf("debezium connect: %v", err)
	}

	snapshotID := fmt.Sprintf("snap-debezium-%d", time.Now().UnixNano())
	if err := pg.SaveSnapshot(snapshotID, "2025-04-15", `{"v":1}`, true); err != nil {
		t.Fatal(err)
	}

	group := fmt.Sprintf("kbl-debezium-test-%d", time.Now().UnixNano())
	cons, err := cdc.NewKafkaConsumer(cdc.KafkaConfig{Brokers: brokers, Topic: topic, GroupID: group})
	if err != nil {
		t.Fatal(err)
	}
	defer cons.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()
	envs, err := cons.Consume(ctx, snapshotID)
	if err != nil {
		t.Fatalf("consume debezium topic %s: %v", topic, err)
	}
	if len(envs) == 0 {
		t.Fatal("no debezium envelopes")
	}

	target, err := store.OpenSQLite(t.TempDir() + "/replica.db")
	if err != nil {
		t.Fatal(err)
	}
	defer target.Close()
	progress, err := cdc.ApplyAll(target, snapshotID, envs)
	if err != nil {
		t.Fatal(err)
	}
	if !progress.SnapshotApplied {
		t.Fatalf("debezium envelopes did not apply snapshot: %+v envs=%d", progress, len(envs))
	}
	_, data, sealed, err := target.GetSnapshot(snapshotID)
	if err != nil || data != `{"v":1}` || !sealed {
		t.Fatalf("replica missing sealed snapshot: err=%v data=%q sealed=%v", err, data, sealed)
	}
}

func assertDebeziumConnector(connectURL string) error {
	client := &http.Client{Timeout: 3 * time.Second}
	resp, err := client.Get(connectURL + "/connectors/kbl-snapshots")
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("connect GET %s: %s", resp.Status, body)
	}
	var spec struct {
		Name   string            `json:"name"`
		Config map[string]string `json:"config"`
	}
	if err := json.Unmarshal(body, &spec); err != nil {
		return err
	}
	class := spec.Config["connector.class"]
	if class != "io.debezium.connector.postgresql.PostgresConnector" {
		return fmt.Errorf("connector.class=%q want PostgresConnector", class)
	}

	stResp, err := client.Get(connectURL + "/connectors/kbl-snapshots/status")
	if err != nil {
		return err
	}
	defer stResp.Body.Close()
	stBody, _ := io.ReadAll(stResp.Body)
	if stResp.StatusCode != http.StatusOK {
		return fmt.Errorf("connect status %s: %s", stResp.Status, stBody)
	}
	var st struct {
		Connector struct {
			State string `json:"state"`
		} `json:"connector"`
	}
	if err := json.Unmarshal(stBody, &st); err != nil {
		return err
	}
	if st.Connector.State != "RUNNING" {
		return fmt.Errorf("connector state %s (want RUNNING): %s", st.Connector.State, stBody)
	}
	return nil
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
