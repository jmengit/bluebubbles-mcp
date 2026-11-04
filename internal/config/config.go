package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os"
	"strings"
	"time"
)

type Config struct {
	Mode                 string
	BlueBubblesURL       *url.URL
	BlueBubblesPassword  string
	BlueBubblesAuthQuery string
	BlueBubblesHeaders   map[string]string
	MCPAuthToken         string
	ListenAddr           string
	AllowedHosts         map[string]struct{}
	AllowedOrigins       map[string]struct{}
	ReadTimeout          time.Duration
	WriteTimeout         time.Duration
	MaxRequestBytes      int64
}

func Load() (Config, error) {
	mode := strings.ToLower(strings.TrimSpace(os.Getenv("BLUEBUBBLES_MODE")))
	if mode == "" {
		mode = "readonly"
	}
	if mode != "readonly" && mode != "interactive" {
		return Config{}, fmt.Errorf("BLUEBUBBLES_MODE must be readonly or interactive")
	}
	rawURL := strings.TrimRight(strings.TrimSpace(os.Getenv("BLUEBUBBLES_URL")), "/")
	if rawURL == "" {
		return Config{}, errors.New("BLUEBUBBLES_URL is required")
	}
	u, err := url.Parse(rawURL)
	if err != nil || u.Scheme == "" || u.Host == "" || u.User != nil {
		return Config{}, errors.New("BLUEBUBBLES_URL must be an absolute URL without userinfo")
	}
	password, err := secret("BLUEBUBBLES_PASSWORD_FILE", "BLUEBUBBLES_PASSWORD")
	if err != nil {
		return Config{}, err
	}
	token, err := secret("MCP_AUTH_TOKEN_FILE", "MCP_AUTH_TOKEN")
	if err != nil {
		return Config{}, err
	}
	q := strings.TrimSpace(os.Getenv("BLUEBUBBLES_AUTH_QUERY"))
	if q == "" {
		q = "password"
	}
	if q != "password" && q != "guid" && q != "token" {
		return Config{}, errors.New("BLUEBUBBLES_AUTH_QUERY must be password, guid, or token")
	}
	headers, err := headerFile(os.Getenv("BLUEBUBBLES_HEADERS_FILE"), os.Getenv("BLUEBUBBLES_HEADERS_JSON"))
	if err != nil {
		return Config{}, err
	}
	addr := os.Getenv("MCP_LISTEN_ADDR")
	if addr == "" {
		addr = "127.0.0.1:8080"
	}
	allowedHosts, err := listSet(os.Getenv("MCP_ALLOWED_HOSTS"), false)
	if err != nil {
		return Config{}, fmt.Errorf("MCP_ALLOWED_HOSTS: %w", err)
	}
	allowedOrigins, err := listSet(os.Getenv("MCP_ALLOWED_ORIGINS"), true)
	if err != nil {
		return Config{}, fmt.Errorf("MCP_ALLOWED_ORIGINS: %w", err)
	}
	return Config{Mode: mode, BlueBubblesURL: u, BlueBubblesPassword: password, BlueBubblesAuthQuery: q, BlueBubblesHeaders: headers, MCPAuthToken: token, ListenAddr: addr, AllowedHosts: allowedHosts, AllowedOrigins: allowedOrigins, ReadTimeout: 15 * time.Second, WriteTimeout: 30 * time.Second, MaxRequestBytes: 4 << 20}, nil
}

func secret(fileEnv, directEnv string) (string, error) {
	if p := strings.TrimSpace(os.Getenv(fileEnv)); p != "" {
		b, err := os.ReadFile(p)
		if err != nil {
			return "", fmt.Errorf("read %s: %w", fileEnv, err)
		}
		v := strings.TrimSpace(string(b))
		if v == "" {
			return "", fmt.Errorf("%s is empty", fileEnv)
		}
		return v, nil
	}
	v := strings.TrimSpace(os.Getenv(directEnv))
	if v == "" {
		return "", fmt.Errorf("%s or %s is required", fileEnv, directEnv)
	}
	return v, nil
}
func listSet(raw string, origins bool) (map[string]struct{}, error) {
	result := map[string]struct{}{}
	for _, item := range strings.Split(raw, ",") {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		if item == "*" || item == "null" {
			return nil, errors.New("wildcards and null are not allowed")
		}
		if origins {
			u, err := url.Parse(item)
			if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" || u.Path != "" || u.RawQuery != "" || u.Fragment != "" {
				return nil, fmt.Errorf("invalid exact origin %q", item)
			}
			item = u.Scheme + "://" + u.Host
		} else if strings.ContainsAny(item, "/?#\r\n") {
			return nil, fmt.Errorf("invalid host %q", item)
		}
		result[item] = struct{}{}
	}
	return result, nil
}

func headerFile(path, rawJSON string) (map[string]string, error) {
	var b []byte
	var err error
	if strings.TrimSpace(rawJSON) != "" {
		b = []byte(rawJSON)
	} else if strings.TrimSpace(path) != "" {
		b, err = os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("read BLUEBUBBLES_HEADERS_FILE: %w", err)
		}
	} else {
		return map[string]string{}, nil
	}
	var h map[string]string
	if err := json.Unmarshal(b, &h); err != nil {
		return nil, fmt.Errorf("parse BLUEBUBBLES_HEADERS_FILE: %w", err)
	}
	for k, v := range h {
		lk := strings.ToLower(k)
		if k == "" || strings.ContainsAny(k, "\r\n") || strings.ContainsAny(v, "\r\n") || lk == "host" || lk == "content-length" || lk == "transfer-encoding" || lk == "connection" || lk == "proxy-connection" || lk == "keep-alive" || lk == "upgrade" {
			return nil, fmt.Errorf("invalid upstream header %q", k)
		}
	}
	return h, nil
}
