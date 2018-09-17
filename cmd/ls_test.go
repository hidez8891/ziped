package cmd

import (
	"bytes"
	"strings"
	"testing"
)

func TestLsRender(t *testing.T) {
	tests := []struct {
		args   []string
		output string
	}{
		{
			args: []string{
				"ls",
				"../testcase/test.zip",
			},
			output: strings.Join([]string{
				"dir/",
				"dir/text1.txt",
				"dir/text2.txt",
				"text1.txt",
			}, "\n") + "\n",
		},
		{
			args: []string{
				"ls",
				"../testcase/test.zip",
				"--filter",
				"dir/*.txt",
			},
			output: strings.Join([]string{
				"dir/text1.txt",
				"dir/text2.txt",
			}, "\n") + "\n",
		},
		{
			args: []string{
				"ls",
				"../testcase/test.zip",
				"--regexp",
				"\\d.txt",
			},
			output: strings.Join([]string{
				"dir/text1.txt",
				"dir/text2.txt",
				"text1.txt",
			}, "\n") + "\n",
		},
	}

	for _, tt := range tests {
		stdout := new(bytes.Buffer)
		stderr := new(bytes.Buffer)

		cmd := newRootCmd(stdout, stderr)
		cmd.SetArgs(tt.args)
		if err := cmd.Execute(); err != nil {
			t.Fatal(err)
		}

		if stderr.Len() != 0 {
			t.Fatalf("error output: %q", stderr.String())
		}

		out := stdout.String()
		if out != tt.output {
			t.Fatalf("output=%q, want %q", out, tt.output)
		}
	}
}
