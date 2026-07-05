# Home lab setup (i9 + discrete GPU workstation)

Run the full KBL Kind lab on your **home network machine** (i9, RTX 4080, 32–64 GiB RAM). Use a **compact profile** on an i7 laptop for lighter local dev.

## Quick start on the home machine

```bash
git clone https://github.com/jmjava/uber-lang-of-compute.git
cd uber-lang-of-compute

# Give Docker Desktop / Docker Engine at least 24 GiB RAM and 8+ CPUs
export KBL_LAB_PROFILE=home
chmod +x lab/scripts/*.sh
make lab-up
```

Verify Volcano batch scheduling:

```bash
./lab/scripts/verify-volcano.sh
```

## Profiles

| Profile | Host | Kind nodes | Volcano demo | OpenKruise default |
|---------|------|------------|--------------|-------------------|
| **`home`** (recommended for i9) | 32–64 GiB RAM, 8+ cores | 1 CP + 2 workers | 2-context wheel + parallel burst | on |
| **`compact`** (i7 laptop) | 16–32 GiB RAM | 1 CP + 1 worker | single-context wheel | off |
| **`default`** | same as legacy | 1 CP + 2 workers | 2-context wheel + burst | on |

```bash
# i7 laptop — lighter footprint
KBL_LAB_PROFILE=compact make lab-up

# home i9 — full Volcano Ferris wheel + burst
KBL_LAB_PROFILE=home make lab-up
```

Recreate the cluster when switching profiles:

```bash
./lab/scripts/down.sh
KBL_LAB_PROFILE=home ./lab/scripts/up.sh
```

## What Volcano demo shows

1. **ComputeWheel Ferris wheel** (`julia-finance-wheel`) — rotates **compute-a → compute-b**, creating sequential **Workflow → DominoChain → VCJob** pipelines through queue **`kbl-lab`**. Player-piano `preProvisionNext` pre-stages the next slot.
2. **Parallel burst** — two `DominoChain` resources (`volcano-burst-a/b`) submit VCJobs to the same queue, pinned to different workers via `kubernetes.io/hostname`.
3. **`verify-volcano.sh`** — prints queue depth, VCJob state, `schedulerName: volcano`, and which node each pod landed on.

Pipeline diagram: [docs/diagrams.md §8](../docs/diagrams.md#8-volcano-batch-path-lab-demo).

## Remote kubectl from your i7

On the **home i9** machine, copy kubeconfig after `make lab-up`:

```bash
kind get kubeconfig --name kbl-lab > ~/.kube/kbl-lab-config
# On i7 laptop:
export KUBECONFIG=~/.kube/kbl-lab-config   # or merge into ~/.kube/config
kubectl get nodes
./lab/scripts/verify-volcano.sh
```

Replace `kind get kubeconfig` hostnames if the home machine is reachable on LAN (e.g. `server-linux:6443` instead of `127.0.0.1:6443`).

## GPU label (future dominos)

The **home** Kind config labels worker 1 with `kbl.io/gpu=present` for future GPU-backed domino experiments. Volcano and Julia finance demos do **not** require a GPU today — the label is reserved for later phases.

When NVIDIA Container Toolkit is installed on the host:

```bash
kubectl get nodes -L kbl.io/gpu
# kbl-lab-worker   ...   present
```

## Docker resource hints

| Host | Docker RAM | Docker CPUs |
|------|------------|-------------|
| i7 compact | 12–16 GiB | 4–6 |
| i9 home | 24–48 GiB | 8–16 |

## Skip components

```bash
KBL_LAB_VOLCANO=0 make lab-up          # platform only
KBL_LAB_OPENKURISE=0 make lab-up       # skip hot-swap demo
KBL_LAB_PROFILE=compact KBL_LAB_OPENKURISE=0 make lab-up   # minimal i7
```

See also [lab/README.md](README.md).
