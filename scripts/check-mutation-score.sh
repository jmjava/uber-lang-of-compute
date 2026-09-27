#!/usr/bin/env bash
# Compare a go-mutesting log to a survived ceiling.
# Does not run go-mutesting and does not rewrite the ceiling file.
# go-mutesting "failed" is a survived mutant. "passed" is a killed mutant.
set -euo pipefail

usage() {
  echo "usage: check-mutation-score.sh --log FILE --ceiling FILE" >&2
  exit 2
}

log=""
ceiling=""
while [[ $# -gt 0 ]]; do
  case "$1" in
    --log)
      log="${2:-}"
      shift 2
      ;;
    --ceiling)
      ceiling="${2:-}"
      shift 2
      ;;
    *)
      usage
      ;;
  esac
done

if [[ -z "$log" || -z "$ceiling" ]]; then
  usage
fi
if [[ ! -f "$log" ]]; then
  echo "check-mutation-score: missing log $log" >&2
  exit 1
fi
if [[ ! -f "$ceiling" ]]; then
  echo "check-mutation-score: missing ceiling $ceiling" >&2
  exit 1
fi

allowed="$(awk -F= '/^survived=/ { print $2; exit }' "$ceiling")"
if [[ ! "$allowed" =~ ^[0-9]+$ ]]; then
  echo "check-mutation-score: ceiling must record survived=<n>" >&2
  exit 1
fi

summary="$(grep -E 'The mutation score is ' "$log" | tail -n 1 || true)"
if [[ -z "$summary" ]]; then
  echo "check-mutation-score: go-mutesting produced no mutation score" >&2
  exit 1
fi
echo "check-mutation-score: $summary"

parsed="$(python3 - "$summary" <<'PY'
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
)" || {
  echo "check-mutation-score: could not parse mutation score" >&2
  exit 1
}
read -r killed survived duplicated skipped total <<<"$parsed"

if [[ "$total" -eq 0 || $((killed + survived)) -eq 0 ]]; then
  echo "check-mutation-score: no executed mutant (killed=${killed} survived=${survived} duplicated=${duplicated} skipped=${skipped} total=${total}); refusing to treat that as a score" >&2
  exit 1
fi

if ((survived > allowed)); then
  echo "check-mutation-score: FAIL survived ${survived} is above ceiling ${allowed} (killed=${killed} total=${total})" >&2
  exit 1
fi

echo "check-mutation-score: PASS survived ${survived} (ceiling ${allowed}, killed=${killed}, total=${total})"
