# AWS overlay

Patches `lab/kustomize/base/` image tags from Kind-local `*:lab` to ECR URIs.

After `cdk deploy` (see [infra/aws/cdk/README.md](../../../../infra/aws/cdk/README.md)):

```bash
export KBL_ECR_REGISTRY="$(aws sts get-caller-identity --query Account --output text).dkr.ecr.${AWS_DEFAULT_REGION:-us-east-1}.amazonaws.com"

cd lab/kustomize/overlays/aws
kustomize edit set image \
  "kbl-controller=${KBL_ECR_REGISTRY}/kbl-controller:latest" \
  "kbl-tsdb=${KBL_ECR_REGISTRY}/kbl-tsdb:latest"

kubectl apply -f ../../../crds/
kustomize build . | kubectl apply -f -
```

Apply lab manifests from `lab/manifests/` the same way as Kind.
