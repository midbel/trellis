package main

import (
	"errors"
	"flag"
	"fmt"
	"os"

	"github.com/midbel/cli"
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

type inspectCommand struct{}

func (c inspectCommand) Run(args []string) error {
	set := cli.NewFlagSet("inspect")
	if err := set.Parse(args); err != nil {
		return err
	}
	if set.NArg() != 1 {
		return cli.ErrUsage
	}
	return nil
}
