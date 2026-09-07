#!/usr/bin/env bash
# Stop the research Redpanda broker started by research-kafka.sh.
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
COMPOSE_FILE="${KBL_RESEARCH_COMPOSE:-$ROOT/lab/compose/research.yaml}"
NAME="${KBL_REDPANDA_NAME:-kbl-redpanda}"

if docker compose version >/dev/null 2>&1; then
  docker compose -f "$COMPOSE_FILE" stop redpanda >/dev/null 2>&1 || true
fi
if docker ps -a --format '{{.Names}}' | grep -qx "$NAME"; then
  docker stop "$NAME" >/dev/null
  echo "stopped ${NAME}"
else
  echo "no redpanda container ${NAME}"
fi
