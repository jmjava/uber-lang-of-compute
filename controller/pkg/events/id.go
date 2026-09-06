package events

import "github.com/jmjava/uber-lang-of-compute/controller/pkg/hash"

// EventID is a content address of a completed classical record.
// Wall-clock is excluded (M16/M22): retries of the same worldline agree.
func EventID(snapshotID, worldline, universe, workflow string) (string, error) {
	return hash.Link("kbl.snapshot.completed", snapshotID, worldline, universe, workflow)
}
