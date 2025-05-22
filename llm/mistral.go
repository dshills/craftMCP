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
	mistralURL   = "https://api.mistral.ai/v1/chat/completions"
	mistralModel = "codestral-latest"
)

func callMistral(ctx context.Context, llmReq mcp.LLMRequest) (*mcp.Response, error) {
	if llmReq.Model == "" {
		llmReq.Model = mistralModel
	}

	apiKey := os.Getenv("MISTRAL_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("missing MISTRAL_API_KEY environment variable")
	}

	body := map[string]any{
		"model": llmReq.Model,
		"messages": []map[string]any{
			{
				"role":    "user",
				"content": promptWithContext(llmReq),
			},
		},
	}
	if llmReq.MaxTokens != -1 {
		body["max_tokens"] = llmReq.MaxTokens
	}

	req, err := http.NewRequestWithContext(ctx, "POST", mistralURL, EncodeJSON(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

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
	ext := extractor.MistralExtractor{}
	mcpResponse.Content, _ = ext.ContentFromResult(mcpResponse.Result.(map[string]any))
	return &mcpResponse, nil
}
