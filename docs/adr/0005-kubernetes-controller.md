# ADR 0005: Kubernetes Controller Reconciler

## Status

Accepted

## Context

The MVP CLI (`kbl-compute`) proved deterministic replay, memoization, and node-local storage. The next step toward a Kubernetes-native compute fabric is reconciling Workflow resources in-cluster rather than invoking a standalone binary.

CRDs for Snapshot, Domino, ComputeContext, ComputeWheel, and PluggableUniverse already exist. The engine package already executes domino chains. A reconciler bridges these pieces.

## Decision

Introduce a **Workflow CRD** as the primary reconciliation unit for MVP Phase 2:

- A Workflow embeds an inline Snapshot, Domino chain, and execution config (matching the CLI workflow YAML shape)
- The `kbl-controller` uses controller-runtime to watch Workflow resources
- On reconcile (when `observedGeneration != generation` or not yet Completed):
  1. Open node-local SQLite store at `spec.provisioning.storePath` or `{store-root}/{namespace}/{name}.db`
  2. Execute the domino chain via the existing engine
  3. Update Workflow status: phase, snapshot ID, domino counts, per-domino results
  4. Write replay log JSON to an owned ConfigMap (`{name}-replay`)
- Completed workflows with matching generation are no-ops (idempotent reconcile)

The CLI remains for local development and CI; the controller is the cluster-native path.

## Consequences

- Workflow is the first "composed" CRD; standalone Snapshot/Domino CRDs remain for future fine-grained reconciliation
- Replay logs live in ConfigMaps (status has summary only); large outputs may need object-store offload later
- Controller requires RBAC for workflows, status, finalizers, and configmaps
- Compute Wheel scheduling and standalone Domino reconciliation are deferred to Phase 3+

## References

- MVP scaffold PR #1
- *What can be created with CDK8s?* (Jan 15, 2021) — CRDs/operators as implementation path
