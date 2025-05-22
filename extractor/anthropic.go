package extractor

import (
	"fmt"
	"strings"

	"github.com/dshills/craftMCP/mcp"
)

type AnthropicExtractor struct{}

func (a *AnthropicExtractor) ContentFromResult(result map[string]any) (string, error) {
	contentList, ok := result["content"].([]any)
	if !ok || len(contentList) == 0 {
		return "", fmt.Errorf("anthropic: no content or bad format")
	}
	var b strings.Builder
	for _, part := range contentList {
		partMap, ok := part.(map[string]any)
		if !ok {
			continue
		}
		if text, ok := partMap["text"].(string); ok {
			b.WriteString(text)
		}
	}
	return b.String(), nil
}

func (a *AnthropicExtractor) ToolCallsFromResult(result map[string]any) []mcp.ToolCall {
	return nil
}

func (a *AnthropicExtractor) ExtractLLMResult(result map[string]any) mcp.LLMResult {
	res := mcp.LLMResult{
		ID:           toString(result["id"]),
		Model:        toString(result["model"]),
		FinishReason: toString(result["stop_reason"]),
	}

	if usage, ok := result["usage"].(map[string]any); ok {
		res.InputTokens = toInt(usage["input_tokens"])
		res.OutputTokens = toInt(usage["output_tokens"])
	}

	// Anthropic messages typically have a "content" array
	if contentArr, ok := result["content"].([]any); ok && len(contentArr) > 0 {
		if contentObj, ok := contentArr[0].(map[string]any); ok {
			res.Content = toString(contentObj["text"])
		}
	}

	// Anthropic tool calls are under `content[].tool_calls`
	if contentArr, ok := result["content"].([]any); ok {
		for _, msg := range contentArr {
			if msgMap, ok := msg.(map[string]any); ok {
				if toolCalls, ok := msgMap["tool_calls"].([]any); ok {
					for _, tc := range toolCalls {
						if tcMap, ok := tc.(map[string]any); ok {
							call := mcp.ToolCall{
								ID:       toString(tcMap["id"]),
								Type:     toString(tcMap["type"]),
								Function: toStringFromMap(tcMap, "function", "name"),
								Args:     tcMap["function"],
							}
							res.ToolCalls = append(res.ToolCalls, call)
						}
					}
				}
			}
		}
	}

	return res
}
