package extractor

import (
	"fmt"

	"github.com/dshills/craftMCP/mcp"
)

type OllamaExtractor struct{}

func (o *OllamaExtractor) ContentFromResult(result map[string]any) (string, error) {
	resp, ok := result["response"].(string)
	if !ok {
		return "", fmt.Errorf("ollama: missing response field")
	}
	return resp, nil
}

func (o *OllamaExtractor) ToolCallsFromResult(result map[string]any) []mcp.ToolCall {
	return nil
}

func (o *OllamaExtractor) ExtractLLMResult(result map[string]any) mcp.LLMResult {
	res := mcp.LLMResult{
		ID:    toString(result["id"]),
		Model: toString(result["model"]),
	}

	// Ollama outputs have `done_reason` or `stop_reason`
	res.FinishReason = toString(result["done_reason"])
	if res.FinishReason == "" {
		res.FinishReason = toString(result["stop_reason"])
	}

	// Token usage (optional depending on config)
	if usage, ok := result["usage"].(map[string]any); ok {
		res.InputTokens = toInt(usage["prompt_tokens"])
		res.OutputTokens = toInt(usage["completion_tokens"])
	}

	// Content usually under `message.content`
	if message, ok := result["message"].(map[string]any); ok {
		res.Content = toString(message["content"])
	}

	// Ollama does not natively support tool calls in the same way as OpenAI/Anthropic
	// but we can future-proof this
	if calls, ok := result["tool_calls"].([]any); ok {
		for _, tc := range calls {
			if tcMap, ok := tc.(map[string]any); ok {
				res.ToolCalls = append(res.ToolCalls, mcp.ToolCall{
					ID:       toString(tcMap["id"]),
					Type:     toString(tcMap["type"]),
					Function: toStringFromMap(tcMap, "function", "name"),
					Args:     tcMap["function"],
				})
			}
		}
	}

	return res
}
