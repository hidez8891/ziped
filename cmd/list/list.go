package cmd_list

import (
	"fmt"
	"os"

	"github.com/hidez8891/zip"
	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/transform"
)

type CmdList struct {
}

func NewCommand() *CmdList {
	return &CmdList{}
}

func (o *CmdList) Run(u *zip.Updater) error {
	for _, zf := range u.Files() {
		name := zf.Name

		if zf.NonUTF8 {
			decoder := japanese.ShiftJIS.NewDecoder()
			decodeName, _, err := transform.String(decoder, name)
			if err == nil {
				name = decodeName
			}
		}

		fmt.Fprintln(os.Stdout, name)
	}
	return nil
}
