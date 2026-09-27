#!/usr/bin/env bash
# GitHub Actions entry: go-mutesting on controller/pkg/hash, then the ceiling.
# A missing tool is a failure. This script does not invent a score.
set -euo pipefail

root="$(cd "$(dirname "$0")/.." && pwd)"
ceiling="$root/controller/pkg/hash/mutation.ceiling"

if ! command -v go-mutesting >/dev/null 2>&1; then
  echo "run-mutation: go-mutesting is not installed; refusing to record a score" >&2
  exit 1
fi
if ! command -v go >/dev/null 2>&1; then
  echo "run-mutation: go is not installed; refusing to record a score" >&2
  exit 1
fi

log="$(mktemp)"
trap 'rm -f "$log" "$root/controller/report.json"' EXIT

cd "$root/controller"
echo "run-mutation: go-mutesting ./pkg/hash/"
set +e
go-mutesting --exec-timeout 60 ./pkg/hash/ >"$log" 2>&1
status=$?
set -e
if [[ "$status" -ne 0 ]]; then
  echo "run-mutation: go-mutesting exited ${status}" >&2
  tail -n 40 "$log" >&2
  exit "$status"
fi

"$root/scripts/check-mutation-score.sh" --log "$log" --ceiling "$ceiling"
