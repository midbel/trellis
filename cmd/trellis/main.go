package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"

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
	options, err := spec.Build()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	if err := trellis.Render(os.Stdout, spec.Node, options); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

type CliFlags struct {
	AllocateMode trellis.Allocate
	Width        int
	Height       int
	Reverse      bool
	Border       bool
	Orient       trellis.Orientation
	Output       trellis.Output
	Style        trellis.ConnectorStyle
	File         string
}

func parseArgs(args []string) (CliFlags, *flag.FlagSet, error) {
	var (
		cli = CliFlags{
			AllocateMode: trellis.AllocateEqual,
		}
		fs = flag.NewFlagSet("trellis", flag.ExitOnError)
	)
	fs.IntVar(&cli.Width, "w", 0, "width")
	fs.IntVar(&cli.Height, "h", 0, "height")
	fs.BoolVar(&cli.Reverse, "r", false, "reverse")
	fs.BoolVar(&cli.Border, "b", false, "border")

	fs.Func("a", "allocation", func(str string) error {
		v, err := trellis.ParseAllocate(str)
		cli.AllocateMode = v
		return err
	})
	fs.Func("t", "orientation", func(str string) error {
		v, err := trellis.ParseOrientation(str)
		cli.Orient = v
		return err
	})
	fs.Func("o", "output", func(str string) error {
		v, err := trellis.ParseOutput(str)
		cli.Output = v
		return err
	})
	fs.Func("c", "connector style", func(str string) error {
		v, err := trellis.ParseConnector(str)
		cli.Style = v
		return err
	})

	err := fs.Parse(args)
	if err == nil {
		cli.File = fs.Arg(0)
	}
	return cli, fs, err
}

func applyOverrides(fs *flag.FlagSet, cli *CliFlags, spec *codec.TreeSpec) {
	fs.Visit(func(f *flag.Flag) {
		switch f.Name {
		case "a":
			spec.Options.AllocateMode = cli.AllocateMode
		case "w":
			spec.Options.Width = cli.Width
		case "h":
			spec.Options.Height = cli.Height
		case "t":
			spec.Options.Orient = cli.Orient
		case "o":
			spec.Options.Type = cli.Output
		case "c":
			spec.Options.Style = cli.Style
		case "r":
			spec.Options.Reverse = cli.Reverse
		case "b":
			spec.Options.Border = cli.Border
		default:
		}
	})
}

func loadTree() (*codec.TreeSpec, error) {
	cli, fs, err := parseArgs(os.Args[1:])
	if err != nil {
		return nil, err
	}

	r, err := os.Open(cli.File)
	if err != nil {
		return nil, err
	}
	defer r.Close()

	var format codec.Format
	switch filepath.Ext(cli.File) {
	case ".sexpr":
		format = codec.FormatSexpr
	case ".json":
		format = codec.FormatJson
	default:
		return nil, fmt.Errorf("file type not supported")
	}

	opts := codec.Options{
		Format: format,
	}
	spec, err := codec.Tree(r, opts)
	if err != nil {
		return nil, err
	}
	spec.Options.AllocateMode = trellis.AllocateEqual
	applyOverrides(fs, &cli, spec)
	return spec, nil
}
