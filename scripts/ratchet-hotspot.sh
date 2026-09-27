#!/usr/bin/env bash
# Fail only when a file in the diff is both over the complexity cap and in the
# top git change-frequency set. A change in a quiet file passes even when the
# file is complex. A simple change in a hotspot passes.
set -euo pipefail

root="$(cd "$(dirname "$0")/.." && pwd)"
# shellcheck source=ratchet-common.sh
source "$root/scripts/ratchet-common.sh"
require_cmd gocyclo

# gocyclo reports functions with complexity greater than this value under -over.
complexity_cap=15
# Share of controller Go files, by commit count, treated as the hot set.
top_fraction=10

should_fail() {
  local in_diff="$1" over_cap="$2" in_hot_set="$3"
  [[ "$in_diff" == 1 && "$over_cap" == 1 && "$in_hot_set" == 1 ]]
}

self_check() {
  if should_fail 1 1 0; then
    echo "ratchet-hotspot: self-check failed: a quiet complex file must pass" >&2
    exit 1
  fi
  if should_fail 1 0 1; then
    echo "ratchet-hotspot: self-check failed: a simple hotspot must pass" >&2
    exit 1
  fi
  if should_fail 0 1 1; then
    echo "ratchet-hotspot: self-check failed: an unchanged hotspot must pass" >&2
    exit 1
  fi
  if ! should_fail 1 1 1; then
    echo "ratchet-hotspot: self-check failed: a complex hotspot in the diff must fail" >&2
    exit 1
  fi
}

max_complexity() {
  local file="$1"
  local max=0
  local line n
  while IFS= read -r line; do
    n="${line%% *}"
    if [[ "$n" =~ ^[0-9]+$ ]] && ((n > max)); then
      max=$n
    fi
  done < <(gocyclo "$file" 2>/dev/null || true)
  printf '%s\n' "$max"
}

self_check

cd "$root"
freq="$(mktemp)"
trap 'rm -f "$freq"' EXIT

git log --format=format: --name-only -- controller |
  awk 'NF && $0 ~ /\.go$/ { c[$0]++ } END { for (f in c) print c[f], f }' |
  sort -nr >"$freq"

total="$(wc -l <"$freq" | tr -d ' ')"
if [[ "$total" -eq 0 ]]; then
  echo "ratchet-hotspot: no controller Go history; nothing to gate"
  exit 0
fi

hot_n=$(((total + top_fraction - 1) / top_fraction))
if ((hot_n < 1)); then
  hot_n=1
fi
cutoff="$(awk -v n="$hot_n" 'NR == n { print $1; exit }' "$freq")"
if [[ -z "$cutoff" ]]; then
  echo "ratchet-hotspot: could not compute a change-frequency cutoff" >&2
  exit 1
fi

declare -A hot=()
while read -r count path; do
  [[ -n "${count:-}" && -n "${path:-}" ]] || continue
  if ((count >= cutoff)); then
    hot["$path"]=1
  fi
done <"$freq"

echo "ratchet-hotspot: ${total} controller Go files, top ${top_fraction}% cutoff is ${cutoff} commits (${hot_n} files before ties), complexity cap ${complexity_cap}"

failures=0
while IFS= read -r rel; do
  [[ "$rel" == controller/*.go ]] || continue
  [[ -f "$root/$rel" ]] || continue

  complexity="$(max_complexity "$root/$rel")"
  over=0
  in_hot=0
  if ((complexity > complexity_cap)); then
    over=1
  fi
  if [[ -n "${hot[$rel]:-}" ]]; then
    in_hot=1
  fi

  if should_fail 1 "$over" "$in_hot"; then
    echo "ratchet-hotspot: FAIL $rel complexity ${complexity} > ${complexity_cap} and change frequency is in the top set" >&2
    failures=1
  elif ((over == 1)); then
    echo "ratchet-hotspot: allow $rel complexity ${complexity} > ${complexity_cap}; file is outside the top change-frequency set"
  else
    echo "ratchet-hotspot: allow $rel complexity ${complexity}"
  fi
done < <(changed_files)

if ((failures != 0)); then
  exit 1
fi
echo "ratchet-hotspot: pass"
