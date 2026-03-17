package cli

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"net/url"
	"os"
	"strings"
)

// readLine reads one line from stdin.
// ctx currently is not used (kept for future cancellation support).
func readLine(_ context.Context, prompt string) (string, error) {
	if prompt != "" {
		fmt.Fprint(os.Stdout, prompt)
	}
	br := bufio.NewReader(os.Stdin)
	line, err := br.ReadString('\n')
	if err != nil {
		// allow EOF with partial line
		if strings.TrimSpace(line) != "" {
			return strings.TrimSpace(line), nil
		}
		return "", err
	}
	return strings.TrimSpace(line), nil
}

// parseOAuthPastedInput accepts a full redirect/callback URL (containing code/state)
// or a raw authorization code.
//
// If input parses as URL, it will extract code/state from query first; if missing,
// it will attempt to extract from fragment.
func parseOAuthPastedInput(input string) (code string, state string, err error) {
	s := strings.TrimSpace(input)
	if s == "" {
		return "", "", errors.New("empty input")
	}

	// Try URL parse.
	if u, uerr := url.Parse(s); uerr == nil && u.Scheme != "" {
		q := u.Query()
		code = q.Get("code")
		state = q.Get("state")

		// Some providers return params in fragment.
		if code == "" && u.Fragment != "" {
			if fragQ, ferr := url.ParseQuery(u.Fragment); ferr == nil {
				code = fragQ.Get("code")
				if state == "" {
					state = fragQ.Get("state")
				}
			}
		}
		if code == "" {
			return "", "", errors.New("callback url missing code")
		}
		return code, state, nil
	}

	// Not a URL: treat as raw code.
	return s, "", nil
}
