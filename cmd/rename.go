package cmd

import (
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/hidez8891/zip"
	"github.com/spf13/cobra"
	"gopkg.in/cheggaaa/pb.v1"
	"gopkg.in/go-playground/pool.v3"
)

func newRenameCmd(params *cmdParams) *cobra.Command {
	renamecmd := &rename{
		baseCmd: &baseCmd{params},
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
	cmd.Flags().UintVar(&renamecmd.jobs, "jobs", 1, "parallel job number")
	cmd.Flags().BoolVar(&renamecmd.showProgress, "show-progress", true, "show progress-bar")
	return cmd
}

type rename struct {
	*baseCmd
	from         string
	to           string
	jobs         uint
	showProgress bool
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
