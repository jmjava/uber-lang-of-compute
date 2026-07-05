# shellcheck shell=bash
# Shared GitHub clone/auth helpers for private courseforge + jmjava repos.
# Source from lab scripts:  source "$(dirname "$0")/lib-broad-git.sh"

_broad_token_file() {
  echo "${BROAD_REPO_TOKEN_FILE:-$HOME/.config/courseforge/broad-repo-token}"
}

load_broad_repo_token() {
  if [[ -n "${BROAD_REPO_TOKEN:-}" ]]; then
    export GH_TOKEN="${GH_TOKEN:-$BROAD_REPO_TOKEN}"
    export GITHUB_TOKEN="${GITHUB_TOKEN:-$BROAD_REPO_TOKEN}"
    return 0
  fi
  local f
  f="$(_broad_token_file)"
  if [[ -f "$f" ]]; then
    BROAD_REPO_TOKEN="$(tr -d '[:space:]' < "$f")"
    export BROAD_REPO_TOKEN
    export GH_TOKEN="${GH_TOKEN:-$BROAD_REPO_TOKEN}"
    export GITHUB_TOKEN="${GITHUB_TOKEN:-$BROAD_REPO_TOKEN}"
    return 0
  fi
  return 1
}

require_broad_repo_token() {
  if load_broad_repo_token; then
    return 0
  fi
  cat >&2 <<'EOF'
error: BROAD_REPO_TOKEN is required for private jmjava / courseforge repos.

  export BROAD_REPO_TOKEN='ghp_...'    # suite PAT (same as infrastructure publish)

  Or save once (chmod 600):
  mkdir -p ~/.config/courseforge
  echo 'ghp_...' > ~/.config/courseforge/broad-repo-token
  chmod 600 ~/.config/courseforge/broad-repo-token

EOF
  return 1
}

# git_clone_broad <github.com/org/repo> <destination>
git_clone_broad() {
  local slug="$1"
  local dest="$2"
  require_broad_repo_token || return 1
  slug="${slug#https://github.com/}"
  slug="${slug%.git}"
  local url="https://x-access-token:${BROAD_REPO_TOKEN}@github.com/${slug}.git"
  if [[ -d "$dest/.git" ]]; then
    echo "==> already cloned: $dest"
    return 0
  fi
  mkdir -p "$(dirname "$dest")"
  echo "==> cloning ${slug} → ${dest}"
  git clone --depth 1 "$url" "$dest"
}
