# KBL AWS CDK (Phase 22 scaffold)

AWS deployment foundation for KBL Compute Engine using [AWS CDK](https://docs.aws.amazon.com/cdk/v2/guide/home.html).

## What this stack creates

| Resource | Purpose |
|----------|---------|
| **VPC** | 2 AZ, public + private subnets, single NAT |
| **ECR repos** | `kbl-controller`, `kbl-tsdb`, `kbl-domino-runner`, `kbl-domino-runner-julia` |
| **EKS cluster** | Kubernetes 1.31, managed node group (t3.medium × 2) |

## Prerequisites

- AWS account + credentials (`aws configure`)
- Node.js 18+
- CDK bootstrap: `cdk bootstrap aws://ACCOUNT/REGION`

## Commands

```bash
cd infra/aws/cdk
npm install
npm run build
npm run synth          # preview CloudFormation
npm run deploy         # deploy KblPlatformStack
```

## After deploy

1. **Build and push images** (from repo root):

```bash
AWS_ACCOUNT=$(aws sts get-caller-identity --query Account --output text)
AWS_REGION=${AWS_DEFAULT_REGION:-us-east-1}
REGISTRY="${AWS_ACCOUNT}.dkr.ecr.${AWS_REGION}.amazonaws.com"

aws ecr get-login-password | docker login --username AWS --password-stdin "$REGISTRY"

docker build -f controller/docker/kbl-controller/Dockerfile -t "$REGISTRY/kbl-controller:latest" .
docker push "$REGISTRY/kbl-controller:latest"
# repeat for kbl-tsdb, domino-runner, domino-runner-julia
```

2. **Configure kubectl**:

```bash
aws eks update-kubeconfig --name $(aws cloudformation describe-stacks \
  --stack-name KblPlatformStack --query "Stacks[0].Outputs[?OutputKey=='ClusterName'].OutputValue" \
  --output text)
```

3. **Install KBL** — apply the same manifests as the Kind lab (`lab/kustomize/base/`), retagging images to ECR URIs via `lab/kustomize/overlays/aws/` (see overlay README).

## Deferred (roadmap)

- MSK for Multiverse Kafka sync
- FSx / S3 data staging for path snapshots
- IRSA for controller + TSDB service accounts
- Helm chart packaging
- `cdk8s` or EKS add-ons for node-local TSDB DaemonSet on EC2

## Local lab first

Validate with Kind before AWS spend:

```bash
./lab/scripts/up.sh
```

See [lab/README.md](../../../lab/README.md) and [ADR 0026](../../../docs/adr/0026-kind-lab-aws-cdk.md).
