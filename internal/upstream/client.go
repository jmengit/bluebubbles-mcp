package upstream

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/jmengit/bluebubbles-mcp/internal/config"
)

type Client struct {
	cfg  config.Config
	http *http.Client
}

func New(c config.Config) *Client {
	return &Client{cfg: c, http: &http.Client{Timeout: c.ReadTimeout, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}}
}
func (c *Client) Get(ctx context.Context, path string, query url.Values) (json.RawMessage, error) {
	return c.do(ctx, http.MethodGet, path, query, nil)
}
func (c *Client) PostJSON(ctx context.Context, path string, query url.Values, body any) (json.RawMessage, error) {
	b, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	return c.do(ctx, http.MethodPost, path, query, strings.NewReader(string(b)))
}
func (c *Client) do(ctx context.Context, method, path string, query url.Values, body io.Reader) (json.RawMessage, error) {
	if !strings.HasPrefix(path, "/") || strings.Contains(path, "..") {
		return nil, errors.New("invalid upstream path")
	}
	u := *c.cfg.BlueBubblesURL
	u.Path = strings.TrimRight(u.Path, "/") + "/api/v1" + path
	q := u.Query()
	for k, v := range query {
		for _, x := range v {
			q.Add(k, x)
		}
	}
	q.Set(c.cfg.BlueBubblesAuthQuery, c.cfg.BlueBubblesPassword)
	u.RawQuery = q.Encode()
	req, err := http.NewRequestWithContext(ctx, method, u.String(), body)
	if err != nil {
		return nil, err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	for k, v := range c.cfg.BlueBubblesHeaders {
		req.Header.Set(k, v)
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, errors.New("upstream request failed")
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("upstream returned HTTP %d", resp.StatusCode)
	}
	var envelope struct {
		Status int             `json:"status"`
		Data   json.RawMessage `json:"data"`
	}
	if err := json.NewDecoder(io.LimitReader(resp.Body, c.cfg.MaxRequestBytes)).Decode(&envelope); err != nil {
		return nil, errors.New("invalid upstream response")
	}
	if envelope.Status >= 400 {
		return nil, fmt.Errorf("upstream API returned status %d", envelope.Status)
	}
	if envelope.Data == nil {
		envelope.Data = json.RawMessage("null")
	}
	return envelope.Data, nil
}
