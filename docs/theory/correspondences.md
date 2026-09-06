# Correspondences: Physics Structures and the KBL Fabric

**Abstract.** We construct eight correspondences between named structures in mechanics, information theory, and relativistic causality and the KBL compute fabric. Each correspondence is a 5-tuple: source definition, target definition, morphism, transferred properties, non-transferred properties. Theorems that transfer are implemented as tests in `controller/pkg/theory` and the engine/wheel/routing/replica packages. We claim a *linkage of structure*, not an identity of ontology. KBL is not a physical theory.

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

This is the same methodological move as Lamport's clocks (1978): special relativity is not running on the network; the *causal partial order* is. We keep the name when the transferred structure is the reason the name was chosen, and we drop the physical remainder.

**Rule.** If \(P_+\) is empty, the name is branding and must not appear in a theorem statement.

---

## I. Newtonian determinism → unique discrete trajectories

**Source \(S\).** Classical mechanics: given a sufficiently regular vector field, the Cauchy problem \(\dot x = F(x)\), \(x(t_0)=x_0\) has a unique local solution (Picard–Lindelöf). The "laws" plus the initial data determine the worldline. Time-reversibility is extra structure (existence of a first integral / symplectic form), not implied by uniqueness.

**Target \(T\).** A sealed snapshot \(x\) (initial data) and a chain of dominos \(d_1,\ldots,d_n\), each a map from resolved inputs to an output string. One step of \(\Phi\) is "execute the next domino or reuse its memo." The engine refuses to start unless the snapshot is sealed (`engine.Run`).

**Morphism \(\varphi\).** Initial data \(\mapsto\) sealed snapshot content. Vector field \(\mapsto\) the sequence of declared functions. Integration \(\mapsto\) ordered chain execution. Worldline \(\mapsto\) the sequence of \((\mathrm{inputHash},\mathrm{outputHash})\) pairs (the replay log minus wall-clock timestamps).

**Transfers \(P_+\).**

- **Uniqueness (Thm. D2).** If every \(d_i\) is a function and hashing is injective on the domain of interest, the hash worldline is unique. Test: `TestTheoremD2SealedSnapshotHasUniqueTrajectory`, `TestSnapshotReplayDeterministic`.
- **No evolution without Cauchy data (Thm. D1).** Unsealed snapshots have no trajectory. Test: `TestTheoremD1UnsealedSnapshotHasNoTrajectory`.

**Does not transfer \(P_-\).**

- \(F=ma\), symplectic structure, energy conservation, continuous time, or time-reversal. The replay log is not a reversed integration; it is an audit of a forward unique path.
- Purity of arbitrary containers. `deterministic: true` is a **contract**, not a proof. Julia evidence is pinned-Manifest empirical identity, not a theorem about \(\mathbb{R}\).

**Loose linkage, stated precisely.** The analogue of "Newtonian" here is *the well-posedness of a discrete Cauchy problem*, not classical mechanics as a physical law. That is the same sense in which a functional program is "deterministic."

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

**Does not transfer \(P_-\).**

- Boltzmann entropy of a physical macrostate.
- A numerical drop in \(H_{\mathrm{pay}}\) at `sealed: true`.
- Landauer heat in the cluster. Memoization avoids repeating *work*; it does not imply \(kT\ln 2\) saved per skipped domino (the CPU, the store, and the cache lookup have their own costs). Landauer is Correspondence VII, marked heuristic.

**Loose linkage, stated precisely.** The blog phrase "low-entropy snapshot" is correct if and only if entropy means \(H_{\mathrm{ens}}\). It is false if it means thermodynamic \(S\) or payload compressibility. This revision adopts the information-theoretic reading and rejects the other two as theorems.

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

**Does not transfer \(P_-\).**

- Hilbert space, unitarity, decoherence, interference, "many-worlds interpretation." There is no amplitude. Calling this Everett would be a category error.
- Conservation of probability or a measure on the set of universes.

**Loose linkage, stated precisely.** Keep the name *Multiverse* as a product-and-route construction (many law-sets, weak coupling). Cite Everett only to **decline** that correspondence.

