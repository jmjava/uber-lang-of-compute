#!/usr/bin/env bash
# Compare today's golangci-lint findings to the committed baseline.
# This script never writes the baseline. Delete nothing from it to go green.
set -euo pipefail

root="$(cd "$(dirname "$0")/.." && pwd)"
# shellcheck source=ratchet-common.sh
source "$root/scripts/ratchet-common.sh"
require_golangci

baseline="$root/controller/.golangci-baseline.txt"
if [[ ! -f "$baseline" ]]; then
  echo "ratchet-baseline: missing $baseline" >&2
  exit 1
fi

json="$(mktemp)"
current="$(mktemp)"
trap 'rm -f "$json" "$current"' EXIT

cd "$root/controller"
echo "ratchet-baseline: collecting golangci-lint findings (baseline is read-only)"
golangci-lint run \
  --timeout 10m \
  --max-issues-per-linter 0 \
  --max-same-issues 0 \
  --output.json.path "$json" \
  --output.text.path /dev/null \
  --issues-exit-code 0 \
  ./...

python3 - "$json" "$current" <<'PY'
import json
import sys

src, dest = sys.argv[1], sys.argv[2]
with open(src, encoding="utf-8") as fh:
    data = json.load(fh)
issues = []
if isinstance(data, dict):
    issues = data.get("Issues") or []
elif isinstance(data, list):
    issues = data
lines = []
for issue in issues:
    pos = issue.get("Pos") or {}
    filename = pos.get("Filename") or ""
    text = (issue.get("Text") or "").splitlines()[0] if issue.get("Text") else ""
    lines.append(
        "%s:%s:%s:%s:%s"
        % (
            filename,
            pos.get("Line", 0),
            pos.get("Column", 0),
            issue.get("FromLinter") or "",
            text,
        )
    )
with open(dest, "w", encoding="utf-8") as out:
    out.write("\n".join(sorted(set(lines))))
    if lines:
        out.write("\n")
PY

new="$(mktemp)"
trap 'rm -f "$json" "$current" "$new"' EXIT
comm -13 \
  <(grep -v '^#' "$baseline" | sed '/^$/d' | sort -u) \
  <(sed '/^$/d' "$current" | sort -u) >"$new"

if [[ -s "$new" ]]; then
  echo "ratchet-baseline: findings that are not in the committed baseline:" >&2
  cat "$new" >&2
  exit 1
fi
echo "ratchet-baseline: pass"
