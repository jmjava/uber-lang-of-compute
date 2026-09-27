#!/usr/bin/env bash
# A recorded finding passes; a new one fails; deleting the record while the
# finding remains fails. The checker does not rewrite the file.
set -euo pipefail
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
CHECK="${SCRIPT_DIR}/check-static-baseline.py"
TMP="$(mktemp -d)"
trap 'rm -rf "$TMP"' EXIT

cat > "$TMP/.golangci.yml" <<'EOF'
version: "2"

run:
  timeout: 5m

linters:
  default: none
  enable:
    - staticcheck

issues:
  max-issues-per-linter: 0
  max-same-issues: 0
EOF
cat > "$TMP/p.go" <<'EOF'
package p

func id(v int) int { return v }

func f() {
	if id(1) != id(1) {
	}
}
EOF
(
  cd "$TMP"
  go mod init example.com/p >/dev/null 2>&1
)
BASELINE="$TMP/baseline.txt"
printf 'p.go\tSA4000\n' > "$BASELINE"
STAMP="$(cksum "$BASELINE")"

run_check() {
  python3 "$CHECK" --root "$TMP" --module "$TMP" --config "$TMP/.golangci.yml" --baseline "$BASELINE"
}

if ! run_check; then
  echo "expected PASS when the only finding is recorded" >&2
  exit 1
fi
if [[ "$(cksum "$BASELINE")" != "$STAMP" ]]; then
  echo "checker rewrote the baseline" >&2
  exit 1
fi

cat > "$TMP/p.go" <<'EOF'
package p

func id(v int) int { return v }

func f() {
	if id(1) != id(1) {
	}
	if id(2) != id(2) {
	}
}
EOF
if run_check; then
  echo "expected FAIL on a new finding" >&2
  exit 1
fi

printf '' > "$BASELINE"
cat > "$TMP/p.go" <<'EOF'
package p

func id(v int) int { return v }

func f() {
	if id(1) != id(1) {
	}
}
EOF
if run_check; then
  echo "expected FAIL when the record is deleted and the finding remains" >&2
  exit 1
fi

printf 'p.go\tSA4000\n' > "$BASELINE"
cat > "$TMP/p.go" <<'EOF'
package p

func id(v int) int { return v }
EOF
if ! run_check; then
  echo "expected PASS when the finding is gone and the record is stale" >&2
  exit 1
fi

if python3 "$CHECK" --write-baseline >/dev/null 2>&1; then
  echo "expected refusal to rewrite the baseline" >&2
  exit 1
fi

echo "test-check-static-baseline: PASS"
