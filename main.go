package main

import (
	"os"

	"github.com/hidez8891/ziped/cmd"
)

func main() {
	cmd := cmd.NewCmd()

	cmd.Use = name
	cmd.Short = description
	cmd.Version = version

	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
