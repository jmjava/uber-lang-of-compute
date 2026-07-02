//go:build !unix

package snapshot

import "os"

func readPathBytesForSeal(path string) (pathBytesView, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return pathBytesView{}, err
	}
	return pathBytesView{data: data}, nil
}

func readPathBytesOptimized(path string) ([]byte, error) {
	return os.ReadFile(path)
}
