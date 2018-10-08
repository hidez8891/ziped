package cmd

import (
	"io"
	"os"
	"regexp"
	"strings"

	"github.com/bmatcuk/doublestar"
	"github.com/spf13/cobra"
)

// Execute executes the root command.
func Execute() {
	cmd := newRootCmd(os.Stdout, os.Stderr)
	if err := cmd.Execute(); err != nil {
		cmd.Println(err)
		os.Exit(1)
	}
}

func newRootCmd(stdout, stderr io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use: "ziped",
	}
	cmd.SetOutput(stderr)

	params := &cmdParams{
		stdout: stdout,
		stderr: stderr,
	}

	cmd.AddCommand(newLsCmd(params))
	cmd.AddCommand(newRmCmd(params))
	cmd.AddCommand(newConvertCmd(params))
	cmd.AddCommand(newRenameCmd(params))

	cmd.PersistentFlags().StringVar(&params.pattern, "filter", "", "target filename pattern (support wildcard)")
	cmd.PersistentFlags().StringVar(&params.regexp, "regexp", "", "target filename pattern (support regexp)")
	cmd.PersistentFlags().BoolVar(&params.isOverwrite, "overwrite", false, "overwrite source file")
	cmd.PersistentFlags().StringVar(&params.outFilename, "out", "", "output file name")
	return cmd
}

func expandFilePath(paths []string) ([]string, error) {
	newpaths := make([]string, 0)
	for _, path := range paths {
		if strings.Contains(path, "*") {
			ps, err := doublestar.Glob(path)
			if err != nil {
				return nil, err
			}
			newpaths = append(newpaths, ps...)
		} else {
			newpaths = append(newpaths, path)
		}
	}
	return newpaths, nil
}

type cmdParams struct {
	pattern     string
	regexp      string
	isOverwrite bool
	outFilename string
	stdout      io.Writer
	stderr      io.Writer
}

func (o *cmdParams) generatePathFilter() (func(string) (bool, error), error) {
	filter := func(_ string) (bool, error) {
		return true, nil
	}

	if len(o.regexp) != 0 {
		reg, err := regexp.Compile(o.regexp)
		if err != nil {
			return nil, err
		}
		filter = func(s string) (bool, error) {
			return reg.Match([]byte(s)), nil
		}
	} else if len(o.pattern) != 0 {
		filter = func(s string) (bool, error) {
			return doublestar.Match(o.pattern, s)
		}
	}
	return filter, nil
}
