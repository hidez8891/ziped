package cmd

import (
	"flag"
	"io"

	"github.com/hidez8891/zip"
)

type MetaData struct {
	SrcPath        string
	MultiInputMode bool
	IsLastFile     bool
}

type CommandIO struct {
	In  io.Reader
	Out io.Writer
	Err io.Writer
}

type Command interface {
	Flags() *flag.FlagSet
	Run(*zip.Updater, MetaData) error
	SetCmdIO(CommandIO)
	SetName(string)
}
