#!/usr/bin/env bash
# Fail only when a changed controller Go file is both complex and in the top
# change-frequency set.
#
# Frequency is the number of commits reachable from the merge base with
# origin/main. Commits that exist only on this branch are not counted.
# The set is the first k files by that count (k = ceil(10% of ranked files)).
# Rank is the membership test: files past rank k stay out, including ties at
# the cutoff. A count of 1 is not membership. When the k-th count is no higher
# than the least-changed file, only files inside the top k that sit strictly
# above that floor are included, so a flat history does not flag every file.
# A change in a quiet file passes even when the file is complex.
set -euo pipefail

hotspot_percent=10
complexity_cap=15

hotspot_root() {
  if [[ -n "${HOTSPOT_ROOT:-}" ]]; then
    (cd "$HOTSPOT_ROOT" && pwd)
  else
    (cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)
  fi
}

merge_base_rev() {
  local root="$1"
  local base="${HOTSPOT_BASE:-origin/main}"
  if ! git -C "$root" rev-parse --verify --quiet "${base}^{commit}" >/dev/null; then
    echo "hotspot-gate: cannot find ${base}" >&2
    return 1
  fi
  git -C "$root" merge-base HEAD "$base"
}

# Print "count<TAB>path" for controller Go files present at rev.
# History is `git log rev`, so commits after the merge base are ignored.
churn_counts() {
  local root="$1"
  local rev="$2"
  awk '
    NR == FNR {
      if ($0 ~ /^controller\/.*\.go$/) present[$0] = 1
      next
    }
    NF && ($0 in present) { counts[$0]++ }
    END {
      for (path in counts) printf "%d\t%s\n", counts[path], path
    }
  ' <(git -C "$root" ls-tree -r --name-only "$rev" -- controller) \
    <(git -C "$root" log "$rev" --pretty=format: --name-only -- controller)
}

# stdin: "count<TAB>path"
# arg: optional explicit k (tests). Otherwise k = ceil(n * hotspot_percent / 100).
# stdout: hot paths, one per line, top-k only.
top_k_paths() {
  local k_arg="${1:-}"
  sort -t $'\t' -k1,1nr -k2,2 | awk -F '\t' -v k_arg="$k_arg" -v percent="$hotspot_percent" '
    BEGIN { n = 0 }
    $1 !~ /^[0-9]+$/ || $2 == "" { next }
    {
      n++
      counts[n] = $1 + 0
      paths[n] = $2
    }
    END {
      if (n == 0) exit
      if (k_arg ~ /^[0-9]+$/ && k_arg + 0 >= 1) k = k_arg + 0
      else k = int((n * percent + 99) / 100)
      if (k < 1) k = 1
      if (k > n) k = n
      cutoff = counts[k]
      floor = counts[n]
      for (i = 1; i <= k; i++) {
        # Top-k only. Do not keep walking because count >= cutoff, and do not
        # treat count >= 1 as membership. A cutoff that sits on the floor has
        # no frequency signal past files that are strictly hotter.
        if (cutoff <= floor) {
          if (counts[i] > floor) print paths[i]
        } else {
          print paths[i]
        }
      }
    }
  '
}

changed_go() {
  local root="$1"
  local base="$2"
  {
    git -C "$root" diff --name-only --diff-filter=ACMR "$base" HEAD
    git -C "$root" diff --name-only --diff-filter=ACMR HEAD
    git -C "$root" ls-files --others --exclude-standard
  } | awk '/^controller\/.*\.go$/' | sort -u
}

file_is_complex() {
  local path="$1"
  local out
  out="$(gocyclo -over "$complexity_cap" "$path" 2>/dev/null || true)"
  [[ -n "$out" ]]
}

main() {
  local root base hot_file rel in_hot failures need_cyclo
  root="$(hotspot_root)"
  base="$(merge_base_rev "$root")"
  hot_file="$(mktemp)"
  trap 'rm -f "$hot_file"' RETURN

  churn_counts "$root" "$base" | top_k_paths >"$hot_file"
  local hot_n
  hot_n="$(wc -l <"$hot_file" | tr -d ' ')"
  echo "hotspot-gate: merge-base ${base} top-k size ${hot_n} (cap complexity > ${complexity_cap})"

  failures=0
  need_cyclo=0
  while IFS= read -r rel; do
    [[ -n "$rel" ]] || continue
    if grep -Fxq -- "$rel" "$hot_file"; then
      need_cyclo=1
      break
    fi
  done < <(changed_go "$root" "$base")

  if ((need_cyclo == 1)) && ! command -v gocyclo >/dev/null 2>&1; then
    echo "hotspot-gate: missing gocyclo (CI installs github.com/fzipp/gocyclo/cmd/gocyclo@v0.6.0)" >&2
    return 1
  fi

  local saw=0
  while IFS= read -r rel; do
    [[ -n "$rel" ]] || continue
    [[ -f "$root/$rel" ]] || continue
    saw=1
    in_hot=0
    if grep -Fxq -- "$rel" "$hot_file"; then
      in_hot=1
    fi
    if ((in_hot == 1)) && file_is_complex "$root/$rel"; then
      echo "hotspot-gate: FAIL ${rel} complexity > ${complexity_cap} and change frequency is in the top-k set" >&2
      failures=1
    elif ((in_hot == 1)); then
      echo "hotspot-gate: allow ${rel}; in the top-k set and within the complexity cap"
    else
      echo "hotspot-gate: allow ${rel}; outside the top-k change-frequency set"
    fi
  done < <(changed_go "$root" "$base")

  if ((saw == 0)); then
    echo "hotspot-gate: no changed controller Go files"
  fi
  if ((failures != 0)); then
    return 1
  fi
  echo "hotspot-gate: pass"
}

if [[ "${BASH_SOURCE[0]}" == "$0" ]]; then
  main "$@"
fi
