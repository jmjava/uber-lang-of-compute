//go:build !unix

package snapshot

import "os"

func readPathBytesOptimized(path string) ([]byte, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return data, nil
}
