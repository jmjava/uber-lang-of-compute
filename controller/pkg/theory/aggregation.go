package theory

import (
	"fmt"
	"math"
)

// Node is one cell in a finite window of a k-ary aggregation tree.
//
// Correspondence: the Windowed Mandelbrot pattern copies the *explorer*
// (only a finite viewport of an infinite self-similar object is materialized).
// Coarsening recovers the shallower window (renormalization / multigrid analogue).
// This is nature-inspired design, not a claim that the scheduler is z ↦ z²+c.
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

// LeafCount is the number of leaves of a perfect arity-ary tree of the given depth.
func LeafCount(depth, arity int) int {
	if depth <= 0 || arity <= 0 {
		return 1
	}
	n := 1
	for i := 0; i < depth; i++ {
		n *= arity
	}
	return n
}

// Leaves returns leaf nodes in left-to-right order.
func Leaves(n *Node) []*Node {
	if n == nil {
		return nil
	}
	if len(n.Children) == 0 {
		return []*Node{n}
	}
	var out []*Node
	for _, c := range n.Children {
		out = append(out, Leaves(c)...)
	}
	return out
}

// SimilarityDimension is the similarity (Hutchinson) dimension of a perfect
// k-ary IFS with contraction 1/scale: log(k)/log(scale).
//
// Nature-inspired reading: this is the fractal *budget* of a windowed explorer
// (how densely a scale fills), not the Hausdorff dimension of the Mandelbrot set.
// Additive Unfold uses scale=arity, so the dimension is 1 — consistent with a
// conserved 1-D scalar (value-mass) along the hierarchy.
func SimilarityDimension(arity int, scale float64) float64 {
	if arity <= 0 || scale <= 1 {
		return 0
	}
	return math.Log(float64(arity)) / math.Log(scale)
}

// EscapeTime iterates z → z²+c until |z| exceeds radius or maxIter is reached.
//
// This is the Mandelbrot *explorer* primitive: a finite iteration window on an
// infinite generated object. It is not the scheduler's law of motion and does
// not identify KBL with the Mandelbrot set. It copies the viewport idea the
// way a neural net copies a firing rate.
func EscapeTime(c complex128, maxIter int, radius float64) int {
	z := 0 + 0i
	for i := 0; i < maxIter; i++ {
		z = z*z + c
		re, im := real(z), imag(z)
		if re*re+im*im > radius*radius {
			return i + 1
		}
	}
	return maxIter
}
