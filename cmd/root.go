package cmd

import (
	"io"
	"os"

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
	return cmd
}
