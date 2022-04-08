package cmd_list

import (
	"flag"
	"fmt"
	"io"
	"strings"

	"ziped/cmd"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/hidez8891/zip"
	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/transform"
)

type options struct {
}

type CmdList struct {
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
	`)

	tmpl = strings.ReplaceAll(tmpl, "\t", "    ")
	tmpl = strings.ReplaceAll(tmpl, "{:CMD:}", cmd)
	fmt.Fprintln(writer, tmpl)
}

func NewCommand(name string, cmdIO cmd.CommandIO) *CmdList {
	var opts options

	flags := flag.NewFlagSet(name, flag.ExitOnError)
	flags.Usage = func() {
		usage(flags.Output(), flags.Name())
	}
	flags.SetOutput(cmdIO.Err)

	return &CmdList{
		CommandIO: cmdIO,
		flags:     flags,
		opts:      opts,
	}
}

func (o *CmdList) Flags() *flag.FlagSet {
	return o.flags
}

func (o *CmdList) SetCmdIO(cmdio cmd.CommandIO) {
	o.CommandIO = cmdio
	o.flags.SetOutput(cmdio.Err)
}

func (o *CmdList) SetName(name string) {
	o.flags.Init(name, o.flags.ErrorHandling())
}

func (o *CmdList) Run(u *zip.Updater, metadata cmd.MetaData) (cmd.ResultState, error) {
	if metadata.MultiInputMode {
		fmt.Fprintln(o.Out, metadata.SrcPath)
	}

	for _, zf := range u.Files() {
		name := zf.Name

		if zf.NonUTF8 {
			decoder := japanese.ShiftJIS.NewDecoder()
			decodeName, _, err := transform.String(decoder, name)
			if err == nil {
				name = decodeName
			}
		}

		fmt.Fprintln(o.Out, name)
	}

	if metadata.MultiInputMode && !metadata.IsLastFile {
		fmt.Fprintln(o.Out)
	}
	return cmd.ResultNotUpdated, nil
}
