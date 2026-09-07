package cdc

import (
	"encoding/json"
	"fmt"
)

// unwrapDebezium maps a Kafka Connect / Debezium JSON record onto Envelope.
// Engine-published envelopes have no source.connector and fall through.
func unwrapDebezium(data []byte) (Envelope, error) {
	var top map[string]json.RawMessage
	if err := json.Unmarshal(data, &top); err != nil {
		return Envelope{}, err
	}
	body := data
	if p, ok := top["payload"]; ok && len(p) > 0 && p[0] == '{' {
		body = p
	}
	var msg struct {
		Op     string          `json:"op"`
		After  json.RawMessage `json:"after"`
		Source struct {
			Connector string `json:"connector"`
			Table     string `json:"table"`
		} `json:"source"`
	}
	if err := json.Unmarshal(body, &msg); err != nil {
		return Envelope{}, err
	}
	if msg.Source.Connector == "" && msg.Source.Table == "" {
		return Envelope{}, fmt.Errorf("not a debezium record")
	}
	if msg.Source.Table == "" {
		return Envelope{}, fmt.Errorf("debezium record missing source.table")
	}
	var after interface{}
	if len(msg.After) > 0 && string(msg.After) != "null" {
		if err := json.Unmarshal(msg.After, &after); err != nil {
			return Envelope{}, err
		}
		after = normalizeDebeziumAfter(after)
	}
	return Envelope{Op: msg.Op, Table: msg.Source.Table, After: after}, nil
}

func normalizeDebeziumAfter(after interface{}) interface{} {
	m, ok := after.(map[string]interface{})
	if !ok {
		return after
	}
	if v, exists := m["sealed"]; exists {
		m["sealed"] = coerceBool(v)
	}
	if v, exists := m["reused"]; exists {
		m["reused"] = coerceBool(v)
	}
	return m
}

func coerceBool(v interface{}) bool {
	switch t := v.(type) {
	case bool:
		return t
	case float64:
		return t != 0
	case string:
		return t == "true" || t == "t" || t == "1"
	default:
		return false
	}
}

// DebeziumSourceConnector is the source.connector value emitted by the Postgres connector.
const DebeziumSourceConnector = "postgresql"
