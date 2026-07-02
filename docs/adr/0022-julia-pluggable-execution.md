# ADR 0022: Julia Pluggable Execution Engine

## Status

Accepted

## Context

Phase 6 introduced `PluggableUniverse` with `executionEngine.type` enum values including `julia`, but domino execution only supported Go `builtin:` commands. Operators need to run deterministic dominos in Julia without forking the memoization, replay, or store-first snapshot paths.

## Decision

1. **`pkg/executor`** — central dispatch for `builtin:`, `julia:`, and future `python:` command prefixes
2. **Julia subprocess runtime** — `julia:<script>` runs `controller/julia/scripts/<script>.jl` with file-based JSON handoff (input.json → output.json)
3. **Bundled Julia project** — `controller/julia/Project.toml` pins `JSON.jl` for finance domino scripts
4. **Engine + domino-runner** — both call `executor.Execute`; container entrypoint inherits `KBL_JULIA_*` env vars
5. **PluggableUniverse status** — `executionEngine.type: julia` sets an informative status message; dominos declare explicit `julia:` commands

Python execution remains unimplemented (`python:` returns a clear error).

## Consequences

- Julia must be installed on nodes running `julia:` dominos (or baked into `runtimeImage`)
- First run requires `Pkg.instantiate()` for the bundled Julia project
- Container/long-running Julia workers deferred; subprocess model is simple and deterministic
- Independent of snapshot mmap/TSDB phases — Julia receives JSON strings like builtins

## References

- ADR 0009 — Multiverse Routing
- `controller/pkg/executor/`
- `controller/julia/`
- `examples/julia-domino-chain/`
