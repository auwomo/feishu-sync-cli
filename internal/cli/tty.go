package cli

import (
	"os"

	"golang.org/x/term"
)

func isTTYFile(f *os.File) bool {
	return term.IsTerminal(int(f.Fd()))
}