---

## V. Compute Wheel → discrete flow on a cylinder

**Source \(S\).** A planar rotation is motion on \(S^1\). Combined with a time coordinate that only advances, the orbit lives on a cylinder \(S^1\times\mathbb{R}\) (a helix if time is plotted). Discrete analogue: a cyclic seat index in \(\mathbb{Z}_n\) and a slice in a discrete time lattice.

**Target \(T\).** `wheel.AdvanceAfterCompletion`: seat \(i \mapsto i+1 \bmod n\); when \(i\) wraps, slice \(\mapsto\) slice \(+\Delta t\) and rotation count increments. Unique successor for \(n>0\).

**Morphism \(\varphi\).** Seat \(\mapsto\) ComputeContext. Angle \(2\pi i/n\) \(\mapsto\) index \(i\). Continuous rotation \(\mapsto\) discrete advance after slot completion. Time \(\mapsto\) `CurrentTimeSlice`.

**Transfers \(P_+\).**

- **Unique successor (Thm. W1).** \(\Phi\) is a function. Test: `TestTheoremW1WheelHasUniqueSuccessor`.
- **Period in the angular coordinate.** \(n\) advances wrap the seat index and add exactly one interval to the slice. Test: `TestCylinderPeriodAdvancesTimeOnce`.

**Does not transfer \(P_-\).**

- Angular momentum, torque, a Lagrangian on \(TS^1\), Poincaré recurrence of a Hamiltonian flow, or continuous time. `maxRotations` is an engineering stop, not a conserved quantity.
- Isochronous rotation: slots complete when workflows complete, not after a fixed \(\Delta\theta\).

**Loose linkage, stated precisely.** The Ferris-wheel image is a **discrete cylinder map**. That is enough to keep the name; it is not rotational mechanics.

---

## VI. Windowed Mandelbrot → coarsenable self-similar aggregation

**Source \(S\).** Mandelbrot (1982): sets that are statistically self-similar across scales; the Mandelbrot set itself is the connectedness locus of \(z\mapsto z^2+c\). Wilson renormalization and multigrid methods share a *different* idea that the blog actually needs: a hierarchy that can be coarse-grained, with only a finite window materialized.

**Target \(T\).** `theory.Unfold(depth, arity, …)` builds a finite perfect \(k\)-ary tree (the window). Child values partition the parent additively. `theory.Coarsen` drops one level and restores values by summation.

**Morphism \(\varphi\).** Infinite self-similar hierarchy \(\mapsto\) the inductive tree that *would* continue forever. Viewport \(\mapsto\) finite depth \(D\). Zoom-out / renormalization \(\mapsto\) `Coarsen`. Conservation of a scalar under coarse-graining \(\mapsto\) additive partition + sum.

**Transfers \(P_+\).**

- **Self-similarity of windows (Thm. F1).** `Coarsen(Unfold(d,k))` is shape- and value-equal to `Unfold(d-1,k)`. Tests: `TestCoarsenRecoversShallowerWindow`, `TestTheoremF1WindowedAggregation`.
- **Finiteness of the window.** Height is exactly \(D\). Test: `TestWindowIsFinite`.

**Does not transfer \(P_-\).**

- The map \(z\mapsto z^2+c\), critical-point orbits, Hausdorff dimension of a Julia set, or a fractal scheduler in the Kubernetes controller. The production fabric does **not** yet build compute DAGs by Unfold. The correspondence is a *precise pattern*, now executable, not an implemented fractal scheduler (still a roadmap item in vision.md).

**Loose linkage, stated precisely.** The blog's Mandelbrot is the *windowed explorer* of an infinite self-similar object, plus coarse-graining. We keep the historical name as "Windowed Mandelbrot *pattern*" and define it as Thm. F1. We do not claim the Mandelbrot set.

---

## VII. Landauer / Bennett → memoization as recorded reversibility (heuristic)

