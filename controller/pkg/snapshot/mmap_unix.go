//go:build unix

package snapshot

import (
	"fmt"
	"os"

	"golang.org/x/sys/unix"
)

const mmapReadThreshold = 1 << 20 // 1 MiB

func readPathBytesForSeal(path string) (pathBytesView, error) {
	info, err := os.Stat(path)
	if err != nil {
		return pathBytesView{}, fmt.Errorf("read snapshot path %q: %w", path, err)
	}
	if info.Size() > mmapReadThreshold {
		return mmapPathBytes(path, int(info.Size()))
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return pathBytesView{}, fmt.Errorf("read snapshot path %q: %w", path, err)
	}
	return pathBytesView{data: data}, nil
}

func mmapPathBytes(path string, size int) (pathBytesView, error) {
	f, err := os.Open(path)
	if err != nil {
		return pathBytesView{}, err
	}

	mapped, err := unix.Mmap(int(f.Fd()), 0, size, unix.PROT_READ, unix.MAP_SHARED)
	if err != nil {
		f.Close()
		return pathBytesView{}, err
	}

	return pathBytesView{
		data: mapped,
		release: func() {
			_ = unix.Munmap(mapped)
			f.Close()
		},
	}, nil
}

func readPathBytesOptimized(path string) ([]byte, error) {
	view, err := readPathBytesForSeal(path)
	if err != nil {
		return nil, err
	}
	defer view.close()

	out := make([]byte, len(view.data))
	copy(out, view.data)
	return out, nil
}
