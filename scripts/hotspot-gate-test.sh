#!/usr/bin/env bash
# Ranking and merge-base checks for the hotspot gate. No gocyclo install and
# no mutation.
set -euo pipefail

root="$(cd "$(dirname "$0")/.." && pwd)"
# shellcheck source=hotspot-gate.sh
source "$root/scripts/hotspot-gate.sh"

fail() {
  echo "hotspot-gate-test: FAIL $*" >&2
  exit 1
}

assert_eq() {
  local got="$1" want="$2" label="$3"
  if [[ "$got" != "$want" ]]; then
    echo "hotspot-gate-test: FAIL ${label}" >&2
    echo "  got:  ${got}" >&2
    echo "  want: ${want}" >&2
    exit 1
  fi
}

top_of() {
  local k="$1"
  shift
  printf '%s\n' "$@" | top_k_paths "$k"
}

echo "hotspot-gate-test: top-k excludes ties past rank k"
got="$(top_of 10 \
  $'20\ta.go' \
  $'19\tb.go' \
  $'18\tc.go' \
  $'17\td.go' \
  $'16\te.go' \
  $'15\tf.go' \
  $'14\tg.go' \
  $'13\th.go' \
  $'12\ti.go' \
  $'11\tj.go' \
  $'11\tk.go' \
  $'11\tl.go' \
  $'1\tquiet.go')"
want=$'a.go\nb.go\nc.go\nd.go\ne.go\nf.go\ng.go\nh.go\ni.go\nj.go'
assert_eq "$got" "$want" "ties at the cutoff stay outside the top-k set"
printf '%s\n' "$got" | grep -qx 'quiet.go' && fail "quiet.go entered the top-k set"
printf '%s\n' "$got" | grep -qx 'k.go' && fail "tie k.go entered the top-k set"

echo "hotspot-gate-test: count >= 1 is not membership"
got="$(top_of 5 \
  $'9\thot.go' \
  $'1\tq01.go' \
  $'1\tq02.go' \
  $'1\tq03.go' \
  $'1\tq04.go' \
  $'1\tq05.go' \
  $'1\tq06.go' \
  $'1\tq07.go' \
  $'1\tq08.go' \
  $'1\tq09.go' \
  $'1\tq10.go')"
assert_eq "$got" "hot.go" "only the file above the count-1 floor is hot"
printf '%s\n' "$got" | grep -q 'q01.go' && fail "count >= 1 included q01.go"

echo "hotspot-gate-test: a flat history has no hotspot"
got="$(top_of 4 \
  $'1\ta.go' \
  $'1\tb.go' \
  $'1\tc.go' \
  $'1\td.go' \
  $'1\te.go')"
assert_eq "$got" "" "every file at count 1 stays out"

echo "hotspot-gate-test: equal counts above 1 are not a top set"
got="$(top_of 3 \
  $'3\ta.go' \
  $'3\tb.go' \
  $'3\tc.go' \
  $'3\td.go')"
assert_eq "$got" "" "cutoff equal to the floor is not a hotspot signal"

commit_file() {
  local repo="$1" rel="$2" text="$3" msg="$4"
  mkdir -p "$(dirname "$repo/$rel")"
  printf '%s\n' "$text" >"$repo/$rel"
  git -C "$repo" add -- "$rel"
  git -C "$repo" -c user.email=hotspot-test@example.com -c user.name=hotspot-test commit -m "$msg" >/dev/null
}

echo "hotspot-gate-test: branch-only commits do not make a quiet file hot"
work="$(mktemp -d)"
trap 'rm -rf "$work"' EXIT
repo="$work/repo"
fake="$work/bin"
mkdir -p "$fake"
cat >"$fake/gocyclo" <<'EOF'
#!/bin/sh
file=""
for arg in "$@"; do
  file="$arg"
done
case "$file" in
  *hot.go|*quiet.go) printf '20 %s:1:1 Fn\n' "$file" ;;
esac
exit 0
EOF
chmod +x "$fake/gocyclo"

git init -q -b main "$repo"
for i in 1 2 3 4 5 6 7 8 9; do
  commit_file "$repo" "controller/filler$(printf '%02d' "$i").go" "package filler" "filler $i"
  commit_file "$repo" "controller/filler$(printf '%02d' "$i").go" "package filler // $i" "filler $i again"
done
for n in 1 2 3 4 5; do
  commit_file "$repo" "controller/hot.go" "package hot // $n" "hot $n"
  commit_file "$repo" "controller/simple.go" "package simple // $n" "simple $n"
done
commit_file "$repo" "controller/quiet.go" "package quiet" "quiet once"
git -C "$repo" update-ref refs/remotes/origin/main HEAD
git -C "$repo" checkout -q -b feature
for n in 1 2 3 4 5 6 7 8; do
  commit_file "$repo" "controller/quiet.go" "package quiet // branch $n" "quiet branch $n"
done
# One-line edit on top of the branch history.
printf 'package quiet // one line\n' >"$repo/controller/quiet.go"
printf 'package simple // one line\n' >"$repo/controller/simple.go"

PATH="$fake:$PATH" HOTSPOT_ROOT="$repo" "$root/scripts/hotspot-gate.sh" >"$work/quiet.out"
grep -q 'allow controller/quiet.go; outside the top-k change-frequency set' "$work/quiet.out" \
  || fail "one-line change in quiet.go was not allowed"
grep -q 'allow controller/simple.go; in the top-k set and within the complexity cap' "$work/quiet.out" \
  || fail "simple hotspot was not allowed"
if grep -q 'FAIL controller/quiet.go' "$work/quiet.out"; then
  fail "quiet.go failed the gate"
fi

printf 'package hot // changed\n' >"$repo/controller/hot.go"
set +e
PATH="$fake:$PATH" HOTSPOT_ROOT="$repo" "$root/scripts/hotspot-gate.sh" >"$work/hot.out" 2>"$work/hot.err"
status=$?
set -e
if [[ "$status" -eq 0 ]]; then
  fail "complex hotspot was allowed"
fi
grep -q 'FAIL controller/hot.go' "$work/hot.err" || fail "complex hotspot did not fail"
grep -q 'allow controller/quiet.go; outside the top-k change-frequency set' "$work/hot.out" \
  || fail "quiet.go failed once hot.go was also changed"

echo "hotspot-gate-test: pass"
