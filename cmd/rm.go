package cmd

import (
	"fmt"

	"github.com/hidez8891/zip"
	"github.com/spf13/cobra"
)

func newRmCmd(params *cmdParams) *cobra.Command {
	rmcmd := &rm{
		baseCmd: &baseCmd{params},
		pexe:    &toolParallelCmd{writer: params.stdout},
	}

	var cmd = &cobra.Command{
		Use:   "rm [filepath...]",
		Short: "Remove file",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			rmcmd.run(cmd, args)
		},
	}

	rmcmd.pexe.setFlags(cmd)
	return cmd
}

type rm struct {
	*baseCmd
	pexe *toolParallelCmd
}

func (o *rm) run(cmd *cobra.Command, args []string) {
	paths, err := expandFilePath(args)
	if err != nil {
		fmt.Fprintln(o.stderr, err)
		return
	}

	if ok, err := o.validateOutputFlag(paths); !ok {
		fmt.Fprintln(o.stderr, err.Error())
		return
	}
	if err := o.pexe.flagValidate(); err != nil {
		fmt.Fprintln(o.stderr, err.Error())
		return
	}

	errors := o.pexe.execute(paths, func(filepath string) error {
		return o.execute(filepath)
	})

	if errors != nil {
		for _, err := range errors {
			fmt.Fprintln(o.stderr, err.Error())
		}
	}
}

func (o *rm) execute(filepath string) error {
	return o.editZipFile(filepath, func(zu *zip.Updater) (bool, error) {
		filter, err := o.generatePathFilter()
		if err != nil {
			return false, err
		}

		isModified := false
		for _, header := range zu.Files() {
			ok, err := filter(header.Name)
			if err != nil {
				return false, err
			}
			if !ok {
				continue
			}

			isModified = true
			if err := zu.Remove(header.Name); err != nil {
				return false, err
			}
		}

		return isModified, nil
	})
}
