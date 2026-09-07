#!/usr/bin/env bash
# Start Postgres (WAL) + Redpanda + Debezium Connect and prove the connector
# captures sealed snapshot rows onto the Kafka bus. Compact Kind is not required.
# This is not engine-published CDC envelopes (Phase 39).
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
COMPOSE_FILE="${KBL_RESEARCH_COMPOSE:-$ROOT/lab/compose/research.yaml}"
CONNECTOR_JSON="${ROOT}/lab/compose/debezium-postgres-connector.json"
NETWORK="${KBL_RESEARCH_NET:-kbl-research}"
BROKER="${KAFKA_BROKERS:-127.0.0.1:19092}"
PG_DSN="${POSTGRES_DSN:-postgres://kbl:kbl@127.0.0.1:15432/kbl?sslmode=disable}"
CONNECT_URL="${DEBEZIUM_CONNECT:-http://127.0.0.1:8083}"
REDPANDA_IMAGE="${KBL_REDPANDA_IMAGE:-docker.redpanda.com/redpandadata/redpanda:v24.3.8}"
REDPANDA_NAME="${KBL_REDPANDA_NAME:-kbl-redpanda}"
PG_IMAGE="${KBL_POSTGRES_IMAGE:-postgres:16-alpine}"
PG_NAME="${KBL_POSTGRES_NAME:-kbl-postgres}"
DEBEZIUM_IMAGE="${KBL_DEBEZIUM_IMAGE:-quay.io/debezium/connect:2.7.3.Final}"
DEBEZIUM_NAME="${KBL_DEBEZIUM_NAME:-kbl-debezium}"

wait_tcp() {
  local host="$1" port="$2" name="$3"
  echo "waiting for ${name} at ${host}:${port}..."
  local i
  for i in $(seq 1 90); do
    if (echo >/dev/tcp/"$host"/"$port") >/dev/null 2>&1; then
      return 0
    fi
    sleep 1
  done
  echo "error: ${name} ${host}:${port} did not become reachable" >&2
  return 1
}

ensure_network() {
  docker network inspect "$NETWORK" >/dev/null 2>&1 || docker network create "$NETWORK"
}

container_exists() { docker ps -a --format '{{.Names}}' | grep -qx "$1"; }
container_running() { docker ps --format '{{.Names}}' | grep -qx "$1"; }

ensure_container_network() {
  local name="$1"
  docker network connect "$NETWORK" "$name" >/dev/null 2>&1 || true
}

start_redpanda() {
  if container_running "$REDPANDA_NAME"; then
    ensure_container_network "$REDPANDA_NAME"
    return 0
  fi
  if container_exists "$REDPANDA_NAME"; then
    docker start "$REDPANDA_NAME" >/dev/null
    ensure_container_network "$REDPANDA_NAME"
    return 0
  fi
  docker run -d --name "$REDPANDA_NAME" --hostname "$REDPANDA_NAME" --network "$NETWORK" \
    -p 19092:19092 \
    "$REDPANDA_IMAGE" \
    redpanda start \
    --kafka-addr internal://0.0.0.0:9092,external://0.0.0.0:19092 \
    --advertise-kafka-addr internal://"${REDPANDA_NAME}":9092,external://127.0.0.1:19092 \
    --rpc-addr "${REDPANDA_NAME}:33145" \
    --advertise-rpc-addr "${REDPANDA_NAME}:33145" \
    --mode dev-container \
    --smp 1 \
    --memory 1G \
    --overprovisioned \
    --default-log-level=warn >/dev/null
}

start_postgres() {
  if container_running "$PG_NAME"; then
    ensure_container_network "$PG_NAME"
    return 0
  fi
  if container_exists "$PG_NAME"; then
    docker start "$PG_NAME" >/dev/null
    ensure_container_network "$PG_NAME"
    return 0
  fi
  docker run -d --name "$PG_NAME" --hostname "$PG_NAME" --network "$NETWORK" \
    -e POSTGRES_USER=kbl -e POSTGRES_PASSWORD=kbl -e POSTGRES_DB=kbl \
    -p 15432:5432 \
    -v "${ROOT}/lab/compose/postgres-init.sql:/docker-entrypoint-initdb.d/01-kbl.sql:ro" \
    "$PG_IMAGE" \
    postgres -c wal_level=logical -c max_wal_senders=4 -c max_replication_slots=4 >/dev/null
}

