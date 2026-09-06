# Claims and Evidence

Falsifiable claims only. Metaphors without tests are listed as non-claims.

| ID | Claim | Status | Evidence | Remaining gap |
|----|-------|--------|----------|----------------|
| D1 | Unsealed snapshots do not execute | Proven in engine | `TestTheoremD1…`, `TestUnsealedSnapshotRejected` | None for the gate |
| D2 | Same sealed snapshot + pure chain ⇒ same hashes and output | Proven for builtins; evidenced for pinned Julia | `TestTheoremD2…`, `TestSnapshotReplayDeterministic`, `TestJuliaFinanceModelsWorkflowDeterministic` | H2 for arbitrary containers |
| M1 | Memo hit ≡ recompute observationally | Proven under H2, H4, H5 | `TestTheoremM1…`, `TestMemoizationReusesResults` | Poisoned memo if a non-pure command was cached |
| E1 | Uniform \(N\)-ensemble has \(H=\log_2 N\); seal ⇒ \(H=0\) | Proven for the defined functional | `TestEnsembleEntropyCollapsesOnSeal`, `TestTheoremE1…` | Does not speak to thermodynamics |
| Pay | Sealing does not change payload Shannon entropy | Stated non-claim; checked invariant | `TestEnsembleEntropyCollapsesOnSeal` | — |
| C-past | Readable set at \(d_k\) is snapshot + strict prefix | Proven | `TestCausalPastIsPrefix`, `TestDominoCannotReadFutureOutput` | Container processes could still open the network; not a sandbox theorem |
| R1 | Unsealed snapshots cannot replicate across universes | Proven after patch | `TestTheoremR1…`, `TestMaterializeRejectsUnsealedSnapshot`, `TestApplyRejectsUnsealedSnapshot`, `TestExportFromStoreRejectsUnsealedSnapshot` | Pre-patch hole is closed |
| C1 | Routing is a deterministic function | Proven | `TestTheoremC1…`, `TestRouterPartitionMatch` | — |
| W1 | Wheel has unique successor; \(n_c\) steps advance one slice | Proven | `TestTheoremW1…`, `TestCylinderPeriodAdvancesTimeOnce` | Slot duration is not isochronous |
| F1 | `Coarsen(Unfold(d)) ≅ Unfold(d-1)` in shape and conserved value | Proven for the pattern | `TestCoarsenRecoversShallowerWindow`, `TestTheoremF1…` | Pattern is not yet the cluster scheduler |
| ID128 | Snapshot IDs are 128-bit | Implemented | `hash.SnapshotIDHexLen`, hash/seal tests | Truncation of SHA-256 remains; full 256-bit IDs unused in store keys |
| Landauer | Skipping a domino saves \(\ge kT\ln 2\) | **Non-claim** | — | Would need a thermodynamic accountant |
| Everett | Universes are wavefunction branches | **Non-claim** | — | Declined |
| \(z^2+c\) | Scheduler iterates the Mandelbrot map | **Non-claim** | — | Declined; F1 is the replacement |
| Lifeform | KBL is an organism | **Non-claim** | — | Cybernetic gloss only |
| Purity flag | `deterministic: true` is enforced as a type system | **Open** | Documented as contract | No effect system or sandbox |

## How to run the evidence

```bash
cd controller
go test ./pkg/theory/ ./pkg/engine/ ./pkg/wheel/ ./pkg/routing/ ./pkg/replica/ ./pkg/cdc/ ./pkg/hash/ -count=1
```

Julia determinism (optional, skipped if Julia is missing):

```bash
cd controller
go test ./pkg/engine/ -run TestJuliaFinanceModelsWorkflowDeterministic -count=1
```
