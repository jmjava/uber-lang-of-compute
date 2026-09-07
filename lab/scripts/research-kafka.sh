#!/usr/bin/env bash
# Start Redpanda and prove engine CDC envelopes round-trip on Kafka.
# This is the production bus, not a Debezium connector.
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
COMPOSE_FILE="${KBL_RESEARCH_COMPOSE:-$ROOT/lab/compose/research.yaml}"
BROKER="${KAFKA_BROKERS:-127.0.0.1:19092}"
IMAGE="${KBL_REDPANDA_IMAGE:-docker.redpanda.com/redpandadata/redpanda:v24.3.8}"
NAME="${KBL_REDPANDA_NAME:-kbl-redpanda}"

HOST="${BROKER%%,*}"
HOST="${HOST#*://}"
PORT="${HOST##*:}"
HOST="${HOST%%:*}"
if [[ "$HOST" == "$PORT" || -z "$PORT" ]]; then
  HOST=127.0.0.1
  PORT=19092
fi

start_with_compose() {
  docker compose -f "$COMPOSE_FILE" up -d redpanda
}

container_running() {
  docker ps --format '{{.Names}}' | grep -qx "$NAME"
}

ensure_network() {
  docker network inspect "${KBL_RESEARCH_NET:-kbl-research}" >/dev/null 2>&1 || \
    docker network create "${KBL_RESEARCH_NET:-kbl-research}"
}

start_with_docker_run() {
  local net="${KBL_RESEARCH_NET:-kbl-research}"
  ensure_network
  if container_running; then
    docker network connect "$net" "$NAME" >/dev/null 2>&1 || true
    echo "redpanda container ${NAME} already running"
    return 0
  fi
  if docker ps -a --format '{{.Names}}' | grep -qx "$NAME"; then
    docker start "$NAME"
    docker network connect "$net" "$NAME" >/dev/null 2>&1 || true
    return 0
  fi
  echo "starting ${NAME} from ${IMAGE} (no docker compose plugin)"
  docker run -d --name "$NAME" --hostname "$NAME" --network "$net" \
    -p "${PORT}:19092" \
    "$IMAGE" \
    redpanda start \
    --kafka-addr internal://0.0.0.0:9092,external://0.0.0.0:19092 \
    --advertise-kafka-addr internal://"${NAME}":9092,external://127.0.0.1:19092 \
    --rpc-addr "${NAME}:33145" \
    --advertise-rpc-addr "${NAME}:33145" \
    --mode dev-container \
    --smp 1 \
    --memory 1G \
    --overprovisioned \
    --default-log-level=warn
}

cd "$ROOT"
if docker compose version >/dev/null 2>&1; then
  start_with_compose
else
  start_with_docker_run
fi

echo "waiting for kafka at ${HOST}:${PORT}..."
ready=0
for _ in $(seq 1 60); do
  if (echo >/dev/tcp/"$HOST"/"$PORT") >/dev/null 2>&1; then
    ready=1
    break
  fi
  sleep 1
done
if [[ "$ready" != "1" ]]; then
  echo "error: kafka broker ${HOST}:${PORT} did not become reachable" >&2
  if docker compose version >/dev/null 2>&1; then
    docker compose -f "$COMPOSE_FILE" logs --tail=80 redpanda >&2 || true
  else
    docker logs --tail=80 "$NAME" >&2 || true
  fi
  exit 1
fi
# Port can accept TCP before Kafka is serving.
sleep 3

export KAFKA_BROKERS="${KAFKA_BROKERS:-$BROKER}"
echo "running TestKafkaCDCRoundTrip against ${KAFKA_BROKERS}"
(
  cd "$ROOT/controller"
  go test ./pkg/cdc/ -run TestKafkaCDCRoundTrip -count=1 -timeout 90s -v
)

echo "kafka cdc round-trip passed (broker=${KAFKA_BROKERS})"
echo "  stop: make research-kafka-down"
