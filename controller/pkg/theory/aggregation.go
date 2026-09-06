package theory

import "fmt"

// Node is one cell in a finite window of a k-ary aggregation tree.
//
// Correspondence: the Windowed Mandelbrot pattern is not the Mandelbrot set
// z ↦ z²+c. It is a windowed unfolding of an infinite self-similar hierarchy:
// only depth ≤ D is materialized, and coarsening one level recovers the
// shallower window (renormalization / multigrid analogue).
type Node struct {
	Label    string
	Depth    int
	Value    float64
	Children []*Node
}

// Unfold builds a perfect arity-ary tree of the given depth. Child values
// partition the parent additively (each child gets parent/arity), so the
// scalar is conserved under coarsening-by-sum — a discrete conservation law.
func Unfold(depth, arity int, label string, value float64) *Node {
	n := &Node{Label: label, Depth: depth, Value: value}
	if depth <= 0 || arity <= 0 {
		return n
	}
	share := value / float64(arity)
	n.Children = make([]*Node, arity)
	for i := 0; i < arity; i++ {
		childLabel := fmt.Sprintf("%s.%d", label, i)
		n.Children[i] = Unfold(depth-1, arity, childLabel, share)
	}
	return n
}

// Coarsen drops one level of descendants, replacing each node's children
// with the coarsened children, and restoring Values by summing children.
// A depth-0 node is a leaf and is returned unchanged.
func Coarsen(n *Node) *Node {
	if n == nil {
		return nil
	}
	out := &Node{Label: n.Label, Depth: n.Depth, Value: n.Value}
	if len(n.Children) == 0 {
		return out
	}
	// If children are leaves, fold them into this node.
	allLeaves := true
	for _, c := range n.Children {
		if len(c.Children) > 0 {
			allLeaves = false
			break
		}
	}
	if allLeaves {
		var sum float64
		for _, c := range n.Children {
			sum += c.Value
		}
		out.Value = sum
		out.Depth = 0
		return out
	}
	out.Children = make([]*Node, len(n.Children))
	var sum float64
	maxDepth := 0
	for i, c := range n.Children {
		out.Children[i] = Coarsen(c)
		sum += out.Children[i].Value
		if out.Children[i].Depth > maxDepth {
			maxDepth = out.Children[i].Depth
		}
	}
	out.Value = sum
	out.Depth = maxDepth + 1
	return out
}

// ShapeEqual reports whether two trees have the same arity structure and labels,
// ignoring numeric values.
func ShapeEqual(a, b *Node) bool {
	if a == nil || b == nil {
		return a == b
	}
	if a.Label != b.Label || len(a.Children) != len(b.Children) {
		return false
	}
	for i := range a.Children {
		if !ShapeEqual(a.Children[i], b.Children[i]) {
			return false
		}
	}
	return true
}

// Height is the longest path from the node to a leaf, in edges.
func Height(n *Node) int {
	if n == nil || len(n.Children) == 0 {
		return 0
	}
	max := 0
	for _, c := range n.Children {
		if h := Height(c); h > max {
			max = h
		}
	}
	return max + 1
}

// ValuesEqual reports whether corresponding nodes have the same value.
func ValuesEqual(a, b *Node) bool {
	if a == nil || b == nil {
		return a == b
	}
	if !AlmostEqual(a.Value, b.Value) || len(a.Children) != len(b.Children) {
		return false
	}
	for i := range a.Children {
		if !ValuesEqual(a.Children[i], b.Children[i]) {
			return false
		}
	}
	return true
}
