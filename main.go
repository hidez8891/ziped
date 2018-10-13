package main

import (
	"os"

	"github.com/hidez8891/ziped/cmd"
)

func main() {
	cmd := cmd.NewCmd()

	cmd.Use = _Name
	cmd.Short = _Description
	cmd.Version = _Version

	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
