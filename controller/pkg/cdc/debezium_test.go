package cdc_test

import (
	"encoding/json"
	"testing"

	"github.com/jmjava/uber-lang-of-compute/controller/pkg/cdc"
	"github.com/jmjava/uber-lang-of-compute/controller/pkg/store"
)

func TestUnmarshalDebeziumPostgresEnvelope(t *testing.T) {
	raw := []byte(`{
		"schema": {"type":"struct"},
		"payload": {
			"op": "c",
			"source": {"connector":"postgresql","table":"snapshots"},
			"after": {"snapshot_id":"abc","time_slice":"2025-04-15","data":"{\"v\":1}","sealed":true}
		}
	}`)
	env, err := cdc.UnmarshalEnvelope(raw)
	if err != nil {
		t.Fatal(err)
	}
	if env.Op != cdc.OpCreate || env.Table != cdc.TableSnapshots {
		t.Fatalf("unwrapped envelope: %+v", env)
	}

	target, err := store.OpenSQLite(t.TempDir() + "/replica.db")
	if err != nil {
		t.Fatal(err)
	}
	defer target.Close()
	if err := cdc.Apply(target, env); err != nil {
		t.Fatal(err)
	}
	_, data, sealed, err := target.GetSnapshot("abc")
	if err != nil || data != `{"v":1}` || !sealed {
		t.Fatalf("replica snapshot missing: err=%v data=%q sealed=%v", err, data, sealed)
	}
}

func TestUnmarshalDebeziumUnsealedIsRefused(t *testing.T) {
	raw := []byte(`{
		"payload": {
			"op": "c",
			"source": {"connector":"postgresql","table":"snapshots"},
			"after": {"snapshot_id":"live","time_slice":"2025-04-15","data":"{}","sealed":false}
		}
	}`)
	env, err := cdc.UnmarshalEnvelope(raw)
	if err != nil {
		t.Fatal(err)
	}
	target, err := store.OpenSQLite(t.TempDir() + "/replica.db")
	if err != nil {
		t.Fatal(err)
	}
	defer target.Close()
	if err := cdc.Apply(target, env); err == nil {
		t.Fatal("unsealed debezium snapshot must be refused")
	}
}

func TestUnmarshalEngineEnvelopeStillWorks(t *testing.T) {
	env := cdc.Envelope{
		Op:    cdc.OpCreate,
		Table: cdc.TableSnapshots,
		After: cdc.SnapshotRow{SnapshotID: "x", TimeSlice: "2025-04-15", Data: "{}", Sealed: true},
	}
	body, err := json.Marshal(env)
	if err != nil {
		t.Fatal(err)
	}
	got, err := cdc.UnmarshalEnvelope(body)
	if err != nil {
		t.Fatal(err)
	}
	if got.Table != cdc.TableSnapshots || got.Op != cdc.OpCreate {
		t.Fatalf("engine envelope round-trip: %+v", got)
	}
}
