# ADR 0036: Courseforge Integration Exploration

## Status

Proposed (exploration)

## Context

The **Courseforge** suite ([`courseforge/course-builder`](https://github.com/courseforge/course-builder), [`courseforge/infrastructure`](https://github.com/courseforge/infrastructure)) provides a Course Builder–style **balancer + worker pool** for learner code execution. Worker container images live in the **course-builder** repository.

**KBL Compute Engine** (this repo) now ships Volcano batch scheduling, ComputeWheel time slices, PluggableUniverse runtimes (`runtimeImage`), and Multiverse routing — a substantially richer Kubernetes-native scheduling layer than a single worker URL.

Operators asked whether KBL could integrate with Courseforge: reuse **existing worker images** and act as a **more sophisticated compute scheduler** behind the course-builder balancer.

## Decision (exploration only — no implementation yet)

Proceed with a **phased integration exploration** documented in [docs/explorations/courseforge-integration.md](../explorations/courseforge-integration.md):

1. **Inventory** course-builder worker images and API contracts (requires private repo access).
2. **Map** worker images → `PluggableUniverse.spec.executionEngine.runtimeImage` and `DominoChain.spec.runnerImage`.
3. **PoC** on home Kind profile (i9): Volcano queue + one Workflow/DominoChain using a course-builder worker image (likely via a **domino-runner shim** if APIs differ).
4. **Adapter** from course-builder task dispatch → Workflow CR (balancer unchanged at UI layer).
5. **Optional:** ComputeWheel for module time slices; Multiverse for multi-course routing.

KBL remains the **execution/scheduling plane**; course-builder remains the **balancer and UX**.

## Consequences

- May require a new **`kbl-domino-runner-courseforge`** (or similar) wrapper image.
- `courseforge/infrastructure` Kind installer may gain an optional KBL + Volcano bundle.
- Grading/replay semantics must align with Courseforge expectations (snapshot + replay log as audit artifact).
- Private repo access needed before Phase 32a inventory can complete.

## References

- [docs/explorations/courseforge-integration.md](../explorations/courseforge-integration.md)
- [jmjava/documentation-generator suite handbook](https://github.com/jmjava/documentation-generator/tree/main/docs/suite)
- ADR 0031, ADR 0035
