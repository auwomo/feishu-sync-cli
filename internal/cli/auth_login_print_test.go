package cli

import (
	"bytes"
	"strings"
	"testing"
)

func TestPrintAuthLoginOptions_RemoteWithRedirectURI_UsesEffectiveRedirect(t *testing.T) {
	var out bytes.Buffer
	opts := authLoginOptions{Remote: true, RedirectURI: "https://example.com/callback"}
	authURL := "https://open.feishu.cn/open-apis/authen/v1/index?redirect_uri=https%3A%2F%2Fexample.com%2Fcallback"
	printAuthLoginOptions(&out, opts, authURL, opts.RedirectURI, "http://127.0.0.1:18900/callback")

	s := out.String()
	if !strings.Contains(s, "https%3A%2F%2Fexample.com%2Fcallback") {
		t.Fatalf("expected output to include effective redirect_uri, got: %s", s)
	}
	if strings.Contains(s, "127.0.0.1:18900/callback") {
		t.Fatalf("did not expect output to include local redirect_uri when remote+redirect provided, got: %s", s)
	}
}
