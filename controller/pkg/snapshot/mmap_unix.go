//go:build unix

package snapshot

import (
	"fmt"
	"os"

	"golang.org/x/sys/unix"
)

const mmapReadThreshold = 1 << 20 // 1 MiB

func readPathBytesMmap(path string, size int) ([]byte, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	mapped, err := unix.Mmap(int(f.Fd()), 0, size, unix.PROT_READ, unix.MAP_SHARED)
	if err != nil {
		return nil, err
	}

	out := make([]byte, size)
	copy(out, mapped)
	if err := unix.Munmap(mapped); err != nil {
		return nil, fmt.Errorf("munmap %q: %w", path, err)
	}
	return out, nil
}

func readPathBytesOptimized(path string) ([]byte, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("read snapshot path %q: %w", path, err)
	}
	if info.Size() > mmapReadThreshold {
		return readPathBytesMmap(path, int(info.Size()))
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read snapshot path %q: %w", path, err)
	}
	return data, nil
}
