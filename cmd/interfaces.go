package cmd

import (
	"github.com/hidez8891/zip"
)

type Command interface {
	Run(*zip.Updater) error
}
