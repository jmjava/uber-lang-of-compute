# ADR 0039: OpenKruise ImagePullJob for `runtime: openkruise`

## Status

Accepted — **Phase 36 verified** on compact Kind (`verify-volcano.sh --strict`). `ImagePullJob` then runner-slot Pod.

## Context

Phase 34 installed OpenKruise (`kruise-controller-manager` + `kruise-daemon`) and ran `julia-finance-openkruise` on the compact worker. OpenKruise 1.6 **ContainerRecreateRequest cannot change image or env**, so the chain is a multi-container runner-slot Pod, not pause+CRR.

That path never created an OpenKruise object. `runtime: openkruise` was a vanilla Pod on a cluster that happened to have Kruise installed. Phase 35 catalog work did not change that.

The Julia runner image is large (~1.5 GiB with depot). OpenKruise **ImagePullJob** is the CR that tells `kruise-daemon` to prefetch an image onto selected nodes. It is not a new `kbl.io` kind (ADR 0038).

## Decision

1. **Prefetch, then run.** `DominoChain` with `runtime: openkruise` creates `apps.kruise.io/v1alpha1 ImagePullJob` for each unique runner image, selects nodes from `spec.nodeSelector`, and **does not** create the runner-slot Pod until the job reports `succeeded >= desired`.
2. **Fail closed.** Missing ImagePullJob CRD, or a job that finishes with zero successes, fails the chain. Do not silently fall back to a bare Pod.
3. **Still no CRR swaps.** CRR helpers remain in tree for the 1.6 limitation; the reconciler does not emit them.
4. **No new KBL CRDs.** ImagePullJob stays an OpenKruise object owned by DominoChain.

## Consequences

- `kubectl get imagepulljobs` is the OpenKruise proof for the compact lab, alongside runner-slot logs.
- The lab Helm install enables `KruiseDaemon` and `ImagePullJobGate` (comma in `--set featureGates` must be escaped, or Helm treats `ImagePullJobGate` as a separate key).
- ImagePullJob uses `imagePullPolicy: IfNotPresent` so Kind-loaded images complete without a registry pull.
- Courseforge (Phase 32) and M1–M32 stay untouched.

## References

- ADR 0007 — Hot-Swapped Dominos Implementation
- ADR 0032 — OpenKruise Kind Lab
- ADR 0038 — CRD / Operator Standardization
- [OpenKruise ImagePullJob](https://openkruise.io/docs/user-manuals/imagepulljob)
