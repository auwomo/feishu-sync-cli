package discovery

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/your-org/feishu-sync/internal/feishu"
)

func TestDiscoverDriveFolder_Paginates(t *testing.T) {
	calls := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if r.Header.Get("Authorization") == "" {
			t.Fatalf("missing auth")
		}
		w.Header().Set("Content-Type", "application/json")
		if calls == 1 {
			io.WriteString(w, `{"code":0,"msg":"ok","data":{"has_more":true,"page_token":"p2","files":[{"token":"a","name":"A","type":"docx"}]}}`)
			return
		}
		io.WriteString(w, `{"code":0,"msg":"ok","data":{"has_more":false,"page_token":"","files":[{"token":"b","name":"B","type":"sheet"}]}}`)
	}))
	defer srv.Close()

	c := feishu.NewClient(srv.Client())
	cBase := srv.URL
	// hack: override baseURL via unexported field using struct literal copy
	c.BaseURL = cBase

	items, errs := DiscoverDriveFolder(context.Background(), c, "tenant", "folder")
	if len(errs) != 0 {
		t.Fatalf("unexpected errs: %+v", errs)
	}
	if len(items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(items))
	}
	if calls != 2 {
		t.Fatalf("expected 2 calls, got %d", calls)
	}
}
