package feishu

import (
	"net/http/httptest"
	"testing"
)

func TestParseOAuthCallback(t *testing.T) {
	r := httptest.NewRequest("GET", "http://127.0.0.1/callback?code=abc&state=xyz", nil)
	code, state, err := ParseOAuthCallback(r)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if code != "abc" || state != "xyz" {
		t.Fatalf("got code=%q state=%q", code, state)
	}
}

func TestParseOAuthCallback_ErrorParams(t *testing.T) {
	r := httptest.NewRequest("GET", "http://127.0.0.1/callback?error=access_denied&error_description=nope&state=xyz", nil)
	code, state, err := ParseOAuthCallback(r)
	if err == nil {
		t.Fatalf("expected err")
	}
	if code != "" || state != "xyz" {
		t.Fatalf("got code=%q state=%q", code, state)
	}
}
