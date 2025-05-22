package mcpregistry

import (
	"context"
	"fmt"

	"github.com/dshills/craftMCP/mcp"
)

func NewHTTPServer(name, baseURL string) mcp.MCPServer {
	srv := _mcpServer{
		name:      name,
		baseURL:   baseURL,
		transport: "HTTP",
	}
	ctx := context.Background()
	_ = srv.getToolList(ctx)
	return &srv
}

func (s *_mcpServer) httpCall(_ context.Context, _ mcp.Request) (*mcp.Response, error) {
	return nil, fmt.Errorf("httpCall not implemented")
}
