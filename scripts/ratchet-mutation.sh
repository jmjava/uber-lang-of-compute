#!/usr/bin/env bash
# Mutation-test controller/pkg/hash only. Fail when the killed-mutant score
# drops below the score of the existing hash tests recorded in the baseline.
# go-mutesting labels a killed mutant as "passed".
set -euo pipefail

root="$(cd "$(dirname "$0")/.." && pwd)"
# shellcheck source=ratchet-common.sh
source "$root/scripts/ratchet-common.sh"
require_cmd go-mutesting
require_cmd go

baseline="$root/controller/pkg/hash/mutation.baseline"
if [[ ! -f "$baseline" ]]; then
  echo "ratchet-mutation: missing $baseline" >&2
  exit 1
fi

floor_killed="$(awk -F= '/^killed=/ { print $2 }' "$baseline")"
floor_total="$(awk -F= '/^total=/ { print $2 }' "$baseline")"
if [[ ! "$floor_killed" =~ ^[0-9]+$ || ! "$floor_total" =~ ^[0-9]+$ || "$floor_total" -eq 0 ]]; then
  echo "ratchet-mutation: baseline must record killed=<n> and total=<n> with total > 0" >&2
  exit 1
fi

log="$(mktemp)"
report="$root/controller/report.json"
trap 'rm -f "$log" "$report" "$root/controller/go-mutesting-report.html"' EXIT

cd "$root/controller"
echo "ratchet-mutation: go-mutesting ./pkg/hash/"
set +e
go-mutesting --exec-timeout 60 ./pkg/hash/ >"$log" 2>&1
status=$?
set -e
# Surviving mutants are still a successful tool run. A missing summary is not.
if [[ "$status" -ne 0 ]]; then
  echo "ratchet-mutation: go-mutesting exited ${status}" >&2
  tail -n 40 "$log" >&2
  exit "$status"
fi

summary="$(grep -E 'The mutation score is ' "$log" | tail -n 1 || true)"
if [[ -z "$summary" ]]; then
  echo "ratchet-mutation: go-mutesting produced no mutation score" >&2
  tail -n 40 "$log" >&2
  exit 1
fi
echo "ratchet-mutation: $summary"

read -r killed escaped duplicated skipped total < <(
  python3 - "$summary" <<'PY'
import re, sys
line = sys.argv[1]
match = re.search(
    r"\((\d+) passed, (\d+) failed, (\d+) duplicated, (\d+) skipped, total is (\d+)\)",
    line,
)
if not match:
    sys.exit(1)
print(*match.groups())
PY
)

if [[ "$total" -eq 0 ]]; then
  echo "ratchet-mutation: total mutants is 0; refusing to treat that as a score" >&2
  exit 1
fi

# killed * floor_total < floor_killed * total  =>  killed/total < floor_killed/floor_total
if ((killed * floor_total < floor_killed * total)); then
  echo "ratchet-mutation: killed-mutant score ${killed}/${total} is below baseline ${floor_killed}/${floor_total}" >&2
  echo "ratchet-mutation: escaped=${escaped} duplicated=${duplicated} skipped=${skipped}" >&2
  exit 1
fi

echo "ratchet-mutation: pass (${killed}/${total}, baseline ${floor_killed}/${floor_total})"
