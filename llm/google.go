package llm

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/dshills/craftMCP/extractor"
	"github.com/dshills/craftMCP/mcp"
)

const (
	geminiURL   = "https://generativelanguage.googleapis.com/v1beta"
	geminiModel = "gemini-2.0-flash"
)

func callGoogle(ctx context.Context, llmReq mcp.LLMRequest) (*mcp.Response, error) {
	if llmReq.Model == "" {
		llmReq.Model = geminiModel
	}

	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("missing GEMINI_API_KEY environment variable")
	}

	url := fmt.Sprintf("%s/models/%s:generateContent?key=%s", geminiURL, llmReq.Model, apiKey)

	body := map[string]any{
		"contents": []map[string]any{
			{
				"parts": []map[string]any{
					{"text": promptWithContext(llmReq)},
				},
			},
		},
	}
	if llmReq.MaxTokens != -1 {
		body["generationConfig"] = map[string]any{"maxOutputTokens": llmReq.MaxTokens}
	}

	if len(llmReq.Tools) > 0 {
		var toolList []any
		for _, t := range llmReq.Tools {
			toolList = append(toolList, ToolDefToGeminiTool(t))
		}
		body["tools"] = toolList
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, EncodeJSON(body))
	if err != nil {
		return nil, err
	}
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
	ext := extractor.GoogleExtractor{}
	mcpResponse.Content, _ = ext.ContentFromResult(mcpResponse.Result.(map[string]any))
	return &mcpResponse, nil
}

func ToolDefToGeminiTool(tool mcp.ToolDefinition) map[string]any {
	param := map[string]any{
		"type":       "object",
		"properties": tool.Parameters,
	}
	if len(tool.Required) > 0 {
		param["required"] = tool.Required
	}

	return map[string]any{
		"functionDeclarations": []map[string]any{
			{
				"name":        tool.Name,
				"description": tool.Description,
				"parameters":  param,
			},
		},
	}
}
