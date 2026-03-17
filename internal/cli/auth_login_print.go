package cli

import (
	"fmt"
	"io"
	"os"
)

func printAuthLoginOptions(out io.Writer, opts authLoginOptions, authURL string, effectiveRedirectURI string, localRedirectURI string) {
	st := newTermStyle(out)

	fmt.Fprintln(out, st.heading("Auth login"))
	fmt.Fprintln(out)

	fmt.Fprintln(out, st.heading("Option 1: Local (recommended on your laptop)"))
	fmt.Fprintln(out, "Will start a local callback server and open the browser.")
	fmt.Fprintln(out, authURL)
	fmt.Fprintln(out, st.faint("(will open browser)"))
	if opts.Verbose {
		fmt.Fprintln(out, st.faint("redirect_uri: "+localRedirectURI))
	}
	fmt.Fprintln(out)

	fmt.Fprintln(out, st.heading("Option 2: Remote/manual (recommended on a server)"))
	if opts.RedirectURI != "" {
		fmt.Fprintln(out, "Open this URL to authorize:")
		fmt.Fprintln(out, authURL)
		if opts.Verbose {
			fmt.Fprintln(out, st.faint("redirect_uri: "+effectiveRedirectURI))
		}
	} else {
		fmt.Fprintln(out, "Run:")
		fmt.Fprintln(out, "  feishu-sync auth login --remote --redirect-uri <WHITELISTED_REDIRECT_URI>")
		if opts.Verbose {
			fmt.Fprintln(out, st.faint("effective redirect_uri: "+effectiveRedirectURI))
		}
	}
	fmt.Fprintln(out, st.warn("Note: after you authorize, the browser may show 404/blank — this is normal."))
	fmt.Fprintln(out)

}

func isTTYFile(f *os.File) bool {
	fi, err := f.Stat()
	if err != nil {
		return false
	}
	return (fi.Mode() & os.ModeCharDevice) != 0
}
