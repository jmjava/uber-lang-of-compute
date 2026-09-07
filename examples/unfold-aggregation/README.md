# Unfold aggregation chain — depth-1 binary window (Phase 37/38).

Leaves identity a unit snapshot `{v: 1}`; the parent runs `builtin:coarsen`.
Root `v` equals leaf count (2). Unfold does not spawn ComputeWheel contexts.

YAML 1.1 treats `n` as a boolean — quote the parent name (`"n"`) in Kubernetes manifests.

```bash
# CLI
make build
./controller/bin/kbl-compute --workflow examples/unfold-aggregation/workflow.yaml

# Kind lab (controller image must include builtin:coarsen)
kubectl apply -f lab/manifests/workflow-unfold.yaml
kubectl get wf unfold-lab -o wide
kubectl get cm unfold-lab-replay -o jsonpath='{.data.replay\.json}'
```
