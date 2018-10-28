package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/hidez8891/zip"
)

func TestRmExecuteOverwrite(t *testing.T) {
	tests := []struct {
		file     string
		args     []string
		contents []string
	}{
		{
			file: "../testcase/test.zip",
			args: []string{
				"rm",
				"--overwrite",
				"--filter",
				"text1.txt",
				"--show-progress=false",
			},
			contents: []string{
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
				"--show-progress=false",
			},
			contents: []string{
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
				"--show-progress=false",
			},
			contents: []string{
				"dir/",
			},
		},
	}

	for _, tt := range tests {
		tmpname, err := copyTempFile(tt.file)
		if err != nil {
			t.Fatal(err)
		}
		defer os.Remove(tmpname)

		helperExecuteCommand(t, append(tt.args, tmpname))
		helperRmCheckFileContents(t, tmpname, tt.contents)
	}
}

func TestRmParallelExecute(t *testing.T) {
	tests := []struct {
		files    []string
		args     []string
		contents map[string][]string
	}{
		{
			files: []string{
				"../testcase/test.zip",
				"../testcase/test2.zip",
			},
			args: []string{
				"rm",
				"--overwrite",
				"--filter",
				"*.txt",
				"--jobs=2",
				"--show-progress=false",
			},
			contents: map[string][]string{
				"../testcase/test.zip": []string{
					"dir/",
					"dir/text1.txt",
					"dir/text2.txt",
				},
				"../testcase/test2.zip": []string{
					"dir/",
					"dir/text1.txt",
					"dir/text2.txt",
				},
			},
		},
	}

	for _, tt := range tests {
		tempfilemap := make(map[string]string)
		for _, filename := range tt.files {
			tmpname, err := copyTempFile(filename)
			if err != nil {
				t.Fatal(err)
			}
			tempfilemap[filename] = tmpname
		}
		defer func() {
			for _, tmpname := range tempfilemap {
				os.Remove(tmpname)
			}
		}()

		wildCardPath := filepath.Dir(tempfilemap[tt.files[0]])
		wildCardPath = filepath.Join(wildCardPath, "*.zip")
		helperExecuteCommand(t, append(tt.args, wildCardPath))

		for _, filename := range tt.files {
			tmpname := tempfilemap[filename]
			contents := tt.contents[filename]

			helperRmCheckFileContents(t, tmpname, contents)
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
		"--show-progress=false",
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

	helperExecuteCommand(t, append(args, tmpname))

	st, err = os.Stat(tmpname)
	if err != nil {
		t.Fatal(err)
	}
	time2 := st.ModTime().UnixNano()

	if time1 != time2 {
		t.Fatalf("file was changed unnecessarily")
	}
}

func helperRmCheckFileContents(t *testing.T, filename string, contents []string) {
	zr, err := zip.OpenReader(filename)
	if err != nil {
		t.Fatal(err)
	}
	defer zr.Close()

	if len(zr.File) != len(contents) {
		t.Fatalf("update filename count=%d, want %d", len(zr.File), len(contents))
	}

	for i, zf := range zr.File {
		if zf.Name != contents[i] {
			t.Fatalf("update filename=%q, want %q", zf.Name, contents[i])
		}
	}
}
