package cmd

import (
	"io"
	"os"
	"regexp"
	"strings"

	"github.com/bmatcuk/doublestar"
	"github.com/spf13/cobra"
)

var (
	gPattern string
	gRegexp  string
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

	cmd.AddCommand(newLsCmd(stdout, stderr))
	cmd.AddCommand(newRmCmd(stdout, stderr))
	cmd.AddCommand(newConvertCmd(stdout, stderr))
	cmd.AddCommand(newRenameCmd(stdout, stderr))

	cmd.PersistentFlags().StringVar(&gPattern, "filter", "", "target filename pattern (support wildcard)")
	cmd.PersistentFlags().StringVar(&gRegexp, "regexp", "", "target filename pattern (support regexp)")
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

func generatePathFilter(def func(string) (bool, error)) (func(string) (bool, error), error) {
	filter := def

	if len(gRegexp) != 0 {
		reg, err := regexp.Compile(gRegexp)
		if err != nil {
			return nil, err
		}
		filter = func(s string) (bool, error) {
			return reg.Match([]byte(s)), nil
		}
	} else if len(gPattern) != 0 {
		filter = func(s string) (bool, error) {
			return doublestar.Match(gPattern, s)
		}
	}
	return filter, nil
}
