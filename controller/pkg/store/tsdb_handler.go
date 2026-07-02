package store

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
)

// NewTSDBHandler returns an HTTP handler for the node-local TSDB API.
func NewTSDBHandler(engine *TSDBEngine) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	mux.HandleFunc("/v1/stats", func(w http.ResponseWriter, _ *http.Request) {
		snaps, memo, err := engine.Stats()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		writeJSON(w, map[string]int{"snapshots": snaps, "memo_entries": memo})
	})
	mux.HandleFunc("/v1/snapshots", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var req struct {
			SnapshotID string `json:"snapshot_id"`
			TimeSlice  string `json:"time_slice"`
			Data       string `json:"data"`
			Sealed     bool   `json:"sealed"`
		}
		if err := readJSON(r, &req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if err := engine.SaveSnapshot(req.SnapshotID, req.TimeSlice, req.Data, req.Sealed); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusCreated)
	})
	mux.HandleFunc("/v1/snapshots/", func(w http.ResponseWriter, r *http.Request) {
		rest := strings.TrimPrefix(r.URL.Path, "/v1/snapshots/")
		if rest == "" {
			http.NotFound(w, r)
			return
		}
		if strings.HasSuffix(rest, "/data") && r.Method == http.MethodGet {
			id := strings.TrimSuffix(strings.TrimSuffix(rest, "/data"), "/")
			if id == "" {
				http.NotFound(w, r)
				return
			}
			data, sealed, err := engine.GetSnapshotData(id)
			if err != nil {
				http.NotFound(w, r)
				return
			}
			if !sealed {
				http.Error(w, "snapshot not sealed", http.StatusConflict)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(data))
			return
		}
		if r.Method != http.MethodGet {
			http.NotFound(w, r)
			return
		}
		id := rest
		ts, data, sealed, err := engine.GetSnapshot(id)
		if err != nil {
			http.NotFound(w, r)
			return
		}
		writeJSON(w, snapshotRecord{SnapshotID: id, TimeSlice: ts, Data: data, Sealed: sealed})
	})
	mux.HandleFunc("/v1/memo", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		q := r.URL.Query()
		outHash, out, found, err := engine.LookupMemo(q.Get("snapshot_id"), q.Get("domino_id"), q.Get("input_hash"))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if !found {
			http.NotFound(w, r)
			return
		}
		writeJSON(w, memoRecord{
			SnapshotID: q.Get("snapshot_id"),
			DominoID:   q.Get("domino_id"),
			InputHash:  q.Get("input_hash"),
			OutputHash: outHash,
			Output:     out,
		})
	})
	mux.HandleFunc("/v1/results", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var req struct {
			SnapshotID string `json:"snapshot_id"`
			DominoID   string `json:"domino_id"`
			InputHash  string `json:"input_hash"`
			OutputHash string `json:"output_hash"`
			Output     string `json:"output"`
			Reused     bool   `json:"reused"`
		}
		if err := readJSON(r, &req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if err := engine.SaveResult(req.SnapshotID, req.DominoID, req.InputHash, req.OutputHash, req.Output, req.Reused); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusCreated)
	})
	mux.HandleFunc("/v1/outputs/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/v1/outputs/"), "/")
		if len(parts) != 2 {
			http.NotFound(w, r)
			return
		}
		out, err := engine.GetDominoOutput(parts[0], parts[1])
		if err != nil {
			http.NotFound(w, r)
			return
		}
		writeJSON(w, map[string]string{"output": out})
	})
	mux.HandleFunc("/v1/latest-results/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/v1/latest-results/"), "/")
		if len(parts) != 2 {
			http.NotFound(w, r)
			return
		}
		inHash, outHash, out, err := engine.GetLatestResult(parts[0], parts[1])
		if err != nil {
			http.NotFound(w, r)
			return
		}
		writeJSON(w, map[string]string{
			"input_hash":  inHash,
			"output_hash": outHash,
			"output":      out,
		})
	})
	return mux
}

func readJSON(r *http.Request, v interface{}) error {
	defer r.Body.Close()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return err
	}
	return json.Unmarshal(body, v)
}

func writeJSON(w http.ResponseWriter, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
}
