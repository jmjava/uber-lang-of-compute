#!/usr/bin/env bash
# Compact Kind cluster + Volcano rates desk-day + Julia wheel + OpenKruise hot-swap.
# Fits a 16 GiB research VM. Full home-lab remains: KBL_LAB_PROFILE=home make lab-up
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
chmod +x "$ROOT/lab/scripts/"*.sh
export KBL_LAB_PROFILE="${KBL_LAB_PROFILE:-compact}"
export KBL_LAB_VOLCANO="${KBL_LAB_VOLCANO:-1}"
export KBL_LAB_OPENKURISE="${KBL_LAB_OPENKURISE:-1}"
export KBL_LAB_JULIA="${KBL_LAB_JULIA:-1}"
bash "$ROOT/lab/scripts/install-kind-tools.sh"
if ! docker info >/dev/null 2>&1; then
  echo "docker is not running after toolchain install" >&2
  exit 1
fi
"$ROOT/lab/scripts/up.sh"
"$ROOT/lab/scripts/verify-volcano.sh" --strict
