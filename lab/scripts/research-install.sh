#!/usr/bin/env bash
# Install Go module deps and research binaries (kbl-compute, kbl-review, kbl-tsdb).
# Used by Cursor Cloud Agents (.cursor/environment.json install) and make research-up.
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
cd "$ROOT"

export GOTOOLCHAIN="${GOTOOLCHAIN:-go1.23.7}"
export CGO_ENABLED="${CGO_ENABLED:-1}"

need() {
  command -v "$1" >/dev/null 2>&1 || {
    echo "error: $1 is required" >&2
    exit 1
  }
}

need go
need make
need gcc

echo "Go toolchain: $(go version) GOTOOLCHAIN=${GOTOOLCHAIN}"
echo "Downloading controller modules..."
( cd "$ROOT/controller" && go mod download )

echo "Building research binaries..."
make build
( cd "$ROOT/controller" && go build -o bin/kbl-review ./cmd/kbl-review )

mkdir -p /tmp/kbl-research
echo "Research binaries:"
ls -l "$ROOT/controller/bin/kbl-compute" "$ROOT/controller/bin/kbl-review" "$ROOT/controller/bin/kbl-tsdb"
