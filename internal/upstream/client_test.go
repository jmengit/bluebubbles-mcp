package upstream

import (
	"context"
	"encoding/json"
	"github.com/jmengit/bluebubbles-mcp/internal/config"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"
)

func TestClientBuildsRedactedSafeRequest(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/ping" {
			t.Errorf("path=%s", r.URL.Path)
		}
		if r.URL.Query().Get("password") != "secret" {
			t.Errorf("password absent")
		}
		if r.Header.Get("X-Test") != "ok" {
			t.Errorf("header absent")
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":200,"data":{"ok":true}}`))
	}))
	defer srv.Close()
	u, _ := url.Parse(srv.URL)
	c := config.Config{BlueBubblesURL: u, BlueBubblesPassword: "secret", BlueBubblesAuthQuery: "password", BlueBubblesHeaders: map[string]string{"X-Test": "ok"}, ReadTimeout: time.Second, MaxRequestBytes: 1 << 20}
	got, err := New(c).Get(context.Background(), "/ping", url.Values{})
	if err != nil {
		t.Fatal(err)
	}
	var x map[string]bool
	if err := json.Unmarshal(got, &x); err != nil || !x["ok"] {
		t.Fatalf("got %s", got)
	}
}
