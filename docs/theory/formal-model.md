# Formal Model

This note states the discrete dynamical system that the correspondences act on, the hypotheses they need, and proofs of the theorems named in [correspondences.md](correspondences.md). Notation is computational. Where a physics name is used, it refers to the correspondence, not to a physical quantity.

---

## 1. State space

Fix a finite alphabet of byte strings \(\mathbf{Str}\).

A **snapshot** is a triple \(s = (\tau, x, \sigma)\) where \(\tau\) is a time-slice identifier, \(x\in\mathbf{Str}\) is a payload, and \(\sigma\in\{\mathrm{open},\mathrm{sealed}\}\).

A **domino** is a name \(d\) together with a command interpreted as a partial function \(f_d : \mathbf{Str}\rightharpoonup\mathbf{Str}\) and a list of input selectors (snapshot and/or earlier dominos).

A **chain** is a finite sequence \(D=(d_1,\ldots,d_n)\) of distinct names.

A **memo** is a partial map
\[
M : \mathrm{ID}\times\mathrm{Name}\times\{0,1\}^{256} \rightharpoonup \mathbf{Str}.
\]

A **wheel state** is \((i,t,r)\in\{0,\ldots,n_c-1\}\times T\times\mathbb{N}\), where \(n_c\) is the number of contexts and \(T\) is a discrete time lattice.

A **universe** is a tuple \(u=(\Phi_u, \mathrm{Store}_u, \mathrm{Prov}_u)\): evolution, local store, provisioning.

A **multiverse specification** is a routing function \(\rho\) as implemented by `routing.Router.Resolve`.

The **engine state** relevant to one workflow is \((s, D, M, Y)\) where \(Y\) maps already-executed names to outputs.

---

## 2. Hypotheses

