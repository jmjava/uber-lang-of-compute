# Theory correspondence tests

Executable counterparts of the theorems in `docs/theory/`.

```bash
cd controller
go test ./pkg/theory/ ./pkg/engine/ ./pkg/wheel/ ./pkg/routing/ ./pkg/replica/ ./pkg/cdc/ ./pkg/hash/ -count=1
```

| Theorem | Package |
|---------|---------|
| D1, D2, M1, future-read | `pkg/engine`, `pkg/theory` |
| E1 ensemble collapse | `pkg/theory` |
| W1 cylinder map | `pkg/wheel`, `pkg/theory` |
| F1 windowed aggregation | `pkg/theory` |
| C1 routing | `pkg/routing`, `pkg/theory` |
| R1 sealed-only replica/CDC | `pkg/replica`, `pkg/cdc`, `pkg/theory` |

See [docs/theory/claims-and-evidence.md](../../docs/theory/claims-and-evidence.md).
