package cli

import "testing"

func TestAuthLoginRedirectURI_Defaults(t *testing.T) {
	got := authLoginRedirectURI("", 0, "")
	want := "http://127.0.0.1:18900/callback"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestAuthLoginRedirectURI_Custom(t *testing.T) {
	got := authLoginRedirectURI("0.0.0.0", 18000, "cb")
	want := "http://0.0.0.0:18000/cb"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}