| ID | Hypothesis | Role |
|----|------------|------|
| H1 | **Seal gate.** \(\Phi\) is undefined on \(\sigma=\mathrm{open}\). | D1 |
| H2 | **Purity.** Each \(f_d\) used under `deterministic: true` is a total function of its resolved input string. | D2, M1 |
| H3 | **Input resolution is a function.** `resolveInputs` depends only on \(x\) and \(\{Y(d'): d'\prec d\}\). | D2, causal past |
| H4 | **Collision resistance.** SHA-256 is injective on the payloads that actually occur. Snapshot IDs are 128-bit prefixes (birthday bound \(\sim 2^{64}\)). | identity of worldlines |
| H5 | **Memo integrity.** \(M(s,d,h)=y\) only if some prior execution of \(f_d\) on an input hashing to \(h\) produced \(y\). | M1 |
| H6 | **Replica Cauchy condition.** Materialize/CDC copy only when \(\sigma=\mathrm{sealed}\). | R1 |

H2 is graded, not binary. `CommandRegularity` classifies commands: builtins discharge H2 as a theorem; Julia discharges it as uniqueness of a *pinned discretization* (shadowing of a discrete map); container images leave it as a Picard-style regularity *contract*. Isolation (`SandboxPolicy.Isolated`) admits a contract-grade command; it does **not** discharge H2. Test: `TestIsolatedSandboxImpureIsNotUnique` (M10). See [nature-inspired.md](nature-inspired.md).

---

## 3. Evolution

### 3.1 Domino step

Given sealed \(s\), current \(Y\), and next name \(d\):

1. \(u \leftarrow \mathrm{resolveInputs}(d,x,Y)\)
2. \(h \leftarrow \mathrm{SHA256}(u)\)
3. If \(M(\mathrm{id}(s),d,h)=y\), set \(Y(d)\leftarrow y\) (reuse).
4. Else set \(y\leftarrow f_d(u)\), write \(M\), set \(Y(d)\leftarrow y\).

Chain evolution \(\Phi_{\mathrm{chain}}\) is the \(n\)-fold composition of this step along \(D\).

### 3.2 Wheel step

`AdvanceAfterCompletion` (see `pkg/wheel`):

\[
\Phi_W(i,t,r)=
\begin{cases}
(i+1,t,r) & i+1 < n_c \\
(0,t+\Delta t,r+1) & i+1 = n_c
\end{cases}
\]

with an absorbing stop when \(r+1\) hits `maxRotations`.

### 3.3 Routing

\(\rho(\mathrm{event})\) is the first matching time-slice override, else the first matching partition rule, else the default universe. This is a total function whenever a default is configured.

---

## 4. Theorems

### Theorem D1 (no trajectory without Cauchy data)

Assume H1. If \(\sigma=\mathrm{open}\), `Engine.Run` returns an error and does not write a worldline.

**Proof.** Immediate from the guard in `engine.go` (`snapshot … is not sealed; cannot execute deterministically`). ∎

### Theorem D2 (unique hash worldline)

Assume H1–H4 and a fixed chain \(D\). Let \(s\) be sealed with payload \(x\). Then any two complete executions produce the same sequence of input hashes, output hashes, and final output.

**Proof.** By induction on chain position \(k\).

*Base.* \(k=0\): no outputs. Snapshot ID is \(\mathrm{SHA256}(\tau,x)\) truncated to 128 bits; H4 gives uniqueness on the occurring domain.

*Step.* Resolved input \(u_k\) is a function of \(x\) and \(Y_{<k}\) (H3). Those are unique by induction. \(h_k=\mathrm{SHA256}(u_k)\) is unique. Either the memo returns the unique stored \(y_k\) (H5, and H2 says it equals \(f_{d_k}(u_k)\)) or \(f_{d_k}(u_k)\) is unique (H2). Output hash is unique by H4. ∎

**Remark.** Wall-clock `Timestamp` on the replay log is *not* part of the worldline. Two runs may differ in timestamps and in the `reused` bit; they must not differ in hashes or outputs.

### Theorem M1 (memo observational equivalence)

Assume H2, H4, H5. A memo hit for \((\mathrm{id}(s),d,h)\) yields the same output string and output hash as executing \(f_d\) on any input that hashes to \(h\).

**Proof.** H5 says the stored \(y\) came from such an execution. H2 says all such executions agree. H4 says the output hash agrees. ∎

### Theorem M9 (tamper-evident replay spine)

Assume H4. Let \(L_0=\mathrm{id}(s)\) and \(L_k=\mathrm{SHA256}(L_{k-1}\Vert h_k^{\mathrm{in}}\Vert h_k^{\mathrm{out}})\). A complete run writes \(L_k\) on each replay entry. `VerifySpine` accepts the log iff every stored link equals this recurrence.

**Proof.** Inspection of `attachSpine` and `VerifySpine`: each entry's `PrevLink` is the previous \(L\), and `Link` is `hash.Link` of that prefix plus the two hashes. Mutating an output hash without recomputing subsequent links fails the equality. This is a hash chain (Merkle spine), not a Merkle tree and not a consensus ledger. ∎

### Proposition M10 (isolation does not imply H2)

An isolated sandbox is a sufficient *admission* condition for a contract-grade command, not a sufficient *uniqueness* condition. `sandbox:impure` is admitted by `AllowDeterministic` when `SandboxPolicy.Isolated` holds, and two independent evaluations need not agree.

**Proof.** `AllowDeterministic` returns nil for any contract-grade command under isolation, without inspecting the command body. `sandboxImpure` returns a wall-clock / counter payload that is not a function of the input string. Tests: `TestIsolatedSandboxDoesNotImplyUniqueness`, `TestExecuteSandboxImpureIsNotUnique`, `TestIsolatedSandboxImpureIsNotUnique`. ∎

### Theorem M11 (four-DSL orthogonality)

Assume H1–H4 and a builtin chain. Let \(W\) and \(W'\) be workflows that agree on snapshot payload, time-slice, and execution chain, and differ only in provisioning and routing fields. Then \(\mathrm{id}(s)=\mathrm{id}(s')\) and the hash worldlines coincide.

