# Correspondences: Physics Structures and the KBL Fabric

**Abstract.** We construct eight **nature-inspired** correspondences between named structures in mechanics, information theory, and biology and the KBL compute fabric. The method is the same as neural nets copying the neuron: keep a usable abstraction, discard the wetware, say which is which. Each correspondence is a 5-tuple: source definition, target definition, morphism, transferred properties, remainder left in nature. Theorems that transfer are tests in `controller/pkg/theory`. KBL mimics nature; it does not prove physics.

The method paper is [nature-inspired.md](nature-inspired.md).

---

## 0. Method

A **correspondence** \(C = (S, T, \varphi, P_+, P_-)\) consists of:

| Slot | Meaning |
|------|---------|
| \(S\) | Source structure, defined as a physicist or information theorist would define it |
| \(T\) | Target structure, defined on KBL's discrete state space |
| \(\varphi\) | A mapping of objects and arrows of \(S\) onto objects and arrows of \(T\) |
| \(P_+\) | Properties of \(S\) that \(\varphi\) preserves (theorems) |
| \(P_-\) | Properties of \(S\) that \(\varphi\) does **not** preserve (limits of the analogy) |

This is the same methodological move as an artificial neuron (McCulloch & Pitts 1943): cortex is not running in the matrix multiply; the *weighted combination* is. We keep the name when the copied structure is why the name was chosen, and we leave the biological or physical remainder in nature.

**Rule.** If nothing was copied, the name is branding. If something was copied, grade the mimic (theorem / empirical discretization / contract / bookkeeping / explorer / branching / control) rather than calling it a failed proof of physics.

---

## I. Newtonian determinism → unique discrete trajectories

**Source \(S\).** Classical mechanics: given a sufficiently regular vector field, the Cauchy problem \(\dot x = F(x)\), \(x(t_0)=x_0\) has a unique local solution (Picard–Lindelöf). The "laws" plus the initial data determine the worldline. Time-reversibility is extra structure (existence of a first integral / symplectic form), not implied by uniqueness.

**Target \(T\).** A sealed snapshot \(x\) (initial data) and a chain of dominos \(d_1,\ldots,d_n\), each a map from resolved inputs to an output string. One step of \(\Phi\) is "execute the next domino or reuse its memo." The engine refuses to start unless the snapshot is sealed (`engine.Run`).

**Morphism \(\varphi\).** Initial data \(\mapsto\) sealed snapshot content. Vector field \(\mapsto\) the sequence of declared functions. Integration \(\mapsto\) ordered chain execution. Worldline \(\mapsto\) the sequence of \((\mathrm{inputHash},\mathrm{outputHash})\) pairs (the replay log minus wall-clock timestamps).

**Transfers \(P_+\).**

- **Uniqueness (Thm. D2).** If every \(d_i\) is a function and hashing is injective on the domain of interest, the hash worldline is unique. Test: `TestTheoremD2SealedSnapshotHasUniqueTrajectory`, `TestSnapshotReplayDeterministic`.
- **Tamper-evident spine (Thm. M9).** The replay log is a hash chain from the snapshot ID. Test: `TestReplaySpineIsTamperEvident`.
- **No evolution without Cauchy data (Thm. D1).** Unsealed snapshots have no trajectory. Test: `TestTheoremD1UnsealedSnapshotHasNoTrajectory`.

**Does not transfer \(P_-\) (left in nature, as glia are left out of an ANN).**

- \(F=ma\), symplectic structure, energy conservation, continuous time, or time-reversal. The replay log is not a reversed integration; it is an audit of a forward unique path.

**Loose linkage (nature-inspired).** Uniqueness is *conditional on regularity*, copying Picard–Lindelöf rather than claiming every vector field is Lipschitz. `CommandRegularity` grades that hypothesis: builtins are theorem-grade, Julia is a pinned discretization (shadowing of a discrete map, not uniqueness on \(\mathbb{R}\)), containers are a contract (`deterministic: true`). An isolated sandbox is a Faraday cage: it admits the contract, it does not make an impure command unique (`sandbox:impure`, M10). Tests: `TestCommandRegularityLadder`, `TestIsolatedSandboxImpureIsNotUnique`, D1, D2.

---

## II. Entropy → ensemble collapse at seal time

**Source \(S\).** Shannon (1948): for a discrete random variable \(X\) with law \(\mu\), \(H(X)=-\sum\mu(x)\log_2\mu(x)\). Thermodynamic entropy is a related but distinct functional on a physical ensemble. Landauer (1961) bounds heat from *logically irreversible* bit erasure; it is not a statement about YAML.

**Target \(T\).** Two different functionals, which the original prose conflated:

