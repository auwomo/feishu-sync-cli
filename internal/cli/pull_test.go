package cli

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/your-org/feishu-sync/internal/feishu"
)

func TestRunPull_DryRun_PrintsManifest(t *testing.T) {
	wsDir := t.TempDir()
	cfgDir := filepath.Join(wsDir, ".feishu-sync")
	if err := os.MkdirAll(cfgDir, 0o755); err != nil {
		t.Fatal(err)
	}

	secretEnv := "FEISHU_APP_SECRET_TEST"
	os.Setenv(secretEnv, "sec")
	defer os.Unsetenv(secretEnv)

	cfg := `app:
  id: "app"
  secret_env: "` + secretEnv + `"
  secret_file: ".feishu-sync/secret"
scope:
  mode: "drive"
  drive_folder_tokens: ["folder"]
  wiki_space_ids: []
output:
  dir: "."
`
	if err := os.WriteFile(filepath.Join(cfgDir, "config.yaml"), []byte(cfg), 0o644); err != nil {
		t.Fatal(err)
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case strings.HasPrefix(r.URL.Path, "/open-apis/auth/v3/tenant_access_token/internal"):
			io.WriteString(w, `{"code":0,"msg":"ok","tenant_access_token":"tenant"}`)
		case strings.HasPrefix(r.URL.Path, "/open-apis/drive/v1/files"):
			io.WriteString(w, `{"code":0,"msg":"ok","data":{"has_more":false,"page_token":"","files":[{"token":"a","name":"A","type":"docx"}]}}`)
		default:
			w.WriteHeader(404)
			io.WriteString(w, `not found`)
		}
	}))
	defer srv.Close()

	oldNew := feishuNewClient
	feishuNewClient = func() *feishu.Client {
		c := feishu.NewClient(srv.Client())
		c.BaseURL = srv.URL
		return c
	}
	defer func() { feishuNewClient = oldNew }()

	var buf bytes.Buffer
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	defer func() { os.Stdout = oldStdout }()

	done := make(chan struct{})
	go func() {
		io.Copy(&buf, r)
		close(done)
	}()

	if err := runPull(wsDir, "", true); err != nil {
		w.Close()
		t.Fatalf("unexpected err: %v", err)
	}
	w.Close()
	<-done

	out := buf.String()
	if !strings.Contains(out, "\"drive\"") {
		t.Fatalf("expected manifest JSON, got: %s", out)
	}
	if !strings.Contains(out, "\"token\": \"a\"") {
		t.Fatalf("expected drive item, got: %s", out)
	}
}
