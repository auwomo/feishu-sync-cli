package feishu

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestExchangeUserCode_Request(t *testing.T) {
	var gotPath, gotAuth string
	var gotBody map[string]string

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotAuth = r.Header.Get("Authorization")
		b, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(b, &gotBody)
		io.WriteString(w, `{"code":0,"msg":"ok","data":{"access_token":"u-123","token_type":"Bearer","expires_in":7200,"refresh_token":"r-123","refresh_expires_in":2592000,"scope":"","tenant_key":"t","open_id":"o"}}`)
	}))
	defer srv.Close()

	c := NewClient(srv.Client())
	c.BaseURL = srv.URL

	_, err := c.ExchangeUserCode(context.Background(), "tenant-abc", "code-xyz")
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}

	if gotPath != "/open-apis/authen/v1/access_token" {
		t.Fatalf("unexpected path: %s", gotPath)
	}
	if gotAuth != "Bearer tenant-abc" {
		t.Fatalf("unexpected auth header: %q", gotAuth)
	}
	if gotBody["grant_type"] != "authorization_code" || gotBody["code"] != "code-xyz" {
		t.Fatalf("unexpected body: %#v", gotBody)
	}
}

func TestRefreshUserToken_Request(t *testing.T) {
	var gotPath, gotAuth string
	var gotBody map[string]string

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotAuth = r.Header.Get("Authorization")
		b, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(b, &gotBody)
		io.WriteString(w, `{"code":0,"msg":"ok","data":{"access_token":"u-123","token_type":"Bearer","expires_in":7200,"refresh_token":"r-123","refresh_expires_in":2592000,"scope":""}}`)
	}))
	defer srv.Close()

	c := NewClient(srv.Client())
	c.BaseURL = srv.URL

	_, err := c.RefreshUserToken(context.Background(), "tenant-abc", "rt-xyz")
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}

	if gotPath != "/open-apis/authen/v1/refresh_access_token" {
		t.Fatalf("unexpected path: %s", gotPath)
	}
	if gotAuth != "Bearer tenant-abc" {
		t.Fatalf("unexpected auth header: %q", gotAuth)
	}
	if gotBody["grant_type"] != "refresh_token" || gotBody["refresh_token"] != "rt-xyz" {
		t.Fatalf("unexpected body: %#v", gotBody)
	}
}
