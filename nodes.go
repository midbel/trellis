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
