package extractor

import (
	"github.com/dshills/craftMCP/mcp"
)

type MistralExtractor struct{}

func (m *MistralExtractor) ContentFromResult(result map[string]any) (string, error) {
	return (&OpenAIExtractor{}).ContentFromResult(result)
}

func (m *MistralExtractor) ToolCallsFromResult(result map[string]any) []mcp.ToolCall {
	return nil
}

func (m *MistralExtractor) ExtractLLMResult(result map[string]any) mcp.LLMResult {
	res := mcp.LLMResult{
		ID:           toString(result["id"]),
		Model:        toString(result["model"]),
		FinishReason: toStringFromArray(result["choices"], "finish_reason"),
		Content:      toStringFromArray(result["choices"], "message", "content"),
	}

	if usage, ok := result["usage"].(map[string]any); ok {
		res.InputTokens = toInt(usage["prompt_tokens"])
		res.OutputTokens = toInt(usage["completion_tokens"])
	}

	// Tool calls (if using tool mode)
	if toolCalls := getArrayFromChoices(result["choices"], "message", "tool_calls"); len(toolCalls) > 0 {
		for _, tc := range toolCalls {
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
