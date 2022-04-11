package cmd_rename

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
	from string
	to   string
}

type CmdRename struct {
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
			    --from  Name before replacement
			    --to    Name after  replacement
	`)

	tmpl = strings.ReplaceAll(tmpl, "\t", "    ")
	tmpl = strings.ReplaceAll(tmpl, "{:CMD:}", cmd)
	fmt.Fprintln(writer, tmpl)
}

func NewCommand(name string, cmdIO cmd.CommandIO) *CmdRename {
	var opts options

	flags := flag.NewFlagSet(name, flag.ExitOnError)
	flags.Usage = func() {
		usage(flags.Output(), flags.Name())
	}
	flags.SetOutput(cmdIO.Err)

	c := &CmdRename{
		CommandIO: cmdIO,
		flags:     flags,
		opts:      opts,
	}

	flags.StringVar(&c.opts.from, "from", "", "")
	flags.StringVar(&c.opts.to, "to", "", "")

	return c
}

func (o *CmdRename) Flags() *flag.FlagSet {
	return o.flags
}

func (o *CmdRename) SetCmdIO(cmdio cmd.CommandIO) {
	o.CommandIO = cmdio
	o.flags.SetOutput(cmdio.Err)
}

func (o *CmdRename) SetName(name string) {
	o.flags.Init(name, o.flags.ErrorHandling())
}

func (o *CmdRename) Run(u *zip.Updater, metadata cmd.MetaData) (cmd.ResultState, error) {
	modified := false

	for _, zf := range u.Files() {
		oldname := zf.Name
		newname := strings.Replace(oldname, o.opts.from, o.opts.to, 1)

		if oldname == newname {
			continue
		}
		modified = true

		if err := u.Rename(oldname, newname); err != nil {
			return cmd.ResultError, err
		}
	}

	if modified {
		return cmd.ResultUpdated, nil
	}
	return cmd.ResultNotUpdated, nil
}
