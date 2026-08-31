package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"slices"
	"strconv"

	"github.com/midbel/cli"
	"github.com/midbel/trellis"
	"github.com/midbel/trellis/codec"
)

var errFail = errors.New("fail")

func main() {
	var (
		set  = cli.NewFlagSet("trellis")
		root = prepare()
	)
	if err := set.Parse(os.Args[1:]); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			root.Help()
			os.Exit(2)
		}
	}
	err := root.Execute(set.Args())
	if err != nil {
		if s, ok := err.(cli.SuggestionError); ok && len(s.Others) > 0 {
			fmt.Fprintln(os.Stderr, "similar command(s)")
			for _, n := range s.Others {
				fmt.Fprintln(os.Stderr, "-", n)
			}
		}
		if !errors.Is(err, errFail) {
			fmt.Fprintln(os.Stderr, err)
		}
		os.Exit(1)
	}
}

func prepare() *cli.CommandTrie {
	root := cli.New()
	root.Register(single("horizontal"), &horizontalCmd)
	root.Register(single("vertical"), &verticalCmd)
	root.Register(single("compact"), &compactCmd)
	root.Register(single("treemap"), &treemapCmd)
	root.Register(single("inspect"), &inspectCmd)
	return root
}

func single(str string) []string {
	return []string{str}
}

var horizontalCmd = cli.Command{
	Name:    "horizontal",
	Summary: "",
	Usage:   "horizontal <file>",
	Handler: &horizontalCommand{},
}

type horizontalCommand struct{}

func (c horizontalCommand) Run(args []string) error {
	set := cli.NewFlagSet("horizontal")
	if err := set.Parse(args); err != nil {
		return err
	}
	if set.NArg() != 1 {
		return cli.ErrUsage
	}
	return nil
}

var verticalCmd = cli.Command{
	Name:    "vertical",
	Summary: "",
	Usage:   "vertical <file>",
	Handler: &verticalCommand{},
}

type verticalCommand struct{}

func (c verticalCommand) Run(args []string) error {
	set := cli.NewFlagSet("vertical")
	if err := set.Parse(args); err != nil {
		return err
	}
	if set.NArg() != 1 {
		return cli.ErrUsage
	}
	return nil
}

var compactCmd = cli.Command{
	Name:    "compact",
	Summary: "",
	Usage:   "compact <file>",
	Handler: &compactCommand{},
}

type compactCommand struct{}

func (c compactCommand) Run(args []string) error {
	set := cli.NewFlagSet("compact")
	if err := set.Parse(args); err != nil {
		return err
	}
	if set.NArg() != 1 {
		return cli.ErrUsage
	}
	return nil
}

var treemapCmd = cli.Command{
	Name:    "treemap",
	Summary: "",
	Usage:   "treemap <file>",
	Handler: &treemapCommand{},
}

type treemapCommand struct{}

func (c treemapCommand) Run(args []string) error {
	set := cli.NewFlagSet("treemap")
	if err := set.Parse(args); err != nil {
		return err
	}
	if set.NArg() != 1 {
		return cli.ErrUsage
	}
	return nil
}

var inspectCmd = cli.Command{
	Name:    "inspect",
	Summary: "",
	Usage:   "inspect <file>",
	Handler: &inspectCommand{},
}

type inspectCommand struct {
	Type    trellis.Orientation
	Reverse bool
	Width   int
	Height  int
	Spacing int
}

func (c inspectCommand) Run(args []string) error {
	set := cli.NewFlagSet("inspect")
	set.BoolVar(&c.Reverse, "r", false, "reverse chart")
	set.IntVar(&c.Width, "w", 0, "width")
	set.IntVar(&c.Height, "h", 0, "height")
	set.Func("k", "type", func(str string) error {
		o, err := trellis.ParseOrientation(str)
		if err == nil {
			c.Type = o
		}
		return err
	})
	if err := set.Parse(args); err != nil {
		return err
	}
	if set.NArg() != 1 {
		return cli.ErrUsage
	}
	r, err := os.Open(set.Arg(0))
	if err != nil {
		cli.FailIO(err)
	}
	defer r.Close()

	spec, err := codec.Tree(r, codec.Options{
		Format: codec.FormatSexpr,
	})
	if err != nil {
		cli.FailData(err)
	}
	tbl1 := cli.Table{
		Headers: []string{
			"value",
			"ideal-x",
			"ideal-y",
			"computed-x",
			"computed-y",
			"start-x",
			"end-x",
			"start-y",
			"end-y",
			"width",
			"height",
		},
	}
	if c.Width > 0 {
		spec.Width = c.Width
	}
	if c.Height > 0 {
		spec.Height = c.Height
	}
	if c.Type > 0 {
		spec.Orient = c.Type
	}
	res := trellis.ComputeLayout(spec.Tree, spec.Options)

	tbl2 := cli.Table{
		Rows: [][]string{
			{"Width", strconv.Itoa(spec.Width), strconv.Itoa(res.Width)},
			{"Height", strconv.Itoa(spec.Height), strconv.Itoa(res.Height)},
		},
	}

	slices.SortFunc(res.Coordinates, func(c1, c2 trellis.Coordinate) int {
		var diff int
		switch c.Type {
		case trellis.VerticalLayout:
			diff = c1.Ideal.Y - c2.Ideal.Y
			if diff == 0 {
				diff = c1.Ideal.X - c2.Ideal.X
			}
		default:
			diff = c1.Ideal.X - c2.Ideal.X
			if diff == 0 {
				diff = c1.Ideal.Y - c2.Ideal.Y
			}
		}
		return diff
	})

	for _, n := range res.Coordinates {
		row := []string{
			n.Value,
			strconv.Itoa(n.Ideal.X),
			strconv.Itoa(n.Ideal.Y),
			strconv.Itoa(n.Computed.X),
			strconv.Itoa(n.Computed.Y),
			strconv.Itoa(n.Width.Start),
			strconv.Itoa(n.Width.End),
			strconv.Itoa(n.Height.Start),
			strconv.Itoa(n.Height.End),
			strconv.Itoa(n.Width.Len()),
			strconv.Itoa(n.Height.Len()),
		}
		tbl1.Rows = append(tbl1.Rows, row)
	}
	rdr := cli.NewTableRenderer(cli.Stdout)
	rdr.Render(tbl1)
	rdr.Empty()
	rdr.Render(tbl2)
	return nil
}
