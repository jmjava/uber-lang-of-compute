# Theory milestones

Nature-inspired copies, implemented and proved in order. Method: [nature-inspired.md](nature-inspired.md). Each milestone has a **proof** (tests that must pass) and a **remainder in nature** (what we still are not claiming).

**Rule.** Finish one milestone, prove it, push it, then start the next. Do not stack unproven work.

| ID | Milestone | Status | Proof |
|----|-----------|--------|-------|
| M1 | Picard regularity gate in the engine | **proved** | `TestDeterministicWorkflowRejectsContractGradeCommand`, `TestDeterministicBuiltinRunRecordsRegularity`, `TestRequireDeterministicRejectsContract` |
| M2 | Bennett logical work on every run | pending | `RunResult` reports evaluations/reuses; replay strictly saves irreversible steps |
| M3 | Explorer window seats the Compute Wheel | pending | \(k^d\) leaves = wheel seats; one full turn visits each leaf once; coarsen ⇒ \(k^{d-1}\) seats |
| M4 | Classical histories on snapshot-completed events | later | routing fan-out carries sealed snapshot ID; `Interfere(liveShare=false)` is false in the event path |
| M5 | Homeostasis on Workflow status | later | spec-phase vs status-phase error is 0 iff Ready |
| M6 | Unfold-driven hierarchical workflows | later | a live chain can be built from `Unfold` (F1 leaves the pattern library) |
| M7 | Dollar/cost accountant (not joules) | later | logical work × unit cost; still not \(kT\ln 2\) |
| M8 | Sandbox for contract-grade images | later | network/FS isolation so H2 can be evidenced for containers |

## Proof command

```bash
cd controller
go test ./pkg/theory/ ./pkg/engine/ ./pkg/wheel/ ./pkg/routing/ ./pkg/replica/ ./pkg/cdc/ ./pkg/hash/ -count=1
```

Or `make theory-prove` from the repository root.
