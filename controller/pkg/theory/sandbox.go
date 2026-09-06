package theory

import "fmt"

// SandboxPolicy is the isolation envelope that lets a contract-grade command
// be evidenced as a function (H2). Nature-inspired reading: a Faraday cage /
// isolated lab bench — not a proof that arbitrary images are pure, but a
// named condition under which we are willing to treat them as such.
type SandboxPolicy struct {
	NetworkNone  bool
	ReadOnlyRoot bool
}

// Isolated reports whether both network and root filesystem are locked down.
func (p SandboxPolicy) Isolated() bool {
	return p.NetworkNone && p.ReadOnlyRoot
}

// AllowDeterministic is the Picard gate with an isolation exception:
// builtin/julia pass; contract-grade passes only inside an isolated sandbox.
// Isolation is a cage, not a uniqueness proof: sandbox:impure is admitted
// and is still not a function of its inputs (M10).
func AllowDeterministic(command string, sandbox SandboxPolicy) error {
	if CommandRegularity(command) != RegularityContract {
		return nil
	}
	if sandbox.Isolated() {
		return nil
	}
	return fmt.Errorf("command %q is contract-grade; deterministic workflows require builtin:, julia:, or an isolated sandbox (network=none, read-only root)", command)
}
