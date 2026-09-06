# Related Work

The author's blog series is the design lineage. Scholarly context has two layers: **nature-inspired computing** (the method) and **systems** (the artifact). Physics-shaped names are used the way "neuron" is used in a neural net: a copied structure, not a laboratory identity. Method: [nature-inspired.md](nature-inspired.md).

---

## Nature-inspired computing (the method)

- **McCulloch, W. & Pitts, W.** (1943). *A logical calculus of the ideas immanent in nervous activity.* Bull. Math. Biophys. The original "copy the neuron, leave the wetware" paper.
- **Rosenblatt, F.** (1958). *The perceptron.* Psychological Review. Same method, learning included.
- **Rumelhart, D., Hinton, G. & Williams, R.** (1986). *Learning representations by back-propagating errors.* Nature. The abstraction remains productive without becoming neuroscience.
- **Holland, J.** (1975). *Adaptation in Natural and Artificial Systems.* Genetic algorithms as evolution-inspired search.
- **Kirkpatrick, S., Gelatt, C. & Vecchi, M.** (1983). *Optimization by simulated annealing.* Science. Metallurgy copied as a search heuristic; joules of the metal are left in nature.
- **Dorigo, M.** *Ant Colony Optimization.* Another explicit mimic with the biological remainder named.

KBL's use of entropy, worlds, wheels, and lifeforms belongs in this family, not in Physical Review.

---

## Determinism, purity, replay

- **Wadler, P.** (1992). *The essence of functional programming.* POPL. Referential transparency is the computational content of Correspondence I.
- **Abadi, M. & Lamport, L.** (1991). *The existence of refinement mappings.* Theoretical Computer Science. Uniqueness of observed behaviors under a spec.
- **Michie, D.** (1968). *Memo functions and machine learning.* Nature. Memoization as recording a function graph — Correspondence VII without thermodynamics.
- **Nix / Dolstra, E.** (2006). *The purely functional software deployment model.* Content-addressed derivation graphs; the closest deployed cousin of snapshot + hash + reuse.
- **Merkle, R.** (1987). *A digital signature based on a conventional encryption function.* CRYPTO. Hash-linked authenticity; replay logs are a shallow Merkle spine (snapshot ID + per-domino hashes), not a full Merkle DAG.
- **Zaharia, M. et al.** (2012). *Resilient Distributed Datasets.* NSDI. Lineage-based replay of immutable partitions. KBL dominos are a coarser, operator-defined lineage.

KBL does not claim novelty for "pure functions of immutable inputs." It claims an operator-native packaging of that idea (CRDs, node-local stores, wheels, routed universes).

---

## Isolation, snapshots, locality

- **Berenson, H. et al.** (1995). *A critique of ANSI SQL isolation levels.* SIGMOD. Snapshot isolation as a named isolation level — the database ancestor of ADR 0002, distinct from thermodynamic entropy.
- **Dean, J. & Ghemawat, S.** (2004). *MapReduce.* OSDI. Bring compute to data; ordered map then reduce. KBL generalizes the "bring compute to data" slogan to time-sliced CRDs rather than a single shuffle.
- **Gray, J. & Putzolu, G.** (1987). *The 5 minute rule for trading memory for disc accesses.* SIGMOD. Locality as an economic physics of the memory hierarchy, closer to Correspondence III than relativity is.
- **Kubernetes / Borg** (Burns et al., 2016, *Borg, Omega, and Kubernetes*, ACM Queue). Desired-state reconciliation — the control-theoretic content of Correspondence VIII.

---

## Causality in distributed systems (the honest "relativity" literature)

- **Lamport, L.** (1978). *Time, clocks, and the ordering of events in a distributed system.* CACM. The standard correspondence from Minkowski causal order to happens-before. Correspondence III is an instance, not a new relativity.
- **Fidge, C.** (1988) / **Mattern, F.** (1989). Vector clocks. KBL does not currently ship vector clocks; replica recency is weaker (eventual, sealed).
- **Petri, C. A.** (1962). *Kommunikation mit Automaten.* Causal occurrence nets; another discrete light-cone.
- **Lloyd, W. et al.** (2011). *Don't settle for eventual: scalable causal consistency for wide-area storage with COPS.* SOSP. What sealed-only replication is *not*: we replicate completed snapshots, we do not implement a causal-consistency store for live keys.

---

## Information, entropy, thermodynamic computation

- **Shannon, C. E.** (1948). *A mathematical theory of communication.* Bell System Technical Journal. Ensemble entropy \(H_{\mathrm{ens}}\) in Correspondence II.
- **Cover, T. & Thomas, J.** *Elements of Information Theory.* Definitions and the Dirac-mass limit.
- **Landauer, R.** (1961). *Irreversibility and heat generation in the computing process.* IBM J. Res. Dev. Cited to **bound** Correspondence VII: we do not measure joules.
- **Bennett, C. H.** (1973). *Logical reversibility of computation.* IBM J. Res. Dev.; **(1982)** *The thermodynamics of computation — a review.* Int. J. Theor. Phys. Recording intermediates to avoid erasure — the ancestor of memoization-as-physics-talk.
- **Feynman, R.** *Feynman Lectures on Computation* (1996). Computation as a physical process; useful as a warning not to confuse the metaphor with a laboratory claim.

---

## Self-similarity, coarse-graining, "Mandelbrot"

- **Mandelbrot, B.** (1982). *The Fractal Geometry of Nature.* Source of the name; the set \(z\mapsto z^2+c\) is **declined** as a correspondence.
- **Wilson, K. G.** (1971). *Renormalization group and critical phenomena.* Phys. Rev. B. Coarse-graining as a map from a fine theory to a coarser one — the structure actually used in Theorem F1.
- **Brandt, A.** (1977). Multi-level adaptive solutions to boundary-value problems. *Math. Comp.* Multigrid restriction/prolongation.
- **Okabe, A. et al.** *Spatial Tessellations* / standard quadtree and pyramid encodings in graphics: windowed views of hierarchical data, the "Mandelbrot explorer" UX the blog is pointing at.

---

## Many systems, not many-worlds

- **Everett, H.** (1957). *Relative state formulation of quantum mechanics.* Rev. Mod. Phys. Source of the branching *image*. KBL copies non-interfering classical histories with sealed records, not amplitudes — the neural-net move.
- **Coupled map lattices / multi-physics codes** (Kaneko; engineering co-simulation). Product of systems with different local rules and weak interface coupling — Correspondence IV.

---

## Cybernetics (lifeform, weakly)

- **Ashby, W. R.** (1956). *An Introduction to Cybernetics.* Ultrastability and essential variables. Optional reading of Kubernetes operators; not biology.
- **Wiener, N.** (1948). *Cybernetics.* Feedback. Same caution.

---

## What we are not

KBL is not a contribution to statistical mechanics, general relativity, or fractal geometry. It is a **nature-inspired systems artifact**: it copies named structures from those fields the way a neural net copies the neuron, and it names the remainder left in nature.
