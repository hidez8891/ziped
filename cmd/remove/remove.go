package cmd_remove

import (
	"flag"
	"fmt"
	"io"
	"strings"

	"ziped/cmd"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/hidez8891/zip"
)

type options struct {
	name string
}

type CmdRemove struct {
	cmd.CommandIO
	flags *flag.FlagSet
	opts  options
}

func usage(writer io.Writer, cmd string) {
	tmpl := heredoc.Doc(`
		Usage:
			{:CMD:} [OPTIONS] <FILES...>

		Options:
			-h, --help  Show help information
			    --name  Target file name
	`)

	tmpl = strings.ReplaceAll(tmpl, "\t", "    ")
	tmpl = strings.ReplaceAll(tmpl, "{:CMD:}", cmd)
	fmt.Fprintln(writer, tmpl)
}

func NewCommand(name string, cmdIO cmd.CommandIO) *CmdRemove {
	var opts options

	flags := flag.NewFlagSet(name, flag.ExitOnError)
	flags.Usage = func() {
		usage(flags.Output(), flags.Name())
	}
	flags.SetOutput(cmdIO.Err)

	c := &CmdRemove{
		CommandIO: cmdIO,
		flags:     flags,
		opts:      opts,
	}

	flags.StringVar(&c.opts.name, "name", "", "")

	return c
}

func (o *CmdRemove) Flags() *flag.FlagSet {
	return o.flags
}

func (o *CmdRemove) SetCmdIO(cmdio cmd.CommandIO) {
	o.CommandIO = cmdio
	o.flags.SetOutput(cmdio.Err)
}

func (o *CmdRemove) SetName(name string) {
	o.flags.Init(name, o.flags.ErrorHandling())
}

func (o *CmdRemove) Run(u *zip.Updater, metadata cmd.MetaData) (cmd.ResultState, error) {
	modified := false

	if o.opts.name == "" {
		return cmd.ResultError, fmt.Errorf("required parameter 'name' not set")
	}

	for _, zf := range u.Files() {
		if !strings.Contains(zf.Name, o.opts.name) {
			continue
		}
		modified = true

		if err := u.Remove(zf.Name); err != nil {
			return cmd.ResultError, err
		}
	}

	if modified {
		return cmd.ResultUpdated, nil
	}
	return cmd.ResultNotUpdated, nil
}
