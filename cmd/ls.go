package cmd

import (
	"fmt"

	"github.com/hidez8891/zip"
	"github.com/spf13/cobra"
)

func newLsCmd(params *cmdParams) *cobra.Command {
	lscmd := &ls{
		baseCmd: &baseCmd{params},
	}

	var cmd = &cobra.Command{
		Use:   "ls [filepath...]",
		Short: "Show file list",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			lscmd.run(cmd, args)
		},
	}

	return cmd
}

type ls struct {
	*baseCmd
}

func (o *ls) run(cmd *cobra.Command, args []string) {
	paths, err := expandFilePath(args)
	if err != nil {
		fmt.Fprintln(o.stderr, err)
		return
	}

	if len(paths) == 1 {
		files, err := o.execute(paths[0])
		if err != nil {
			fmt.Fprintln(o.stderr, err)
			return
		}
		for _, file := range files {
			fmt.Fprintln(o.stdout, file)
		}
	} else {
		for i, filepath := range paths {
			files, err := o.execute(filepath)
			if err != nil {
				fmt.Fprintln(o.stderr, err)
				return
			}
			if len(files) == 0 {
				continue
			}

			fmt.Fprintf(o.stdout, "%s:\n", filepath)
			for _, file := range files {
				fmt.Fprintln(o.stdout, file)
			}
			if i != len(paths)-1 {
				fmt.Fprintln(o.stdout)
			}
		}
	}
}

func (o *ls) execute(filepath string) ([]string, error) {
	result := make([]string, 0)

	zr, err := zip.OpenReader(filepath)
	if err != nil {
		return nil, err
	}
	defer zr.Close()

	filter, err := o.generatePathFilter()
	if err != nil {
		return nil, err
	}

	for _, zf := range zr.File {
		ok, err := filter(zf.Name)
		if err != nil {
			return nil, err
		}
		if ok {
			result = append(result, zf.Name)
		}
	}
	return result, nil
}
