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
	"gopkg.in/cheggaaa/pb.v1"
	"gopkg.in/go-playground/pool.v3"
)

const usageTemplate = `
{{- if .Version}}
{{- "Name:"}}
  {{.CommandPath}} - {{.Short}} [{{.Version}}]
  {{- "\n"}}
{{- end}}
{{- "\n"}}

{{- "Usage:"}}
  {{.CommandPath}}
  {{- if .HasAvailableSubCommands}}
	{{- " [command]"}}
  {{- end }}
  {{- " [options] files..."}}
  {{- "\n"}}

{{- if .HasAvailableSubCommands}}
Commands:
  {{- "\n"}}
  {{- range .Commands}}
	{{- if (or .IsAvailableCommand (eq .Name "help"))}}
	  {{- "  "}}
      {{- rpad .Name .NamePadding }} {{.Short}}
      {{- "\n"}}
    {{- end}}
  {{- end}}
{{- end}}

{{- if .HasAvailableLocalFlags}}
Options:
  {{- "\n"}}
  {{- .LocalFlags.FlagUsages | trimTrailingWhitespaces}}
  {{- "\n"}}
{{- end}}

{{- if .HasAvailableInheritedFlags}}
Global Options:
  {{- "\n"}}
  {{- .InheritedFlags.FlagUsages | trimTrailingWhitespaces}}
  {{- "\n"}}
{{- end}}
`

func NewCmd() *cobra.Command {
	return newRootCmd(os.Stdout, os.Stderr)
}

func newRootCmd(stdout, stderr io.Writer) *cobra.Command {
	cmd := &cobra.Command{}
	cmd.SetOutput(stderr)

	params := &cmdParams{
		stdout: stdout,
		stderr: stderr,
	}

	cmd.AddCommand(newLsCmd(params))
	cmd.AddCommand(newRmCmd(params))
	cmd.AddCommand(newConvertCmd(params))
	cmd.AddCommand(newRenameCmd(params))

	cmd.PersistentFlags().StringVar(&params.pattern, "filter", "", "target filename pattern (support wildcard)")
	cmd.PersistentFlags().StringVar(&params.regexp, "regexp", "", "target filename pattern (support regexp)")
	cmd.PersistentFlags().BoolVar(&params.isOverwrite, "overwrite", false, "overwrite source file")
	cmd.PersistentFlags().StringVar(&params.outFilename, "out", "", "output file name")

	cmd.SetUsageTemplate(usageTemplate)
	cmd.SetHelpTemplate(usageTemplate)

	return cmd
}

func expandFilePath(paths []string) ([]string, error) {
	newpaths := make([]string, 0)
	for _, path := range paths {
		if strings.Contains(path, "*") {
			ps, err := doublestar.Glob(path)
			if err != nil {
				return nil, err
			}
			newpaths = append(newpaths, ps...)
		} else {
			newpaths = append(newpaths, path)
		}
	}
	return newpaths, nil
}

type cmdParams struct {
	pattern     string
	regexp      string
	isOverwrite bool
	outFilename string
	stdout      io.Writer
	stderr      io.Writer
}

func (o *cmdParams) generatePathFilter() (func(string) (bool, error), error) {
	filter := func(_ string) (bool, error) {
		return true, nil
	}

	if len(o.regexp) != 0 {
		reg, err := regexp.Compile(o.regexp)
		if err != nil {
			return nil, err
		}
		filter = func(s string) (bool, error) {
			return reg.Match([]byte(s)), nil
		}
	} else if len(o.pattern) != 0 {
		filter = func(s string) (bool, error) {
			return doublestar.Match(o.pattern, s)
		}
	}
	return filter, nil
}

func (o *cmdParams) validateOutputFlag(paths []string) (bool, error) {
	if !o.isOverwrite && len(o.outFilename) == 0 {
		return false, fmt.Errorf("output file name is required")
	}
	if !o.isOverwrite && len(paths) > 1 {
		return false, fmt.Errorf("for multiple files, only overwrite mode is supported")
	}
	return true, nil
}

type baseCmd struct {
	*cmdParams
}

func (o *baseCmd) openZipUpdater(filepath string) (*os.File, *zip.Updater, error) {
	st, err := os.Stat(filepath)
	if err != nil {
		return nil, nil, err
	}

	file, err := os.Open(filepath)
	if err != nil {
		return nil, nil, err
	}

	zu, err := zip.NewUpdater(file, st.Size())
	if err != nil {
		file.Close()
		return nil, nil, err
	}

	return file, zu, nil
}

func (o *baseCmd) openOutput(filepath string) (*os.File, error) {
	if o.isOverwrite {
		filename := path.Base(filepath)
		return ioutil.TempFile("", filename)
	}
	return os.OpenFile(o.outFilename, os.O_CREATE|os.O_WRONLY|os.O_EXCL, 0666)
}

func (o *baseCmd) overWriteFile(filepath string, data *os.File) error {
	file, err := os.OpenFile(filepath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)
	if err != nil {
		return err
	}
	defer file.Close()

	data.Seek(0, os.SEEK_SET)
	if _, err := io.Copy(file, data); err != nil {
		return err
	}

	return nil
}

func (o *baseCmd) editZipFile(filepath string, editor func(*zip.Updater) (bool, error)) error {
	file, zu, err := o.openZipUpdater(filepath)
	if err != nil {
		return err
	}
	defer close(file)
	defer close(zu)

	if ok, err := editor(zu); !ok {
		return err
	}

	outfile, err := o.openOutput(filepath)
	if err != nil {
		return err
	}
	defer close(outfile)

	if err := zu.SaveAs(outfile); err != nil {
		return err
	}
	zu.Close()
	file.Close()
	zu = nil
	file = nil

	if o.isOverwrite {
		if err := o.overWriteFile(filepath, outfile); err != nil {
			return err
		}

		outfile.Close()
		os.Remove(outfile.Name())
		outfile = nil
	}

	return nil
}

type toolParallelCmd struct {
	writer       io.Writer
	jobs         uint
	showProgress bool
}

func (o *toolParallelCmd) setFlags(cmd *cobra.Command) {
	cmd.Flags().UintVar(&o.jobs, "jobs", 1, "parallel job number")
	cmd.Flags().BoolVar(&o.showProgress, "show-progress", true, "show progress-bar")
}

func (o *toolParallelCmd) flagValidate() error {
	if o.jobs < 1 {
		o.jobs = 1
	}
	return nil
}

func (o *toolParallelCmd) execute(paths []string, executer func(string) error) []error {
	progress := pb.New(len(paths))
	progress.Output = o.writer
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
				err := executer(filepath)
				return nil, err
			})
		}
		worker.QueueComplete()
	}()

	errors := make([]error, 0)
	for result := range worker.Results() {
		if err := result.Error(); err != nil {
			worker.Cancel()
			errors = append(errors, err)
		}
	}
	progress.Finish()

	if len(errors) == 0 {
		return nil
	}
	return errors
}

func close(closer io.Closer) error {
	if closer != nil {
		return closer.Close()
	}
	return nil
}
