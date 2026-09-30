package trellis

type Node struct {
	Value string
	Nodes []*Node
}

func NewNode(value string) *Node {
	return &Node{
		Value: value,
	}
}

func (n *Node) Leaf() bool {
	return len(n.Nodes) == 0
}

func (n *Node) Weight() int {
	if n.Leaf() {
		return 1
	}
	var count int
	for _, c := range n.Nodes {
		count += c.Weight()
	}
	return count + 1
}

func (n *Node) Count() int {
	if n.Leaf() {
		return 1
	}
	var count int
	for i := range n.Nodes {
		count += n.Nodes[i].Count()
	}
	return count
}

func (n *Node) Depth() int {
	if n.Leaf() {
		return 0
	}
	var depth int
	for i := range n.Nodes {
		z := n.Nodes[i].Depth()
		depth = max(depth, z)
	}
	return depth + 1
}

func traverse(node *Node, target int) []*Node {
	if target == 0 {
		return []*Node{node}
	}
	return traverseDepth(node, 0, target-1)
}

func traverseDepth(node *Node, currDepth, targetDepth int) []*Node {
	if currDepth >= targetDepth {
		return node.Nodes
	}
	var all []*Node
	for _, n := range node.Nodes {
		ns := traverseDepth(n, currDepth+1, targetDepth)
		all = append(all, ns...)
	}
	return all
}
