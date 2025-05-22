package mcpregistry

import (
	"context"
	"fmt"
	"io"
	"os/exec"
	"sync"

	"github.com/dshills/craftMCP/extractor"
	"github.com/dshills/craftMCP/mcp"
)

type _mcpServer struct {
	name      string
	baseURL   string
	command   string
	env       []string
	args      []string
	m         sync.Mutex
	stdin     io.WriteCloser
	stdout    io.ReadCloser
	stderr    io.ReadCloser
	isRunning bool
	cmd       *exec.Cmd
	tools     []mcp.ToolDefinition
	transport string
}

func (s *_mcpServer) Name() string {
	return s.name
}

func (s *_mcpServer) Tools() []mcp.ToolDefinition {
	return s.tools
}

func (s *_mcpServer) Call(ctx context.Context, req mcp.Request) (*mcp.Response, error) {
	if s.transport == "STDIO" {
		return s.stdioCall(ctx, req)
	}
	return s.httpCall(ctx, req)
}

func (s *_mcpServer) Stop() error {
	if s.transport == "STDIO" {
		return s.stopStdioServer()
	}
	return nil
}

func (s *_mcpServer) initialize(ctx context.Context) error {
	req := mcp.Request{
		JSONRPC: "2.0",
		ID:      "init",
		Method:  "initialize",
		Params: map[string]any{
			"protocolVersion": "2024-11-05",
			"capabilities": map[string]any{
				"roots": map[string]any{
					"listChanged": false,
				},
				"sampling": map[string]any{},
			},
			"clientInfo": map[string]any{
				"name":    "craftMCP",
				"version": "0.1.0",
			},
		},
	}
	resp, err := s.Call(ctx, req)
	if err != nil {
		return err
	}
	if resp.Error != nil {
		return fmt.Errorf("error initializing server %s: %s", s.name, resp.Error.Message)
	}
	return nil
}

func (s *_mcpServer) getToolList(ctx context.Context) error {
	req := mcp.Request{
		JSONRPC: "2.0",
		ID:      "tool-list",
		Method:  "tools/list",
	}
	resp, err := s.Call(ctx, req)
	if err != nil {
		return err
	}
	//log.Printf("Response from %s/%s: %+v", s.name, req.Method, resp.Result)
	resMap, ok := resp.Result.(map[string]any)
	if !ok {
		return fmt.Errorf("invalid response format: %T", resp.Result)
	}
	s.tools, err = extractor.ConvertAllToolsFromResult(s.name, resMap)
	return err
}
