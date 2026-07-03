# ADR 0032: OpenKruise Hot-Swap in Kind Lab

## Status

Accepted

## Context

ADR 0007 and Phase 4 implemented `runtime: openkruise` on DominoChain: placeholder Pod slots plus OpenKruise **ContainerRecreateRequest** (CRR) hot-swap per domino step. Examples shipped under `examples/julia-domino-chain/dominochain-openkruise.yaml`, but the Kind lab had no OpenKruise install and no runnable demo — operators could not validate player-piano hot-swap locally.

Phases 25–27 added Volcano to the lab with install scripts and end-to-end demos. OpenKruise is the blog's complementary runtime for in-place domino swaps without init-container chains.

## Decision

### 1. OpenKruise install script

`lab/scripts/install-openkruise.sh`:

- Installs Helm 3 if missing
- `helm upgrade --install kruise openkruise/kruise` pinned to `KBL_OPENKURISE_VERSION` (default `1.6.4`)
- `featureGates=KruiseDaemon=false` and `manager.replicas=1` for Kind resource footprint
- Waits for `kruise-controller-manager` in `kruise-system`

Skip with `KBL_LAB_OPENKURISE=0` in `up.sh`.

### 2. Lab demo manifest

`lab/manifests/openkruise/dominochain-julia-finance.yaml`:

- `DominoChain` `julia-finance-openkruise` with `runtime: openkruise`
- Julia finance chain (`julia:identity` → `julia:interpolate` → `julia:greeks`)
- `runnerImage: kbl-domino-runner-julia:lab`, `nodeSelector: kbl.io/lab-role=compute`

Controller creates placeholder Pod + sequential CRRs; no static manifests beyond the DominoChain CR.

### 3. Lab wiring

`lab/scripts/up.sh` invokes OpenKruise install and applies the demo after the KBL platform (and optional Volcano demo). Waits for `DominoChain.status.phase=Completed`.

## Consequences

- Lab validates both batch (Volcano) and hot-swap (OpenKruise) provisioning paths on the same multi-node Kind cluster
- Helm is required for OpenKruise install (auto-installed by script if absent)
- CRR webhook must be healthy before domino chains reconcile; install script waits for controller-manager
- ComputeWheel + openkruise per time slice remains future work (apply DominoChain or Workflow with `runtime: openkruise` today)

## References

- ADR 0007 — Hot-Swapped Dominos Implementation
- ADR 0025 — Julia Production Wiring (OpenKruise env parity)
- [OpenKruise installation](https://openkruise.io/docs/installation)
- `examples/julia-domino-chain/dominochain-openkruise.yaml`
