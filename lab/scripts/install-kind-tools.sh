#!/usr/bin/env bash
# Install Docker, Kind, kubectl, and kustomize for the compact Kind+Volcano lab.
# Works on systemd hosts and on research VMs whose PID 1 is not systemd
# (dockerd is started directly; nested overlay uses fuse-overlayfs).
set -euo pipefail

KIND_VERSION="${KBL_KIND_VERSION:-v0.27.0}"
KUSTOMIZE_VERSION="${KBL_KUSTOMIZE_VERSION:-v5.6.0}"
DOCKERD_LOG="${KBL_DOCKERD_LOG:-/tmp/kbl-research/dockerd.log}"

need_sudo() {
  if [[ "$(id -u)" -eq 0 ]]; then
    "$@"
  else
    sudo "$@"
  fi
}

root_fstype() {
  findmnt -n -o FSTYPE / 2>/dev/null || stat -f -c %T / 2>/dev/null || echo unknown
}

ensure_docker_sock() {
  if [[ -S /var/run/docker.sock ]]; then
    need_sudo chmod 666 /var/run/docker.sock || true
  fi
}

write_nested_daemon_json() {
  need_sudo mkdir -p /etc/docker
  if [[ -f /etc/docker/daemon.json ]] && grep -q fuse-overlayfs /etc/docker/daemon.json 2>/dev/null; then
    return
  fi
  echo '{"storage-driver":"fuse-overlayfs","iptables":true,"live-restore":false}' \
    | need_sudo tee /etc/docker/daemon.json >/dev/null
}

start_dockerd() {
  if docker info >/dev/null 2>&1; then
    ensure_docker_sock
    return
  fi
  need_sudo service docker start 2>/dev/null || true
  need_sudo systemctl start docker 2>/dev/null || true
  if docker info >/dev/null 2>&1; then
    ensure_docker_sock
    return
  fi

  echo "starting dockerd without systemd..."
  mkdir -p "$(dirname "$DOCKERD_LOG")"
  if [[ "$(root_fstype)" == overlay* ]]; then
    echo "root filesystem is overlay; using fuse-overlayfs storage driver"
    need_sudo apt-get update -y
    need_sudo DEBIAN_FRONTEND=noninteractive apt-get install -y --no-install-recommends fuse-overlayfs iptables
    write_nested_daemon_json
  fi
  ensure_docker_sock
  need_sudo sh -c "nohup dockerd >'${DOCKERD_LOG}' 2>&1 &"
  for _ in $(seq 1 60); do
    if docker info >/dev/null 2>&1; then
      ensure_docker_sock
      return
    fi
    sleep 0.25
  done
  echo "dockerd failed to start; last log:" >&2
  tail -n 80 "$DOCKERD_LOG" >&2 || true
  exit 1
}

install_docker() {
  if command -v docker >/dev/null 2>&1; then
    start_dockerd
    docker info >/dev/null
    echo "docker is ready ($(docker version --format '{{.Server.Version}}' 2>/dev/null || echo ok))"
    return
  fi
  echo "Installing docker.io..."
  need_sudo apt-get update -y
  need_sudo DEBIAN_FRONTEND=noninteractive apt-get install -y --no-install-recommends \
    docker.io ca-certificates curl iptables fuse-overlayfs
  if [[ "$(id -u)" -ne 0 ]]; then
    need_sudo usermod -aG docker "$(id -un)" || true
  fi
  if [[ "$(root_fstype)" == overlay* ]]; then
    write_nested_daemon_json
  fi
  start_dockerd
  docker info >/dev/null
  echo "docker is ready"
}

install_bin() {
  local name="$1" url="$2"
  if command -v "$name" >/dev/null 2>&1; then
    echo "$name already installed: $($name version 2>/dev/null | head -1 || echo ok)"
    return
  fi
  echo "Installing $name from $url"
  tmp="$(mktemp)"
  curl -fsSL "$url" -o "$tmp"
  chmod +x "$tmp"
  need_sudo mv "$tmp" "/usr/local/bin/$name"
}

install_kustomize() {
  if command -v kustomize >/dev/null 2>&1; then
    echo "kustomize already installed: $(kustomize version)"
    return
  fi
  echo "Installing kustomize ${KUSTOMIZE_VERSION}..."
  tmpdir="$(mktemp -d)"
  curl -fsSL "https://github.com/kubernetes-sigs/kustomize/releases/download/kustomize%2F${KUSTOMIZE_VERSION}/kustomize_${KUSTOMIZE_VERSION}_linux_amd64.tar.gz" \
    | tar -xz -C "$tmpdir"
  need_sudo mv "$tmpdir/kustomize" /usr/local/bin/kustomize
  rm -rf "$tmpdir"
}

install_docker

K8S_VER="$(curl -fsSL https://dl.k8s.io/release/stable.txt)"
install_bin kubectl "https://dl.k8s.io/release/${K8S_VER}/bin/linux/amd64/kubectl"
install_bin kind "https://kind.sigs.k8s.io/dl/${KIND_VERSION}/kind-linux-amd64"
install_kustomize

echo "Kind toolchain:"
docker version --format '{{.Server.Version}}'
kind version
kubectl version --client
kustomize version
