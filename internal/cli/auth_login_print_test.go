package cli

import (
	"bytes"
	"strings"
	"testing"
)

func TestPrintAuthLoginOptions_WithRedirectURI_UsesEffectiveRedirect(t *testing.T) {
	var out bytes.Buffer
	opts := authLoginOptions{RedirectURI: "https://example.com/callback"}
	authURL := "https://open.feishu.cn/open-apis/authen/v1/index?redirect_uri=https%3A%2F%2Fexample.com%2Fcallback"
	printAuthLoginOptions(&out, opts, authURL, opts.RedirectURI, "http://127.0.0.1:18900/callback")

	s := out.String()
	if !strings.Contains(s, "https%3A%2F%2Fexample.com%2Fcallback") {
		t.Fatalf("expected output to include effective redirect_uri, got: %s", s)
	}
	// local redirect is always displayed as the fixed built-in callback;
	// a custom redirect is shown as the effective redirect.
}
