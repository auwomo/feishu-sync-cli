package cli

import (
	"io"
	"os"
)

type termStyle struct {
	color bool
}

func newTermStyle(out io.Writer) termStyle {
	// NO_COLOR is a de-facto standard.
	if _, ok := os.LookupEnv("NO_COLOR"); ok {
		return termStyle{color: false}
	}
	f, ok := out.(*os.File)
	if !ok {
		return termStyle{color: false}
	}
	return termStyle{color: isTTYFile(f)}
}

func (s termStyle) heading(t string) string {
	if !s.color {
		return t
	}
	return "\x1b[1;36m" + t + "\x1b[0m" // bold cyan
}

func (s termStyle) warn(t string) string {
	if !s.color {
		return t
	}
	return "\x1b[33m" + t + "\x1b[0m" // yellow
}

func (s termStyle) faint(t string) string {
	if !s.color {
		return t
	}
	return "\x1b[2m" + t + "\x1b[0m"
}

