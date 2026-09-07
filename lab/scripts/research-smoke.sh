#!/usr/bin/env bash
# Smoke the research environment: workshop receipt + live TSDB memo replay.
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
cd "$ROOT"
chmod +x "$ROOT/lab/scripts/"research-*.sh

bash "$ROOT/lab/scripts/research-up.sh"

echo
echo "== workshop receipt (make review) =="
make review

SMOKE_DIR="${KBL_RESEARCH_SMOKE:-/tmp/kbl-research/smoke}"
mkdir -p "$SMOKE_DIR"
WF="$ROOT/examples/finance-curve-snapshot/workflow.yaml"
STORE="${KBL_TSDB_URL:-http://127.0.0.1:9090}"

echo
echo "== live TSDB first run =="
"$ROOT/controller/bin/kbl-compute" \
  --workflow "$WF" \
  --store "$STORE" \
  --replay-log "$SMOKE_DIR/replay-1.json"

echo
echo "== live TSDB memo replay =="
"$ROOT/controller/bin/kbl-compute" \
  --workflow "$WF" \
  --store "$STORE" \
  --replay-log "$SMOKE_DIR/replay-2.json"

python3 - "$SMOKE_DIR/replay-1.json" "$SMOKE_DIR/replay-2.json" <<'PY'
import json, sys
first = json.load(open(sys.argv[1]))
second = json.load(open(sys.argv[2]))
if first.get("snapshot_id") != second.get("snapshot_id"):
    raise SystemExit("snapshot id diverged on TSDB replay")
if first.get("head_link") != second.get("head_link"):
    raise SystemExit("head link diverged on TSDB replay")
entries = second.get("entries") or []
if not entries or any(not e.get("reused") for e in entries):
    raise SystemExit("second TSDB run must reuse every domino")
print("TSDB memo replay ok: snapshot=%s reuses=%d" % (second["snapshot_id"], len(entries)))
PY

echo
echo "Research smoke passed."
