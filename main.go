package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/MakeNowJust/heredoc/v2"
)

func usage(writer io.Writer, cmd, version string) {
	tmpl := heredoc.Doc(`
		Name:
			{:CMD:} - [{:VERSION:}]

		Usage:
			(1) single command
				{:CMD:} [OPTIONS] COMMAND <FILES...>

			(2) multiple command
				{:CMD:} [OPTIONS] exec <FILES...>
					-- COMMAND-1 [CMD-OPTIONS-1]
					-- COMMAND-2 [CMD-OPTIONS-2]
					...

		Commands:
			convert    Convert file contents
			ls         Show file list
			rename     Rename file name
			rm         Remove file

		Options:
			-h, --help      help for {:CMD:}
			-v, --version   version for {:CMD:}
	`)

	tmpl = strings.ReplaceAll(tmpl, "\t", "    ")
	tmpl = strings.ReplaceAll(tmpl, "{:CMD:}", cmd)
	tmpl = strings.ReplaceAll(tmpl, "{:VERSION:}", version)
	fmt.Fprintln(writer, tmpl)
}

func main() {
	flags := flag.NewFlagSet("ziped", flag.ExitOnError)
	flags.Usage = func() {
		usage(flags.Output(), "ziped", "0.0.0-dev")
	}

	if err := flags.Parse(os.Args[1:]); err != nil {
		if err == flag.ErrHelp {
			os.Exit(0)
		}
		os.Exit(2)
	}
}
