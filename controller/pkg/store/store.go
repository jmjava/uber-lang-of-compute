package store

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
)

// Store provides node-local persistence for snapshots, domino outputs, and memo cache.
type Store struct {
	db *sql.DB
}

// Open creates or opens a SQLite store at the given path.
func Open(path string) (*Store, error) {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("create store directory: %w", err)
	}

	db, err := sql.Open("sqlite3", path+"?_journal_mode=WAL&_foreign_keys=on")
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}

	s := &Store{db: db}
	if err := s.migrate(); err != nil {
		db.Close()
		return nil, err
	}
	return s, nil
}

func (s *Store) migrate() error {
	schema := `
	CREATE TABLE IF NOT EXISTS snapshots (
		snapshot_id TEXT PRIMARY KEY,
		time_slice  TEXT NOT NULL,
		data        TEXT NOT NULL,
		sealed      INTEGER NOT NULL DEFAULT 0,
		created_at  TEXT NOT NULL DEFAULT (datetime('now'))
	);

	CREATE TABLE IF NOT EXISTS domino_results (
		id          INTEGER PRIMARY KEY AUTOINCREMENT,
		snapshot_id TEXT NOT NULL,
		domino_id   TEXT NOT NULL,
		input_hash  TEXT NOT NULL,
		output_hash TEXT NOT NULL,
		output      TEXT NOT NULL,
		reused      INTEGER NOT NULL DEFAULT 0,
		created_at  TEXT NOT NULL DEFAULT (datetime('now')),
		UNIQUE(snapshot_id, domino_id, input_hash)
	);

	CREATE TABLE IF NOT EXISTS replay_log (
		id          INTEGER PRIMARY KEY AUTOINCREMENT,
		snapshot_id TEXT NOT NULL,
		domino_id   TEXT NOT NULL,
		input_hash  TEXT NOT NULL,
		output_hash TEXT NOT NULL,
		reused      INTEGER NOT NULL,
		output      TEXT,
		created_at  TEXT NOT NULL DEFAULT (datetime('now'))
	);

	CREATE INDEX IF NOT EXISTS idx_memo ON domino_results(snapshot_id, domino_id, input_hash);
	`
	_, err := s.db.Exec(schema)
	return err
}

// SaveSnapshot persists snapshot data.
func (s *Store) SaveSnapshot(snapshotID, timeSlice, data string, sealed bool) error {
	sealedInt := 0
	if sealed {
		sealedInt = 1
	}
	_, err := s.db.Exec(
		`INSERT OR REPLACE INTO snapshots (snapshot_id, time_slice, data, sealed) VALUES (?, ?, ?, ?)`,
		snapshotID, timeSlice, data, sealedInt,
	)
	return err
}

// GetSnapshot retrieves snapshot data by ID.
func (s *Store) GetSnapshot(snapshotID string) (timeSlice, data string, sealed bool, err error) {
	row := s.db.QueryRow(
		`SELECT time_slice, data, sealed FROM snapshots WHERE snapshot_id = ?`, snapshotID,
	)
	var sealedInt int
	err = row.Scan(&timeSlice, &data, &sealedInt)
	sealed = sealedInt == 1
	return
}

// LookupMemo checks if a domino result exists for the given key.
func (s *Store) LookupMemo(snapshotID, dominoID, inputHash string) (outputHash, output string, found bool, err error) {
	row := s.db.QueryRow(
		`SELECT output_hash, output FROM domino_results
		 WHERE snapshot_id = ? AND domino_id = ? AND input_hash = ?`,
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

// SaveResult stores a domino result and appends to replay log.
func (s *Store) SaveResult(snapshotID, dominoID, inputHash, outputHash, output string, reused bool) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if !reused {
		_, err = tx.Exec(
			`INSERT OR IGNORE INTO domino_results
			 (snapshot_id, domino_id, input_hash, output_hash, output, reused)
			 VALUES (?, ?, ?, ?, ?, 0)`,
			snapshotID, dominoID, inputHash, outputHash, output,
		)
		if err != nil {
			return err
		}
	}

	reusedInt := 0
	if reused {
		reusedInt = 1
	}
	_, err = tx.Exec(
		`INSERT INTO replay_log (snapshot_id, domino_id, input_hash, output_hash, reused, output)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		snapshotID, dominoID, inputHash, outputHash, reusedInt, output,
	)
	if err != nil {
		return err
	}
	return tx.Commit()
}

// GetDominoOutput retrieves the latest output for a domino in a snapshot chain run.
func (s *Store) GetDominoOutput(snapshotID, dominoID string) (string, error) {
	row := s.db.QueryRow(
		`SELECT output FROM domino_results
		 WHERE snapshot_id = ? AND domino_id = ?
		 ORDER BY created_at DESC LIMIT 1`,
		snapshotID, dominoID,
	)
	var output string
	err := row.Scan(&output)
	return output, err
}

// Close closes the database connection.
func (s *Store) Close() error {
	return s.db.Close()
}