1. **Ensemble (epistemic) entropy** \(H_{\mathrm{ens}}(\mu)\): uncertainty about *which payload* a computation will see, before sealing.
2. **Payload entropy** \(H_{\mathrm{pay}}(x)\): Shannon entropy of the byte histogram of a single payload.

**Morphism \(\varphi\).** Live mutable data \(\mapsto\) a prior \(\mu\) on candidate payloads. Sealing to content \(x^*\) \(\mapsto\) conditioning, \(\mu'=\delta_{x^*}\). Memoization \(\mapsto\) not repeating a completed function application (Bennett-style reuse of recorded results — computational, not thermodynamic).

**Transfers \(P_+\).**

- **Collapse (Thm. E1).** If \(\mu\) is uniform on \(N\) distinct payloads, \(H_{\mathrm{ens}}=\log_2 N\). After sealing one member, \(H_{\mathrm{ens}}=0\). Test: `TestEnsembleEntropyCollapsesOnSeal`. Distinct sealed payloads produce distinct snapshot IDs: `TestTheoremE1EngineSealCollapsesInputEnsemble`.
- **Payload entropy is invariant under sealing.** Sealing does not compress \(x^*\). The tests treat this as a non-claim, not a theorem to "prove entropy went down."

**Does not transfer \(P_-\) (left in nature).**

- Boltzmann entropy of a physical macrostate, and any drop in payload Shannon entropy at `sealed: true`.
- Heat in joules. See VII: we copy Bennett's *bookkeeping*, not Landauer's laboratory bound.

**Loose linkage (nature-inspired).** "Low-entropy snapshot" is the information-theoretic mimic of "freeze the ensemble before you integrate." Sealing is not a refrigerator.

---

## III. Relativistic locality → causal past during compute

**Source \(S\).** In Minkowski spacetime, influences propagate inside the light cone. A Cauchy surface is a spacelike hypersurface that determines the future evolution. Lamport (1978) extracted the *partial order* of this idea for distributed systems without keeping the metric.

**Target \(T\).** Event sorts: `seal(s)`, `exec(d)`, `complete(chain)`, `route(e)`, `materialize(replica)`. During `exec(d)`, readable data is the snapshot plus outputs of strictly earlier dominos (`engine.resolveInputs`). Cross-universe copies occur only after completion, and only for sealed snapshots (`replica.Materialize`, `cdc.Apply`).

**Morphism \(\varphi\).** Light-cone \(\mapsto\) Lamport-style causal past of `exec(d)`. Cauchy surface \(\mapsto\) a sealed snapshot (a frozen input hypersurface). Spacelike separation during compute \(\mapsto\) universes do not read each other's live stores on the hot path. Signals \(\mapsto\) sealed events on the bus.

**Transfers \(P_+\).**

- **Causal past is a prefix (Thm. C-past).** `CausalPast(chain, d)` is `{snapshot} ∪ prefix of chain through d`. Test: `TestCausalPastIsPrefix`.
- **Future reads forbidden.** A domino cannot declare `fromDomino` of a later step. Tests: `TestFutureReadRejected`, `TestDominoCannotReadFutureOutput`.
- **No signaling of live data (Thm. R1).** Unsealed snapshots cannot materialize or CDC-export. Tests: `TestTheoremR1ReplicaRequiresSealedCauchyView`, `TestExportFromStoreRejectsUnsealedSnapshot`.

**Does not transfer \(P_-\).**

- Lorentz invariance, a metric \(ds^2\), a finite invariant speed \(c\), relativity of simultaneity as physics. The event bus has an engineering latency, not a Minkowski causal structure.
- Global snapshot isolation across *all* universes in wall-clock time (replicas are eventually consistent).

**Loose linkage, stated precisely.** This is Lamport's correspondence, instantiated on KBL's engine and replica path. It is the strongest physics-shaped invariant in the system because it is already a theorem of the code: the readable set is a causal past, and only completed Cauchy data is allowed to leave a universe.

---

## IV. "Multiverse" → product of independent dynamical systems

**Source \(S\).** Two very different physical ideas share the name:

- Everett (1957): branching of a universal wavefunction.
- Statistical mechanics / multi-physics simulation: many systems, possibly with different Hamiltonians, weakly coupled.

**Target \(T\).** A `Multiverse` is a routing table over `PluggableUniverse` objects. Each universe has its own engine, store, and provisioning ("laws"). Routing \(\rho\) is a function of time-slice overrides, partition labels, and a default (`pkg/routing`). Universes do not call each other's controllers; they emit and consume sealed events.

**Morphism \(\varphi\).** Universe \(\mapsto\) a pair \((S_u,\Phi_u)\). Different laws \(\mapsto\) different \(\Phi_u\) (Julia vs Go, different images, different stores). Weak coupling \(\mapsto\) \(\rho\) plus replica materialization. Parallel time slices \(\mapsto\) independent initial-value problems, not branches of one amplitude.

**Transfers \(P_+\).**

- **Routing is a function (Thm. C1).** Same event, same spec \(\Rightarrow\) same target. Test: `TestTheoremC1RoutingIsAFunction`.
- **Independence of hot-path compute.** A universe computes only against its local store. Coupling is after-the-fact and sealed.

**Does not transfer \(P_-\) (left in nature).**

- Hilbert space, unitarity, interference of amplitudes. There is no inner product.

**Loose linkage (nature-inspired).** What is worth copying from Everett is **non-interfering branches with classical records** — the same abstraction as an ensemble of independently trained nets. `Branch` fans a sealed snapshot into other universes; `Interfere(..., liveShare=false)` is always false. We copy the branching *bookkeeping*, not the wavefunction. Tests: `TestHistoriesDoNotInterfereWhenSealed`.

---

## V. Compute Wheel → discrete flow on a cylinder

**Source \(S\).** A planar rotation is motion on \(S^1\). Combined with a time coordinate that only advances, the orbit lives on a cylinder \(S^1\times\mathbb{R}\) (a helix if time is plotted). Discrete analogue: a cyclic seat index in \(\mathbb{Z}_n\) and a slice in a discrete time lattice.

**Target \(T\).** `wheel.AdvanceAfterCompletion`: seat \(i \mapsto i+1 \bmod n\); when \(i\) wraps, slice \(\mapsto\) slice \(+\Delta t\) and rotation count increments. Unique successor for \(n>0\). `Lookahead` names that successor without advancing the live wheel (player-piano / `preProvisionNext`).

**Morphism \(\varphi\).** Seat \(\mapsto\) ComputeContext. Angle \(2\pi i/n\) \(\mapsto\) index \(i\). Continuous rotation \(\mapsto\) discrete advance after slot completion. Time \(\mapsto\) `CurrentTimeSlice`.

**Transfers \(P_+\).**

- **Unique successor (Thm. W1).** \(\Phi\) is a function. Test: `TestTheoremW1WheelHasUniqueSuccessor`.
- **Period in the angular coordinate.** \(n\) advances wrap the seat index and add exactly one interval to the slice. Test: `TestCylinderPeriodAdvancesTimeOnce`.
- **Lookahead is a function (Thm. M12).** The next Workflow name is determined by current state. Tests: `TestLookaheadIsAFunctionOfState`, `TestLookaheadNameMatchesBuiltWorkflow`.

**Does not transfer \(P_-\).**

- Angular momentum, torque, a Lagrangian on \(TS^1\), Poincaré recurrence of a Hamiltonian flow, or continuous time. `maxRotations` is an engineering stop, not a conserved quantity.
- Isochronous rotation: slots complete when workflows complete, not after a fixed \(\Delta\theta\).

**Loose linkage, stated precisely.** The Ferris-wheel image is a **discrete cylinder map**. That is enough to keep the name; it is not rotational mechanics.

---

## VI. Windowed Mandelbrot → coarsenable self-similar aggregation

**Source \(S\).** Mandelbrot (1982): sets that are statistically self-similar across scales; the Mandelbrot set itself is the connectedness locus of \(z\mapsto z^2+c\). Wilson renormalization and multigrid methods share a *different* idea that the blog actually needs: a hierarchy that can be coarse-grained, with only a finite window materialized.

**Target \(T\).** `theory.Unfold(depth, arity, …)` builds a finite perfect \(k\)-ary tree (the window). Child values partition the parent additively. `theory.Coarsen` drops one level and restores values by summation. The live engine chain (`WorkflowFromUnfold`) identity-copies the snapshot at the leaves and `builtin:coarsen`s parents (sum of child `v`).

**Morphism \(\varphi\).** Infinite self-similar hierarchy \(\mapsto\) the inductive tree that *would* continue forever. Viewport \(\mapsto\) finite depth \(D\). Zoom-out / renormalization \(\mapsto\) `Coarsen`. Conservation of a scalar under coarse-graining \(\mapsto\) additive partition + sum.

**Transfers \(P_+\).**

- **Self-similarity of windows (Thm. F1).** `Coarsen(Unfold(d,k))` is shape- and value-equal to `Unfold(d-1,k)`. Tests: `TestCoarsenRecoversShallowerWindow`, `TestTheoremF1WindowedAggregation`.
- **Live fold-up (Phase 37).** Root `v` = `LeafCount(d,k)` × leaf `v`; `agg(d,k) = agg(d-1,k)·k`. Tests: `TestWorkflowFromUnfoldAggregatesToLeafCount`.
- **Finiteness of the window.** Height is exactly \(D\). Test: `TestWindowIsFinite`.

**Does not transfer \(P_-\) (left in nature).**

- The connectedness locus of \(z\mapsto z^2+c\), and a fractal *Kubernetes* scheduler that iterates that map. The production controller does not yet spawn pods from `Unfold`.

**Loose linkage (nature-inspired).** The blog's Mandelbrot is the **explorer**: an infinite generated object of which only a window is drawn. We copy that viewport.

- Finite unfolding + coarsening (theorem F1).
- Similarity dimension of the windowed IFS (`SimilarityDimension`); additive trees have \(D=1\), matching a conserved scalar.
- Leaf count as a wheel's seats (`LeafCount`) — the current window can be the Ferris-wheel circuit.
- Escape-time (`EscapeTime`) as the explorer's iteration budget: a finite window on \(z\mapsto z^2+c\), used the way a sigmoid is used (a shape from nature, not a proof of cortex or of complex dynamics).

Tests: F1, `TestSimilarityDimensionAdditiveTreeIsOne`, `TestWheelWindowLeafCount`, `TestEscapeTimeIsAFiniteWindow`.

---

## VII. Landauer / Bennett → recorded intermediates (bookkeeping mimic)

**Source \(S\).** Landauer (1961): erasing one bit at temperature \(T\) costs at least \(kT\ln 2\). Bennett (1973, 1982): keep intermediates and you need not pay that at every step.

**Target \(T\).** Memo table keyed by \((snapshotID, dominoID, inputHash)\). `LogicalWork` counts evaluations versus reuses.

**Morphism \(\varphi\).** "Keep the intermediate instead of erasing it" \(\mapsto\) memoization. Irreversible re-evaluation \(\mapsto\) a cache miss. Joules \(\mapsto\) *steps* (the ANN move: "energy" as a loss, not ATP).

**Transfers \(P_+\).**

- **Observational equivalence (Thm. M1).** A hit matches a recompute in outputs and hashes, given purity and collision resistance.
- **Replay saves irreversible steps.** `ReplaySaves(Work(first), Work(second))`. Tests: M1, `TestLogicalWorkReplaySaves`.

**Left in nature.** A bound in joules. The cluster is not a thermodynamically accounted engine. The mimic is the *accounting shape*.

---

## VIII. Organism / lifeform → homeostatic control (control mimic)

**Source \(S\).** Ashby (1956): ultrastable systems that keep essential variables within bounds by feedback. Autopoiesis is stronger and is left in nature.

**Target \(T\).** Kubernetes reconciliation. `ReconcileError(desired, observed)` is zero iff spec matches status (`Homeostatic`).

**Morphism \(\varphi\).** Essential variables \(\mapsto\) CR status vs spec. Feedback \(\mapsto\) reconciler. Organism boundary \(\mapsto\) a universe's stores and engines.

**Transfers \(P_+\).** The control loop is a fair copy of homeostasis. Test: `TestHomeostasisMatchesSpec`. It is Kubernetes' loop as much as KBL's; we still name the mimic.

**Left in nature.** Metabolism, reproduction, evolution, DNA. **KBL** keeps "lifeform" the way "neural net" keeps "neural."

---

## Summary table

| # | Name in the blog | Copied structure | Grade | Primary tests |
|---|------------------|------------------|-------|----------------|
| I | Newtonian determinism | Conditional unique Cauchy trajectory | Theorem (builtins); empirical discretization (Julia); contract (containers) | D1, D2, regularity ladder |
| II | Low-entropy snapshot | Ensemble Shannon collapse at seal | Theorem for \(H_{\mathrm{ens}}\) | E1 |
| III | Locality / isolation | Causal past + sealed-only signaling | Theorem | causal past, R1 |
| IV | Multiverse | Product of systems + non-interfering classical branches | Theorem for routing; branching mimic for Everett-shaped fan-out | C1, histories |
| V | Ferris / Compute Wheel | Discrete cylinder map + player-piano lookahead | Theorem | W1, M12 |
| VI | Windowed Mandelbrot | Explorer viewport + coarsenable IFS | Theorem F1; explorer mimic for \(z^2+c\) | F1, dimension, escape-time, wheel leaves |
| VII | Minimize entropy / maximize caching | Recorded intermediates | Theorem for outputs; bookkeeping mimic for Landauer | M1, logical work |
| VIII | Lifeform | Spec/status homeostasis | Control mimic | homeostasis |

Wiring F1 into ComputeWheel as a live scheduler remains an implementation step, not a reason to drop the explorer mimic. Joules remain in nature; steps are what we count.
