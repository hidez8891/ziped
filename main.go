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

const (
	NAME    = "ziped"
	VERSION = "0.0.0-dev"
)

type options struct {
	showVersion bool
}

var (
	flags = flag.NewFlagSet(NAME, flag.ExitOnError)

	opts options

	cmds = map[string]cmd.Command{
		"ls": cmd_list.NewCommand(),
	}
)

func usage(writer io.Writer) {
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
			    --version   version for {:CMD:}
	`)

	tmpl = strings.ReplaceAll(tmpl, "\t", "    ")
	tmpl = strings.ReplaceAll(tmpl, "{:CMD:}", NAME)
	tmpl = strings.ReplaceAll(tmpl, "{:VERSION:}", VERSION)
	fmt.Fprintln(writer, tmpl)
}

func showVersion(writer io.Writer) {
	tmpl := fmt.Sprintf("%s %s", NAME, VERSION)
	fmt.Fprintln(writer, tmpl)
}

func setupFlags() {
	flags.Usage = func() {
		usage(flags.Output())
	}

	flags.BoolVar(&opts.showVersion, "version", false, "")
}

func setupSubcommands() {
	for tag, subcmd := range cmds {
		subcmdName := fmt.Sprintf("%s %s", NAME, tag)
		subcmd.Flags().Init(subcmdName, flag.ContinueOnError)
	}
}

func main() {
	setupFlags()
	setupSubcommands()

	xargs := slices.Split(os.Args[1:], "--")
	multiCommandMode := len(xargs) != 1

	if err := flags.Parse(xargs[0]); err != nil {
		if err == flag.ErrHelp {
			os.Exit(0)
		}
		os.Exit(1)
	}

	if opts.showVersion {
		showVersion(flags.Output())
		os.Exit(0)
	}

	var subcmds []cmd.Command
	var files []string

	if !multiCommandMode {
		args := flags.Args()
		subcmd, err := parseSingleCommandMode(args[0], args[1:])
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(1)
		}

		subcmds = []cmd.Command{subcmd}
		files = subcmd.Flags().Args()
	} else {
		var err error
		subcmds, err = parseMultiCommandMode(xargs[1:])
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(1)
		}

		files = flags.Args()
	}

	if len(files) == 0 {
		fmt.Fprintln(os.Stderr, "No file specified")
		os.Exit(1)
	}

	for _, file := range files {
		if err := runSubcommands(subcmds, file); err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(1)
		}
	}
}

func parseSingleCommandMode(subcmd string, args []string) (cmd.Command, error) {
	exec, ok := cmds[subcmd]
	if !ok {
		return nil, fmt.Errorf("Undefined command: %s", subcmd)
	}

	if err := exec.Flags().Parse(args); err != nil {
		return nil, err
	}

	return exec, nil
}

func parseMultiCommandMode(xargs [][]string) ([]cmd.Command, error) {
	subcmds := make([]cmd.Command, 0)

	for _, args := range xargs {
		subcmd, args := args[0], args[1:]

		exec, ok := cmds[subcmd]
		if !ok {
			return nil, fmt.Errorf("Undefined command: %s", subcmd)
		}

		if err := exec.Flags().Parse(args); err != nil {
			return nil, err
		}
		if exec.Flags().NArg() != 0 {
			return nil, fmt.Errorf("Undefined option: %s", exec.Flags().Arg(0))
		}

		subcmds = append(subcmds, exec)
	}

	return subcmds, nil
}

func runSubcommands(subcmds []cmd.Command, file string) error {
	st, err := os.Stat(file)
	if err != nil {
		return err
	}

	f, err := os.Open(file)
	if err != nil {
		return err
	}
	defer f.Close()

	u, err := zip.NewUpdater(f, st.Size())
	if err != nil {
		return err
	}
	defer u.Close()

	for _, subcmd := range subcmds {
		if err := subcmd.Run(u); err != nil {
			return err
		}
	}

	return nil
}
