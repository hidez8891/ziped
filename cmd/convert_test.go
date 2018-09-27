package cmd

import (
	"bytes"
	"io"
	"io/ioutil"
	"os"
	"testing"

	"github.com/hidez8891/zip"
)

func TestConvertExecuteOverwrite(t *testing.T) {
	tests := []struct {
		file  string
		args  []string
		contents map[string]string
	}{
		{
			file: "../testcase/test2.zip",
			args: []string{
				"convert",
				"--overwrite",
				"--filter",
				"text1.txt",
				"--cmd",
				"sort"
			},
			contents: map[string]string{
				"text1.txt": "hello1\nhello2\nhello3",
				"dir/text1.txt": "test 3\ntest 2\ntest 1",
				"dir/text2.txt": "test 2",
			},
		},
		{
			file: "../testcase/test2.zip",
			args: []string{
				"convert",
				"--overwrite",
				"--filter",
				"*.txt",
				"--cmd",
				"sort"
			},
			contents: map[string]string{
				"text1.txt": "hello1\nhello2\nhello3",
				"dir/text1.txt": "test 1\ntest 2\ntest 3",
				"dir/text2.txt": "test 2",
			},
		},
		{
			file: "../testcase/test.zip",
			args: []string{
				"convert",
				"--overwrite",
				"--regexp",
				"\\d.txt",
			},
			contents: map[string]string{
				"text1.txt": "hello1\nhello2\nhello3",
				"dir/text1.txt": "test 1\ntest 2\ntest 3",
				"dir/text2.txt": "test 2",
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

		for _, zf := range zr.File {
			if txt, ok := tt.contents[zf.Name]; ok {
				r, err := zf.Open()
				if err != nil {
					t.Fatal(err)
				}
				defer r.Close()

				body := new(bytes.Buffer)
				if _, err := io.Copy(body, r); err != nil {
					t.Fatal(err)
				}

				bodyStr := body.String()
				if bodyStr != txt {
					t.Fatalf("update file %s content=%q, want %q", zf.Name, bodyStr, txt)
				}
			}
		}
	}
}
