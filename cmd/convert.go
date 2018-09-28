package cmd

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	path "path/filepath"
	"regexp"

	"github.com/bmatcuk/doublestar"
	"github.com/hidez8891/zip"
	"github.com/mattn/go-shellwords"
	"github.com/spf13/cobra"
)

func newConvertCmd(stdout, stderr io.Writer) *cobra.Command {
	convcmd := &convert{
		stdout: stdout,
		stderr: stderr,
	}

	var cmd = &cobra.Command{
		Use:   "convert [filepath...]",
		Short: "Convert file contents",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			convcmd.run(cmd, args)
		},
	}

	cmd.Flags().StringVar(&convcmd.pattern, "filter", "", "remove filename pattern")
	cmd.Flags().StringVar(&convcmd.regexp, "regexp", "", "remove filename pattern")
	cmd.Flags().BoolVar(&convcmd.isOverwrite, "overwrite", false, "overwrite source file")
	cmd.Flags().StringVar(&convcmd.outFilename, "out", "", "output file name")
	cmd.Flags().StringVar(&convcmd.command, "cmd", "", "convert command")
	return cmd
}

type convert struct {
	stdout      io.Writer
	stderr      io.Writer
	pattern     string
	regexp      string
	isOverwrite bool
	outFilename string
	command     string
}

func (o *convert) run(cmd *cobra.Command, args []string) {
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
	if len(o.command) == 0 {
		fmt.Fprintln(o.stderr, "execute command is required")
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

func (o *convert) execute(filepath string) error {
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

	var filter func(s string) (bool, error)
	if len(o.regexp) != 0 {
		reg, err := regexp.Compile(o.regexp)
		if err != nil {
			return err
		}
		filter = func(s string) (bool, error) {
			return reg.Match([]byte(s)), nil
		}
	} else if len(o.pattern) != 0 {
		filter = func(s string) (bool, error) {
			return doublestar.Match(o.pattern, s)
		}
	} else {
		filter = func(s string) (bool, error) {
			return true, nil
		}
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
		if err := o.executeShell(zu, header.Name); err != nil {
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

func (o *convert) executeShell(zu *zip.Updater, name string) error {
	args, err := shellwords.Parse(o.command)
	if err != nil {
		return err
	}
	sh := exec.Command(args[0], args[1:]...)

	stdin, err := sh.StdinPipe()
	if err != nil {
		return err
	}
	defer func() {
		if stdin != nil {
			stdin.Close()
		}
	}()

	r, err := zu.Open(name)
	if err != nil {
		return err
	}
	defer func() {
		if r != nil {
			r.Close()
		}
	}()

	if _, err := io.Copy(stdin, r); err != nil {
		return err
	}
	stdin.Close()
	stdin = nil
	r.Close()
	r = nil

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
