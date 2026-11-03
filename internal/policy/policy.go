package policy

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"
)

type Mode string

const (
	Readonly    Mode = "readonly"
	Interactive Mode = "interactive"
)

var chatGUID = regexp.MustCompile(`^[A-Za-z0-9:_+@.-]{1,512}$`)

func (m Mode) AllowsMutation() bool { return m == Interactive }
func ValidateMode(v string) (Mode, error) {
	m := Mode(strings.ToLower(strings.TrimSpace(v)))
	if m != Readonly && m != Interactive {
		return "", fmt.Errorf("unsupported mode %q", v)
	}
	return m, nil
}
func ValidateChatGUID(v string) error {
	if !chatGUID.MatchString(v) {
		return errors.New("invalid chat GUID")
	}
	return nil
}
func SafePath(path string) bool { return strings.HasPrefix(path, "/") && !strings.Contains(path, "..") }
func AllowRead(method, path string) bool {
	if !SafePath(path) || method != http.MethodGet {
		return false
	}
	switch {
	case path == "/ping", path == "/server/info", path == "/chat/count":
		return true
	case strings.HasPrefix(path, "/chat/") && strings.HasSuffix(path, "/message"):
		return true
	case strings.HasPrefix(path, "/chat/") && !strings.Contains(path, "/read") && !strings.Contains(path, "/unread") && !strings.Contains(path, "/typing") && strings.Count(path, "/") == 2:
		return true
	case path == "/message/count", path == "/message/count/updated", path == "/message/count/me":
		return true
	case strings.HasPrefix(path, "/message/") && !strings.HasSuffix(path, "/text") && !strings.HasSuffix(path, "/attachment") && !strings.HasSuffix(path, "/multipart") && !strings.HasSuffix(path, "/react") && !strings.HasSuffix(path, "/edit") && !strings.HasSuffix(path, "/unsend"):
		return true
	default:
		return false
	}
}
func AllowReadQuery(method, path string) bool {
	return method == http.MethodPost && (path == "/chat/query" || path == "/message/query")
}
func BuildQuery(authKey, password string, extra url.Values) url.Values {
	q := url.Values{}
	for k, vs := range extra {
		for _, v := range vs {
			q.Add(k, v)
		}
	}
	q.Set(authKey, password)
	return q
}