start_debezium() {
  if container_running "$DEBEZIUM_NAME"; then
    ensure_container_network "$DEBEZIUM_NAME"
    return 0
  fi
  if container_exists "$DEBEZIUM_NAME"; then
    docker start "$DEBEZIUM_NAME" >/dev/null
    ensure_container_network "$DEBEZIUM_NAME"
    return 0
  fi
  docker run -d --name "$DEBEZIUM_NAME" --hostname "$DEBEZIUM_NAME" --network "$NETWORK" \
    -p 8083:8083 \
    -e BOOTSTRAP_SERVERS="${REDPANDA_NAME}:9092" \
    -e GROUP_ID=kbl-connect \
    -e CONFIG_STORAGE_TOPIC=kbl_connect_configs \
    -e OFFSET_STORAGE_TOPIC=kbl_connect_offsets \
    -e STATUS_STORAGE_TOPIC=kbl_connect_statuses \
    -e CONFIG_STORAGE_REPLICATION_FACTOR=1 \
    -e OFFSET_STORAGE_REPLICATION_FACTOR=1 \
    -e STATUS_STORAGE_REPLICATION_FACTOR=1 \
    -e KEY_CONVERTER=org.apache.kafka.connect.json.JsonConverter \
    -e VALUE_CONVERTER=org.apache.kafka.connect.json.JsonConverter \
    -e CONNECT_KEY_CONVERTER_SCHEMAS_ENABLE=false \
    -e CONNECT_VALUE_CONVERTER_SCHEMAS_ENABLE=false \
    "$DEBEZIUM_IMAGE" >/dev/null
}

register_connector() {
  local code
  code="$(curl -sS -o /tmp/kbl-debezium-conn.out -w '%{http_code}' \
    -X POST "${CONNECT_URL}/connectors" \
    -H 'Content-Type: application/json' \
    --data @"$CONNECTOR_JSON" || true)"
  if [[ "$code" == "201" || "$code" == "200" ]]; then
    echo "registered connector kbl-snapshots"
    return 0
  fi
  if [[ "$code" == "409" ]]; then
    echo "connector kbl-snapshots already exists"
    return 0
  fi
  echo "error: register connector HTTP ${code}: $(cat /tmp/kbl-debezium-conn.out 2>/dev/null || true)" >&2
  return 1
}

wait_connector() {
  echo "waiting for connector kbl-snapshots RUNNING..."
  local i body
  for i in $(seq 1 60); do
    body="$(curl -sf "${CONNECT_URL}/connectors/kbl-snapshots/status" || true)"
    if echo "$body" | grep -q '"state":"RUNNING"' && echo "$body" | grep -q '"id":0'; then
      echo "$body"
      return 0
    fi
    sleep 2
  done
  echo "error: connector not RUNNING: ${body:-'(no status)'}" >&2
  docker logs --tail=80 "$DEBEZIUM_NAME" >&2 || true
  return 1
}

cd "$ROOT"
ensure_network

if docker compose version >/dev/null 2>&1; then
  docker compose -f "$COMPOSE_FILE" up -d redpanda postgres debezium
else
  echo "starting postgres + redpanda + debezium with docker run (no compose plugin)"
  start_redpanda
  start_postgres
  start_debezium
fi

wait_tcp 127.0.0.1 19092 kafka
wait_tcp 127.0.0.1 15432 postgres
wait_tcp 127.0.0.1 8083 debezium-connect
# Connect REST can listen before plugins finish loading.
sleep 5

export POSTGRES_DSN="$PG_DSN"
export GOTOOLCHAIN="${GOTOOLCHAIN:-go1.23.7}"
echo "ensuring postgres schema (store migrate / init)"
(
  cd "$ROOT/controller"
  go test ./pkg/store/ -run TestPostgresSealedSnapshotIsWriteOnce -count=1 -timeout 60s
)

register_connector
wait_connector

export KAFKA_BROKERS="${KAFKA_BROKERS:-$BROKER}"
export DEBEZIUM_CONNECT="$CONNECT_URL"
export DEBEZIUM_TOPIC="${DEBEZIUM_TOPIC:-kbl.public.snapshots}"
export DEBEZIUM_PROOF=1
echo "running TestDebeziumCapturesSealedSnapshot"
(
  cd "$ROOT/controller"
  go test ./pkg/cdc/ -run 'TestDebeziumCapturesSealedSnapshot|TestUnmarshalDebezium' -count=1 -timeout 120s -v
)

echo "debezium capture passed (connect=${CONNECT_URL} topic=${DEBEZIUM_TOPIC})"
echo "  stop: make research-debezium-down"
