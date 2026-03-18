package cli

import "golang.org/x/term"

func readPasswordFromFD(fd int) ([]byte, error) {
	return term.ReadPassword(fd)
}
