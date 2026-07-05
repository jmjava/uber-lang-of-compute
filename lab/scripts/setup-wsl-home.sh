#!/usr/bin/env bash
# One-shot KBL Kind lab setup for Windows 11 + WSL2 (i9 home profile).
#
# Requires BROAD_REPO_TOKEN for private jmjava / courseforge GitHub repos.
#
# Run inside Ubuntu WSL (not PowerShell):
#   export BROAD_REPO_TOKEN='ghp_...'
#   ./lab/scripts/setup-wsl-home.sh --install-deps --clone
#
# Options:
#   --install-deps       Install kind, kubectl, kustomize (Linux amd64) if missing
#   --check-only         Prerequisites check only; do not start the lab
#   --skip-up            Prepare env but do not run lab/scripts/up.sh
#   --clone              Clone uber-lang-of-compute (uses BROAD_REPO_TOKEN)
#   --with-courseforge   Also clone courseforge/course-builder + infrastructure
#
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=lib-broad-git.sh
source "${SCRIPT_DIR}/lib-broad-git.sh"

INSTALL_DEPS=0
CHECK_ONLY=0
SKIP_UP=0
DO_CLONE=0
WITH_COURSEFORGE=0
KBL_REPO_SLUG="${KBL_REPO_SLUG:-jmjava/uber-lang-of-compute}"
CLONE_DIR="${KBL_CLONE_DIR:-$HOME/src/uber-lang-of-compute}"
COURSEFORGE_DIR="${COURSEFORGE_DIR:-$HOME/src/course-builder}"
INFRA_DIR="${COURSEFORGE_INFRA_DIR:-$HOME/src/courseforge-infrastructure}"

usage() {
  sed -n '2,16p' "$0" | sed 's/^# \{0,1\}//'
  exit "${1:-0}"
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --install-deps) INSTALL_DEPS=1 ;;
    --check-only) CHECK_ONLY=1 ;;
    --skip-up) SKIP_UP=1 ;;
    --clone) DO_CLONE=1 ;;
    --with-courseforge) WITH_COURSEFORGE=1; DO_CLONE=1 ;;
    -h|--help) usage 0 ;;
    *) echo "unknown option: $1" >&2; usage 1 ;;
  esac
  shift
done

info() { echo "==> $*"; }
warn() { echo "warning: $*" >&2; }
die() { echo "error: $*" >&2; exit 1; }

is_wsl() {
  grep -qiE 'microsoft|wsl' /proc/version 2>/dev/null
}

find_repo_root() {
  local dir="$PWD"
  while [[ "$dir" != "/" ]]; do
    if [[ -f "$dir/lab/scripts/up.sh" ]]; then
      echo "$dir"
      return 0
    fi
    dir="$(dirname "$dir")"
  done
  return 1
}

install_kind() {
  local ver="${KIND_VERSION:-v0.26.0}"
  info "Installing kind ${ver}..."
  curl -fsSL "https://kind.sigs.k8s.io/dl/${ver}/kind-linux-amd64" -o /tmp/kind
  chmod +x /tmp/kind
  sudo install -m 0755 /tmp/kind /usr/local/bin/kind
}

install_kubectl() {
  local ver="${KUBECTL_VERSION:-v1.31.0}"
  info "Installing kubectl ${ver}..."
  curl -fsSL "https://dl.k8s.io/release/${ver}/bin/linux/amd64/kubectl" -o /tmp/kubectl
  chmod +x /tmp/kubectl
  sudo install -m 0755 /tmp/kubectl /usr/local/bin/kubectl
}

install_kustomize() {
  local ver="${KUSTOMIZE_VERSION:-v5.5.0}"
  info "Installing kustomize ${ver}..."
  curl -fsSL "https://github.com/kubernetes-sigs/kustomize/releases/download/kustomize%2F${ver#v}/kustomize_${ver#v}_linux_amd64.tar.gz" \
    | tar -xz -C /tmp
  sudo install -m 0755 /tmp/kustomize /usr/local/bin/kustomize
}

ensure_tool() {
  local bin="$1"
  local installer="$2"
  if command -v "$bin" >/dev/null 2>&1; then
    return 0
  fi
  if [[ "$INSTALL_DEPS" == "1" ]]; then
    "$installer"
    return 0
  fi
  die "$bin not found. Re-run with --install-deps or install manually."
}

check_docker() {
  if ! command -v docker >/dev/null 2>&1; then
    die "docker not in PATH. Install Docker Desktop for Windows and enable WSL integration for this distro."
  fi
  if ! docker info >/dev/null 2>&1; then
    cat >&2 <<'EOF'
error: docker daemon not reachable from WSL.

Windows 11 checklist:
  1. Install Docker Desktop (WSL2 backend).
  2. Settings → Resources → assign at least 24 GiB RAM and 8 CPUs (i9 home lab).
  3. Settings → Resources → WSL Integration → enable your Ubuntu distro.
  4. Restart WSL: in PowerShell run  wsl --shutdown  then reopen Ubuntu.
EOF
    exit 1
  fi
}

