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
	Width   int
	Height  int
	Reverse bool
	Border  bool
	Orient  trellis.Orientation
	Output  trellis.Output
	Style   trellis.ConnectorStyle
	File    string
}

// func (c CliFlags) Build() (trellis.Options, error) {
// 	base := trellis.RenderOptions{
// 		Orient:  c.Orient,
// 		Reverse: c.Reverse,
// 		Size:    trellis.NewDimension(c.Width, c.Height),
// 		Padding: trellis.PaddingM,
// 		Margin:  trellis.SpacingM,
// 		Spacing: trellis.SpacingM,
// 		AlignX:  trellis.AlignCenter,
// 		AlignY:  trellis.AlignCenter,
// 	}
// 	var opts trellis.Options
// 	switch c.Output {
// 	case trellis.OutputScreen:
// 		opts = &trellis.ScreenOptions{
// 			RenderOptions: base,
// 			Border:        c.Border,
// 			Style:         c.Style,
// 		}
// 	case trellis.OutputSvg:
// 		opts = &trellis.SvgOptions{
// 			RenderOptions: base,
// 			Border:        c.Border,
// 			Style:         c.Style,
// 			Path:          trellis.ManathanPath,
// 		}
// 	case trellis.OutputXml:
// 		opts = &trellis.XmlOptions{
// 			RenderOptions: base,
// 		}
// 	case trellis.OutputJson:
// 		opts = &trellis.JsonOptions{
// 			RenderOptions: base,
// 		}
// 	default:
// 		return nil, fmt.Errorf("unsupported output type")
// 	}
// 	return opts, nil
// }

func parseArgs(args []string) (CliFlags, *flag.FlagSet, error) {
	var (
		cli CliFlags
		fs  = flag.NewFlagSet("trellis", flag.ExitOnError)
	)
	fs.IntVar(&cli.Width, "w", 0, "width")
	fs.IntVar(&cli.Height, "h", 0, "height")
	fs.BoolVar(&cli.Reverse, "r", false, "reverse")
	fs.BoolVar(&cli.Border, "b", false, "border")

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

	opts := codec.Options{
		Format: codec.FormatSexpr,
	}
	spec, err := codec.Tree(r, opts)
	if err != nil {
		return nil, err
	}
	applyOverrides(fs, &cli, spec)
	return spec, nil
}
