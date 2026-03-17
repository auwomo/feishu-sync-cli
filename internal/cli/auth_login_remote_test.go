package cli

import (
	"bytes"
	"context"
	"errors"
	"testing"
	"time"
)

func TestAuthLogin_Remote_DoesNotListen(t *testing.T) {
	called := false
	old := netListen
	defer func() { netListen = old }()

	netListen = func(network, address string) (listener, error) {
		called = true
		return nil, errors.New("should not listen")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	var out, errOut bytes.Buffer
	_ = runAuthLogin(ctx, ".", "", authLoginOptions{Remote: true, RedirectURI: "https://example.com/callback", NoBrowser: true}, &out, &errOut)
	if called {
		t.Fatalf("remote mode attempted to listen")
	}
}
