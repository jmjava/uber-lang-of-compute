package theory

import "fmt"

// UniqueNames reports whether every label in a chain is distinct.
// Nature-inspired reading: one name, one particle — Pauli-style exclusion
// in a discrete state space, not a spin-statistics theorem.
func UniqueNames(chain []string) error {
	seen := make(map[string]struct{}, len(chain))
	for _, n := range chain {
		if n == "" {
			return fmt.Errorf("empty chain name")
		}
		if _, ok := seen[n]; ok {
			return fmt.Errorf("duplicate chain name %q", n)
		}
		seen[n] = struct{}{}
	}
	return nil
}
