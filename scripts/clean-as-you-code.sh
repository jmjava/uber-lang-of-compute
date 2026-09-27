#!/usr/bin/env bash
# Clean as You Code: staticcheck grades Go lines that are new or changed
# since the base revision. Findings on unchanged lines are printed so they
# stay visible, and they do not fail this gate.
set -euo pipefail

root="$(cd "$(dirname "$0")/.." && pwd)"

if ! command -v golangci-lint >/dev/null 2>&1; then
  echo "clean-as-you-code: missing golangci-lint (CI installs v2.13.2)" >&2
  exit 1
fi

resolve_base() {
  local base="${CLEAN_BASE:-origin/main}"
  if [[ "$base" == "0000000000000000000000000000000000000000" ]]; then
    base="origin/main"
  fi
  if git -C "$root" rev-parse --verify --quiet "${base}^{commit}" >/dev/null; then
    git -C "$root" rev-parse "${base}^{commit}"
    return
  fi
  if git -C "$root" rev-parse --verify --quiet origin/main >/dev/null; then
    git -C "$root" rev-parse origin/main
    return
  fi
  echo "clean-as-you-code: cannot find base revision (CLEAN_BASE=${CLEAN_BASE:-unset})" >&2
  exit 1
}

base="$(resolve_base)"
cd "$root/controller"

common=(
  --color never
  --timeout 10m
  --enable-only staticcheck
  --max-issues-per-linter 0
  --max-same-issues 0
)

echo "clean-as-you-code: existing staticcheck findings (visible, not a failure)"
set +e
golangci-lint run "${common[@]}" --issues-exit-code 0 ./...
report_status=$?
set -e
if [[ "$report_status" -ne 0 ]]; then
  echo "clean-as-you-code: existing-code report exited ${report_status}; unchanged findings do not fail this gate" >&2
fi

echo "clean-as-you-code: staticcheck --new-from-rev ${base}"
golangci-lint run "${common[@]}" --new-from-rev "$base" ./...

# --new-from-rev follows git diff and misses untracked files. Every line in
# those files is new, so the same linter has to see them.
untracked=()
while IFS= read -r rel; do
  [[ -n "$rel" ]] || continue
  untracked+=("$rel")
done < <(git -C "$root" ls-files --others --exclude-standard | awk '/^controller\/.*\.go$/ { sub(/^controller\//, ""); print }')

if ((${#untracked[@]} > 0)); then
  echo "clean-as-you-code: staticcheck on untracked Go files"
  golangci-lint run "${common[@]}" "${untracked[@]}"
fi

echo "clean-as-you-code: pass"
