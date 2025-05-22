package llm

import (
	"context"
	"net/http"
	"os"

	"github.com/dshills/craftMCP/extractor"
	"github.com/dshills/craftMCP/mcp"
)

const (
	anthropicURL   = "https://api.anthropic.com/v1/messages"
	anthropicModel = "claude-3-5-haiku-latest"
)

func callAnthropic(ctx context.Context, llmReq mcp.LLMRequest) (*mcp.Response, error) {
	if llmReq.Model == "" {
		llmReq.Model = anthropicModel
	}

	body := map[string]any{
		"model":      llmReq.Model,
		"max_tokens": llmReq.MaxTokens,
		"messages": []map[string]string{
			{"role": "user", "content": promptWithContext(llmReq)},
		},
	}

	if len(llmReq.Tools) > 0 {
		var toolList []any
		for _, t := range llmReq.Tools {
			toolList = append(toolList, ToolDefToAnthropicTool(t))
		}
		body["tools"] = toolList
	}

	req, err := http.NewRequestWithContext(ctx, "POST", anthropicURL, EncodeJSON(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("x-api-key", os.Getenv("ANTHROPIC_API_KEY"))
	req.Header.Set("anthropic-version", "2023-06-01")
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	mcpResponse := mcp.Response{}
	mcpResponse.Result, err = DecodeJSON(resp)
	if err != nil {
		return nil, err
	}
	ext := extractor.AnthropicExtractor{}
	mcpResponse.Content, _ = ext.ContentFromResult(mcpResponse.Result.(map[string]any))
	return &mcpResponse, nil
}

func ToolDefToAnthropicTool(tool mcp.ToolDefinition) map[string]any {
	schema := map[string]any{
		"type":       "object",
		"properties": tool.Parameters,
	}
	if len(tool.Required) > 0 {
		schema["required"] = tool.Required
	}

	return map[string]any{
		"name":         tool.Name,
		"description":  tool.Description,
		"input_schema": schema,
	}
}
