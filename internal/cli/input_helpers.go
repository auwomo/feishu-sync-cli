package cli

import (
	"bufio"
	"io"
	"strings"
)

// readLineFrom reads a single line from the provided reader.
// It trims trailing \r\n and allows EOF with a partial line.
func readLineFrom(r io.Reader) (string, error) {
	br := bufio.NewReader(r)
	line, err := br.ReadString('\n')
	if err == io.EOF {
		return strings.TrimRight(line, "\r\n"), nil
	}
	if err != nil {
		return "", err
	}
	return strings.TrimRight(line, "\r\n"), nil
}
