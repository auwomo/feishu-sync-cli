package feishu

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestTenantAccessToken_OK(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/open-apis/auth/v3/tenant_access_token/internal" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		b, _ := io.ReadAll(r.Body)
		if len(b) == 0 {
			t.Fatalf("expected body")
		}
		w.Header().Set("Content-Type", "application/json")
		io.WriteString(w, `{"code":0,"msg":"ok","tenant_access_token":"t-123","expire":7200}`)
	}))
	defer srv.Close()

	c := NewClient(srv.Client())
	c.BaseURL = srv.URL

	tok, err := c.TenantAccessToken(context.Background(), "app", "sec")
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if tok != "t-123" {
		t.Fatalf("unexpected token: %s", tok)
	}
}

func TestTenantAccessToken_CodeNonZero(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		io.WriteString(w, `{"code":100,"msg":"bad"}`)
	}))
	defer srv.Close()

	c := NewClient(srv.Client())
	c.BaseURL = srv.URL

	_, err := c.TenantAccessToken(context.Background(), "app", "sec")
	if err == nil {
		t.Fatalf("expected err")
	}
}
