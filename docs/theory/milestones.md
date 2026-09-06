# Theory milestones

Nature-inspired copies, implemented and proved in order. Method: [nature-inspired.md](nature-inspired.md). Each milestone has a **proof** (tests that must pass) and a **remainder in nature** (what we still are not claiming).

**Rule.** Finish one milestone, prove it, push it, then start the next. Do not stack unproven work.

| ID | Milestone | Status | Proof |
|----|-----------|--------|-------|
| M1 | Picard regularity gate in the engine | **proved** | `TestDeterministicWorkflowRejectsContractGradeCommand`, `TestDeterministicBuiltinRunRecordsRegularity`, `TestRequireDeterministicRejectsContract` |
| M2 | Bennett logical work on every run | **proved** | `TestSnapshotReplayDeterministic` work fields; `TestTheoremM1MemoObservationallyEquivalent` |
| M3 | Explorer window seats the Compute Wheel | **proved** | `TestWindowSeatsMatchExplorerLeaves`, `TestFullTurnVisitsEveryWindowLeaf`, `TestCoarsenReducesSeatsByArity` |
| M4 | Classical histories on snapshot-completed events | **proved** | `TestFanoutCarriesSealedHistoryWithoutInterference`, `TestWorldlineFromRun` |
| M5 | Homeostasis on Workflow status | **proved** | `TestWorkflowPhaseHomeostasis`, `TestWorkflowReconcilerExecutesChain` (`status.homeostatic`) |
| M6 | Unfold-driven hierarchical workflows | **proved** | `TestWorkflowFromUnfoldRunsDeterministically`, `TestNodeCountPerfectTree` |
| M7 | Dollar/cost accountant (not joules) | **proved** | `TestChargeUSDIsEvaluationsTimesPrice`, `TestSnapshotReplayDeterministic` cost fields |
| M8 | Sandbox for contract-grade images | **proved** | `TestAllowDeterministicWithIsolatedSandbox`, `TestSandboxedContractCommandRequiresIsolation`, `TestExecuteSandboxIdentity` |
| M9 | Tamper-evident Merkle-style replay spine | **proved** | `TestReplaySpineIsTamperEvident`, `TestVerifySpineRejectsTamperedOutput`, `TestLinkIsDeterministicAndOrderSensitive` |
| M10 | Isolation does not imply uniqueness | **proved** | `TestIsolatedSandboxImpureIsNotUnique`, `TestExecuteSandboxImpureIsNotUnique`, `TestIsolatedSandboxDoesNotImplyUniqueness` |
| M11 | Four-DSL orthogonality for builtins | **proved** | `TestProvisioningOrthogonalToBuiltinWorldline` |
| M12 | Player-piano lookahead is a function of wheel state | **proved** | `TestLookaheadIsAFunctionOfState`, `TestLookaheadNameMatchesBuiltWorkflow`, `TestLookaheadStopsWhenDone` |
| M13 | Unique names in an execution chain | **proved** | `TestRejectsDuplicateChainNames`, `TestUniqueNamesRejectsDuplicates` |
| M14 | Engine-enforced causal past | **proved** | `TestDominoCannotReadFutureOutput`, `TestRejectsDependsOnOutsideCausalPast`, `TestFutureReadRejected` |
| M15 | Partial sandbox isolation is not a cage | **proved** | `TestPartialSandboxIsolationIsNotACage`, `TestPartialSandboxIsolationRejectedByEngine` |
| M16 | Wall-clock is not in the worldline | **proved** | `TestWallClockNotInWorldline` |
| M17 | Sealed snapshots are write-once | **proved** | `TestSealedSnapshotIsWriteOnce`, `TestSealedSnapshotIsWriteOnceTSDB` |
| M18 | Memo key conflict is H5, not ignore | **proved** | `TestMemoRejectsConflictingOutputForSameKey` |
| M19 | Collision-safe WorkflowName truncation | **proved** | `TestWorkflowNameTruncationDistinguishesEqualLength` |
| M20 | RunSingle continues the replay spine | **proved** | `TestRunSingleSpineChainsAcrossSteps` |
| M21 | Store persists the replay spine | **proved** | `TestReplayLogRoundTripsSpine` |
| M22 | Snapshot-completed EventID is content-addressed | **proved** | `TestSnapshotEventIDIsFunctionOfPayload` |
| M23 | CDC refuses orphan domino results | **proved** | `TestApplyDominoResultRequiresSealedParent` |
| M24 | Materialize/export fail closed on missing dominos | **proved** | `TestMaterializeFailsOnMissingDomino`, `TestExportFromStoreFailsOnMissingDomino` |
| M25 | Routing priority: time-slice beats partition | **proved** | `TestRoutingPriorityTimeSliceBeatsPartition` |
| M26 | Ambiguous partition matches are rejected | **proved** | `TestAmbiguousPartitionMatchRejected` |
| M27 | CRD sandbox flags reach the Picard cage | **proved** | `TestConvertPassesSandboxIntoEngineGate` |
| M28 | Workflow status exposes HeadLink and work | **proved** | `TestWorkflowReconcilerExecutesChain` (`status.headLink`, `workEvaluations`, `workCostUSD`) |
| M29 | Fan-out events carry HeadLink | **proved** | `TestFanoutCarriesSealedHistoryWithoutInterference` |
| M30 | Fan-out refuses incomplete events | planned | empty snapshot ID / worldline does not branch |
| M31 | Loaded snapshot matches its content address | planned | corrupt store bytes fail closed |
| M32 | Explorer window can seat a live ComputeWheel | planned | declared depth/arity must match context count |

## Proof command

```bash
cd controller
go test ./pkg/theory/ ./pkg/engine/ ./pkg/wheel/ ./pkg/routing/ ./pkg/replica/ ./pkg/cdc/ ./pkg/hash/ ./pkg/convert/ ./pkg/executor/ ./pkg/store/ ./pkg/events/ ./internal/controller/ -count=1
```

Or `make theory-prove` from the repository root. Store invariants (M17+) are included.
