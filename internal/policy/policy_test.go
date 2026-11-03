package policy

import (
	"net/http"
	"net/url"
	"testing"
)

func TestReadonlyManifest(t *testing.T) {
	for _, p := range []string{"/ping", "/server/info", "/chat/count", "/chat/abc/message", "/message/count", "/message/abc"} {
		if !AllowRead(http.MethodGet, p) {
			t.Errorf("expected read %s", p)
		}
	}
	for _, p := range []string{"/server/restart/hard", "/chat/abc/read", "/message/text", "/message/abc/edit", "/foo/../bar"} {
		if AllowRead(http.MethodGet, p) {
			t.Errorf("unexpected read %s", p)
		}
	}
	if !AllowReadQuery(http.MethodPost, "/chat/query") || AllowReadQuery(http.MethodPost, "/message/text") {
		t.Fatal("query policy")
	}
}
func TestBuildQuery(t *testing.T) {
	v := url.Values{"limit": {"2"}}
	q := BuildQuery("password", "secret", v)
	if q.Get("password") != "secret" || q.Get("limit") != "2" {
		t.Fatal(q)
	}
}
