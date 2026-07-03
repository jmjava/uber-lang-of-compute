# ADR 0033: Documentation Phase

## Status

Accepted

## Context

Phases 22–28 added Kind lab, Volcano, ComputeWheel queue wiring, OpenKruise demos, and Julia finance chains. Documentation grew across ADRs, `lab/README.md`, and the root README, but lacked:

- A central documentation index
- A unified getting-started path (CLI → Kind lab)
- A single reference comparing provisioning runtimes (`kubernetes-init`, `openkruise`, `volcano-init`)
- Updated architecture text (still marked OpenKruise as "Future")
- Cross-links from examples to lab demos

Operators had to assemble the story from scattered ADRs and README sections.

## Decision

### 1. Documentation hub (`docs/README.md`)

Central index with:

- Start-here guides
- Blog-to-repo concept mapping
- Examples index
- Kind lab demo matrix
- ADRs grouped by topic
- Phase roadmap summary

### 2. Getting Started (`docs/getting-started.md`)

Three paths: CLI only, Kind lab full stack, manual Kubernetes. Verification commands for Volcano and OpenKruise demos.

### 3. Provisioning Runtimes (`docs/provisioning-runtimes.md`)

Runtime comparison table, mermaid flow, field propagation (Workflow → DominoChain), Kind lab matrix, links to examples.

### 4. Updates to existing docs

- `docs/architecture.md` — provisioning runtime section, Kind lab stack, OpenKruise no longer "Future"
- `examples/compute-wheel/README.md` — Volcano wheel + lab integration
- `examples/julia-domino-chain/README.md` — all runtimes + lab demo names
- Root `README.md` — documentation section points to hub; Phase 29 row
- `lab/README.md` — link to getting-started and provisioning-runtimes

## Consequences

- New contributors have a single entry point (`docs/README.md`)
- Runtime choice is documented without reading multiple ADRs
- Architecture doc reflects current implementation state
- Future phases should update hub + getting-started when adding lab demos

## References

- [docs/README.md](../README.md)
- [docs/getting-started.md](../getting-started.md)
- [docs/provisioning-runtimes.md](../provisioning-runtimes.md)
