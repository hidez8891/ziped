package cmd

import (
	"fmt"

	"github.com/hidez8891/zip"
	"github.com/spf13/cobra"
	"gopkg.in/go-playground/pool.v3"
)

func newRmCmd(params *cmdParams) *cobra.Command {
	rmcmd := &rm{
		baseCmd: &baseCmd{params},
	}

	var cmd = &cobra.Command{
		Use:   "rm [filepath...]",
		Short: "Remove file",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			rmcmd.run(cmd, args)
		},
	}

	cmd.Flags().UintVar(&rmcmd.jobs, "jobs", 1, "parallel job number")
	return cmd
}

type rm struct {
	*baseCmd
	jobs uint
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
	if o.jobs < 1 {
		o.jobs = 1
	}

	threads := pool.NewLimited(o.jobs)
	defer threads.Close()

	worker := threads.Batch()
	go func() {
		for _, filepath := range paths {
			filepath := filepath

			worker.Queue(func(wu pool.WorkUnit) (interface{}, error) {
				if wu.IsCancelled() {
					return nil, nil
				}
				err := o.execute(filepath)
				return nil, err
			})
		}
		worker.QueueComplete()
	}()

	for result := range worker.Results() {
		if err := result.Error(); err != nil {
			worker.Cancel()
			fmt.Fprintln(o.stderr, err)
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
