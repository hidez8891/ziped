package cmd

import (
	"bytes"
	"io"
	"io/ioutil"
	"os"
	"testing"
)

func copyTempFile(path string) (string, error) {
	r, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer r.Close()

	tmp, err := ioutil.TempFile("", "test*.zip")
	if err != nil {
		return "", err
	}
	defer tmp.Close()

	if _, err := io.Copy(tmp, r); err != nil {
		return "", err
	}
	return tmp.Name(), nil
}

func helperExecuteCommand(t *testing.T, args []string) {
	t.Helper()

	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)

	cmd := newRootCmd(stdout, stderr)
	cmd.SetArgs(args)
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}

	if stderr.Len() != 0 {
		t.Fatalf("error output: %q", stderr.String())
	}
	if stdout.Len() != 0 {
		t.Fatalf("stdout output: %q", stdout.String())
	}
}
