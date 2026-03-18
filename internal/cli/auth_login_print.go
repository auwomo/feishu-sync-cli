package cli

import (
	"fmt"
	"io"
)

func printAuthLoginOptions(out io.Writer, opts authLoginOptions, authURL string, effectiveRedirectURI string, localRedirectURI string) {
	st := newTermStyle(out)

	fmt.Fprintln(out, st.heading("Login"))
	fmt.Fprintln(out)

	fmt.Fprintln(out, st.heading("Authorize URL"))
	fmt.Fprintln(out, st.heading(authURL))
	if opts.Verbose {
		fmt.Fprintln(out, st.faint("redirect_uri: "+effectiveRedirectURI))
	}
	fmt.Fprintln(out)

	fmt.Fprintln(out, st.heading("How to use"))
	fmt.Fprintln(out, "1) Local (recommended): we will open the URL and wait for the callback.")
	fmt.Fprintln(out, st.faint("   callback: "+localRedirectURI))
	fmt.Fprintln(out, "2) Remote/manual: open the same URL on another machine.")
	fmt.Fprintln(out, st.faint("   after login it may show 404/blank — ")+st.warn("this is normal")+st.faint(". Copy the FULL URL and paste it back (must include state)."))
	fmt.Fprintln(out)

	fmt.Fprintln(out, st.heading("Paste callback URL (optional)"))
	fmt.Fprintln(out, "If you used remote/manual login: paste the FULL callback URL here (?code=...&state=...).")
	fmt.Fprintln(out, st.faint("Leave empty and press Enter to keep waiting for the local callback."))
	fmt.Fprintln(out)
}
