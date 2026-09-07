package theory

import (
	"github.com/jmjava/uber-lang-of-compute/controller/pkg/types"
)

// WorkflowFromUnfold builds a live deterministic workflow whose dominos are
// the nodes of a windowed aggregation tree. Leaves identity the snapshot;
// parents run builtin:coarsen (sum of child value/v) in post-order. This is
// F1 leaving the pattern library: Unfold is an executable aggregation chain.
func WorkflowFromUnfold(depth, arity int, snap types.Snapshot) *types.Workflow {
	root := Unfold(depth, arity, "n", 1)
	dominos, chain := dominosFromNode(root)
	if snap.Metadata.Name == "" {
		snap.Metadata.Name = "unfold-snap"
	}
	return &types.Workflow{
		Kind:     "Workflow",
		Metadata: types.ObjectMeta{Name: "unfold"},
		Spec: types.WorkflowSpec{
			Snapshot: snap,
			Dominos:  dominos,
			Execution: types.ExecutionConfig{
				Chain:         chain,
				Deterministic: true,
			},
		},
	}
}

func dominosFromNode(n *Node) ([]types.Domino, []string) {
	var dominos []types.Domino
	var chain []string
	var walk func(*Node)
	walk = func(n *Node) {
		if n == nil {
			return
		}
		for _, c := range n.Children {
			walk(c)
		}
		d := types.Domino{
			Metadata: types.ObjectMeta{Name: n.Label},
			Spec:     types.DominoSpec{Command: "builtin:identity"},
		}
		if len(n.Children) == 0 {
			d.Spec.Inputs = []types.DominoInput{{FromSnapshot: "snap"}}
		} else {
			d.Spec.Command = "builtin:coarsen"
			for _, c := range n.Children {
				d.Spec.DependsOn = append(d.Spec.DependsOn, c.Label)
				d.Spec.Inputs = append(d.Spec.Inputs, types.DominoInput{FromDomino: c.Label})
			}
		}
		dominos = append(dominos, d)
		chain = append(chain, n.Label)
	}
	walk(n)
	return dominos, chain
}

// NodeCount is the number of cells in a perfect k-ary tree of the given depth.
func NodeCount(depth, arity int) int {
	if depth <= 0 || arity <= 0 {
		return 1
	}
	// (k^{d+1} - 1) / (k - 1); arity==1 is a path of length depth.
	if arity == 1 {
		return depth + 1
	}
	pow := 1
	for i := 0; i <= depth; i++ {
		pow *= arity
	}
	return (pow - 1) / (arity - 1)
}
