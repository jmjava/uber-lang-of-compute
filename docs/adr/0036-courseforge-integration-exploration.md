# ADR 0036: Courseforge Integration Exploration

## Status

Proposed (exploration)

## Context

The **Courseforge** suite ([`courseforge/course-builder`](https://github.com/courseforge/course-builder), [`courseforge/infrastructure`](https://github.com/courseforge/infrastructure)) runs:

- **CourseForge API** (`tools/courseforge/backend/`) — UI/backend balancer
- **Orchestrator** (`tools/automation-workers/orchestrator/`) — Postgres stages, HTTP dispatch
- **Stable worker Services** — `blender-worker` / `unreal-worker` on `:8080`, **job packages** (`packageUri` + `job.yaml`)

**KBL** ships Volcano, ComputeWheel, PluggableUniverse, Multiverse, snapshots, and memoization.

Operators asked whether KBL could reuse **existing worker images** and act as a **more sophisticated compute scheduler** behind Courseforge.

## Decision (exploration only — no implementation yet)

Proceed per [docs/explorations/courseforge-integration.md](../explorations/courseforge-integration.md):

1. **Keep** heavy worker Deployments and the job-package HTTP contract.
2. **Add** KBL domino `courseforge:package` → `POST /jobs` on worker Services; Volcano fair-share queues.
3. **Map** orchestrator stage DAG ↔ Workflow / DominoChain.
4. **PoC** on home Kind: KBL + `automation-workers` namespace from course-builder.

KBL upgrades **scheduling + audit**; CourseForge API, worker images, and artifact storage stay.

## Consequences

- New **`courseforge:` executor** in controller (HTTP client to worker Services).
- Optional thin **`kbl-domino-runner-courseforge`** if domino steps run in-cluster.
- `courseforge/infrastructure` Kind installer may bundle KBL + Volcano alongside existing workers.
- Artifact/replay semantics: KBL snapshot + replay log complement Postgres job records.

## References

- [docs/explorations/courseforge-integration.md](../explorations/courseforge-integration.md)
- `course-builder/tools/automation-workers/README.md`
- `courseforge/infrastructure/docs/stable-worker-job-package-pattern/stable-worker-spec.md`
- ADR 0031, ADR 0035
