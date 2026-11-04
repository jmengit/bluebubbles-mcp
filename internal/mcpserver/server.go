package mcpserver

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/jmengit/bluebubbles-mcp/internal/config"
	"github.com/jmengit/bluebubbles-mcp/internal/policy"
	"github.com/jmengit/bluebubbles-mcp/internal/upstream"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type Server struct {
	cfg      config.Config
	upstream *upstream.Client
}

func New(c config.Config) *Server { return &Server{cfg: c, upstream: upstream.New(c)} }
func (s *Server) MCP() *mcp.Server {
	out := mcp.NewServer(&mcp.Implementation{Name: "bluebubbles-mcp", Version: "0.1.0"}, nil)
	ro, destructive, open := true, false, true
	mcp.AddTool(out, &mcp.Tool{Name: "ping", Description: "Return gateway availability and configured capability mode.", Annotations: &mcp.ToolAnnotations{ReadOnlyHint: ro, DestructiveHint: &destructive, IdempotentHint: true, OpenWorldHint: &open}}, s.ping)
	if s.cfg.Mode == "readonly" {
		s.addReadTools(out)
	} else {
		s.addReadTools(out)
		s.addInteractiveTools(out)
	}
	return out
}
func (s *Server) ping(ctx context.Context, req *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, any, error) {
	b, _ := json.Marshal(map[string]string{"status": "ok", "mode": s.cfg.Mode})
	return textResult(string(b)), nil, nil
}

type listChatsInput struct {
	Limit  int `json:"limit,omitempty" jsonschema:"maximum number of chats to return"`
	Offset int `json:"offset,omitempty" jsonschema:"number of chats to skip"`
}
type chatInput struct {
	ChatGUID string `json:"chat_guid" jsonschema:"BlueBubbles chat GUID"`
	Limit    int    `json:"limit,omitempty"`
	Offset   int    `json:"offset,omitempty"`
	Sort     string `json:"sort,omitempty"`
	After    int64  `json:"after,omitempty"`
	Before   int64  `json:"before,omitempty"`
}
type messageInput struct {
	MessageGUID string `json:"message_guid" jsonschema:"BlueBubbles message GUID"`
}
type searchInput struct {
	Query    string `json:"query,omitempty"`
	ChatGUID string `json:"chat_guid,omitempty"`
	Limit    int    `json:"limit,omitempty"`
	Offset   int    `json:"offset,omitempty"`
	Sort     string `json:"sort,omitempty"`
	After    int64  `json:"after,omitempty"`
	Before   int64  `json:"before,omitempty"`
}

