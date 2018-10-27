package cmd

import (
	"fmt"
	"io"
	"os"
	"os/exec"

	"github.com/hidez8891/zip"
	"github.com/mattn/go-shellwords"
	"github.com/spf13/cobra"
	"gopkg.in/go-playground/pool.v3"
)

func newConvertCmd(params *cmdParams) *cobra.Command {
	convcmd := &convert{
		baseCmd: &baseCmd{params},
	}

	var cmd = &cobra.Command{
		Use:   "convert [filepath...]",
		Short: "Convert file contents",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			convcmd.run(cmd, args)
		},
	}

	cmd.Flags().StringVar(&convcmd.command, "cmd", "", "convert command")
	cmd.Flags().UintVar(&convcmd.jobs, "jobs", 1, "parallel job number")
	return cmd
}

type convert struct {
	*baseCmd
	command string
	jobs    uint
}

func (o *convert) run(cmd *cobra.Command, args []string) {
	paths, err := expandFilePath(args)
	if err != nil {
		fmt.Fprintln(o.stderr, err)
		return
	}

	if ok, err := o.validateOutputFlag(paths); !ok {
		fmt.Fprintln(o.stderr, err.Error())
		return
	}
	if len(o.command) == 0 {
		fmt.Fprintln(o.stderr, "execute command is required")
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

func (o *convert) execute(filepath string) error {
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
			if err := o.executeShell(zu, header.Name); err != nil {
				return false, err
			}
		}

		return isModified, nil
	})
}

func (o *convert) executeShell(zu *zip.Updater, name string) error {
	args, err := shellwords.Parse(o.command)
	if err != nil {
		return err
	}

	sh := exec.Command(args[0], args[1:]...)
	sh.Stderr = os.Stderr

	stdin, err := sh.StdinPipe()
	if err != nil {
		return err
	}
	defer close(stdin)

	r, err := zu.Open(name)
	if err != nil {
		return err
	}
	defer close(r)

	go func() {
		defer func() {
			stdin.Close()
			r.Close()
			stdin = nil
			r = nil
		}()

		io.Copy(stdin, r)
	}()

	out, err := sh.Output()
	if err != nil {
		return err
	}

	w, err := zu.Update(name)
	if err != nil {
		return err
	}
	defer w.Close()

	if _, err := w.Write(out); err != nil {
		return err
	}

	return nil
}
