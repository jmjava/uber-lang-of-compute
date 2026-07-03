# ADR 0023: Julia Deployment Models — Multi-Container vs Single-Container Multi-Process

## Status

Accepted (evaluation); implementation deferred to Phase 20+

## Context

Phase 14 added Julia domino execution via **local subprocess** (`julia:<script>` → file-based JSON handoff). ADR 0007 already defines **in-cluster** execution through `domino-runner` inside DominoChain pods (init-container daisy chain or OpenKruise hot-swap).

Engineers asked whether Julia workloads should run as:

1. **Many small containers** — one domino step (one process) per container slot, or
2. **One large container** — multiple Julia worker processes inside a shared runtime image.

Both models must preserve KBL invariants: sealed snapshots, deterministic dominos, input/output hashing, memoization, and replay logs. The Go engine (or controller finalize path) remains authoritative for hashes even when dominos run out-of-process.

## Evaluation criteria

| Criterion | Question |
|-----------|----------|
| **Isolation** | Can a failed or runaway domino corrupt neighbors or shared state? |
| **Cold-start latency** | Time from chain step scheduled → Julia ready to accept JSON input |
| **Memory** | RSS per chain run; JIT/precompile duplication across steps |
| **Ops complexity** | Image size, K8s objects, debugging, upgrades, resource limits |
| **Determinism / replay** | Same input JSON → same output JSON; replay log still valid |
| **Data locality** | Compatibility with node-local TSDB, path snapshots, store-first hot path |
| **Hot-swap fit** | Works with OpenKruise player-piano and init-container handoff |
| **Multiverse fit** | PluggableUniverse `runtimeImage` per universe |

## Models evaluated

### A. Local subprocess (Phase 14 — current dev default)

```
kbl-compute / engine.Run
  └─ executor.Execute("julia:interpolate")
       └─ julia --project=... script.jl input.json output.json
```

| Pros | Cons |
|------|------|
| Simplest; no cluster required | New Julia process per domino (slow JIT warm-up) |
| Maximum isolation between steps | Not suitable for production throughput |
| Identical command surface as builtins | No shared precompiled cache |

**Verdict:** Keep for **local dev, CI, and bootstrap**. Not the production in-cluster target.

---

### B. Multi-container — one domino per container (recommended production default)

```
DominoChain pod
  ├─ initContainer / slot-N: domino-runner  (KBL_COMMAND=julia:interpolate)
  │     └─ julia subprocess inside container (single step, then exit)
  └─ emptyDir handoff + snapshot ConfigMap
```

Each domino step runs in its **own container boundary** (init container or hot-swapped slot). `domino-runner` already dispatches `julia:*` via `pkg/executor` (Phase 14).

| Pros | Cons |
|------|------|
| Aligns with ADR 0007 hot-swap and init-chain designs | Julia JIT still cold per step unless image warms caches |
| Strong failure isolation; K8s resource limits per step | More container starts than model C |
| Independent `runnerImage` / domino `image` overrides | Larger total image pull surface if every step uses a different image |
| OpenKruise player-piano maps 1:1 to domino slots | |
| Memo/replay unchanged — JSON handoff files + engine finalize | |

**Verdict:** **Default for in-cluster Julia** and multiverse `runtimeImage` wiring. Extends existing DominoChain + domino-runner with no new long-lived supervisor.

---

### C. Single-container multi-process — shared Julia runtime inside one pod

```
DominoChain pod (one long-lived container)
  ├─ julia-supervisor (PID 1)
  │     ├─ worker 1..N (preloaded JSON.jl + scripts)
  │     └─ IPC: Unix socket / HTTP / stdin RPC
  └─ domino-runner OR thin Go shim calls supervisor per step
```

One **fat container** stays up for the whole chain (or node lifetime). A supervisor keeps Julia workers warm; domino steps are **in-container RPC** instead of process spawn.

| Pros | Cons |
|------|------|
| Amortizes JIT and `Pkg.instantiate()` once per chain/node | Weaker isolation — memory leak or bad package state affects all workers |
| Lowest step-to-step latency for long chains | New component: supervisor protocol, health, restarts |
| Smaller K8s churn (one container vs many init/swap cycles) | Harder to map OpenKruise hot-swap slots (designed for per-slot containers) |
| Good for batch chains with many Julia steps | Determinism requires pinned Julia version + fixed worker pool semantics |
| | Blurs “one domino = one container” mental model from ADR 0004 |

**Verdict:** **Spike candidate (Phase 20+)** for latency-sensitive, same-node finance chains with 3+ Julia steps. Not the default until supervisor + replay semantics are proven.

---

## Comparison matrix

| | A. Subprocess (local) | B. Multi-container | C. Single-container multi-process |
|--|----------------------|-------------------|-----------------------------------|
| Isolation | High | High | Medium |
| Cold-start | Poor | Medium | Good (after warm-up) |
| Memory | Low peak, repeated JIT | Medium, repeated JIT | Best amortization |
| Ops complexity | Low | Low (existing CRDs) | High (new supervisor) |
| Hot-swap fit | N/A | Excellent | Poor |
| Implementation today | **Shipped (Phase 14)** | **Partial** (domino-runner + executor; Julia in cluster untested E2E) | Not started |

## Decision

1. **Production in-cluster default → Model B (multi-container)**  
   Use DominoChain + `domino-runner` + `julia:*` commands. One container per domino step; Julia may still subprocess inside the container (acceptable — container is the isolation boundary).

2. **Local / CI → Model A (subprocess)**  
   Unchanged Phase 14 path via `kbl-compute`.

3. **Optional optimization → Model C (single-container multi-process)**  
   Defer until Phase 20 spike validates supervisor IPC, worker pool sizing, and replay-equivalent outputs against Model B on the finance curve chain.

4. **PluggableUniverse mapping**  
   - `executionEngine.type: julia` + `runtimeImage` → Model B image (includes Julia + `controller/julia` project)  
   - Future `executionEngine.mode: pooled` (CRD extension) → Model C if spike succeeds

## Migration path (Phase 14 → Model B)

| Step | Work |
|------|------|
| 1 | Publish `runtimeImage` with Julia + pre-instantiated `controller/julia` project |
| 2 | E2E test: `workflow-container.yaml` with `julia:*` dominos + init chain |
| 3 | Document `KBL_JULIA_*` env in DominoChain pod templates |
| 4 | OpenKruise chain: verify hot-swap with Julia cold-start budgets |

No engine API changes required — command prefix and JSON handoff stay the same.

## Migration path (Model B → Model C, if needed)

| Step | Work |
|------|------|
| 1 | Spike `julia-supervisor` — HTTP or Unix socket `{command, input}` → `{output}` |
| 2 | Add `executor` backend `julia:pooled` or env `KBL_JULIA_SUPERVISOR_URL` |
| 3 | Prove byte-identical outputs vs Model B on finance curve workflow |
| 4 | CRD field `executionEngine.mode: pooled` on PluggableUniverse |

## Consequences

- Model B requires a **Julia-enabled domino-runner image** for cluster runs (not only host-installed Julia)
- Model C is **not** compatible with OpenKruise slot-per-domino without redesigning the player-piano model
- Memoization and replay remain in the Go store path regardless of model
- Python pluggable execution can follow the same B/C evaluation when implemented

## References

- ADR 0007 — Hot-Swapped Dominos Implementation
- ADR 0022 — Julia Pluggable Execution
- `controller/cmd/domino-runner`
- `examples/hot-swap-domino-chain/`
- `examples/julia-domino-chain/`