**Proof.** `hash.SnapshotID` is a function of \((\tau,x)\) only. `Engine.Run` uses provisioning solely as the Picard-gate sandbox policy; builtins skip that gate. Routing fields are not read on the hot path. Therefore \(\Phi_{\mathrm{chain}}\) is independent of those axes. Changing the data payload changes \(\mathrm{id}(s)\). Test: `TestProvisioningOrthogonalToBuiltinWorldline`. ∎

### Theorem E1 (ensemble collapse)

Let \(E=\{x_1,\ldots,x_N\}\) be distinct payloads, \(\mu\) uniform on \(E\). Then \(H_{\mathrm{ens}}(\mu)=\log_2 N\). After sealing a member \(x^*\in E\), the posterior is \(\delta_{x^*}\) and \(H_{\mathrm{ens}}=0\).

**Proof.** Shannon entropy of the uniform law on \(N\) atoms is \(\log_2 N\). A Dirac mass has entropy 0. `SealEnsemble` implements the conditioning. ∎

**Corollary (non-claim).** \(H_{\mathrm{pay}}(x^*)\) is independent of the sealing predicate. Sealing is not a compression algorithm.

### Theorem W1 (unique cylinder successor)

For \(n_c>0\), \(\Phi_W\) is a total function. After \(n_c\) steps from \((0,t,r)\), the state is \((0,t+\Delta t,r+1)\) if the rotation cap allows.

**Proof.** Inspection of `AdvanceAfterCompletion`. Period: the index increments \(n_c-1\) times without wrapping, then wraps once. ∎

### Theorem M12 (player-piano lookahead)

Assume \(n_c>0\). `Lookahead` is a function of \((\mathrm{wheelName},\mathrm{contexts},\mathrm{state},\Delta t,\mathrm{maxRotations})\). If the successor is not absorbing, its name equals `WorkflowName` of the state produced by \(\Phi_W\), which is the Workflow `BuildWorkflow` would create for that slot.

**Proof.** `Lookahead` is `AdvanceAfterCompletion` followed by `ActiveContextName` and `WorkflowName`. Two calls on equal inputs therefore agree. When \(\Phi_W\) is absorbing (`Done`), the name is empty. Tests: `TestLookaheadIsAFunctionOfState`, `TestLookaheadNameMatchesBuiltWorkflow`, `TestLookaheadStopsWhenDone`. ∎

### Theorem M13 (unique chain names)

An execution chain is a sequence of distinct names. Duplicate labels are rejected before any step of \(\Phi_{\mathrm{chain}}\).

**Proof.** `theory.UniqueNames` scans the chain into a set; a repeated or empty name is an error. `Engine.Run` also refuses a duplicate in the domino catalog (map overwrite would silently change which \(f_d\) runs). Tests: `TestUniqueNamesRejectsDuplicates`, `TestRejectsDuplicateChainNames`. ∎

### Theorem M14 (engine-enforced causal past)

Assume the causal-past theorem. `Engine.Run` calls `AllowedReads` on every `fromDomino` and `dependsOn` name before resolving inputs. A future or unknown name fails closed even if a later map lookup might have succeeded.

**Proof.** Inspection of the `Run` loop. Tests: `TestDominoCannotReadFutureOutput`, `TestRejectsDependsOnOutsideCausalPast`. ∎

### Proposition M15 (partial isolation is not a cage)

`SandboxPolicy.Isolated` is the conjunction of network-none and read-only-root. Either flag alone does not admit a contract-grade command.

**Proof.** Inspection of `Isolated` and `AllowDeterministic`. Tests: `TestPartialSandboxIsolationIsNotACage`, `TestPartialSandboxIsolationRejectedByEngine`. ∎

### Theorem M16 (wall-clock gauge)

