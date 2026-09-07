#!/usr/bin/env bash
# Compact Kind cluster + Volcano + builtin finance wheel (no Julia, no OpenKruise).
# Fits a 16 GiB research VM. Full home-lab remains: KBL_LAB_PROFILE=home make lab-up
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
chmod +x "$ROOT/lab/scripts/"*.sh
export KBL_LAB_PROFILE="${KBL_LAB_PROFILE:-compact}"
export KBL_LAB_VOLCANO="${KBL_LAB_VOLCANO:-1}"
export KBL_LAB_OPENKURISE=0
export KBL_LAB_JULIA="${KBL_LAB_JULIA:-0}"
bash "$ROOT/lab/scripts/install-kind-tools.sh"
if ! docker info >/dev/null 2>&1; then
  echo "docker is not running after toolchain install" >&2
  exit 1
fi
"$ROOT/lab/scripts/up.sh"
"$ROOT/lab/scripts/verify-volcano.sh" --strict
