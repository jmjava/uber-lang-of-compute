package store

// Backend provides node-local persistence for snapshots, domino outputs, and memo cache.
// GetSnapshot returns bytes as persisted at seal time (direct-bytes staging); callers on the
// hot path should prefer store reads over re-resolving HTTP/path sources.
type Backend interface {
	SaveSnapshot(snapshotID, timeSlice, data string, sealed bool) error
	GetSnapshot(snapshotID string) (timeSlice, data string, sealed bool, err error)
	LookupMemo(snapshotID, dominoID, inputHash string) (outputHash, output string, found bool, err error)
	SaveResult(snapshotID, dominoID, inputHash, outputHash, output string, reused bool) error
	GetDominoOutput(snapshotID, dominoID string) (string, error)
	GetLatestResult(snapshotID, dominoID string) (inputHash, outputHash, output string, err error)
	Close() error
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