`Timestamp` on a replay entry is not an argument of `hash.Link` or `Worldline`. Mutating it leaves `VerifySpine` and the hash worldline invariant. Independent sealed builtin runs share `HeadLink`.

**Proof.** `attachSpine` hashes `(prev, inputHash, outputHash)` only. Tests: `TestWallClockNotInWorldline`. ∎

### Theorem M17 (sealed write-once)

A sealed snapshot identity may be rewritten only with identical \((\tau,x)\). A different payload, time-slice, or an unseal is rejected.

**Proof.** `SaveSnapshot` on SQLite and TSDB consults the existing row/file; `refuseSealedMutation` returns `ErrSealedOverwrite` unless the write is a no-op. Tests: `TestSealedSnapshotIsWriteOnce`. ∎

### Theorem M18 (memo integrity)

Assume H5. Writing \(M(s,d,h)=y'\) when \(M(s,d,h)=y\) already and \(y'\neq y\) is an error. Identical rewrite is idempotent.

**Proof.** `SaveResult` looks up the key; `refuseMemoConflict` returns `ErrMemoConflict` on disagreement. Tests: `TestMemoRejectsConflictingOutputForSameKey`. ∎

### Theorem M19 (collision-safe truncation)

If `WorkflowName` exceeds 63 characters, the suffix is a content hash of the full name, not its length. Equal-length distinct names therefore remain distinct after truncation.

**Proof.** SHA-256 of the full name, 8 hex characters, prefix truncated to fit. Test: `TestWorkflowNameTruncationDistinguishesEqualLength`. ∎

### Theorem M20 (stepwise spine)

`RunSingleFrom(..., prevLink)` writes `PrevLink=prevLink` so a sequence of standalone steps is a single hash chain from the snapshot ID.

**Proof.** `attachSpine(prevLink, entry)` after the step. Test: `TestRunSingleSpineChainsAcrossSteps`. ∎

### Theorem M21 (persisted spine)

A completed `Engine.Run` writes `PrevLink`/`Link` on each replay row. `ListReplay` reloads a chain that `VerifySpine` accepts.

**Proof.** `SaveResult` inserts the links; `ListReplay` orders by row id. Test: `TestReplayLogRoundTripsSpine`. ∎

### Theorem M22 (content-addressed event identity)

`events.EventID` is `hash.Link` of type, snapshot ID, worldline, universe, and workflow name. `OccurredAt` is excluded.

**Proof.** Inspection of `EventID`. Test: `TestSnapshotEventIDIsFunctionOfPayload`. ∎

### Theorem M23 (no orphan excitations)

A CDC domino-result envelope applies only when the target already holds a sealed parent snapshot.

**Proof.** `applyDominoResult` calls `GetSnapshot` and requires `sealed`. Test: `TestApplyDominoResultRequiresSealedParent`. ∎

### Theorem M24 (complete copy)

`Materialize` and `ExportFromStore` error if a named chain member has no stored result. They do not return success with a short count.

**Proof.** The missing-result branch returns an error instead of `continue`. Tests: `TestMaterializeFailsOnMissingDomino`, `TestExportFromStoreFailsOnMissingDomino`. ∎

### Theorem M25 (routing priority)

Time-slice overrides are evaluated before partition rules. An event that matches both is routed to the time-slice target.

**Proof.** Inspection of `Router.Resolve`. Test: `TestRoutingPriorityTimeSliceBeatsPartition`. ∎

### Theorem M26 (unique partition outcome)

If two universes match the same partition labels, `Resolve` returns an error rather than silently taking the first.

**Proof.** The partition loop collects matches; `len>1` is an error. Test: `TestAmbiguousPartitionMatchRejected`. ∎

### Theorem F1 (windowed self-similarity)

Let \(U(d,k,v)=\mathrm{Unfold}(d,k,\mathrm{root},v)\) with arity \(k\ge 1\) and additive child partition. Then \(\mathrm{Coarsen}(U(d,k,v))\) is shape-equal and value-equal to \(U(d-1,k,v)\) for all \(d\ge 1\).

