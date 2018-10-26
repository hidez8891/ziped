package cmd

import (
	"io"
	"io/ioutil"
	"os"
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
