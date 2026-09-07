# ADR 0038: CRD / Operator Standardization (after Julia + OpenKruise)

## Status

Proposed — **Phase 35 verified** on compact Kind (`verify-volcano.sh --strict`). Catalog Snapshot/Domino CRs, PluggableUniverse fills `runnerImage`, fail-closed julia runner + unsealed snapshots.

## Context

`kbl-controller` already reconciles every `kbl.io` kind: Snapshot, Domino, Workflow, DominoChain, ComputeContext, ComputeWheel, PluggableUniverse, Multiverse, ReadReplica.

Volcano and OpenKruise are **provisioning backends** of `DominoChain.spec.runtime` (ADR 0030, ADR 0032), not missing KBL APIs. Compact-lab wheels still **inline** instruments and `julia:*` steps. PluggableUniverse records `executionEngine` / `dataLayer` but does not yet select runner image or runtime for child chains.

Adding `KblVolcanoJob`, `KblCRR`, book, queue, or replay-log CRDs would split the four-DSL model (ADR 0001). Theory milestones **M1–M32 stay frozen**; this is an engineering phase, not M33.

## Decision

### Phase 34 (prerequisite — current lab work)

Compact Kind cluster runs, on one worker:

- `rates-desk-wheel` (desk-day book)
- `julia-finance-wheel` (`volcano-init`, Julia runner)
- `julia-finance-openkruise` (`runtime: openkruise`)
- OpenKruise `kruise-controller-manager` ready

No new CRDs in Phase 34.

### Phase 35 (this ADR — after Phase 34)

Standardize operator work **on the existing kinds**:

1. **Catalog, not inline.** Sealed `Snapshot` CRs plus reusable `Domino` CRs (`julia:identity`, `julia:interpolate`, `julia:greeks`, builtin twins). Wheels and the OpenKruise demo **reference** them (Phase 10/13 pattern) instead of copying instrument lists and step lists.
2. **Thicken PluggableUniverse.** Reconciler selects `runtimeImage` and default DominoChain runtime for child work (`julia` vs `builtin`). Status-only stamps are not enough.
3. **Fail closed in admission or reconcile.** Volcano job names/labels ≤ 63 chars; julia commands require the Julia runner image; unsealed snapshots cannot run.
4. **Non-goals.** No new KBL CRDs for Volcano Jobs, OpenKruise CRR, queues, books, or replay logs. Do not install Argo Workflows, Tekton, Airflow, Spark Operator, Kueue, JobSet-as-a-second-batch-layer, or Knative.

### Later (same theme, not a new theory milestone)

Adopt **without wrapping** as `kbl.io` kinds:

| When | Operator / CRD | Why |
|------|----------------|-----|
| Webhooks need certs | cert-manager | Validating/conversion webhooks |
| CDC leaves MemoryBus | Strimzi + real Debezium | ADR 0012 envelopes already exist |
| Home-lab CUDA dominos | NVIDIA GPU Operator | Compact Kind does not need this |
| Wheel/queue SLOs | Prometheus Operator | Ops, not compute |
| ComputeContext hardcoded Service blocks | optional Data-DSL store claim (`NodeStore`) | Only if the TSDB Deployment pin is the bottleneck |

## Consequences

- Operators describe work as Snapshot + Domino + Wheel/Chain + Universe, same objects on NY, LN, Julia, and OpenKruise.
- Runtime diversity stays a field on DominoChain / PluggableUniverse, not a growing CRD zoo.
- Courseforge (Phase 32) and physics correspondence (Phase 33 / M1–M32) are independent; this phase does not reopen them.
- Phase 36 uses OpenKruise **ImagePullJob** as a provisioning object, still not wrapped as a `kbl.io` kind ([ADR 0039](0039-openkruise-imagepulljob.md)).

## References

- ADR 0001 Four-DSL Model
- ADR 0010 Standalone Snapshot + Domino
- ADR 0013 Workflow CR references
- ADR 0016 ComputeWheel CR references
- ADR 0023 Julia deployment models (`executionEngine.mode` stays a field)
- ADR 0030 Controller Volcano emission
- ADR 0032 OpenKruise Kind lab
