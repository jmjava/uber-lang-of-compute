#!/usr/bin/env bash
# Install toolchain + binaries and start kbl-tsdb (background).
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
chmod +x "$ROOT/lab/scripts/"research-*.sh
bash "$ROOT/lab/scripts/research-install.sh"
bash "$ROOT/lab/scripts/research-start.sh"
echo
echo "Research assets are up."
echo "  make review          # workshop receipt (SQLite)"
echo "  make research-test   # receipt + live TSDB memo replay"
echo "  make research-kafka-test  # Redpanda + TestKafkaCDCRoundTrip"
echo "  make research-debezium-test  # Postgres WAL + Debezium Connect"
echo "  make research-down   # stop kbl-tsdb"
echo "  TSDB health: curl -sf http://127.0.0.1:9090/healthz"