func bounded(limit, offset int) (int, int, error) {
	if limit == 0 {
		limit = 25
	}
	if limit < 1 || limit > 100 {
		return 0, 0, fmt.Errorf("limit must be between 1 and 100")
	}
	if offset < 0 {
		return 0, 0, fmt.Errorf("offset must not be negative")
	}
	return limit, offset, nil
}
func (s *Server) addReadTools(out *mcp.Server) {
	ro, destructive, open := true, false, false
	mcp.AddTool(out, &mcp.Tool{Name: "list_chats", Description: "List BlueBubbles chats.", Annotations: &mcp.ToolAnnotations{ReadOnlyHint: ro, DestructiveHint: &destructive, IdempotentHint: true, OpenWorldHint: &open}}, func(ctx context.Context, req *mcp.CallToolRequest, in listChatsInput) (*mcp.CallToolResult, any, error) {
		limit, offset, err := bounded(in.Limit, in.Offset)
		if err != nil {
			return nil, nil, err
		}
		data, err := s.upstream.PostJSON(ctx, "/chat/query", url.Values{}, map[string]any{"limit": limit, "offset": offset, "sort": "lastmessage", "with": []string{"lastmessage", "participants"}})
		if err != nil {
			return nil, nil, err
		}
		return jsonResult(data), nil, nil
	})
	mcp.AddTool(out, &mcp.Tool{Name: "get_chat", Description: "Get one BlueBubbles chat by GUID.", Annotations: &mcp.ToolAnnotations{ReadOnlyHint: ro, DestructiveHint: &destructive, IdempotentHint: true, OpenWorldHint: &open}}, func(ctx context.Context, req *mcp.CallToolRequest, in chatInput) (*mcp.CallToolResult, any, error) {
		if err := policy.ValidateChatGUID(in.ChatGUID); err != nil {
			return nil, nil, err
		}
		data, err := s.upstream.Get(ctx, "/chat/"+in.ChatGUID, url.Values{})
		if err != nil {
			return nil, nil, err
		}
		return jsonResult(data), nil, nil
	})
	mcp.AddTool(out, &mcp.Tool{Name: "get_chat_messages", Description: "Get bounded messages from one BlueBubbles chat.", Annotations: &mcp.ToolAnnotations{ReadOnlyHint: ro, DestructiveHint: &destructive, IdempotentHint: true, OpenWorldHint: &open}}, func(ctx context.Context, req *mcp.CallToolRequest, in chatInput) (*mcp.CallToolResult, any, error) {
		limit, offset, err := bounded(in.Limit, in.Offset)
		if err != nil {
			return nil, nil, err
		}
		if err := policy.ValidateChatGUID(in.ChatGUID); err != nil {
			return nil, nil, err
		}
		sort := in.Sort
		if sort == "" {
			sort = "DESC"
		}
		if sort != "ASC" && sort != "DESC" {
			return nil, nil, fmt.Errorf("sort must be ASC or DESC")
		}
		q := url.Values{"limit": {fmt.Sprint(limit)}, "offset": {fmt.Sprint(offset)}, "sort": {sort}}
		if in.After != 0 {
			q.Set("after", fmt.Sprint(in.After))
		}
		if in.Before != 0 {
			q.Set("before", fmt.Sprint(in.Before))
		}
		data, err := s.upstream.Get(ctx, "/chat/"+in.ChatGUID+"/message", q)
		if err != nil {
			return nil, nil, err
		}
		return jsonResult(data), nil, nil
	})
	mcp.AddTool(out, &mcp.Tool{Name: "get_message", Description: "Get one BlueBubbles message by GUID.", Annotations: &mcp.ToolAnnotations{ReadOnlyHint: ro, DestructiveHint: &destructive, IdempotentHint: true, OpenWorldHint: &open}}, func(ctx context.Context, req *mcp.CallToolRequest, in messageInput) (*mcp.CallToolResult, any, error) {
		if err := policy.ValidateChatGUID(in.MessageGUID); err != nil {
			return nil, nil, err
		}
		data, err := s.upstream.Get(ctx, "/message/"+in.MessageGUID, url.Values{"with": {"chats,attachments"}})
		if err != nil {
			return nil, nil, err
		}
		return jsonResult(data), nil, nil
	})
	mcp.AddTool(out, &mcp.Tool{Name: "search_messages", Description: "Search BlueBubbles messages with bounded filters.", Annotations: &mcp.ToolAnnotations{ReadOnlyHint: ro, DestructiveHint: &destructive, IdempotentHint: true, OpenWorldHint: &open}}, func(ctx context.Context, req *mcp.CallToolRequest, in searchInput) (*mcp.CallToolResult, any, error) {
		limit, offset, err := bounded(in.Limit, in.Offset)
		if err != nil {
			return nil, nil, err
		}
		if in.ChatGUID != "" {
			if err := policy.ValidateChatGUID(in.ChatGUID); err != nil {
				return nil, nil, err
			}
		}
		sort := in.Sort
		if sort == "" {
			sort = "DESC"
		}
		if sort != "ASC" && sort != "DESC" {
			return nil, nil, fmt.Errorf("sort must be ASC or DESC")
		}
		body := map[string]any{"limit": limit, "offset": offset, "sort": sort, "with": []string{"chats", "attachment"}}
		if in.ChatGUID != "" {
			body["chatGuid"] = in.ChatGUID
		}
		if in.Query != "" {
			if len(in.Query) > 256 {
				return nil, nil, fmt.Errorf("query too long")
			}
			body["where"] = []map[string]any{{"statement": "message.text LIKE :query", "args": map[string]string{"query": "%" + in.Query + "%"}}}
		}
		if in.After != 0 {
			body["after"] = in.After
		}
		if in.Before != 0 {
			body["before"] = in.Before
		}
		data, err := s.upstream.PostJSON(ctx, "/message/query", url.Values{}, body)
		if err != nil {
			return nil, nil, err
		}
		return jsonResult(data), nil, nil
	})
}
func (s *Server) addInteractiveTools(out *mcp.Server) {
	// Intentionally empty until mutation policy and freshness tests land.
}
func textResult(v string) *mcp.CallToolResult {
	return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: v}}}
}
func jsonResult(v json.RawMessage) *mcp.CallToolResult { return textResult(string(v)) }
