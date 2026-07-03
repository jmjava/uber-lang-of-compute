# KBL Compute Lab (Kind)

Runnable local lab for the full in-cluster stack: **CRDs**, **kbl-controller**, **kbl-tsdb**, **Volcano** batch scheduling, **OpenKruise** hot-swap dominos, and sample finance workloads.

AWS production deployment is scaffolded separately under `infra/aws/cdk/` (see [ADR 0026](../docs/adr/0026-kind-lab-aws-cdk.md)). Volcano integration is in [ADR 0029](../docs/adr/0029-volcano-kind-lab.md); OpenKruise in [ADR 0032](../docs/adr/0032-openkruise-kind-lab.md).

## Prerequisites

- [Docker](https://docs.docker.com/get-docker/)
- [Kind](https://kind.sigs.k8s.io/docs/user/quick-start/#installation)
- [kubectl](https://kubernetes.io/docs/tasks/tools/)
- [Kustomize](https://kubectl.docs.kubernetes.io/installation/kustomize/) (`kubectl kustomize` also works)

Recommended host: **64 GiB RAM / 20 CPU** for the 3-node Kind cluster plus Volcano demo Job.

## Quick start

```bash
chmod +x lab/scripts/*.sh
./lab/scripts/up.sh
```

Tear down:

```bash
./lab/scripts/down.sh
```

Skip Volcano or OpenKruise:

```bash
KBL_LAB_VOLCANO=0 ./lab/scripts/up.sh
KBL_LAB_OPENKURISE=0 ./lab/scripts/up.sh
KBL_LAB_VOLCANO=0 KBL_LAB_OPENKURISE=0 ./lab/scripts/up.sh   # platform only
```

Pin Volcano / OpenKruise versions:

```bash
KBL_VOLCANO_VERSION=v1.9.0 KBL_OPENKURISE_VERSION=1.6.4 ./lab/scripts/up.sh
```

If upgrading from an older single-node lab cluster, delete and recreate:

```bash
./lab/scripts/down.sh && ./lab/scripts/up.sh
```

## What gets deployed

| Component | Namespace | Notes |
|-----------|-----------|-------|
| CRDs | cluster-scoped | `crds/*.yaml` |
| Volcano | `volcano-system` | Scheduler + controllers (unless `KBL_LAB_VOLCANO=0`) |
| OpenKruise | `kruise-system` | Hot-swap controller (unless `KBL_LAB_OPENKURISE=0`) |
| kbl-controller | `kbl-system` | In-cluster operator with RBAC |
| kbl-tsdb | `kbl-system` | Deployment on `kbl.io/tsdb-node=true` worker |
| ComputeContext `default-context` | `default` | Points at TSDB service |
| Workflow `finance-lab` | `default` | 3-step finance chain |
| Queue `kbl-lab` | cluster | Volcano queue (20 CPU / 64 GiB) |
| ComputeWheel `julia-finance-wheel` | `default` | `volcanoQueue: kbl-lab`, `runtime: volcano-init` per time slice |
| DominoChain `julia-finance-openkruise` | `default` | `runtime: openkruise` Julia hot-swap chain |

Images are built locally as `*:lab` and loaded into Kind (`kind load docker-image`).

### Cluster topology

```
control-plane   kbl.io/lab-role=control-plane
worker w1       kbl.io/lab-role=compute
worker w2       kbl.io/lab-role=compute, kbl.io/tsdb-node=true  ← TSDB pinned here
```

## Verify

```bash
kubectl get nodes -L kbl.io/lab-role,kbl.io/tsdb-node
kubectl get workflows finance-lab -o wide
kubectl -n kbl-system get pods -o wide
kubectl get wheel julia-finance-wheel -o wide
kubectl get wf -l kbl.io/computewheel=julia-finance-wheel
kubectl get dchain,vcjob -l kbl.io/volcano-demo=true
kubectl get dchain julia-finance-openkruise -o wide
kubectl get pods -l kbl.io/openkruise-demo=true
kubectl get containerrecreaterequests.apps.kruise.io -l kbl.io/dominochain=julia-finance-openkruise
kubectl -n volcano-system get pods
kubectl -n kruise-system get pods
kubectl -n kbl-system logs deployment/kbl-controller --tail=50
curl -s http://localhost:9090/healthz   # after port-forward
kubectl port-forward -n kbl-system svc/kbl-tsdb 9090:9090
```

Inspect Volcano Job init-chain output:

```bash
POD=$(kubectl get pods -l kbl.io/volcano-demo=true -o jsonpath='{.items[0].metadata.name}')
kubectl logs "$POD" -c slot-2-compute-greeks
```

## Julia / domino-runner

Images `kbl-domino-runner:lab` and `kbl-domino-runner-julia:lab` are loaded for DominoChain workflows:

```bash
kubectl apply -f examples/julia-domino-chain/dominochain-init.yaml
# Edit runnerImage to kbl-domino-runner-julia:lab for Julia chains in Kind
```

The Volcano demo applies a **ComputeWheel** that rotates one time slice and drives Workflow → DominoChain → VCJob ([ADR 0031](../docs/adr/0031-computewheel-volcano-queue.md)).

The OpenKruise demo applies a **DominoChain** with placeholder slots and sequential ContainerRecreateRequests ([ADR 0032](../docs/adr/0032-openkruise-kind-lab.md)).

```bash
kubectl logs -l kbl.io/openkruise-demo=true -c slot-2-compute-greeks
```

## Layout

```
lab/
  kind/kind-config.yaml       # 1 control-plane + 2 workers
  kustomize/base/             # controller + TSDB + RBAC
  kustomize/overlays/kind/    # Kind overlay (TSDB node pin)
  kustomize/overlays/aws/     # ECR image patch overlay
  manifests/                  # lab ComputeContext + Workflow
  manifests/volcano/          # Queue + ComputeWheel volcano-init demo
  manifests/openkruise/     # Julia hot-swap DominoChain demo
  scripts/up.sh | down.sh | install-volcano.sh | install-openkruise.sh
```

## AWS (CDK)

See [infra/aws/cdk/README.md](../infra/aws/cdk/README.md) for the CDK scaffold targeting EKS + ECR. The Kind lab validates manifests and images before AWS deploy.
