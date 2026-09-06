# Claims and Evidence

Grades follow [nature-inspired.md](nature-inspired.md): theorem, empirical discretization, contract, bookkeeping mimic, explorer mimic, branching mimic, control mimic. We mimic nature; we do not prove physics.

| ID | Claim | Grade | Evidence | Remainder in nature |
|----|-------|-------|----------|---------------------|
| D1 | Unsealed snapshots do not execute | Theorem | `TestTheoremD1…`, `TestUnsealedSnapshotRejected` | — |
| D2 | Same sealed snapshot + builtin chain ⇒ same hashes | Theorem | `TestTheoremD2…`, `TestSnapshotReplayDeterministic` | Continuum mechanics |
| D2j | Pinned Julia is unique as a discretization | Empirical discretization | `TestJuliaFinanceModelsWorkflowDeterministic` | Uniqueness on \(\mathbb{R}\) |
| H2c | Container commands are functions | Contract | `CommandRegularity` → `RegularityContract` | Lipschitz for arbitrary images |
| M1 | Deterministic workflows reject contract-grade commands | Theorem of the gate | `TestDeterministicWorkflowRejectsContractGradeCommand` | Purity of the image itself |
| M1 | Memo hit ≡ recompute observationally | Theorem under H2–H5 | `TestTheoremM1…` | — |
| Bennett | Replay pays fewer irreversible *steps* | Bookkeeping mimic | `ReplaySaves`, `TestLogicalWorkReplaySaves` | Joules / \(kT\ln 2\) |
| E1 | Uniform \(N\)-ensemble has \(H=\log_2 N\); seal ⇒ \(H=0\) | Theorem for \(H_{\mathrm{ens}}\) | `TestEnsembleEntropyCollapsesOnSeal` | Thermodynamic \(S\) |
| C-past | Readable set at \(d_k\) is snapshot + strict prefix | Theorem | `TestCausalPastIsPrefix`, `TestDominoCannotReadFutureOutput` | Sandbox of the OS |
| R1 | Unsealed snapshots cannot replicate | Theorem | `TestTheoremR1…` | — |
| C1 | Routing is a deterministic function | Theorem | `TestTheoremC1…` | — |
| W1 | Wheel has unique successor; \(n_c\) steps advance one slice | Theorem | `TestTheoremW1…` | Torque / isochrony |
| F1 | `Coarsen(Unfold(d)) ≅ Unfold(d-1)` | Theorem (pattern) | `TestCoarsenRecoversShallowerWindow` | Live fractal scheduler |
| F-dim | Additive window has similarity dimension 1 | Explorer mimic | `TestSimilarityDimensionAdditiveTreeIsOne` | Hausdorff dim. of the Mandelbrot set |
| F-esc | Escape-time is a finite iteration window | Explorer mimic | `TestEscapeTimeIsAFiniteWindow` | Connectedness locus of \(z^2+c\) |
| F-wheel | Window leaves can be wheel seats | Explorer mimic | `TestWheelWindowLeafCount` | — |
| Hist | Sealed fan-out does not interfere | Branching mimic | `TestHistoriesDoNotInterfereWhenSealed` | Hilbert space |
| Homeo | Spec=status is homeostatic | Control mimic | `TestHomeostasisMatchesSpec` | Metabolism |
| ID128 | Snapshot IDs are 128-bit | Implemented | `hash.SnapshotIDHexLen` | Full 256-bit store keys |

## How to run the evidence

```bash
cd controller
go test ./pkg/theory/ ./pkg/engine/ ./pkg/wheel/ ./pkg/routing/ ./pkg/replica/ ./pkg/cdc/ ./pkg/hash/ -count=1
```

Julia discretization (optional, skipped if Julia is missing):

```bash
cd controller
go test ./pkg/engine/ -run TestJuliaFinanceModelsWorkflowDeterministic -count=1
```
