package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/hidez8891/zip"
)

func TestRenameExecuteOverwrite(t *testing.T) {
	tests := []struct {
		file     string
		args     []string
		contents []string
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
				"--show-progress=false",
			},
			contents: []string{
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
				"--show-progress=false",
			},
			contents: []string{
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
				"--show-progress=false",
			},
			contents: []string{
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
				"--show-progress=false",
			},
			contents: []string{
				"dir/",
				"dir/text1.md",
				"dir/text2.md",
				"text1.txt",
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
		helperRenameCheckFileContents(t, tmpname, tt.contents)
	}
}

func TestRenameParallelExecute(t *testing.T) {
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
				"rename",
				"--overwrite",
				"--from",
				".txt",
				"--to",
				".md",
				"--jobs=2",
				"--show-progress=false",
			},
			contents: map[string][]string{
				"../testcase/test.zip": []string{
					"dir/",
					"dir/text1.md",
					"dir/text2.md",
					"text1.md",
				},
				"../testcase/test2.zip": []string{
					"dir/",
					"dir/text1.md",
					"dir/text2.md",
					"text1.md",
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

			helperRenameCheckFileContents(t, tmpname, contents)
		}
	}
}

func helperRenameCheckFileContents(t *testing.T, filename string, contents []string) {
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
