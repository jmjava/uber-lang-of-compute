# Research test environment

Spin up the **engine test assets** without Kind, Volcano, or a 64 GiB workstation. Future research (humans, CI, Cursor Cloud Agents) should start here.

## What gets started

| Asset | Role |
|-------|------|
| Go 1.23 + gcc | `mattn/go-sqlite3` (CGO) and `make theory-prove` |
| `kbl-compute` / `kbl-review` | Workshop receipt: seal → chain → memo → lookahead → HeadLink fan-out |
| `kbl-tsdb` on `:9090` | Live node-local store for TSDB-backed replay |

This is not the Kind lab. Kind remains `make lab-up` ([lab/README.md](../lab/README.md)).

## One command

```bash
make research-up      # install binaries + start kbl-tsdb
make research-test    # make review + finance chain against live TSDB
make research-down    # stop kbl-tsdb
```

`make research-test` must exit 0. It checks:

1. `make review` JSON receipt (`first_evaluations: 3`, `second_reuses: 3`)
2. `kbl-compute --store http://127.0.0.1:9090` twice; second replay log has `"reused": true` on every domino

## Cursor Cloud Agents

`.cursor/environment.json` builds a Go 1.23 image, runs `lab/scripts/research-install.sh` on each environment build, and starts `kbl-tsdb` on agent boot. New agents inherit that without dashboard clicks.

## Docker Compose (optional)

If Docker is available and you do not want the native binary:

```bash
docker compose -f lab/compose/research.yaml up --build
make review
./controller/bin/kbl-compute \
  --workflow examples/finance-curve-snapshot/workflow.yaml \
  --store http://127.0.0.1:9090 \
  --replay-log /tmp/kbl-research/replay.json
```

## Explicit `--store`

`kbl-compute --store` now wins over `spec.provisioning.storePath` in the YAML. That is how the smoke test points the finance example at live TSDB instead of `/tmp/kbl-finance/store.db`.
