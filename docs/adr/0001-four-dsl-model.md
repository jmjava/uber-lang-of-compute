# ADR 0001: Four-DSL Model

## Status

Accepted

## Context

The Uber Language of Compute (January 2020) defines computation as four orthogonal concerns that should be described separately and composed together:

1. **Execution** — the logic that transforms data
2. **Data** — the shape, location, and lifecycle of inputs/outputs
3. **Provisioning** — how compute and storage resources are allocated
4. **Routing** — how work is directed across universes, nodes, and time slices

Mixing these concerns in a single configuration language leads to tight coupling and makes pluggable universes impossible.

## Decision

We adopt the four-DSL model as the foundational specification layer:

- Each DSL is a separate YAML schema in `specs/`
- CRDs compose fields from one or more DSLs
- Controllers reconcile each DSL independently where possible
- CDK8s API objects serve as the initial implementation path; CRDs/operators follow

## Consequences

- New compute environments (Pluggable Universes) require only new Provisioning + Data bindings; Execution and Routing DSLs remain stable
- DSL schemas can evolve independently with versioned compatibility checks
- Users must understand four concepts instead of one — mitigated by unified workflow examples in `specs/workflow-example.yaml`

## References

- *The Uber Language of Compute* (Jan 11, 2020)
- *What can be created with CDK8s?* (Jan 15, 2021)
- *Visualization: A Multiverse of Computation* (Apr 11, 2025)
