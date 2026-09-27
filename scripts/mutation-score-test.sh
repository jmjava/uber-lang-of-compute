#!/usr/bin/env bash
# Proving test for the survived ceiling. Does not run go-mutesting.
set -euo pipefail

root="$(cd "$(dirname "$0")/.." && pwd)"
check="$root/scripts/check-mutation-score.sh"
workflow="$root/.github/workflows/mutation.yml"
committed="$root/controller/pkg/hash/mutation.ceiling"
tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT

score_line() {
  local killed="$1" survived="$2" total="$3"
  printf 'The mutation score is 0.000000 (%s passed, %s failed, 0 duplicated, 0 skipped, total is %s)\n' \
    "$killed" "$survived" "$total"
}

expect_fail() {
  local label="$1"
  shift
  if "$@"; then
    echo "mutation-score-test: FAIL ${label} (expected non-zero)" >&2
    exit 1
  fi
}

echo "mutation-score-test: survived at the ceiling passes and does not rewrite it"
printf 'survived=2\n' >"$tmp/ceiling"
score_line 8 2 10 >"$tmp/log"
before="$(cksum "$tmp/ceiling")"
"$check" --log "$tmp/log" --ceiling "$tmp/ceiling"
after="$(cksum "$tmp/ceiling")"
if [[ "$before" != "$after" ]]; then
  echo "mutation-score-test: FAIL checker rewrote the ceiling" >&2
  exit 1
fi

echo "mutation-score-test: a higher survived count fails"
score_line 8 3 11 >"$tmp/log"
expect_fail "survived above ceiling" "$check" --log "$tmp/log" --ceiling "$tmp/ceiling"

echo "mutation-score-test: zero mutants is not a score"
score_line 0 0 0 >"$tmp/log"
expect_fail "total 0" "$check" --log "$tmp/log" --ceiling "$tmp/ceiling"

echo "mutation-score-test: skipped-only output is not a score"
printf 'The mutation score is 0.000000 (0 passed, 0 failed, 0 duplicated, 4 skipped, total is 4)\n' >"$tmp/log"
expect_fail "no executed mutant" "$check" --log "$tmp/log" --ceiling "$tmp/ceiling"

echo "mutation-score-test: a log with no score fails"
printf 'go-mutesting: install failed\n' >"$tmp/log"
expect_fail "missing score" "$check" --log "$tmp/log" --ceiling "$tmp/ceiling"

echo "mutation-score-test: committed ceiling is frozen at the first measured score"
got="$(awk -F= '/^survived=/ { print $2; exit }' "$committed")"
if [[ "$got" != "8" ]]; then
  echo "mutation-score-test: FAIL committed survived ceiling is ${got}; frozen at 8 from run 36340786220 and must not be raised" >&2
  exit 1
fi
if ! grep -q '36340786220' "$committed"; then
  echo "mutation-score-test: FAIL ceiling comment does not name run 36340786220" >&2
  exit 1
fi

if ! grep -q 'go-mutesting' "$workflow"; then
  echo "mutation-score-test: FAIL workflow does not run go-mutesting" >&2
  exit 1
fi
if ! grep -q 'pkg/hash' "$workflow"; then
  echo "mutation-score-test: FAIL workflow is not scoped to pkg/hash" >&2
  exit 1
fi
if grep -E 'mutation\.ceiling' "$workflow" | grep -q '>'; then
  echo "mutation-score-test: FAIL workflow rewrites the ceiling" >&2
  exit 1
fi

echo "mutation-score-test: PASS"
