package store

// Backend provides node-local persistence for snapshots, domino outputs, and memo cache.
// GetSnapshot returns bytes as persisted at seal time (direct-bytes staging); callers on the
// hot path should prefer store reads over re-resolving HTTP/path sources.
type Backend interface {
	SaveSnapshot(snapshotID, timeSlice, data string, sealed bool) error
	GetSnapshot(snapshotID string) (timeSlice, data string, sealed bool, err error)
	LookupMemo(snapshotID, dominoID, inputHash string) (outputHash, output string, found bool, err error)
	SaveResult(snapshotID, dominoID, inputHash, outputHash, output string, reused bool, prevLink, link string) error
	GetDominoOutput(snapshotID, dominoID string) (string, error)
	GetLatestResult(snapshotID, dominoID string) (inputHash, outputHash, output string, err error)
	ListReplay(snapshotID string) ([]ReplayEntry, error)
	Close() error
}

// ReplayEntry is a persisted replay-log row (M21).
type ReplayEntry struct {
	SnapshotID string `json:"snapshot_id"`
	DominoID   string `json:"domino_id"`
	InputHash  string `json:"input_hash"`
	OutputHash string `json:"output_hash"`
	Output     string `json:"output"`
	Reused     bool   `json:"reused"`
	PrevLink   string `json:"prev_link,omitempty"`
	Link       string `json:"link,omitempty"`
}

// SpineHead is the last replay link for a snapshot, or snapshotID if the log is empty.
func SpineHead(rows []ReplayEntry, snapshotID string) string {
	if n := len(rows); n > 0 && rows[n-1].Link != "" {
		return rows[n-1].Link
	}
	return snapshotID
}

// LastSpineResult is the last replay row for a named domino on a snapshot.
func LastSpineResult(rows []ReplayEntry, dominoID string) (ReplayEntry, bool) {
	for i := len(rows) - 1; i >= 0; i-- {
		if rows[i].DominoID == dominoID {
			return rows[i], true
		}
	}
	return ReplayEntry{}, false
}

// Type identifies a store backend implementation.
type Type string

const (
	TypeSQLite Type = "sqlite"
	TypeTSDB   Type = "tsdb"
)

// Config describes how to open a store backend.
type Config struct {
	Type     Type
	Path     string // SQLite file path or TSDB data directory (server-side)
	Endpoint string // TSDB HTTP endpoint (client-side), e.g. http://127.0.0.1:9090
}

// OpenBackend opens the configured store backend.
func OpenBackend(cfg Config) (Backend, error) {
	switch cfg.Type {
	case TypeTSDB:
		if cfg.Endpoint == "" {
			return nil, errMissingEndpoint
		}
		return OpenTSDBClient(cfg.Endpoint)
	default:
		if cfg.Path == "" {
			return nil, errMissingPath
		}
		return OpenSQLite(cfg.Path)
	}
}

// Open opens a SQLite store at path (backward-compatible default).
func Open(path string) (Backend, error) {
	return OpenSQLite(path)
}
