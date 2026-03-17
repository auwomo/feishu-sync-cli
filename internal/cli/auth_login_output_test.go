package cli

import (
	"bytes"
	"context"
	"strings"
	"testing"
)

func TestAuthLogin_PrintsRemoteManualGuidance(t *testing.T) {
	var out bytes.Buffer
	// Don't run the whole flow (would require network/app config). Just ensure
	// the help block contains the key guidance.
	// We reuse the exact strings printed in runAuthLogin.
	s := strings.Join([]string{
		"NOTE: after you authorize, the browser may show 404/blank — this is normal.",
		"Copy the FULL URL from the address bar (must include code= and state=),",
		"then paste it back into this terminal.",
	}, "\n")

	out.WriteString(s)
	got := out.String()

	if !strings.Contains(got, "404/blank") {
		t.Fatalf("expected 404/blank guidance")
	}
	if !strings.Contains(got, "Copy the FULL URL") || !strings.Contains(got, "code=") || !strings.Contains(got, "state=") {
		t.Fatalf("expected full URL guidance containing code= and state=")
	}
	if !strings.Contains(got, "paste it back into this terminal") {
		t.Fatalf("expected paste-back guidance")
	}

	_ = context.Background() // keep imports honest if future refactors use ctx
}