**Proof.** By induction on \(d\). For \(d=1\), children are leaves; Coarsen sums them to \(v\) and drops children, matching \(U(0,k,v)\). For \(d>1\), Coarsen acts as \(U(d-1)\) on each child subtree (induction), and the parent value is the sum of child values, which is \(v\) by construction of Unfold. Shape is the perfect \(k\)-ary tree of height \(d-1\). ∎

### Theorem (causal past)

Let \(D=(d_1,\ldots,d_n)\). The readable names at \(d_k\) are \(\{\mathrm{snapshot}\}\cup\{d_1,\ldots,d_{k-1}\}\). In particular \(d_j\) for \(j\ge k\) is not readable.

**Proof.** `resolveInputs` looks up `fromDomino` in `priorOutputs`, which only contains previously executed names in chain order. `AllowedReads` is the static version of the same prefix check. ∎

### Theorem R1 (sealed-only signaling)

Assume H6. `Materialize` and CDC export/apply fail on \(\sigma=\mathrm{open}\). Therefore a replica store cannot observe a live source snapshot.

**Proof.** Guards in `replica.Materialize`, `cdc.ExportFromStore`, `cdc.applySnapshot`. ∎

### Theorem C1 (routing is functional)

`Router.Resolve` is deterministic: it has no hidden entropy; the same `MultiverseSpec` and `SnapshotEvent` yield the same `Target` or the same error.

**Proof.** The procedure is a sequence of equality tests on the spec and event. ∎

---

## 5. Remainder left in nature (inspired-by, not theorems of physics)

These are kept as **mimics**, the way an ANN keeps "neuron" without keeping glia. Each has an executable counterpart in `pkg/theory` at a named grade. None is a laboratory identity.

| Remainder | Grade of the mimic | Code |
|-----------|--------------------|------|
| Container purity | Contract (Picard regularity hypothesis) | `CommandRegularity` |
| Julia / floating point | Empirical discretization / shadowing | pinned Manifest tests; `RegularityPinned` |
| Landauer heat (joules) | Bookkeeping of irreversible *steps* | `LogicalWork`, `ReplaySaves` |
| Everett amplitudes | Classical non-interfering histories | `History`, `Branch`, `Interfere` |
| Mandelbrot set as a polynomial | Explorer viewport + IFS dimension + escape-time window | `Unfold`/`Coarsen`, `SimilarityDimension`, `EscapeTime`, `LeafCount` |
| Biological life | Control homeostasis | `ReconcileError`, `Homeostatic` |

The method is [nature-inspired.md](nature-inspired.md). We mimic nature; we do not prove it.

---

## 6. Mapping onto code

| Object | Code |
|--------|------|
| Seal gate | `controller/pkg/engine/engine.go` `Run` |
| Hash / snapshot ID | `controller/pkg/hash` |
| Memo | `store.LookupMemo` / `SaveResult` |
| Wheel \(\Phi_W\) | `controller/pkg/wheel/rotation.go` |
| Player-piano lookahead | `wheel.Lookahead`, `preProvisionNext` |
| Routing \(\rho\) | `controller/pkg/routing/router.go` |
| Causal past | `controller/pkg/theory/causal.go` |
| Ensemble entropy | `controller/pkg/theory/entropy.go` |
| Windowed aggregation / explorer | `controller/pkg/theory/aggregation.go` |
| Regularity ladder | `controller/pkg/theory/regularity.go` |
| Logical work (Bennett bookkeeping) | `controller/pkg/theory/work.go` |
| Classical histories | `controller/pkg/theory/histories.go` |
| Replay spine | `controller/pkg/theory/spine.go`, `hash.Link` |
| Homeostasis | `controller/pkg/theory/homeostasis.go` |
| Cauchy replica | `controller/pkg/replica/materialize.go`, `controller/pkg/cdc` |
