package cmd

import (
	"fmt"
	"io"
	"regexp"

	"github.com/bmatcuk/doublestar"
	"github.com/hidez8891/zip"
	"github.com/spf13/cobra"
)

var (
	findtext   string
	findregexp string
)

func newLsCmd(stdout, stderr io.Writer) *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "ls [filepath...]",
		Short: "Show file list",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			lscmd := &ls{stdout, stderr}
			lscmd.run(cmd, args)
		},
	}

	cmd.Flags().StringVar(&findtext, "filter", "", "Show filename pattern (support wildcard)")
	cmd.Flags().StringVar(&findregexp, "regexp", "", "Show filename pattern (support regexp)")
	return cmd
}

type ls struct {
	stdout io.Writer
	stderr io.Writer
}

func (o *ls) run(cmd *cobra.Command, args []string) {
	if len(args) == 1 {
		err := o.render(args[0])
		if err != nil {
			fmt.Fprintln(o.stderr, err)
			return
		}
	} else {
		for i, filepath := range args {
			fmt.Fprintf(o.stdout, "%s:\n", filepath)

			err := o.render(filepath)
			if err != nil {
				fmt.Fprintln(o.stderr, err)
				return
			}
			if i != len(args)-1 {
				fmt.Fprintln(o.stdout)
			}
		}
	}
}

func (o *ls) render(filepath string) error {
	zr, err := zip.OpenReader(filepath)
	if err != nil {
		return err
	}
	defer zr.Close()

	filter := func(_ string) (bool, error) {
		return true, nil
	}
	if len(findregexp) != 0 {
		reg, err := regexp.Compile(findregexp)
		if err != nil {
			return err
		}
		filter = func(s string) (bool, error) {
			return reg.Match([]byte(s)), nil
		}
	} else if len(findtext) != 0 {
		filter = func(s string) (bool, error) {
			return doublestar.Match(findtext, s)
		}
	}

	for _, zf := range zr.File {
		ok, err := filter(zf.Name)
		if err != nil {
			return err
		}
		if ok {
			fmt.Fprintln(o.stdout, zf.Name)
		}
	}
	return nil
}
