package cmd

import (
	"flag"

	"github.com/hidez8891/zip"
)

type Command interface {
	Flags() *flag.FlagSet
	Run(*zip.Updater) error
}
