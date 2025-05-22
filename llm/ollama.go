package llm

import (
	"context"
	"net/http"
	"os"

	"github.com/dshills/craftMCP/extractor"
	"github.com/dshills/craftMCP/mcp"
)

const (
	ollamaModel = "codellama:13b"
)

func callOllama(ctx context.Context, llmReq mcp.LLMRequest) (*mcp.Response, error) {
	if llmReq.Model == "" {
		llmReq.Model = ollamaModel
	}

	body := map[string]any{
		"model":  llmReq.Model,
		"prompt": promptWithContext(llmReq),
		"stream": false,
	}
	if llmReq.MaxTokens != -1 {
		body["options"] = map[string]any{"num_predict": llmReq.MaxTokens}
	}

	url := os.Getenv("OLLAMA_API_URL")
	if url == "" {
		url = "http://localhost:11434/api/generate"
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
	ext := extractor.OllamaExtractor{}
	mcpResponse.Content, _ = ext.ContentFromResult(mcpResponse.Result.(map[string]any))
	return &mcpResponse, nil
}
