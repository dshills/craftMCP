package mcp

import (
	"context"
	"fmt"
	"strings"
)

type Request struct {
	JSONRPC string `json:"jsonrpc"`
	ID      any    `json:"id"`
	Method  string `json:"method"`
	Params  any    `json:"params,omitempty"`
}

type Response struct {
	JSONRPC   string     `json:"jsonrpc"`
	ID        any        `json:"id"`
	Result    any        `json:"result,omitempty"`
	Error     *Error     `json:"error,omitempty"`
	Content   string     `json:"content,omitempty"`
	ToolCalls []ToolCall `json:"tool_calls,omitempty"`
}

type Error struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type MCPServer interface {
	Name() string
	Tools() []ToolDefinition
	Call(ctx context.Context, req Request) (*Response, error)
	Stop() error
}

type ToolDefinition struct {
	Source      string         `json:"source"`
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Parameters  map[string]any `json:"input_schema"`
	Required    []string       `json:"required"`
}

func PrettyPrintParams(m map[string]any, indent string) string {
	builder := strings.Builder{}
	for k, v := range m {
		k = strings.TrimSpace(k)
		switch val := v.(type) {
		case map[string]any:
			builder.WriteString(fmt.Sprintf("%s%s:\n", indent, k))
			PrettyPrintParams(val, indent+"  ")
		case []any:
			fmt.Printf("%s%s:\n", indent, k)
			for i, item := range val {
				builder.WriteString(fmt.Sprintf("%s  - [%d]:\n", indent, i))
				if itemMap, ok := item.(map[string]any); ok {
					PrettyPrintParams(itemMap, indent+"    ")
				} else {
					builder.WriteString(fmt.Sprintf("%s    %v\n", indent, item))
				}
			}
		default:
			builder.WriteString(fmt.Sprintf("%s%s: %v\n", indent, k, val))
		}
	}
	return builder.String()
}

type LLMRequest struct {
	BaseURL     string           `json:"base_url"`
	Provider    string           `json:"provider"`
	Model       string           `json:"model"`
	MaxTokens   int              `json:"max_tokens"`
	Prompt      string           `json:"prompt"`
	Tools       []ToolDefinition `json:"tools"`
	Context     []string         `json:"context"`
	ToolResults []ToolResult     `json:"tool_results"`
}

type ToolResult struct {
	ID      string `json:"id"`
	Content string `json:"content"`
}

type LLMResult struct {
	ID           string     `json:"id"`
	Model        string     `json:"model"`
	FinishReason string     `json:"finish_reason"`
	InputTokens  int        `json:"input_tokens"`
	OutputTokens int        `json:"output_tokens"`
	Content      string     `json:"content"`
	ToolCalls    []ToolCall `json:"tool_calls"`
	Error        string     `json:"error"`
	ErrorCode    int        `json:"error_code"`
}

type ToolCall struct {
	ID       string `json:"id"`
	Type     string `json:"type"`
	Function string `json:"function"`
	Args     any    `json:"args"`
}
