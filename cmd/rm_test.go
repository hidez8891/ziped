package cmd

import (
	"bytes"
	"io"
	"io/ioutil"
	"os"
	"testing"

	"github.com/hidez8891/zip"
)

func TestRmExecuteOverwrite(t *testing.T) {
	tests := []struct {
		file  string
		args  []string
		files []string
	}{
		{
			file: "../testcase/test.zip",
			args: []string{
				"rm",
				"--overwrite",
				"--filter",
				"text1.txt",
			},
			files: []string{
				"dir/",
				"dir/text1.txt",
				"dir/text2.txt",
			},
		},
		{
			file: "../testcase/test.zip",
			args: []string{
				"rm",
				"--overwrite",
				"--filter",
				"*.txt",
			},
			files: []string{
				"dir/",
				"dir/text1.txt",
				"dir/text2.txt",
			},
		},
		{
			file: "../testcase/test.zip",
			args: []string{
				"rm",
				"--overwrite",
				"--regexp",
				"\\d.txt",
			},
			files: []string{
				"dir/",
			},
		},
	}

	for _, tt := range tests {
		stdout := new(bytes.Buffer)
		stderr := new(bytes.Buffer)

		tmpname, err := copyTempFile(tt.file)
		if err != nil {
			t.Fatal(err)
		}
		defer os.Remove(tmpname)

		cmd := newRootCmd(stdout, stderr)
		cmd.SetArgs(append(tt.args, tmpname))
		if err := cmd.Execute(); err != nil {
			t.Fatal(err)
		}

		if stderr.Len() != 0 {
			t.Fatalf("error output: %q", stderr.String())
		}
		if stdout.Len() != 0 {
			t.Fatalf("stdout output: %q", stdout.String())
		}

		zr, err := zip.OpenReader(tmpname)
		if err != nil {
			t.Fatal(err)
		}
		defer zr.Close()

		if len(zr.File) != len(tt.files) {
			t.Fatalf("update filename count=%d, want %d", len(zr.File), len(tt.files))
		}

		for i, zf := range zr.File {
			if zf.Name != tt.files[i] {
				t.Fatalf("update filename=%q, want %q", zf.Name, tt.files[i])
			}
		}
	}
}

func copyTempFile(path string) (string, error) {
	r, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer r.Close()

	tmp, err := ioutil.TempFile("", "test")
	if err != nil {
		return "", err
	}
	defer tmp.Close()

	if _, err := io.Copy(tmp, r); err != nil {
		return "", err
	}
	return tmp.Name(), nil
}
