# Claims and Evidence

Grades follow [nature-inspired.md](nature-inspired.md): theorem, empirical discretization, contract, bookkeeping mimic, explorer mimic, branching mimic, control mimic. We mimic nature; we do not prove physics.

| ID | Claim | Grade | Evidence | Remainder in nature |
|----|-------|-------|----------|---------------------|
| D1 | Unsealed snapshots do not execute | Theorem | `TestTheoremD1…`, `TestUnsealedSnapshotRejected` | — |
| D2 | Same sealed snapshot + builtin chain ⇒ same hashes | Theorem | `TestTheoremD2…`, `TestSnapshotReplayDeterministic` | Continuum mechanics |
| D2j | Pinned Julia is unique as a discretization | Empirical discretization | `TestJuliaFinanceModelsWorkflowDeterministic` | Uniqueness on \(\mathbb{R}\) |
| H2c | Container commands are functions | Contract | `CommandRegularity` → `RegularityContract` | Lipschitz for arbitrary images |
| M1 | Deterministic workflows reject contract-grade commands | Theorem of the gate | `TestDeterministicWorkflowRejectsContractGradeCommand` | Purity of the image itself |
| M8 | Isolated sandbox admits contract-grade stubs | Theorem of the cage | `TestSandboxedContractCommandRequiresIsolation` | Arbitrary container images / real kernel namespaces |
| M10 | Isolated `sandbox:impure` is not unique | Falsification of cage⇒H2 | `TestIsolatedSandboxImpureIsNotUnique`, `TestExecuteSandboxImpureIsNotUnique` | Real kernel namespaces / arbitrary images |
| M11 | Provisioning/routing mutations do not change builtin worldlines | Theorem of the axes | `TestProvisioningOrthogonalToBuiltinWorldline` | Container images whose purity depends on the host |
| M12 | Next wheel slot name is a function of current state | Theorem of the piano roll | `TestLookaheadIsAFunctionOfState`, `TestLookaheadNameMatchesBuiltWorkflow` | Isochronous mechanical timing / real pre-warm of pods |
| M13 | Execution chain names are unique | Theorem of exclusion | `TestRejectsDuplicateChainNames`, `TestUniqueNamesRejectsDuplicates` | Namespaced aliases |
| M14 | Engine rejects reads outside the causal past | Theorem of the cone | `TestDominoCannotReadFutureOutput`, `TestRejectsDependsOnOutsideCausalPast` | OS sandbox of the process |
| M15 | Partial isolation is not a Faraday cage | Falsification of XOR⇒H2-admission | `TestPartialSandboxIsolationIsNotACage`, `TestPartialSandboxIsolationRejectedByEngine` | Real kernel namespaces |
| M16 | Wall-clock is excluded from the worldline | Theorem of the gauge | `TestWallClockNotInWorldline` | Relativistic proper time |
| M17 | Sealed snapshots are write-once | Theorem of crystallization | `TestSealedSnapshotIsWriteOnce`, `TestSealedSnapshotIsWriteOnceTSDB` | Physical WORM media |
| M18 | Memo key conflict is rejected | Theorem of H5 | `TestMemoRejectsConflictingOutputForSameKey` | Byzantine multi-writer consensus |
| M19 | Long Workflow names truncate by content hash | Theorem of labels | `TestWorkflowNameTruncationDistinguishesEqualLength` | DNS-1035 aesthetics |
| M20 | Sequential RunSingle continues the spine | Theorem of the geodesic | `TestRunSingleSpineChainsAcrossSteps` | Distributed consensus of spine heads |
| M21 | Store persists the replay spine | Theorem of the fossil record | `TestReplayLogRoundTripsSpine` | Full Merkle DAG / blockchain |
| M1/M2 | Memo hit ≡ recompute observationally | Theorem under H2–H5 | `TestTheoremM1…` | — |
| M9 | Replay log is a hash chain from the snapshot ID | Theorem of the spine | `TestReplaySpineIsTamperEvident`, `TestVerifySpineRejectsTamperedOutput` | Full Merkle DAG / blockchain |
| Cost | Evaluations × unit price in USD | Bookkeeping mimic | `TestChargeUSDIsEvaluationsTimesPrice`, `RunResult.WorkCostUSD` | \(kT\ln 2\) / joules |
| E1 | Uniform \(N\)-ensemble has \(H=\log_2 N\); seal ⇒ \(H=0\) | Theorem for \(H_{\mathrm{ens}}\) | `TestEnsembleEntropyCollapsesOnSeal` | Thermodynamic \(S\) |
| C-past | Readable set at \(d_k\) is snapshot + strict prefix | Theorem | `TestCausalPastIsPrefix`, `TestDominoCannotReadFutureOutput` | Sandbox of the OS |
| R1 | Unsealed snapshots cannot replicate | Theorem | `TestTheoremR1…` | — |
| C1 | Routing is a deterministic function | Theorem | `TestTheoremC1…` | — |
| W1 | Wheel has unique successor; \(n_c\) steps advance one slice | Theorem | `TestTheoremW1…` | Torque / isochrony |
| F1 | `Coarsen(Unfold(d)) ≅ Unfold(d-1)` | Theorem (pattern + live chain) | `TestCoarsenRecoversShallowerWindow`, `TestWorkflowFromUnfoldRunsDeterministically` | Fractal Kubernetes pod spawn |
| F-dim | Additive window has similarity dimension 1 | Explorer mimic | `TestSimilarityDimensionAdditiveTreeIsOne` | Hausdorff dim. of the Mandelbrot set |
| F-esc | Escape-time is a finite iteration window | Explorer mimic | `TestEscapeTimeIsAFiniteWindow` | Connectedness locus of \(z^2+c\) |
| F-wheel | Window leaves can be wheel seats | Explorer mimic | `TestWheelWindowLeafCount`, `TestWindowSeatsMatchExplorerLeaves`, `TestFullTurnVisitsEveryWindowLeaf` | Live CRD field for depth/arity |
| Hist | Sealed fan-out does not interfere | Branching mimic | `TestHistoriesDoNotInterfereWhenSealed`, `TestFanoutCarriesSealedHistoryWithoutInterference` | Hilbert space |
| Homeo | Spec=status is homeostatic | Control mimic | `TestHomeostasisMatchesSpec`, `TestWorkflowPhaseHomeostasis`, Workflow `status.homeostatic` | Metabolism |
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
