package cmd_list

import (
	"flag"
	"fmt"
	"io"
	"os"
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
	flags  *flag.FlagSet
	stdout io.Writer
}

var opts options

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

func NewCommand() *CmdList {
	flags := flag.NewFlagSet("", flag.ExitOnError)
	flags.Usage = func() {
		usage(flags.Output(), flags.Name())
	}

	return &CmdList{
		flags:  flags,
		stdout: os.Stdout,
	}
}

func (o *CmdList) Flags() *flag.FlagSet {
	return o.flags
}

func (o *CmdList) SetOutput(outputer io.Writer) {
	o.stdout = outputer
}

func (o *CmdList) Run(u *zip.Updater, metadata cmd.MetaData) error {
	if metadata.MultiInputMode {
		fmt.Fprintln(o.stdout, metadata.SrcPath)
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

		fmt.Fprintln(o.stdout, name)
	}

	if metadata.MultiInputMode && !metadata.IsLastFile {
		fmt.Fprintln(o.stdout)
	}
	return nil
}
