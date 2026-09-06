# Nature-inspired computing (not a physics proof)

KBL borrows structures from mechanics, information theory, and biology the same way a neural net borrows the neuron: **as a usable abstraction of nature, not as a claim to be nature.**

McCulloch & Pitts (1943) and Rosenblatt (1958) did not prove a theory of cortex. They copied a *shape* — weighted combination, threshold, layered composition — and discarded glia, metabolism, spikes-as-biology, and embodiment. The copy was productive because it was explicit about what was kept.

This repository does the same with snapshots, dominos, wheels, and universes. The academic job is to say **what was copied, what was left on the cutting-room floor, and whether the copy still hangs together as engineering.** That is the method. We are not trying to derive Kubernetes from \(F=ma\).

| Nature (source) | What ANNs copied | What KBL copies |
|-----------------|------------------|-----------------|
| Neuron | weighted sum + nonlinearity | Domino: a function of declared inputs |
| Synaptic memory | stored weights | Memo table: stored outputs keyed by input hash |
| Tissue locality | layers, no all-to-all brain | Node-local store; compute moves to data |
| Distinct brains / environments | ensemble of nets | Pluggable universes with different \(\Phi_u\) |
| Visual zoom / fractal explorer | (CNN receptive fields, image pyramids) | Windowed unfolding: only depth \(D\) exists |
| Homeostasis | residual / target loops in control nets | Reconciler: spec versus status |
| Thermodynamic frugality | (implicit: don't recompute activations) | Bennett-style recording: don't re-evaluate a known function |

None of those rows is a laboratory identity. Each is a **mimic of a structure that nature already found useful.**

## How to read a leftover analogy

Items that are not theorems of the engine are not failures of the method. They are the glial cells: present in nature, not implemented in the abstraction, still allowed to *guide* the design if the guidance is named.

Grade the mimic, do not hide it:

| Grade | Meaning | Example |
|-------|---------|---------|
| **Theorem** | The copied structure is an invariant of the code | Sealed snapshot ⇒ unique builtin worldline |
| **Empirical discretization** | Unique under a pinned runtime, like a fixed-step integrator | Pinned Julia Manifest |
| **Contract** | Assumed regularity, like Picard needing Lipschitz | `deterministic: true` on a container image |
| **Bookkeeping mimic** | Same *accounting shape* as the natural process, different units | Logical work vs \(kT\ln 2\) |
| **Explorer mimic** | Same viewport move as the natural object | Finite Mandelbrot iteration window; coarsenable tree |
| **Branching mimic** | Independent histories, classical records, no amplitudes | Multiverse routing fan-out |
| **Control mimic** | Feedback that holds an essential variable | Reconciler homeostasis |

The remainder of [correspondences.md](correspondences.md) uses these grades instead of “declined.”

## Residual mimics (the former non-theorems)

### Purity of containers — Picard regularity, copied

Newtonian uniqueness is *conditional*: Picard–Lindelöf needs a Lipschitz field. KBL uniqueness is conditional on a **regularity class** (`CommandRegularity`):

- `builtin:*` → theorem grade
- `julia:*` → pinned discretization (the map is the floating-point program, not \(\mathbb{R}\))
- anything else → contract

We copy nature's *conditional well-posedness*, not a proof that every container is a function. An isolated sandbox is the Faraday-cage analogue: it is an admission condition, not Lipschitz. `sandbox:impure` is allowed inside the cage and is still not unique (M10). Tests: `TestCommandRegularityLadder`, `TestIsolatedSandboxImpureIsNotUnique`.

### Julia bit-identity — shadowing, copied

A pinned Manifest is a **fixed discretization**. Shadowing lemmas in numerical dynamics say a pseudo-orbit of a discrete map stays close to a true orbit *of that map*. We copy that: uniqueness is claimed for the discretized Julia program, not for continuum greeks. `TestJuliaFinanceModelsWorkflowDeterministic` is evidence at that grade.

### Windowed Mandelbrot — explorer, copied

The blog's Mandelbrot is the *explorer UX*: an infinite generated object of which only a window is drawn. We copy

1. finite unfolding (`Unfold`) and zoom-out (`Coarsen`) — theorem F1
2. similarity dimension of the IFS window (`SimilarityDimension`); additive trees have \(D=1\), matching a conserved scalar
3. leaf-count as wheel seats (`LeafCount`) — the current window's leaves can be the Ferris-wheel circuit
4. escape-time as the explorer's iteration budget (`EscapeTime`) — a finite window on \(z\mapsto z^2+c\), **not** a claim that the scheduler is that map

The polynomial remains the *illustration of a generated infinite object*, the way a sigmoid is an illustration of a firing curve.

### Landauer / Bennett — recorded intermediates, copied

Nature (and Bennett) says: keep the intermediate, do not erase and redo. Memoization copies that bookkeeping. `LogicalWork` counts irreversible evaluations; a replay that hits memo has fewer of them (`ReplaySaves`). Units are steps, not joules — as ANN "energy" is usually a loss, not ATP.

### Everett — independent histories, copied

What is useful in many-worlds *as a design image* is **non-interfering branches with classical records**. Routing fan-out (`Branch`) copies a sealed snapshot into other universes; `Interfere(..., liveShare=false)` is always false. We do not copy Hilbert space. We copy the ensemble-of-worlds idea the way dropout copies an ensemble of thinned nets.

### Lifeform — homeostasis, copied

Ashby's essential variables: keep spec and status aligned. `ReconcileError` / `Homeostatic` copy the loop. Metabolism and DNA stay in nature.

## What this changes about peer review

A *physics* referee should still not receive this as a claim about nature.

A *nature-inspired computing* referee (the right audience, analogous to neural-net papers) should ask: is the borrowed structure named, is the discarded remainder named, and does the copy still work? That is what [correspondences.md](correspondences.md) and `controller/pkg/theory` are for.
