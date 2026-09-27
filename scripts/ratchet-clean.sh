#!/usr/bin/env bash
# Clean as You Code: lint only code new or changed versus the base revision.
# Existing findings stay in the static baseline and do not fail this gate.
set -euo pipefail

root="$(cd "$(dirname "$0")/.." && pwd)"
# shellcheck source=ratchet-common.sh
source "$root/scripts/ratchet-common.sh"
require_golangci

base="$(base_rev)"
cd "$root/controller"

echo "ratchet-clean: golangci-lint --new-from-rev ${base}"
golangci-lint run \
  --timeout 10m \
  --max-issues-per-linter 0 \
  --max-same-issues 0 \
  --new-from-rev "$base" \
  ./...

# --new-from-rev follows git diff and misses untracked files. Those files are
# entirely new, so the same linters have to see them before the commit exists.
untracked=()
while IFS= read -r rel; do
  [[ -n "$rel" ]] || continue
  untracked+=("$rel")
done < <(git -C "$root" ls-files --others --exclude-standard | awk '/^controller\/.*\.go$/ { sub(/^controller\//, ""); print }')

if ((${#untracked[@]} > 0)); then
  echo "ratchet-clean: linting untracked Go files"
  golangci-lint run \
    --timeout 10m \
    --max-issues-per-linter 0 \
    --max-same-issues 0 \
    "${untracked[@]}"
fi

echo "ratchet-clean: pass"
