package store

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// TSDBEngine is a file-backed time-series store for node-local snapshot/memo/replay data.
type TSDBEngine struct {
	root string
	mu   sync.RWMutex
}

// OpenTSDBEngine opens or creates a TSDB data directory.
func OpenTSDBEngine(root string) (*TSDBEngine, error) {
	dirs := []string{
		filepath.Join(root, "snapshots"),
		filepath.Join(root, "memo"),
		filepath.Join(root, "replay"),
	}
	for _, d := range dirs {
		if err := os.MkdirAll(d, 0o755); err != nil {
			return nil, fmt.Errorf("create tsdb dir %s: %w", d, err)
		}
	}
	return &TSDBEngine{root: root}, nil
}

func (t *TSDBEngine) SaveSnapshot(snapshotID, timeSlice, data string, sealed bool) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	rec := snapshotRecord{SnapshotID: snapshotID, TimeSlice: timeSlice, Data: data, Sealed: sealed, CreatedAt: time.Now().UTC()}
	body, err := json.Marshal(rec)
	if err != nil {
		return err
	}
	return os.WriteFile(t.snapshotPath(snapshotID), body, 0o644)
}

func (t *TSDBEngine) GetSnapshot(snapshotID string) (timeSlice, data string, sealed bool, err error) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	body, err := os.ReadFile(t.snapshotPath(snapshotID))
	if err != nil {
		return "", "", false, err
	}
	var rec snapshotRecord
	if err := json.Unmarshal(body, &rec); err != nil {
		return "", "", false, err
	}
	return rec.TimeSlice, rec.Data, rec.Sealed, nil
}

func (t *TSDBEngine) LookupMemo(snapshotID, dominoID, inputHash string) (outputHash, output string, found bool, err error) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	body, err := os.ReadFile(t.memoPath(snapshotID, dominoID, inputHash))
	if os.IsNotExist(err) {
		return "", "", false, nil
	}
	if err != nil {
		return "", "", false, err
	}
	var rec memoRecord
	if err := json.Unmarshal(body, &rec); err != nil {
		return "", "", false, err
	}
	return rec.OutputHash, rec.Output, true, nil
}

func (t *TSDBEngine) SaveResult(snapshotID, dominoID, inputHash, outputHash, output string, reused bool) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if !reused {
		rec := memoRecord{
			SnapshotID: snapshotID,
			DominoID:   dominoID,
			InputHash:  inputHash,
			OutputHash: outputHash,
			Output:     output,
			CreatedAt:  time.Now().UTC(),
		}
		body, err := json.Marshal(rec)
		if err != nil {
			return err
		}
		if err := os.MkdirAll(filepath.Dir(t.memoPath(snapshotID, dominoID, inputHash)), 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(t.memoPath(snapshotID, dominoID, inputHash), body, 0o644); err != nil {
			return err
		}
	}

	replay := replayRecord{
		SnapshotID: snapshotID,
		DominoID:   dominoID,
		InputHash:  inputHash,
		OutputHash: outputHash,
		Reused:     reused,
		Output:     output,
		CreatedAt:  time.Now().UTC(),
	}
	body, err := json.Marshal(replay)
	if err != nil {
		return err
	}
	name := fmt.Sprintf("%d-%s-%s.json", time.Now().UnixNano(), dominoID, inputHash[:8])
	return os.WriteFile(filepath.Join(t.root, "replay", name), body, 0o644)
}

func (t *TSDBEngine) GetDominoOutput(snapshotID, dominoID string) (string, error) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	dir := filepath.Join(t.root, "memo", snapshotID, dominoID)
	entries, err := os.ReadDir(dir)
	if err != nil {
		return "", err
	}
	if len(entries) == 0 {
		return "", os.ErrNotExist
	}
	latest := entries[len(entries)-1]
	body, err := os.ReadFile(filepath.Join(dir, latest.Name()))
	if err != nil {
		return "", err
	}
	var rec memoRecord
	if err := json.Unmarshal(body, &rec); err != nil {
		return "", err
	}
	return rec.Output, nil
}

func (t *TSDBEngine) Close() error { return nil }

func (t *TSDBEngine) Stats() (snapshots, memoEntries int, err error) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	snapEntries, err := os.ReadDir(filepath.Join(t.root, "snapshots"))
	if err != nil {
		return 0, 0, err
	}
	count := 0
	_ = filepath.Walk(filepath.Join(t.root, "memo"), func(path string, info os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if !info.IsDir() {
			count++
		}
		return nil
	})
	return len(snapEntries), count, nil
}

func (t *TSDBEngine) snapshotPath(id string) string {
	return filepath.Join(t.root, "snapshots", id+".json")
}

func (t *TSDBEngine) memoPath(snapshotID, dominoID, inputHash string) string {
	return filepath.Join(t.root, "memo", snapshotID, dominoID, inputHash+".json")
}

type snapshotRecord struct {
	SnapshotID string    `json:"snapshot_id"`
	TimeSlice  string    `json:"time_slice"`
	Data       string    `json:"data"`
	Sealed     bool      `json:"sealed"`
	CreatedAt  time.Time `json:"created_at"`
}

type memoRecord struct {
	SnapshotID string    `json:"snapshot_id"`
	DominoID   string    `json:"domino_id"`
	InputHash  string    `json:"input_hash"`
	OutputHash string    `json:"output_hash"`
	Output     string    `json:"output"`
	CreatedAt  time.Time `json:"created_at"`
}

type replayRecord struct {
	SnapshotID string    `json:"snapshot_id"`
	DominoID   string    `json:"domino_id"`
	InputHash  string    `json:"input_hash"`
	OutputHash string    `json:"output_hash"`
	Reused     bool      `json:"reused"`
	Output     string    `json:"output,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
}
