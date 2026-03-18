package cli

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/your-org/feishu-sync/internal/feishu"
)

// netListen is overridden in tests.
var netListen = net.Listen

// listener is the minimal interface we need from net.Listener.
type listener interface {
	Close() error
	Addr() net.Addr
}

func tryManualAuth(ctx context.Context, out ioWriter, expectedState string) (string, error) {

	input, err := readLine(ctx, "")
	if err != nil {
		return "", err
	}
	if strings.TrimSpace(input) == "" {
		return "", errors.New("skipped")
	}
	code, state, err := parseOAuthPastedInput(input)
	if err != nil {
		return "", err
	}
	if state == "" {
		return "", errors.New("callback url missing state (copy full url)")
	}
	if state != expectedState {
		return "", errors.New("invalid oauth state")
	}
	return code, nil
}

func tryLocalAuth(ctx context.Context, out ioWriter, authURL string, expectedState string, noBrowser bool, timeout time.Duration) (string, error) {
	addr := "127.0.0.1:18900"
	callbackPath := "/callback"

	ln, err := netListen("tcp", addr)
	if err != nil {
		return "", fmt.Errorf("cannot listen on %s: %w (port may be in use)", addr, err)
	}
	defer ln.Close()

	codeCh := make(chan string, 1)
	errCh := make(chan error, 1)

	mux := http.NewServeMux()
	mux.HandleFunc(callbackPath, func(w http.ResponseWriter, r *http.Request) {
		code, st, err := feishu.ParseOAuthCallback(r)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte("Auth failed: " + err.Error()))
			errCh <- err
			return
		}
		if st != expectedState {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte("Auth failed: invalid state"))
			errCh <- errors.New("invalid oauth state")
			return
		}
		_, _ = w.Write([]byte("OK. You can close this tab and return to the terminal."))
		codeCh <- code
	})

	srv := &http.Server{Handler: mux}
	go func() {
		_ = srv.Serve(ln.(net.Listener))
	}()

	if !noBrowser {
		go func() { _ = openBrowser(authURL) }()
	}

	if timeout <= 0 {
		timeout = 2 * time.Minute
	}
	ctx2, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	select {
	case code := <-codeCh:
		_ = srv.Shutdown(context.Background())
		return code, nil
	case err := <-errCh:
		_ = srv.Shutdown(context.Background())
		return "", err
	case <-ctx2.Done():
		_ = srv.Shutdown(context.Background())
		return "", fmt.Errorf("auth timeout after %s", timeout)
	}
}

// ioWriter matches io.Writer (to keep helper file imports small).
type ioWriter interface {
	Write(p []byte) (n int, err error)
}
