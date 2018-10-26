package cmd

import (
	"bytes"
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

func TestRmNotModified(t *testing.T) {
	testfile := "../testcase/test.zip"
	args := []string{
		"rm",
		"--overwrite",
		"--filter",
		"dummy",
	}

	tmpname, err := copyTempFile(testfile)
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpname)

	st, err := os.Stat(tmpname)
	if err != nil {
		t.Fatal(err)
	}
	time1 := st.ModTime().UnixNano()

	cmd := newRootCmd(ioutil.Discard, ioutil.Discard)
	cmd.SetArgs(append(args, tmpname))
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}

	st, err = os.Stat(tmpname)
	if err != nil {
		t.Fatal(err)
	}
	time2 := st.ModTime().UnixNano()

	if time1 != time2 {
		t.Fatalf("file was changed unnecessarily")
	}
}
