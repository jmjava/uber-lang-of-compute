package theory

import (
	"fmt"

	"github.com/jmjava/uber-lang-of-compute/controller/pkg/hash"
	"github.com/jmjava/uber-lang-of-compute/controller/pkg/types"
)

// VerifySpine checks that a replay log is a hash chain from the snapshot ID.
// Nature-inspired reading: a worldline you can audit, like a Merkle spine,
// not a full Merkle DAG and not a blockchain.
func VerifySpine(snapshotID string, entries []types.ReplayLogEntry) error {
	prev := snapshotID
	for i, e := range entries {
		if e.PrevLink != prev {
			return fmt.Errorf("entry %d prev-link mismatch", i)
		}
		want, err := hash.Link(e.PrevLink, e.InputHash, e.OutputHash)
		if err != nil {
			return err
		}
		if e.Link != want {
			return fmt.Errorf("entry %d link mismatch", i)
		}
		prev = e.Link
	}
	return nil
}
