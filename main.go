package main

import (
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"ziped/cmd"
	cmd_list "ziped/cmd/list"
	cmd_rename "ziped/cmd/rename"
	"ziped/pkg/slices"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/hidez8891/zip"
)

const (
	NAME    = "ziped"
	VERSION = "0.0.0-dev"
)

type options struct {
	outputPath  string
	overwrite   bool
	showVersion bool
}

var (
	cmds  map[string]cmd.Command
	flags *flag.FlagSet
	opts  options
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
			-h, --help         help for {:CMD:}
				--output PATH  save to PATH file (only use single command mode)
			    --overwrite    overwrite to source file
			    --version      version for {:CMD:}
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
	flags = flag.NewFlagSet(NAME, flag.ExitOnError)
	flags.Usage = func() {
		usage(flags.Output())
	}

	flags.StringVar(&opts.outputPath, "output", "", "")
	flags.BoolVar(&opts.overwrite, "overwrite", false, "")
	flags.BoolVar(&opts.showVersion, "version", false, "")
}

func setupSubcommands() {
	cmdIO := cmd.CommandIO{
		In:  os.Stdin,
		Out: os.Stdout,
		Err: os.Stderr,
	}

	cmds = map[string]cmd.Command{
		"ls":     cmd_list.NewCommand("", cmdIO),
		"rename": cmd_rename.NewCommand("", cmdIO),
	}

	for tag, subcmd := range cmds {
		subcmdName := fmt.Sprintf("%s %s", NAME, tag)
		subcmd.SetName(subcmdName)
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

	if err := runSubcommands(subcmds, files); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
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

func runSubcommands(subcmds []cmd.Command, files []string) error {
	for i, file := range files {
		metadata := cmd.MetaData{
			SrcPath:        file,
			MultiInputMode: len(files) != 1,
			IsLastFile:     i == len(files)-1,
		}

		outpath, err := runSubcommandsImpl(subcmds, file, metadata)
		if err != nil {
			return err
		}

		if outpath != "" && opts.overwrite {
			outdir := filepath.Dir(outpath)

			swapfile, err := os.CreateTemp(outdir, "tmpswap_")
			if err != nil {
				return err
			}
			swapname := swapfile.Name()
			swapfile.Close()

			if err := os.Rename(file, swapname); err != nil {
				return err
			}
			if err := os.Rename(outpath, file); err != nil {
				return err
			}
			if err := os.Remove(swapname); err != nil {
				return err
			}
		}
	}

	return nil
}

func runSubcommandsImpl(subcmds []cmd.Command, file string, metadata cmd.MetaData) (string, error) {
	st, err := os.Stat(file)
	if err != nil {
		return "", err
	}

	f, err := os.Open(file)
	if err != nil {
		return "", err
	}
	defer f.Close()

	u, err := zip.NewUpdater(f, st.Size())
	if err != nil {
		return "", err
	}
	defer u.Close()

	state := cmd.ResultNotUpdated

	for _, subcmd := range subcmds {
		var err error

		state, err = subcmd.Run(u, metadata)
		if err != nil {
			return "", err
		}
	}

	if state != cmd.ResultUpdated {
		return file, nil
	}

	var tmpfile *os.File
	if len(opts.outputPath) > 0 && len(subcmds) == 1 {
		var err error
		if tmpfile, err = os.Create(opts.outputPath); err != nil {
			return "", err
		}
	} else if opts.overwrite {
		abspath, err := filepath.Abs(file)
		if err != nil {
			return "", err
		}
		basedir := filepath.Dir(abspath)

		tmpfile, err = ioutil.TempFile(basedir, "tmp")
		if err != nil {
			return "", err
		}
	} else {
		return "", fmt.Errorf("output path or overwrite permissions not set")
	}
	defer tmpfile.Close()

	if err := u.SaveAs(tmpfile); err != nil {
		return "", err
	}
	return tmpfile.Name(), nil
}
