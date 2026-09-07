#!/usr/bin/env bash
# Stop Debezium Connect and Postgres started by research-debezium.sh.
# Leaves Redpanda running unless KBL_STOP_REDPANDA=1.
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
COMPOSE_FILE="${KBL_RESEARCH_COMPOSE:-$ROOT/lab/compose/research.yaml}"
PG_NAME="${KBL_POSTGRES_NAME:-kbl-postgres}"
DEBEZIUM_NAME="${KBL_DEBEZIUM_NAME:-kbl-debezium}"
REDPANDA_NAME="${KBL_REDPANDA_NAME:-kbl-redpanda}"

if docker compose version >/dev/null 2>&1; then
  docker compose -f "$COMPOSE_FILE" stop debezium postgres >/dev/null 2>&1 || true
  if [[ "${KBL_STOP_REDPANDA:-0}" == "1" ]]; then
    docker compose -f "$COMPOSE_FILE" stop redpanda >/dev/null 2>&1 || true
  fi
fi

for name in "$DEBEZIUM_NAME" "$PG_NAME"; do
  if docker ps -a --format '{{.Names}}' | grep -qx "$name"; then
    docker stop "$name" >/dev/null || true
    echo "stopped ${name}"
  fi
done
if [[ "${KBL_STOP_REDPANDA:-0}" == "1" ]] && docker ps -a --format '{{.Names}}' | grep -qx "$REDPANDA_NAME"; then
  docker stop "$REDPANDA_NAME" >/dev/null || true
  echo "stopped ${REDPANDA_NAME}"
fi
