# KBL Compute Lab (Kind)

Runnable local lab for the full in-cluster stack: **CRDs**, **kbl-controller**, **kbl-tsdb**, and a sample finance **Workflow**.

AWS production deployment is scaffolded separately under `infra/aws/cdk/` (see [ADR 0026](../docs/adr/0026-kind-lab-aws-cdk.md)).

## Prerequisites

- [Docker](https://docs.docker.com/get-docker/)
- [Kind](https://kind.sigs.k8s.io/docs/user/quick-start/#installation)
- [kubectl](https://kubernetes.io/docs/tasks/tools/)
- [Kustomize](https://kubectl.docs.kubernetes.io/installation/kustomize/) (`kubectl kustomize` also works)

## Quick start

```bash
chmod +x lab/scripts/*.sh
./lab/scripts/up.sh
```

Tear down:

```bash
./lab/scripts/down.sh
```

## What gets deployed

| Component | Namespace | Notes |
|-----------|-----------|-------|
| CRDs | cluster-scoped | `crds/*.yaml` |
| kbl-controller | `kbl-system` | In-cluster operator with RBAC |
| kbl-tsdb | `kbl-system` | Deployment + ClusterIP Service `:9090` |
| ComputeContext `default-context` | `default` | Points at TSDB service |
| Workflow `finance-lab` | `default` | 3-step finance chain |

Images are built locally as `*:lab` and loaded into Kind (`kind load docker-image`).

## Verify

```bash
kubectl get workflows finance-lab -o wide
kubectl -n kbl-system logs deployment/kbl-controller --tail=50
curl -s http://localhost:9090/healthz   # only if you port-forward TSDB
kubectl port-forward -n kbl-system svc/kbl-tsdb 9090:9090
```

## Julia / domino-runner

Images `kbl-domino-runner:lab` and `kbl-domino-runner-julia:lab` are loaded for DominoChain workflows:

```bash
kubectl apply -f examples/julia-domino-chain/dominochain-init.yaml
# Edit runnerImage to kbl-domino-runner-julia:lab for Julia chains in Kind
```

## Layout

```
lab/
  kind/kind-config.yaml       # Kind cluster definition
  kustomize/base/             # controller + TSDB + RBAC
  kustomize/overlays/kind/    # Kind overlay
  kustomize/overlays/aws/     # ECR image patch overlay
  manifests/                  # lab ComputeContext + Workflow
  scripts/up.sh | down.sh
```

## AWS (CDK)

See [infra/aws/cdk/README.md](../infra/aws/cdk/README.md) for the CDK scaffold targeting EKS + ECR. The Kind lab validates manifests and images before AWS deploy.
