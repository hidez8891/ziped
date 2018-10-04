package cmd

import (
	"io"
	"os"
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

	cmd.AddCommand(newLsCmd(stdout, stderr))
	cmd.AddCommand(newRmCmd(stdout, stderr))
	cmd.AddCommand(newConvertCmd(stdout, stderr))
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
