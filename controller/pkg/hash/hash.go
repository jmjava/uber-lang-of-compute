package hash

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"sort"
)

// Compute returns a deterministic SHA-256 hex hash of the given data.
func Compute(data interface{}) (string, error) {
	normalized, err := normalize(data)
	if err != nil {
		return "", err
	}
	h := sha256.Sum256(normalized)
	return hex.EncodeToString(h[:]), nil
}

func normalize(data interface{}) ([]byte, error) {
	switch v := data.(type) {
	case string:
		return []byte(v), nil
	default:
		return json.Marshal(v)
	}
}

// ChainKey builds a memoization key from snapshot ID, domino ID, and input hash.
func ChainKey(snapshotID, dominoID, inputHash string) string {
	return snapshotID + ":" + dominoID + ":" + inputHash
}

// SnapshotID computes a deterministic ID from snapshot content.
func SnapshotID(timeSlice string, data interface{}) (string, error) {
	payload := map[string]interface{}{
		"timeSlice": timeSlice,
		"data":      data,
	}
	h, err := Compute(payload)
	if err != nil {
		return "", err
	}
	return h[:16], nil
}

// SortedKeys returns sorted keys for deterministic map iteration.
func SortedKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
