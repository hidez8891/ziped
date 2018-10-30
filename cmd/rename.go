package cmd

import (
	"fmt"
	"strings"

	"github.com/hidez8891/zip"
	"github.com/spf13/cobra"
)

func newRenameCmd(params *cmdParams) *cobra.Command {
	renamecmd := &rename{
		baseCmd: &baseCmd{params},
		pexe:    &toolParallelCmd{writer: params.stdout},
	}

	var cmd = &cobra.Command{
		Use:   "rename [filepath...]",
		Short: "Rename file contents",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			renamecmd.run(cmd, args)
		},
	}

	cmd.Flags().StringVar(&renamecmd.from, "from", "", "text before replacement")
	cmd.Flags().StringVar(&renamecmd.to, "to", "", "text after replacement")
	renamecmd.pexe.setFlags(cmd)
	return cmd
}

type rename struct {
	*baseCmd
	pexe *toolParallelCmd
	from string
	to   string
}

func (o *rename) run(cmd *cobra.Command, args []string) {
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

func (o *rename) execute(filepath string) error {
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

			oldname := header.Name
			newname := strings.Replace(oldname, o.from, o.to, 1)

			if oldname == newname {
				continue
			}
			isModified = true

			if err := zu.Rename(oldname, newname); err != nil {
				return false, err
			}
		}

		return isModified, nil
	})
}
