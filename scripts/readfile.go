package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/midbel/cli"
	"github.com/midbel/trellis"
	"github.com/midbel/trellis/codec"
)

func main() {
	flag.Parse()

	tree, err := loadTree(flag.Arg(0))
	if err != nil {
		cli.FailData(err)
	}
	if err := renderTree(tree); err != nil {
		cli.FailInternal(err)
	}
}

func renderTree(spec *codec.TreeSpec) error {
	switch spec.Type {
	case "", "horizontal":
		return trellis.HorizontalTree(os.Stdout, spec.Tree, spec.Options)
	case "vertical":
		return trellis.VerticalTree(os.Stdout, spec.Tree, spec.Options)
	case "compact":
		return trellis.CompactTree(os.Stdout, spec.Tree, spec.Options)
	default:
		return nil
	}
}

func loadTree(file string) (*codec.TreeSpec, error) {
	r, err := os.Open(file)
	if err != nil {
		cli.FailIO(err)
	}
	defer r.Close()

	opts := codec.Options{
		Format: codec.FormatSexpr,
	}
	return codec.Tree(r, opts)
}
