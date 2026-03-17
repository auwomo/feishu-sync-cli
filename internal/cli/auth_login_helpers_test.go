package cli

import "testing"

func TestParseOAuthPastedInput_URLQuery(t *testing.T) {
	code, state, err := parseOAuthPastedInput("https://example.com/callback?code=abc&state=st123")
	if err != nil {
		t.Fatalf("expected nil err, got %v", err)
	}
	if code != "abc" || state != "st123" {
		t.Fatalf("unexpected result: code=%q state=%q", code, state)
	}
}

func TestParseOAuthPastedInput_URLFragment(t *testing.T) {
	code, state, err := parseOAuthPastedInput("https://example.com/callback#code=abc&state=st123")
	if err != nil {
		t.Fatalf("expected nil err, got %v", err)
	}
	if code != "abc" || state != "st123" {
		t.Fatalf("unexpected result: code=%q state=%q", code, state)
	}
}

func TestParseOAuthPastedInput_RawCode(t *testing.T) {
	code, state, err := parseOAuthPastedInput("abc")
	if err != nil {
		t.Fatalf("expected nil err, got %v", err)
	}
	if code != "abc" || state != "" {
		t.Fatalf("unexpected result: code=%q state=%q", code, state)
	}
}

func TestParseOAuthPastedInput_MissingCode(t *testing.T) {
	_, _, err := parseOAuthPastedInput("https://example.com/callback?state=st123")
	if err == nil {
		t.Fatalf("expected error")
	}
}
