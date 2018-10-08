package cmd

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	path "path/filepath"

	"github.com/hidez8891/zip"
	"github.com/spf13/cobra"
)

func newRmCmd(params *cmdParams) *cobra.Command {
	rmcmd := &rm{
		cmdParams: params,
	}

	var cmd = &cobra.Command{
		Use:   "rm [filepath...]",
		Short: "Remove file",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			rmcmd.run(cmd, args)
		},
	}

	return cmd
}

type rm struct {
	*cmdParams
}

func (o *rm) run(cmd *cobra.Command, args []string) {
	paths, err := expandFilePath(args)
	if err != nil {
		fmt.Fprintln(o.stderr, err)
		return
	}

	if !o.isOverwrite && len(o.outFilename) == 0 {
		fmt.Fprintln(o.stderr, "output file name is required")
		return
	}
	if !o.isOverwrite && len(paths) > 1 {
		fmt.Fprintln(o.stderr, "for multiple files, only overwrite mode is supported")
		return
	}

	for _, filepath := range paths {
		err := o.execute(filepath)
		if err != nil {
			fmt.Fprintln(o.stderr, err)
			return
		}
	}
}

func (o *rm) execute(filepath string) error {
	file, err := os.Open(filepath)
	if err != nil {
		return err
	}
	defer func() {
		if file != nil {
			file.Close()
			file = nil
		}
	}()

	st, err := os.Stat(filepath)
	if err != nil {
		return err
	}

	zu, err := zip.NewUpdater(file, st.Size())
	if err != nil {
		return err
	}
	defer func() {
		if zu != nil {
			zu.Close()
			zu = nil
		}
	}()

	filter, err := o.generatePathFilter()
	if err != nil {
		return err
	}

	var isModified = false
	for _, header := range zu.Files() {
		ok, err := filter(header.Name)
		if err != nil {
			return err
		}
		if !ok {
			continue
		}

		isModified = true
		if err := zu.Remove(header.Name); err != nil {
			return err
		}
	}
	if !isModified {
		return nil
	}

	var outfile *os.File
	if o.isOverwrite {
		filename := path.Base(filepath)
		outfile, err = ioutil.TempFile("", filename)
		if err != nil {
			return err
		}
	} else {
		outfile, err = os.OpenFile(o.outFilename, os.O_CREATE|os.O_WRONLY|os.O_EXCL, 0666)
		if err != nil {
			return err
		}
	}
	defer func() {
		if outfile != nil {
			outfile.Close()
		}
	}()

	if err := zu.SaveAs(outfile); err != nil {
		return err
	}

	if o.isOverwrite {
		// overwrite file
		zu.Close()
		zu = nil
		file.Close()
		file = nil

		file, err = os.OpenFile(filepath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)
		if err != nil {
			return err
		}
		outfile.Seek(0, os.SEEK_SET)
		if _, err := io.Copy(file, outfile); err != nil {
			return err
		}

		outfile.Close()
		os.Remove(outfile.Name())
		outfile = nil

		file.Close()
		file = nil
	}

	return nil
}
