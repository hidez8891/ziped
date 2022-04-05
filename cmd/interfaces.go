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

type Command interface {
	Flags() *flag.FlagSet
	SetOutput(io.Writer)
	Run(*zip.Updater, MetaData) error
}
