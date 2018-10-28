package cmd

import (
	"fmt"
	"io/ioutil"

	"github.com/hidez8891/zip"
	"github.com/spf13/cobra"
	"gopkg.in/cheggaaa/pb.v1"
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
	cmd.Flags().BoolVar(&rmcmd.showProgress, "show-progress", true, "show progress-bar")
	return cmd
}

type rm struct {
	*baseCmd
	jobs         uint
	showProgress bool
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

	progress := pb.New(len(paths))
	progress.Output = o.stderr
	if !o.showProgress {
		progress.Output = ioutil.Discard
	}
	progress.Start()

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
				progress.Increment()
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
	progress.Finish()
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
