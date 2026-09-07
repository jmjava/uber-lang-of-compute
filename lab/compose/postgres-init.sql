-- Schema Debezium tails. Matches controller/pkg/store/postgres.go migrate().
CREATE TABLE IF NOT EXISTS snapshots (
    snapshot_id TEXT PRIMARY KEY,
    time_slice  TEXT NOT NULL,
    data        TEXT NOT NULL,
    sealed      BOOLEAN NOT NULL DEFAULT FALSE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS domino_results (
    id          BIGSERIAL PRIMARY KEY,
    snapshot_id TEXT NOT NULL,
    domino_id   TEXT NOT NULL,
    input_hash  TEXT NOT NULL,
    output_hash TEXT NOT NULL,
    output      TEXT NOT NULL,
    reused      BOOLEAN NOT NULL DEFAULT FALSE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE(snapshot_id, domino_id, input_hash)
);

CREATE TABLE IF NOT EXISTS replay_log (
    id          BIGSERIAL PRIMARY KEY,
    snapshot_id TEXT NOT NULL,
    domino_id   TEXT NOT NULL,
    input_hash  TEXT NOT NULL,
    output_hash TEXT NOT NULL,
    reused      BOOLEAN NOT NULL,
    output      TEXT,
    prev_link   TEXT,
    link        TEXT,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_pg_memo ON domino_results(snapshot_id, domino_id, input_hash);
