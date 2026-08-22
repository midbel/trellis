package main

import (
	"flag"
	"os"

	"github.com/midbel/trellis"
)

func main() {
	layoutOpts := trellis.LayoutOptions{
		VerticalGap:   trellis.DefaultVerticalGapSize,
		HorizontalGap: trellis.DefaultHorizontalGapSize,
		Position:      trellis.ParentAlignFirst,
		Reverse:       false,
	}
	styleOpts := trellis.StyleOptions{
		Align: trellis.AlignCenter,
	}
	renderOpts := trellis.RenderOptions{
		Border: false,
	}
	opts := trellis.Options{
		LayoutOptions: layoutOpts,
		StyleOptions:  styleOpts,
		RenderOptions: renderOpts,
	}
	var kind string
	flag.IntVar(&opts.Width, "w", 0, "width")
	flag.IntVar(&opts.Height, "h", 0, "height")
	flag.IntVar(&opts.VerticalGap, "g", opts.VerticalGap, "vertical gap")
	flag.IntVar(&opts.HorizontalGap, "t", opts.HorizontalGap, "horizontal gap")
	flag.BoolVar(&opts.Border, "b", false, "add border")
	flag.BoolVar(&opts.Reverse, "r", false, "reverse tree")
	flag.IntVar(&opts.Padding, "i", 1, "value padding")
	flag.StringVar(&kind, "k", "", "tree rendering type")
	flag.Parse()

	sub1 := trellis.NewNode("dockit")
	sub1.Nodes = []*trellis.Node{
		trellis.NewNode("oxml"),
		trellis.NewNode("ods"),
		trellis.NewNode("flat"),
		trellis.NewNode("formula"),
	}
	sub2 := trellis.NewNode("angle")
	sub2.Nodes = []*trellis.Node{
		trellis.NewNode("xpath"),
		trellis.NewNode("xslt"),
		trellis.NewNode("relax"),
	}
	sub3 := trellis.NewNode("curly")
	sub3.Nodes = []*trellis.Node{
		trellis.NewNode("jsonata"),
	}
	root := trellis.NewNode("Midbel")
	root.Nodes = []*trellis.Node{
		// trellis.NewNode("trellis"),
		sub2,
		sub3,
		trellis.NewNode("cli"),
		sub1,
		trellis.NewNode("probe"),
		// trellis.NewNode("packit"),
		// trellis.NewNode("sweet"),
	}
	tree := trellis.NewTree(root)
	switch kind {
	case "", "vertical":
		trellis.VerticalTree(os.Stdout, tree, &opts)
	case "horizontal":
		trellis.HorizontalTree(os.Stdout, tree, &opts)
	case "compact":
		trellis.CompactTree(os.Stdout, tree, &opts)
	default:
	}
}
