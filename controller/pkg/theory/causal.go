package theory

import "fmt"

// CausalPast returns the closed predecessor set of a domino in a chain:
// the snapshot token plus every earlier name in the execution order.
//
// Correspondence: during compute the readable set is the causal past of the
// event exec(d), analogous to a discrete light-cone (Lamport happens-before),
// not a Lorentzian metric. Cross-universe materialization is excluded from
// this past by construction — those events occur only after chain completion.
func CausalPast(chain []string, domino string) ([]string, error) {
	past := []string{"snapshot"}
	for _, name := range chain {
		past = append(past, name)
		if name == domino {
			return past, nil
		}
	}
	return nil, fmt.Errorf("domino %q not in chain", domino)
}

// AllowedReads reports whether every requested input name lies in the causal
// past of the executing domino. "snapshot" is always allowed; other names
// must be strictly earlier in the chain (not the current domino).
func AllowedReads(chain []string, domino string, inputs []string) error {
	idx := indexOf(chain, domino)
	if idx < 0 {
		return fmt.Errorf("domino %q not in chain", domino)
	}
	for _, in := range inputs {
		if in == "snapshot" || in == "" {
			continue
		}
		src := indexOf(chain, in)
		if src < 0 {
			return fmt.Errorf("input %q is not in the chain (outside causal past of %q)", in, domino)
		}
		if src >= idx {
			return fmt.Errorf("input %q is not strictly before %q (spacelike / future read)", in, domino)
		}
	}
	return nil
}

func indexOf(chain []string, name string) int {
	for i, n := range chain {
		if n == name {
			return i
		}
	}
	return -1
}
