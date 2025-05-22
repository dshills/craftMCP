package llm

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"strings"

	"github.com/dshills/craftMCP/extractor"
	"github.com/dshills/craftMCP/mcp"
)

const (
	openAIURL   = "https://api.openai.com/v1/chat/completions"
	openaiModel = "o4-mini"
)

type openAIRequest struct {
	Model         string          `json:"model,omitempty"`
	Messages      []openAIMessage `json:"messages,omitempty"`
	Tools         []openAITool    `json:"tools,omitempty"`
	ToolChoice    string          `json:"tool_choice,omitempty"`
	MaxCompTokens int             `json:"max_completion_tokens,omitempty"`
}

type openAIMessage struct {
	Role       string `json:"role,omitempty"`
	Content    string `json:"content,omitempty"`
	ToolCallID string `json:"tool_call_id,omitempty"`
}

type openAITool struct {
	Type     string `json:"type,omitempty"`
	Function struct {
		Name        string         `json:"name,omitempty"`
		Description string         `json:"description,omitempty"`
		Parameters  map[string]any `json:"parameters"`
	} `json:"function"`
}

func toolDefToOpenAITool(def mcp.ToolDefinition) openAITool {
	param := map[string]any{
		"type":       "object",
		"properties": def.Parameters,
	}
	if len(def.Required) > 0 {
		param["required"] = def.Required
	}
	tool := openAITool{}
	tool.Type = "function"
	tool.Function.Name = def.Name
	tool.Function.Description = def.Description
	tool.Function.Parameters = param
	return tool
}

func callOpenAI(ctx context.Context, llmReq mcp.LLMRequest) (*mcp.Response, error) {
	if llmReq.Model == "" {
		llmReq.Model = openaiModel
	}

	reqBody := openAIRequest{
		Model: llmReq.Model,
	}
	msg := openAIMessage{
		Role:    "user",
		Content: promptWithContext(llmReq),
	}
	reqBody.Messages = append(reqBody.Messages, msg)
	for _, tr := range llmReq.ToolResults {
		msg := openAIMessage{Role: "tool", ToolCallID: tr.ID, Content: tr.Content}
		reqBody.Messages = append(reqBody.Messages, msg)
	}
	for _, t := range llmReq.Tools {
		reqBody.Tools = append(reqBody.Tools, toolDefToOpenAITool(t))
	}
	if llmReq.MaxTokens != -1 {
		reqBody.MaxCompTokens = llmReq.MaxTokens
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", openAIURL, strings.NewReader(string(jsonBody)))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+os.Getenv("OPENAI_API_KEY"))
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
	ext := extractor.OpenAIExtractor{}
	mcpResponse.Content, _ = ext.ContentFromResult(mcpResponse.Result.(map[string]any))
	mcpResponse.ToolCalls = ext.ToolCallsFromResult(mcpResponse.Result.(map[string]any))
	return &mcpResponse, nil
}
