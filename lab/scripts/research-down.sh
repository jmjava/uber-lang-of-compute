#!/usr/bin/env bash
# Stop the research kbl-tsdb started by research-start.sh / research-up.sh.
set -euo pipefail
PIDFILE="${KBL_TSDB_PID:-/tmp/kbl-research/tsdb.pid}"
if [[ -f "$PIDFILE" ]]; then
  pid="$(cat "$PIDFILE")"
  if kill -0 "$pid" 2>/dev/null; then
    kill "$pid" 2>/dev/null || true
    sleep 0.2
    kill -9 "$pid" 2>/dev/null || true
    echo "stopped kbl-tsdb pid=${pid}"
  fi
  rm -f "$PIDFILE"
else
  echo "no pidfile ${PIDFILE}"
fi