check_wsl_environment() {
  if is_wsl; then
    info "WSL environment detected"
  else
    warn "not running under WSL — script is tuned for Windows 11 + WSL2"
  fi

  case "$PWD" in
    /mnt/*)
      warn "You are on a Windows drive (/mnt/c/...). Use ~/src for faster Docker/Kind I/O."
      ;;
  esac
}

verify_github_access() {
  require_broad_repo_token || exit 1
  info "GitHub API (BROAD_REPO_TOKEN): courseforge/course-builder"
  if ! GH_TOKEN="$BROAD_REPO_TOKEN" gh repo view courseforge/course-builder --json name -q .name >/dev/null 2>&1; then
    die "BROAD_REPO_TOKEN cannot read courseforge/course-builder — check PAT scopes (repo org access)"
  fi
  info "GitHub API: ${KBL_REPO_SLUG}"
  if ! GH_TOKEN="$BROAD_REPO_TOKEN" gh repo view "$KBL_REPO_SLUG" --json name -q .name >/dev/null 2>&1; then
    die "BROAD_REPO_TOKEN cannot read ${KBL_REPO_SLUG}"
  fi
}

clone_repos() {
  require_broad_repo_token || exit 1
  git_clone_broad "$KBL_REPO_SLUG" "$CLONE_DIR"
  if [[ "$WITH_COURSEFORGE" == "1" ]]; then
    git_clone_broad "courseforge/course-builder" "$COURSEFORGE_DIR"
    git_clone_broad "courseforge/infrastructure" "$INFRA_DIR"
    info "Courseforge workers: $COURSEFORGE_DIR/tools/automation-workers/"
  fi
}

print_docker_hints() {
  cat <<'EOF'

Docker Desktop (Windows) — recommended for i9 home lab:
  • Resources → Memory: 24–48 GiB
  • Resources → CPUs: 8–16
  • Resources → WSL Integration: ON for this distro

EOF
}

main() {
  check_wsl_environment

  if [[ "$DO_CLONE" == "1" || "$WITH_COURSEFORGE" == "1" ]]; then
    verify_github_access
    clone_repos
  fi

  ROOT=""
  if ROOT="$(find_repo_root)"; then
    info "Using repo at $ROOT"
  elif [[ -d "$CLONE_DIR/lab/scripts/up.sh" ]] || [[ -f "$CLONE_DIR/lab/scripts/up.sh" ]]; then
    ROOT="$CLONE_DIR"
    info "Using cloned repo at $ROOT"
  elif [[ "$DO_CLONE" == "1" ]]; then
    ROOT="$CLONE_DIR"
  else
    die "Not inside uber-lang-of-compute. Pass --clone (requires BROAD_REPO_TOKEN)"
  fi

  cd "$ROOT"

  if [[ "$INSTALL_DEPS" == "1" ]]; then
    sudo apt-get update -qq
    sudo apt-get install -y -qq curl git make ca-certificates gh >/dev/null 2>&1 || true
  fi

  command -v gh >/dev/null 2>&1 || warn "gh CLI not installed (optional; apt install gh or github.com/cli/cli)"

  check_docker
  ensure_tool kind install_kind
  ensure_tool kubectl install_kubectl
  ensure_tool kustomize install_kustomize
  command -v make >/dev/null 2>&1 || die "make not found (sudo apt install build-essential)"

  if [[ "$DO_CLONE" == "1" || "$WITH_COURSEFORGE" == "1" ]]; then
    verify_github_access
  fi

  info "Tool versions:"
  docker version --format '  docker: {{.Server.Version}}' 2>/dev/null || docker --version
  kind version 2>/dev/null | sed 's/^/  /' || true
  kubectl version --client --short 2>/dev/null | sed 's/^/  /' || kubectl version --client 2>/dev/null | head -1 | sed 's/^/  /'
  kustomize version 2>/dev/null | sed 's/^/  /' || true

  print_docker_hints

  mkdir -p /tmp/kbl-lab/cp /tmp/kbl-lab/w1 /tmp/kbl-lab/w2
  chmod +x lab/scripts/*.sh 2>/dev/null || true

  export KBL_LAB_PROFILE="${KBL_LAB_PROFILE:-home}"
  export KBL_KIND_CLUSTER="${KBL_KIND_CLUSTER:-kbl-lab}"

  info "Profile: $KBL_LAB_PROFILE (i9 home — 2 workers, Volcano Ferris wheel + burst)"

  if [[ "$CHECK_ONLY" == "1" ]]; then
    info "Prerequisites OK (--check-only)"
    exit 0
  fi

  if [[ "$SKIP_UP" == "1" ]]; then
    info "Skipping lab up (--skip-up). Start manually: make lab-up"
    exit 0
  fi

  info "Starting Kind lab (first run builds images — may take 15–30+ minutes)..."
  make lab-up

  info "Verifying Volcano demo..."
  make lab-verify-volcano || true

  cat <<EOF

Lab is up on WSL.

  kubectl get nodes -L kbl.io/lab-role,kbl.io/tsdb-node,kbl.io/gpu
  ./lab/scripts/verify-volcano.sh
  kubectl get wheel julia-finance-wheel -o wide

Courseforge (if cloned with --with-courseforge):
  cd ${COURSEFORGE_DIR}/tools/automation-workers && docker compose config

Remote kubectl from i7 laptop:
  kind get kubeconfig --name ${KBL_KIND_CLUSTER} > ~/kbl-lab-kubeconfig.yaml

Tear down:
  make lab-down

EOF
}

main "$@"
