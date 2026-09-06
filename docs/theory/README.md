# Theory: Physics Correspondences for the KBL Compute Fabric

This directory is the academic layer of the repository. It treats the Uber Language of Compute not as a physical theory and not as branding, but as a **family of structure-preserving correspondences**: named structures from mechanics, information theory, and relativistic causality, each mapped onto a discrete computational counterpart that can be stated, bounded, and tested.

| Document | Role |
|----------|------|
| [peer-review.md](peer-review.md) | Simulated journal referee report: what fails, what survives, what was patched |
| [correspondences.md](correspondences.md) | The mapping paper: source domain, target domain, transferred properties, limits |
| [formal-model.md](formal-model.md) | Discrete dynamical system, theorems, proofs, assumptions |
| [related-work.md](related-work.md) | Literature that actually bears on the claims |
| [claims-and-evidence.md](claims-and-evidence.md) | Falsifiable claims → tests / remaining gaps |

Implementation: `controller/pkg/theory/` (entropy, causal past, windowed aggregation) plus engine, wheel, routing, and replica tests named after the theorems.

**How to read the physics words.** Every physical term in this project is a *name of a correspondence*, not an identity claim. "Entropy" means ensemble Shannon entropy of candidate inputs, not \(k\log W\) of a thermodynamic macrostate. "Multiverse" means a product of independent dynamical systems with a routing morphism, not Everett branching. The correspondence paper states both what transfers and what does not.

ADR: [0037 Physics Correspondence Layer](../adr/0037-physics-correspondence-layer.md).
