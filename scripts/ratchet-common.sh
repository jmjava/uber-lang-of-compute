#!/usr/bin/env bash
# Shared helpers for maintainability ratchets. CI compares; it does not rewrite baselines.
set -euo pipefail

repo_root() {
  git rev-parse --show-toplevel
}

base_rev() {
  local base="${RATCHET_BASE:-origin/main}"
  if git rev-parse --verify --quiet "$base" >/dev/null; then
    printf '%s\n' "$base"
    return
  fi
  if git rev-parse --verify --quiet main >/dev/null; then
    printf '%s\n' main
    return
  fi
  echo "ratchet: cannot find base revision origin/main or main" >&2
  exit 1
}

require_cmd() {
  if ! command -v "$1" >/dev/null 2>&1; then
    echo "ratchet: missing command: $1" >&2
    exit 1
  fi
}

require_golangci() {
  require_cmd golangci-lint
  local got
  got="$(golangci-lint version 2>/dev/null | head -n 1 || true)"
  if [[ "$got" != *"2.13.2"* ]]; then
    echo "ratchet: golangci-lint 2.13.2 is required so the baseline stays stable (got: ${got:-unknown})" >&2
    exit 1
  fi
}

changed_files() {
  local root base
  root="$(repo_root)"
  base="$(base_rev)"
  {
    git -C "$root" diff --name-only --diff-filter=ACMR "$base" --
    git -C "$root" diff --name-only --diff-filter=ACMR HEAD --
    git -C "$root" diff --name-only --cached --diff-filter=ACMR --
    git -C "$root" ls-files --others --exclude-standard
  } | awk 'NF && !seen[$0]++'
}
