package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"ziped/cmd"
	cmd_list "ziped/cmd/list"
	"ziped/pkg/slices"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/hidez8891/zip"
)

type options struct {
}

var (
	opts options

	cmds = map[string]cmd.Command{
		"ls": cmd_list.NewCommand(),
	}
)

func usage(writer io.Writer, cmd, version string) {
	tmpl := heredoc.Doc(`
		Name:
			{:CMD:} - [{:VERSION:}]

		Usage:
			(1) single command
				{:CMD:} [OPTIONS] COMMAND <FILES...>

			(2) multiple command
				{:CMD:} [OPTIONS] <FILES...>
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

	xargs := slices.Split(os.Args[1:], "--")
	multiCommandMode := len(xargs) != 1

	if err := flags.Parse(xargs[0]); err != nil {
		if err == flag.ErrHelp {
			os.Exit(0)
		}
		os.Exit(1)
	}

	files := flags.Args()
	if !multiCommandMode {
		if len(files) < 2 {
			fmt.Fprintln(os.Stderr, "No file specified")
			os.Exit(1)
		}

		cmd, files := files[0], files[1:]
		singleCommandRun(cmd, files)
	} else {
		if len(files) < 1 {
			fmt.Fprintln(os.Stderr, "No file specified")
			os.Exit(1)
		}

		multiCommandRun(xargs[1:], files)
	}
}

func singleCommandRun(cmd string, files []string) {
	runner, ok := cmds[cmd]
	if !ok {
		fmt.Fprintf(os.Stderr, "Undefined command: %s\n", cmd)
		os.Exit(1)
	}

	st, err := os.Stat(files[0])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	f, err := os.Open(files[0])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer f.Close()

	u, err := zip.NewUpdater(f, st.Size())
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer u.Close()

	if err := runner.Run(u); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func multiCommandRun(xargs [][]string, files []string) {
	// TODO
}
