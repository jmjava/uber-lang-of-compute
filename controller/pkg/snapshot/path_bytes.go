package snapshot

// pathBytesView holds file bytes that may be mmap-backed; call release when done.
type pathBytesView struct {
	data    []byte
	release func()
}

func (v pathBytesView) close() {
	if v.release != nil {
		v.release()
	}
}
