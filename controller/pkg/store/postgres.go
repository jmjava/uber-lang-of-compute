package store

import (
	"database/sql"
	"fmt"
	"strings"

	_ "github.com/lib/pq"
)

// PostgresBackend implements Backend on PostgreSQL so Debezium can tail WAL.
type PostgresBackend struct {
	db *sql.DB
}

// OpenPostgres opens a PostgreSQL store. dsn is a postgres:// URL or lib/pq keyword DSN.
func OpenPostgres(dsn string) (*PostgresBackend, error) {
	if strings.TrimSpace(dsn) == "" {
		return nil, fmt.Errorf("postgres dsn required")
	}
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("open postgres: %w", err)
	}
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("ping postgres: %w", err)
	}
	s := &PostgresBackend{db: db}
	if err := s.migrate(); err != nil {
		db.Close()
		return nil, err
	}
	return s, nil
}

func (s *PostgresBackend) migrate() error {
	schema := `
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
	`
	_, err := s.db.Exec(schema)
	return err
}

func (s *PostgresBackend) SaveSnapshot(snapshotID, timeSlice, data string, sealed bool) error {
	existingTime, existingData, existingSealed, getErr := s.GetSnapshot(snapshotID)
	if err := refuseSealedMutation(getErr, existingTime, existingData, existingSealed, timeSlice, data, sealed); err != nil {
		return err
	}
	if getErr == nil && existingSealed {
		return nil
	}
	_, err := s.db.Exec(
		`INSERT INTO snapshots (snapshot_id, time_slice, data, sealed)
		 VALUES ($1, $2, $3, $4)
		 ON CONFLICT (snapshot_id) DO UPDATE SET
		   time_slice = EXCLUDED.time_slice,
		   data = EXCLUDED.data,
		   sealed = EXCLUDED.sealed`,
		snapshotID, timeSlice, data, sealed,
	)
	return err
}

func (s *PostgresBackend) GetSnapshot(snapshotID string) (timeSlice, data string, sealed bool, err error) {
	row := s.db.QueryRow(
		`SELECT time_slice, data, sealed FROM snapshots WHERE snapshot_id = $1`, snapshotID,
	)
	err = row.Scan(&timeSlice, &data, &sealed)
	return
}

func (s *PostgresBackend) GetSnapshotData(snapshotID string) (string, bool, error) {
	_, data, sealed, err := s.GetSnapshot(snapshotID)
	return data, sealed, err
}

func (s *PostgresBackend) LookupMemo(snapshotID, dominoID, inputHash string) (outputHash, output string, found bool, err error) {
	row := s.db.QueryRow(
		`SELECT output_hash, output FROM domino_results
		 WHERE snapshot_id = $1 AND domino_id = $2 AND input_hash = $3`,
		snapshotID, dominoID, inputHash,
	)
	err = row.Scan(&outputHash, &output)
	if err == sql.ErrNoRows {
		return "", "", false, nil
	}
	if err != nil {
		return "", "", false, err
	}
	return outputHash, output, true, nil
}

func (s *PostgresBackend) SaveResult(snapshotID, dominoID, inputHash, outputHash, output string, reused bool, prevLink, link string) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if !reused {
		storedHash, storedOut, found, err := s.LookupMemo(snapshotID, dominoID, inputHash)
		if err != nil {
			return err
		}
		if err := refuseMemoConflict(found, storedHash, storedOut, outputHash, output); err != nil {
			return err
		}
		if !found {
			_, err = tx.Exec(
				`INSERT INTO domino_results
				 (snapshot_id, domino_id, input_hash, output_hash, output, reused)
				 VALUES ($1, $2, $3, $4, $5, FALSE)`,
				snapshotID, dominoID, inputHash, outputHash, output,
			)
			if err != nil {
				return err
			}
		}
	}

	_, err = tx.Exec(
		`INSERT INTO replay_log (snapshot_id, domino_id, input_hash, output_hash, reused, output, prev_link, link)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		snapshotID, dominoID, inputHash, outputHash, reused, output, prevLink, link,
	)
	if err != nil {
		return err
	}
	return tx.Commit()
}

func (s *PostgresBackend) GetDominoOutput(snapshotID, dominoID string) (string, error) {
	_, _, output, err := s.GetLatestResult(snapshotID, dominoID)
	return output, err
}

func (s *PostgresBackend) GetLatestResult(snapshotID, dominoID string) (inputHash, outputHash, output string, err error) {
	row := s.db.QueryRow(
		`SELECT input_hash, output_hash, output FROM replay_log
		 WHERE snapshot_id = $1 AND domino_id = $2
		 ORDER BY id DESC LIMIT 1`,
		snapshotID, dominoID,
	)
	err = row.Scan(&inputHash, &outputHash, &output)
	if err == sql.ErrNoRows {
		row = s.db.QueryRow(
			`SELECT input_hash, output_hash, output FROM domino_results
			 WHERE snapshot_id = $1 AND domino_id = $2
			 ORDER BY id DESC LIMIT 1`,
			snapshotID, dominoID,
		)
		err = row.Scan(&inputHash, &outputHash, &output)
	}
	return
}

func (s *PostgresBackend) ListReplay(snapshotID string) ([]ReplayEntry, error) {
	rows, err := s.db.Query(
		`SELECT snapshot_id, domino_id, input_hash, output_hash, COALESCE(output,''), reused,
		        COALESCE(prev_link,''), COALESCE(link,'')
		 FROM replay_log WHERE snapshot_id = $1 ORDER BY id ASC`,
		snapshotID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []ReplayEntry
	for rows.Next() {
		var e ReplayEntry
		if err := rows.Scan(&e.SnapshotID, &e.DominoID, &e.InputHash, &e.OutputHash, &e.Output, &e.Reused, &e.PrevLink, &e.Link); err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	return out, rows.Err()
}

func (s *PostgresBackend) Close() error {
	return s.db.Close()
}

// IsPostgresDSN reports whether path is a PostgreSQL connection string.
func IsPostgresDSN(path string) bool {
	return strings.HasPrefix(path, "postgres://") || strings.HasPrefix(path, "postgresql://")
}
