# ADR 0026: Kind Lab and AWS CDK Deployment Foundation

## Status

Accepted

## Context

KBL Compute Engine grew through Phases 1–21 with CRDs, examples, and Dockerfiles for domino runners, but **no runnable in-cluster lab**: the controller ran as a local binary, RBAC was not shipped, and TSDB/controller images lacked Docker build paths. Production target is **AWS** with infrastructure-as-code.

## Decision

### 1. Kind lab (`lab/`)

Runnable local stack:

- **Kind cluster** config with `/var/kbl` node mount
- **Kustomize** base: namespace, RBAC, `kbl-controller` Deployment, `kbl-tsdb` Deployment + Service
- **Scripts** `lab/scripts/up.sh` and `down.sh` — build images, load into Kind, apply CRDs + platform + sample workflow
- **Lab manifests** — `default-context` ComputeContext + `finance-lab` Workflow using TSDB service DNS

Kind uses ClusterIP TSDB (not hostNetwork) so in-cluster controller and workflows share a stable service endpoint.

### 2. Docker images

| Dockerfile | Image |
|------------|-------|
| `controller/docker/kbl-controller/Dockerfile` | Operator |
| `controller/docker/kbl-tsdb/Dockerfile` | Node-local store server |
| existing domino-runner Dockerfiles | Chain execution |

### 3. AWS CDK scaffold (`infra/aws/cdk/`)

TypeScript CDK app `KblPlatformStack`:

- VPC (2 AZ)
- ECR repositories for all four images
- EKS cluster + small managed node group

MSK, FSx, IRSA, and Helm install deferred to Phase 22+.

### 4. Path parity

Kind lab manifests are the reference for AWS/EKS deploy; `lab/kustomize/overlays/aws/` patches image URIs to ECR.

## Consequences

- Operators can validate end-to-end without AWS cost
- CDK synth/deploy requires AWS credentials; EKS creates billable resources
- Production TSDB may remain DaemonSet + hostPath on EC2; Kind uses Deployment for simplicity
- `deploy/node-local-tsdb/daemonset.yaml` remains valid for bare-metal/node-local production; lab overlay is Kind-specific

## References

- ADR 0008 — Node-Local TSDB
- ADR 0005 — Kubernetes Controller
- `lab/README.md`
- `infra/aws/cdk/README.md`
