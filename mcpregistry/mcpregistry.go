package mcpregistry

import (
	"fmt"
	"sync"

	"github.com/dshills/craftMCP/mcp"
)

type MCPRegistry struct {
	mu      sync.RWMutex
	servers map[string]mcp.MCPServer
	tools   map[string]string
}

func New() *MCPRegistry {
	return &MCPRegistry{
		servers: make(map[string]mcp.MCPServer),
		tools:   make(map[string]string),
	}
}

func (r *MCPRegistry) Add(servers ...mcp.MCPServer) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, ser := range servers {
		r.servers[ser.Name()] = ser
		for _, tool := range ser.Tools() {
			r.tools[tool.Name] = ser.Name()
		}
	}
}

func (r *MCPRegistry) Get(name string) (mcp.MCPServer, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if srv, ok := r.servers[name]; ok {
		return srv, nil
	}
	return nil, fmt.Errorf("server %s not found", name)
}

func (r *MCPRegistry) GetByTool(toolName string) mcp.MCPServer {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.servers[r.tools[toolName]]
}

func (r *MCPRegistry) List() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	names := make([]string, 0, len(r.servers))
	for k := range r.servers {
		names = append(names, k)
	}
	return names
}

func (r *MCPRegistry) Tools() []mcp.ToolDefinition {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var allTools []mcp.ToolDefinition
	for _, server := range r.servers {
		allTools = append(allTools, server.Tools()...)
	}
	return allTools
}

func (r *MCPRegistry) StopAll() error {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var errs []string
	for name, srv := range r.servers {
		if err := srv.Stop(); err != nil {
			errs = append(errs, fmt.Sprintf("error stopping server %s: %v\n", name, err))
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("errors stopping servers: %v", errs)
	}
	return nil
}
