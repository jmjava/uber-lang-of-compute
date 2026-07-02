package store

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// TSDBClient implements Backend over the kbl-tsdb HTTP API.
type TSDBClient struct {
	base   string
	client *http.Client
}

// OpenTSDBClient connects to a node-local TSDB HTTP endpoint.
func OpenTSDBClient(endpoint string) (*TSDBClient, error) {
	endpoint = strings.TrimRight(endpoint, "/")
	c := &TSDBClient{
		base: endpoint,
		client: &http.Client{Timeout: 10 * time.Second},
	}
	if err := c.ping(); err != nil {
		return nil, fmt.Errorf("tsdb health check: %w", err)
	}
	return c, nil
}

func (c *TSDBClient) ping() error {
	resp, err := c.client.Get(c.base + "/healthz")
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status %d", resp.StatusCode)
	}
	return nil
}

func (c *TSDBClient) SaveSnapshot(snapshotID, timeSlice, data string, sealed bool) error {
	body, _ := json.Marshal(map[string]interface{}{
		"snapshot_id": snapshotID,
		"time_slice":  timeSlice,
		"data":        data,
		"sealed":      sealed,
	})
	return c.post("/v1/snapshots", body)
}

func (c *TSDBClient) GetSnapshot(snapshotID string) (timeSlice, data string, sealed bool, err error) {
	resp, err := c.client.Get(c.base + "/v1/snapshots/" + url.PathEscape(snapshotID))
	if err != nil {
		return "", "", false, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		return "", "", false, fmt.Errorf("snapshot not found")
	}
	if resp.StatusCode != http.StatusOK {
		return "", "", false, c.readError(resp)
	}
	var rec snapshotRecord
	if err := json.NewDecoder(resp.Body).Decode(&rec); err != nil {
		return "", "", false, err
	}
	return rec.TimeSlice, rec.Data, rec.Sealed, nil
}

func (c *TSDBClient) LookupMemo(snapshotID, dominoID, inputHash string) (outputHash, output string, found bool, err error) {
	q := url.Values{
		"snapshot_id": {snapshotID},
		"domino_id":   {dominoID},
		"input_hash":  {inputHash},
	}
	resp, err := c.client.Get(c.base + "/v1/memo?" + q.Encode())
	if err != nil {
		return "", "", false, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		return "", "", false, nil
	}
	if resp.StatusCode != http.StatusOK {
		return "", "", false, c.readError(resp)
	}
	var rec memoRecord
	if err := json.NewDecoder(resp.Body).Decode(&rec); err != nil {
		return "", "", false, err
	}
	return rec.OutputHash, rec.Output, true, nil
}

func (c *TSDBClient) SaveResult(snapshotID, dominoID, inputHash, outputHash, output string, reused bool) error {
	body, _ := json.Marshal(map[string]interface{}{
		"snapshot_id": snapshotID,
		"domino_id":   dominoID,
		"input_hash":  inputHash,
		"output_hash": outputHash,
		"output":      output,
		"reused":      reused,
	})
	return c.post("/v1/results", body)
}

func (c *TSDBClient) GetDominoOutput(snapshotID, dominoID string) (string, error) {
	path := fmt.Sprintf("/v1/outputs/%s/%s", url.PathEscape(snapshotID), url.PathEscape(dominoID))
	resp, err := c.client.Get(c.base + path)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", c.readError(resp)
	}
	var out struct {
		Output string `json:"output"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return "", err
	}
	return out.Output, nil
}

func (c *TSDBClient) Close() error {
	return nil
}

func (c *TSDBClient) post(path string, body []byte) error {
	resp, err := c.client.Post(c.base+path, "application/json", bytes.NewReader(body))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return c.readError(resp)
	}
	return nil
}

func (c *TSDBClient) readError(resp *http.Response) error {
	b, _ := io.ReadAll(resp.Body)
	return fmt.Errorf("tsdb api %s: %s", resp.Status, strings.TrimSpace(string(b)))
}
