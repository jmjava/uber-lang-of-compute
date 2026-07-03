# Getting Started

Three paths into the KBL Compute Engine, from fastest local validation to full in-cluster lab.

## Prerequisites

| Path | Requires |
|------|----------|
| CLI only | Go 1.22+, `make build` |
| Julia examples | Julia 1.10+, `julia --project=controller/julia -e 'using Pkg; Pkg.instantiate()'` |
| Kind lab | Docker, Kind, kubectl, Kustomize (~64 GiB RAM / 20 CPU recommended) |

---

## Path 1: CLI (5 minutes)

Prove snapshot isolation, memoization, and replay logging without Kubernetes.

```bash
make build

./controller/bin/kbl-compute \
  --workflow examples/finance-curve-snapshot/workflow.yaml \
  --replay-log /tmp/replay.json

# Second run — dominos reused from memo cache
./controller/bin/kbl-compute \
  --workflow examples/finance-curve-snapshot/workflow.yaml \
  --replay-log /tmp/replay-2.json
```

Inspect the replay log for `reused: true` on the second run.

### Julia finance chain (local)

```bash
julia --project=controller/julia -e 'using Pkg; Pkg.instantiate()'
./controller/bin/kbl-compute --workflow examples/julia-domino-chain/workflow.yaml
```

See [examples/julia-domino-chain/README.md](../examples/julia-domino-chain/README.md).

---

## Path 2: Kind lab (full stack)

Runnable in-cluster stack: CRDs, controller, TSDB, Volcano, OpenKruise, and demo workloads.

```bash
chmod +x lab/scripts/*.sh
make lab-up          # or: ./lab/scripts/up.sh
```

### Verify platform

```bash
kubectl get nodes -L kbl.io/lab-role,kbl.io/tsdb-node
kubectl get workflows finance-lab -o wide
kubectl -n kbl-system get pods
```

### Verify Volcano path (Phase 25–27)

```bash
kubectl get wheel julia-finance-wheel -o wide
kubectl get wf -l kbl.io/computewheel=julia-finance-wheel
kubectl get vcjob -l kbl.io/volcano-demo=true
```

### Verify OpenKruise path (Phase 28)

```bash
kubectl get dchain julia-finance-openkruise -o wide
kubectl get containerrecreaterequests.apps.kruise.io \
  -l kbl.io/dominochain=julia-finance-openkruise
kubectl logs -l kbl.io/openkruise-demo=true -c slot-2-compute-greeks
```

### Tear down

```bash
make lab-down
```

Detailed lab options: [lab/README.md](../lab/README.md).

---

## Path 3: Manual Kubernetes (single cluster)

For clusters without the Kind lab scripts:

```bash
kubectl apply -f crds/
kubectl apply -f examples/finance-curve-snapshot/workflow-crd.yaml

# Build and deploy controller + TSDB via kustomize, or run locally:
make build
./controller/bin/kbl-controller --store-root /tmp/kbl-store

kubectl get workflows -o wide
```

For container runtimes (`kubernetes-init`, `openkruise`, `volcano-init`), apply a `DominoChain` or a `Workflow` with `spec.execution.runtime` set. See [Provisioning Runtimes](provisioning-runtimes.md).

---

## Choose a provisioning runtime

| Runtime | When to use | Lab demo |
|---------|-------------|----------|
| `local` (default) | Dev, CI, no cluster | `Workflow/finance-lab` |
| `kubernetes-init` | Standard K8s, init-container chain | `dominochain-init.yaml` |
| `openkruise` | Hot-swap slots, player-piano | `DominoChain/julia-finance-openkruise` |
| `volcano-init` | Batch scheduler, gang scheduling | `ComputeWheel/julia-finance-wheel` |

Full comparison: [provisioning-runtimes.md](provisioning-runtimes.md).

---

## Next steps

- [Architecture](architecture.md) — system layers and data flow
- [Documentation index](README.md) — ADRs, examples, phase map
- [Compute Wheel example](../examples/compute-wheel/README.md) — time-slice rotation
- [AWS CDK scaffold](../infra/aws/cdk/README.md) — production path
