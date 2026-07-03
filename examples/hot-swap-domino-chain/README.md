# Hot-Swapped Domino Chain

In-cluster domino execution with emptyDir handoff between steps (ADR 0004 / 0007).

## Runtimes

| Runtime | Description |
|---------|-------------|
| `kubernetes-init` | Init container daisy chain — works on any Kubernetes cluster |
| `openkruise` | Placeholder pod + OpenKruise `ContainerRecreateRequest` hot-swap |
| `local` | CLI/engine path (default for Workflow without `runtime`) |

## domino-runner

Container entrypoint at `controller/cmd/domino-runner`. Reads:

- `KBL_COMMAND` — e.g. `builtin:identity`, `julia:interpolate`
- `KBL_INPUT` — input JSON path
- `KBL_OUTPUT` — output JSON path

Build:

```bash
cd controller
go build -o bin/domino-runner ./cmd/domino-runner
```

## Deploy init chain (any cluster)

```bash
kubectl apply -f ../../crds/
kubectl apply -f dominochain-init.yaml
./../../controller/bin/kbl-controller --store-root /var/kbl/store
kubectl get dominochains -w
kubectl get pods -l kbl.io/dominochain=simple-init-chain
```

## Deploy OpenKruise chain

Requires [OpenKruise](https://openkruise.io/) installed:

```bash
kubectl apply -f dominochain-openkruise.yaml
```

The controller creates a placeholder pod and issues `ContainerRecreateRequest` objects to hot-swap each slot to `domino-runner`.

## Workflow with container runtime

```bash
kubectl apply -f workflow-container.yaml
```

When `spec.execution.runtime` is set, the Workflow reconciler creates an owned `DominoChain` and waits for completion.

Julia dominos (`julia:*` commands) follow the multi-container model in [ADR 0023](../../docs/adr/0023-julia-deployment-models.md): each step runs `domino-runner` in its own container with Julia installed in the runner image.

## Handoff layout

```
/kbl/input/snapshot.json   ← ConfigMap volume (immutable snapshot)
/kbl/handoff/output.json   ← emptyDir (passed between dominos)
```
