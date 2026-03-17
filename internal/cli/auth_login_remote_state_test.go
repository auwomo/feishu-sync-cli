package cli

import "testing"

func TestRemoteStateValidation_Mismatch(t *testing.T) {
	expected := "expected"
	input := "https://example.com/callback?code=abc&state=wrong"
	_, st, err := parseOAuthPastedInput(input)
	if err != nil {
		t.Fatalf("unexpected parse err: %v", err)
	}
	if st == "" {
		t.Fatalf("expected state")
	}
	if st == expected {
		t.Fatalf("expected mismatch")
	}
}
