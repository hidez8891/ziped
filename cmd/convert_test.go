package cmd

import (
	"bytes"
	"io"
	"os"
	"runtime"
	"strings"
	"testing"

	"github.com/hidez8891/zip"
)

func TestConvertExecuteOverwrite(t *testing.T) {
	tests := []struct {
		file     string
		os       string
		args     []string
		contents map[string]string
	}{
		{
			file: "../testcase/test2.zip",
			os:   "",
			args: []string{
				"convert",
				"--overwrite",
				"--filter",
				"text1.txt",
				"--cmd",
				"sort",
			},
			contents: map[string]string{
				"text1.txt":     "hello1\r\nhello2\r\nhello3",
				"dir/text1.txt": "test 3\r\ntest 2\r\ntest 1",
				"dir/text2.txt": "test 2",
			},
		},
		{
			file: "../testcase/test2.zip",
			os:   "linux",
			args: []string{
				"convert",
				"--overwrite",
				"--filter",
				"*.txt",
				"--cmd",
				"sort -r",
			},
			contents: map[string]string{
				"text1.txt":     "hello1\r\nhello2\r\nhello3",
				"dir/text1.txt": "test 3\r\ntest 2\r\ntest 1",
				"dir/text2.txt": "test 2",
			},
		},
		{
			file: "../testcase/test2.zip",
			os:   "windows",
			args: []string{
				"convert",
				"--overwrite",
				"--filter",
				"*.txt",
				"--cmd",
				"sort /r",
			},
			contents: map[string]string{
				"text1.txt":     "hello3\r\nhello2\r\nhello1",
				"dir/text1.txt": "test 3\r\ntest 2\r\ntest 1",
				"dir/text2.txt": "test 2",
			},
		},
		{
			file: "../testcase/test2.zip",
			os:   "",
			args: []string{
				"convert",
				"--overwrite",
				"--regexp",
				"\\d.txt",
				"--cmd",
				"sort",
			},
			contents: map[string]string{
				"text1.txt":     "hello1\r\nhello2\r\nhello3",
				"dir/text1.txt": "test 1\r\ntest 2\r\ntest 3",
				"dir/text2.txt": "test 2",
			},
		},
	}

	for _, tt := range tests {
		if len(tt.os) != 0 && tt.os != runtime.GOOS {
			continue
		}

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

				// windows's sort command adds a new line.
				bodyStr = strings.Trim(bodyStr, "\r\n")

				if bodyStr != txt {
					t.Fatalf("update file %s content=%q, want %q", zf.Name, bodyStr, txt)
				}
			}
		}
	}
}