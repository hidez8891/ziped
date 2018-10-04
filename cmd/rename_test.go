package cmd

import (
	"bytes"
	"os"
	"testing"

	"github.com/hidez8891/zip"
)

func TestRenameExecuteOverwrite(t *testing.T) {
	tests := []struct {
		file  string
		args  []string
		files []string
	}{
		{
			file: "../testcase/test.zip",
			args: []string{
				"rename",
				"--overwrite",
				"--from",
				".txt",
				"--to",
				".md",
			},
			files: []string{
				"dir/",
				"dir/text1.md",
				"dir/text2.md",
				"text1.md",
			},
		},
		{
			file: "../testcase/test.zip",
			args: []string{
				"rename",
				"--overwrite",
				"--from",
				"text",
				"--to",
				"texttext",
			},
			files: []string{
				"dir/",
				"dir/texttext1.txt",
				"dir/texttext2.txt",
				"texttext1.txt",
			},
		},
		{
			file: "../testcase/test.zip",
			args: []string{
				"rename",
				"--overwrite",
				"--filter",
				"*.txt",
				"--from",
				".txt",
				"--to",
				".md",
			},
			files: []string{
				"dir/",
				"dir/text1.txt",
				"dir/text2.txt",
				"text1.md",
			},
		},
		{
			file: "../testcase/test.zip",
			args: []string{
				"rename",
				"--overwrite",
				"--regexp",
				"dir/.+\\.txt",
				"--from",
				".txt",
				"--to",
				".md",
			},
			files: []string{
				"dir/",
				"dir/text1.md",
				"dir/text2.md",
				"text1.txt",
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
