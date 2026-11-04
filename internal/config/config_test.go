package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadReadonlyAndFiles(t *testing.T) {
	d := t.TempDir()
	pw := filepath.Join(d, "pw")
	tok := filepath.Join(d, "tok")
	hdr := filepath.Join(d, "headers.json")
	for p, v := range map[string]string{pw: "bb-pass\n", tok: "mcp-token\n", hdr: `{"X-Test":"ok"}`} {
		if err := os.WriteFile(p, []byte(v), 0600); err != nil {
			t.Fatal(err)
		}
	}
	t.Setenv("BLUEBUBBLES_MODE", "")
	t.Setenv("BLUEBUBBLES_URL", "https://bluebubbles.example")
	t.Setenv("BLUEBUBBLES_PASSWORD_FILE", pw)
	t.Setenv("BLUEBUBBLES_PASSWORD", "")
	t.Setenv("MCP_AUTH_TOKEN_FILE", tok)
	t.Setenv("MCP_AUTH_TOKEN", "")
	t.Setenv("BLUEBUBBLES_HEADERS_FILE", hdr)
	t.Setenv("MCP_LISTEN_ADDR", "")
	c, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if c.Mode != "readonly" || c.BlueBubblesPassword != "bb-pass" || c.MCPAuthToken != "mcp-token" || c.BlueBubblesHeaders["X-Test"] != "ok" {
		t.Fatalf("bad config: %+v", c)
	}
}
func TestLoadRejectsUserInfoAndBadHeader(t *testing.T) {
	t.Setenv("BLUEBUBBLES_URL", "https://bluebubbles.example?existing=not-allowed")
	if _, err := Load(); err == nil {
		t.Fatal("expected URL query error")
	}
	for _, bad := range []string{"Host", "Authorization", "Proxy-Authorization", "Cookie", "Forwarded", "X-Forwarded-Host"} {
		t.Setenv("BLUEBUBBLES_HEADERS_JSON", `{"`+bad+`":"evil"}`)
		if _, err := Load(); err == nil {
			t.Fatalf("expected bad header rejection for %s", bad)
		}
	}
}
