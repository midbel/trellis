package trellis

import (
	"bufio"
	"fmt"
	"io"
	"strings"
)

type compact struct {
	w *bufio.Writer
}

func NewCompact(w io.Writer) Renderer {
	return compact{
		w: bufio.NewWriter(w),
	}
}

func (c compact) Render(tree *Tree, options *Options) error {
	defer c.w.Flush()
	var (
		opts  = prepareOptions(options)
		maker = makeLayout(opts.VerticalGap, opts.Position)
		root  = maker.Single(tree.Root)
	)
	if opts.Padding == 0 {
		opts.Padding++
	}
	c.printNode(root, 0, options)
	return nil
}

func (c compact) printNode(node *layoutNode, depth int, opts *Options) {
	var padding string
	if depth > 0 {
		padding = strings.Repeat(" ", depth*opts.Padding)
	}
	fmt.Fprint(c.w, padding)
	if depth > 0 {
		c.w.WriteByte(verticalBarAscii)
		c.w.WriteByte(horizontalBarAscii)
	}
	str := opts.paddedValue(node.Value)
	fmt.Fprintln(c.w, string(str))
	for _, x := range node.Children {
		c.printNode(x, depth+1, opts)
	}
}
