# ADR 0034: Architecture Diagrams

## Status

Accepted

## Context

Phase 29 added documentation hub and prose guides, but operators and contributors still lacked **visual** maps for:

- Four DSL → CRD → reconciler relationships
- Compute Wheel rotation and Volcano pipeline
- Kind lab node topology
- Per-runtime sequences (init chain, OpenKruise CRR, Volcano VCJob)
- Lab troubleshooting flow

Architecture.md retained ASCII art from MVP; provisioning-runtimes had one mermaid chart only.

## Decision

### 1. Central diagram reference (`docs/diagrams.md`)

Single page with 13 Mermaid diagrams:

1. Four DSLs → Kubernetes
2. End-to-end domino execution sequence
3. Compute Wheel state + rotation
4. Kind lab 3-node topology
5. Provisioning runtime comparison
6. kubernetes-init pod anatomy
7. OpenKruise hot-swap sequence
8. Volcano batch pipeline sequence
9. Julia finance chain demos
10. Memoization flow
11. Multiverse routing (target)
12. Kind lab troubleshooting decision tree
13. AWS CDK target scaffold

### 2. Cross-links

- `docs/README.md` — prominent link to diagrams
- `docs/architecture.md` — mermaid layer diagram + link to full set
- `docs/getting-started.md` — path overview diagram
- `lab/README.md` — topology + troubleshooting links
- Root README — Phase 30 row

Diagrams use Mermaid only (no binary assets) for GitHub-native rendering.

## Consequences

- Diagrams must be updated when lab topology or reconciler flow changes
- Mermaid render quality varies by viewer; ASCII retained in architecture.md where useful
- Future phases add diagrams to `diagrams.md` rather than scattering new images

## References

- [docs/diagrams.md](../diagrams.md)
- [ADR 0033 Documentation Phase](0033-documentation-phase.md)
