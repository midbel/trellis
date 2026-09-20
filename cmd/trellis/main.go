package main

import (
	"errors"
	"flag"
	"fmt"
	"os"

	"github.com/midbel/trellis"
	"github.com/midbel/trellis/codec"
)

var errFail = errors.New("fail")

func main() {
	spec, err := loadTree()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	if err := trellis.Render(os.Stdout, spec.Node, spec.Options); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func loadTree() (*codec.TreeSpec, error) {
	var (
		width   = flag.Int("w", 0, "width")
		height  = flag.Int("h", 0, "height")
		reverse = flag.Bool("r", false, "reverse")
		border = flag.Bool("b", false, "border")
		orient  trellis.Orientation
		output  trellis.Output
		style   trellis.ConnectorStyle
	)
	flag.Func("t", "orientation", func(str string) error {
		v, err := trellis.ParseOrientation(str)
		if err == nil {
			orient = v
		}
		return err
	})
	flag.Func("o", "output", func(str string) error {
		v, err := trellis.ParseOutput(str)
		if err == nil {
			output = v
		}
		return err
	})
	flag.Func("c", "connector style", func(str string) error {
		v, err := trellis.ParseConnector(str)
		if err == nil {
			style = v
		}
		return err
	})
	flag.Parse()

	r, err := os.Open(flag.Arg(0))
	if err != nil {
		return nil, err
	}
	defer r.Close()

	opts := codec.Options{
		Format: codec.FormatSexpr,
	}
	spec, err := codec.Tree(r, opts)
	if err != nil {
		return nil, err
	}
	if *width > 0 {
		spec.Options.Width = *width
	}
	if *height > 0 {
		spec.Options.Height = *height
	}
	if orient > 0 {
		spec.Options.Orient = orient
	}
	if output > 0 {
		spec.Options.Output = output
	}
	if *reverse != spec.Options.Reverse {
		spec.Options.Reverse = *reverse
	}
	if *border != spec.Options.Border {
		spec.Options.Border = *border
	}
	if style > 0 {
		spec.Options.Style = style
	}
	return spec, nil
}