**Source \(S\).** Landauer (1961): erasing one bit in a computer at temperature \(T\) dissipates at least \(kT\ln 2\). Bennett (1973, 1982): computation can be logically reversible if intermediate results are kept; then the thermodynamic bound need not be paid at each step.

**Target \(T\).** Memo table keyed by \((snapshotID, dominoID, inputHash)\). A hit returns the recorded output without re-executing \(d\).

**Morphism \(\varphi\).** "Keep the intermediate result instead of erasing it and recomputing" \(\mapsto\) memoization. Irreversible re-evaluation \(\mapsto\) a cache miss.

**Transfers \(P_+\) (weak).**

- **Observational equivalence (Thm. M1).** A hit is indistinguishable in outputs and hashes from a recompute, given purity and collision resistance. Test: `TestTheoremM1MemoObservationallyEquivalent`, `TestMemoizationReusesResults`.
- Work (number of executions) is nonincreasing on exact replays.

**Does not transfer \(P_-\).**

- Any bound in joules. Skipping a Julia subprocess does not imply Landauer savings; the machine is not a thermodynamically accounted engine. This correspondence is **heuristic**: it explains *why* the entropy-and-caching slogan was chosen, not a measurable \(kT\ln 2\).

**Status.** Do not put Landauer in the abstract. Keep it as the intellectual ancestor of "minimize recompute by storing results."

---

## VIII. Organism / lifeform → homeostatic control (non-theorem)

**Source \(S\).** Ashby (1956): ultrastable systems that keep essential variables within bounds by feedback. Autopoiesis (Maturana & Varela) is stronger (self-production of components) and is **not** used.

**Target \(T\).** Kubernetes reconciliation: observe, diff against desired spec, act. ComputeContext health checks against node-local stores. Optional player-piano lookahead (`preProvisionNext`) is feed-forward, not metabolism.

**Morphism \(\varphi\).** Essential variables \(\mapsto\) CR status vs spec. Feedback \(\mapsto\) reconciler. Organism boundary \(\mapsto\) a universe's stores and engines.

**Transfers \(P_+\).** Control-loop homeostasis is a fair cybernetic reading of operators. It is not implemented as a theorem in `pkg/theory` because it is Kubernetes' theorem, not KBL's.

**Does not transfer \(P_-\).** Metabolism, reproduction, evolution, autopoiesis, or any biological criterion. **KBL / Kubernetes Based Lifeform remains a name**, with a cybernetic gloss, not a result.

---

## Summary table

| # | Name in the blog | Transferred structure | Status | Primary tests |
|---|------------------|----------------------|--------|----------------|
| I | Newtonian determinism | Unique discrete Cauchy trajectory | Theorem under purity + hash assumptions | D1, D2, snapshot replay |
| II | Low-entropy snapshot | Ensemble Shannon collapse at seal | Theorem for \(H_{\mathrm{ens}}\); *not* for \(S_{\mathrm{thermo}}\) or \(H_{\mathrm{pay}}\) | E1, ensemble entropy |
| III | Locality / isolation | Causal past + sealed-only signaling | Theorem of the engine and replica path | causal past, R1, future-read |
| IV | Multiverse | Product of systems + routing function | Theorem for routing; Everett declined | C1 |
| V | Ferris / Compute Wheel | Discrete cylinder map | Theorem of `pkg/wheel` | W1, cylinder period |
| VI | Windowed Mandelbrot | Coarsenable finite unfolding | Theorem of `pkg/theory`; not yet a cluster scheduler | F1 |
| VII | Minimize entropy / maximize caching | Memo observational equivalence | Theorem for outputs; Landauer heat declined | M1 |
| VIII | Lifeform | Reconciler homeostasis | Metaphor + cybernetic gloss | — |

---

## What a later paper may still claim

If the windowed aggregation pattern is wired into ComputeWheel or hierarchical workflows, Correspondence VI moves from "executable pattern" to "implemented scheduler," and fractal-style aggregation becomes a systems result. If a universe-local energy or cost accountant is added, Correspondence VII can be restated in joules-or-dollars without pretending they are \(kT\ln 2\). Neither is required to keep the present linkages honest.
