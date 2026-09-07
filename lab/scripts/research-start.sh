#!/usr/bin/env bash
# Start node-local kbl-tsdb on :9090 for research tests.
# Foreground (Cloud Agent start): KBL_RESEARCH_FOREGROUND=1
# Background (make research-up): default
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
ADDR="${KBL_TSDB_ADDR:-:9090}"
DATA_DIR="${KBL_TSDB_DATA:-/tmp/kbl-research/tsdb}"
LOG="${KBL_TSDB_LOG:-/tmp/kbl-research/tsdb.log}"
PIDFILE="${KBL_TSDB_PID:-/tmp/kbl-research/tsdb.pid}"
HEALTH_URL="${KBL_TSDB_HEALTH:-http://127.0.0.1:9090/healthz}"

mkdir -p "$(dirname "$LOG")" "$DATA_DIR"

bin="$ROOT/controller/bin/kbl-tsdb"
if [[ ! -x "$bin" ]]; then
  echo "kbl-tsdb missing; running research-install.sh"
  bash "$ROOT/lab/scripts/research-install.sh"
fi

healthy() {
  curl -sf "$HEALTH_URL" >/dev/null 2>&1
}

if healthy; then
  echo "kbl-tsdb already healthy at ${HEALTH_URL}"
  if [[ "${KBL_RESEARCH_FOREGROUND:-}" == "1" ]]; then
    # Keep the start process alive for Cloud Agents when a previous
    # instance already owns the port.
    echo "foreground wait on existing kbl-tsdb"
    while healthy; do sleep 30; done
    echo "existing kbl-tsdb went away" >&2
    exit 1
  fi
  exit 0
fi

if [[ "${KBL_RESEARCH_FOREGROUND:-}" == "1" ]]; then
  echo "kbl-tsdb foreground addr=${ADDR} data-dir=${DATA_DIR}"
  exec "$bin" --addr="$ADDR" --data-dir="$DATA_DIR"
fi

echo "starting kbl-tsdb in background addr=${ADDR} data-dir=${DATA_DIR}"
nohup "$bin" --addr="$ADDR" --data-dir="$DATA_DIR" >"$LOG" 2>&1 &
echo $! >"$PIDFILE"

for _ in $(seq 1 50); do
  if healthy; then
    echo "kbl-tsdb ready pid=$(cat "$PIDFILE") ${HEALTH_URL}"
    exit 0
  fi
  sleep 0.1
done

echo "kbl-tsdb failed to become healthy; log:" >&2
tail -n 50 "$LOG" >&2 || true
exit 1
