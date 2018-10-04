package cmd

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	path "path/filepath"
	"regexp"
	"strings"

	"github.com/bmatcuk/doublestar"
	"github.com/hidez8891/zip"
	"github.com/spf13/cobra"
)

func newRenameCmd(stdout, stderr io.Writer) *cobra.Command {
	renamecmd := &rename{
		stdout: stdout,
		stderr: stderr,
	}

	var cmd = &cobra.Command{
		Use:   "rename [filepath...]",
		Short: "Rename file contents",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			renamecmd.run(cmd, args)
		},
	}

	cmd.Flags().StringVar(&renamecmd.pattern, "filter", "", "target filename pattern")
	cmd.Flags().StringVar(&renamecmd.regexp, "regexp", "", "target filename pattern")
	cmd.Flags().BoolVar(&renamecmd.isOverwrite, "overwrite", false, "overwrite source file")
	cmd.Flags().StringVar(&renamecmd.outFilename, "out", "", "output file name")
	cmd.Flags().StringVar(&renamecmd.from, "from", "", "text before replacement")
	cmd.Flags().StringVar(&renamecmd.to, "to", "", "text after replacement")
	return cmd
}

type rename struct {
	stdout      io.Writer
	stderr      io.Writer
	pattern     string
	regexp      string
	isOverwrite bool
	outFilename string
	from        string
	to          string
}

func (o *rename) run(cmd *cobra.Command, args []string) {
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

func (o *rename) execute(filepath string) error {
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

		oldname := header.Name
		newname := strings.Replace(oldname, o.from, o.to, 1)

		if oldname == newname {
			continue
		}
		isModified = true

		if err := zu.Rename(oldname, newname); err != nil {
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
