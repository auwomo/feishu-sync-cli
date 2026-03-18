package cli

import (
	"fmt"
	"io"
)

func printAuthLoginOptions(out io.Writer, opts authLoginOptions, authURL string, effectiveRedirectURI string, localRedirectURI string) {
	st := newTermStyle(out)

	fmt.Fprintln(out, st.heading("Login"))
	fmt.Fprintln(out)
	fmt.Fprintln(out, "This command supports two ways to authorize. Use whichever is easier:")
	fmt.Fprintln(out)

	fmt.Fprintln(out, st.heading("Option 1: Local (auto, recommended)"))
	fmt.Fprintln(out, "- Starts a local callback server and opens your browser")
	fmt.Fprintln(out, "- Callback:", localRedirectURI)
	fmt.Fprintln(out)

	fmt.Fprintln(out, st.heading("Option 2: Remote/manual"))
	fmt.Fprintln(out, "- Use the SAME authorize URL in any browser")
	fmt.Fprintln(out, "- After login it may redirect to 404/blank (normal)")
	fmt.Fprintln(out, "- Copy the FULL URL from the address bar and paste it back here (must include state)")
	if opts.Verbose {
		fmt.Fprintln(out, st.faint("authorize url:"))
		fmt.Fprintln(out, st.faint(authURL))
		fmt.Fprintln(out, st.faint("effective redirect_uri: "+effectiveRedirectURI))
	}
	fmt.Fprintln(out)
}
