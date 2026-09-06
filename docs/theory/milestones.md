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
| M16 | Wall-clock is not in the worldline | planned | delayed rerun: equal hashes/HeadLink, unequal timestamps |
| M17 | Sealed snapshots are write-once | planned | different payload fails; identical rewrite is idempotent |
| M18 | Memo key conflict is H5, not ignore | planned | same key, different output is an error |
| M19 | Collision-safe WorkflowName truncation | planned | two long equal-length names must not collide |
| M20 | RunSingle continues the replay spine | planned | sequential `RunSingle` verifies as one chain |
| M21 | Store persists the replay spine | planned | reload replay rows, `VerifySpine` passes |
| M22 | Snapshot-completed EventID is content-addressed | planned | same payload ⇒ same EventID; timestamp excluded |
| M23 | CDC refuses orphan domino results | planned | result rows require a sealed parent snapshot |
| M24 | Materialize/export fail closed on missing dominos | planned | incomplete chain is an error, not a short success |
| M25 | Routing priority: time-slice beats partition | planned | overlapping rules pick the time-slice target |
| M26 | Ambiguous partition matches are rejected | planned | two universes matching the same labels is an error |
| M27 | CRD sandbox flags reach the Picard cage | planned | convert threads isolation into engine provisioning |
| M28 | Workflow status exposes HeadLink and work | planned | completed reconcile records spine head and USD cost |
| M29 | Fan-out events carry HeadLink | planned | classical record includes the Merkle receipt |
| M30 | Fan-out refuses incomplete events | planned | empty snapshot ID / worldline does not branch |
| M31 | Loaded snapshot matches its content address | planned | corrupt store bytes fail closed |
| M32 | Explorer window can seat a live ComputeWheel | planned | declared depth/arity must match context count |

## Proof command

```bash
cd controller
go test ./pkg/theory/ ./pkg/engine/ ./pkg/wheel/ ./pkg/routing/ ./pkg/replica/ ./pkg/cdc/ ./pkg/hash/ ./pkg/convert/ ./pkg/executor/ ./internal/controller/ -count=1
```

Or `make theory-prove` from the repository root.
