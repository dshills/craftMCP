package extractor

import (
	"fmt"
	"strings"

	"github.com/dshills/craftMCP/mcp"
)

type GoogleExtractor struct{}

func (g *GoogleExtractor) ContentFromResult(result map[string]any) (string, error) {
	candidates, ok := result["candidates"].([]any)
	if !ok || len(candidates) == 0 {
		return "", fmt.Errorf("google: no candidates")
	}
	cand0, ok := candidates[0].(map[string]any)
	if !ok {
		return "", fmt.Errorf("google: invalid candidate format")
	}
	content, ok := cand0["content"].(map[string]any)
	if !ok {
		return "", fmt.Errorf("google: missing content")
	}
	parts, ok := content["parts"].([]any)
	if !ok || len(parts) == 0 {
		return "", fmt.Errorf("google: missing parts")
	}
	var b strings.Builder
	for _, part := range parts {
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

func (g *GoogleExtractor) ToolCallsFromResult(result map[string]any) []mcp.ToolCall {
	return nil
}

func (g *GoogleExtractor) ExtractLLMResult(result map[string]any) mcp.LLMResult {
	res := mcp.LLMResult{
		ID:    toString(result["id"]),
		Model: toString(result["model"]),
	}

	// PaLM-style text is under `candidates[0].content`
	if candidates, ok := result["candidates"].([]any); ok && len(candidates) > 0 {
		if cand, ok := candidates[0].(map[string]any); ok {
			res.FinishReason = toString(cand["finish_reason"])
			res.Content = toString(cand["content"])

			// Tool use may be present under `function_call`
			if funcCall, ok := cand["function_call"].(map[string]any); ok {
				res.ToolCalls = append(res.ToolCalls, mcp.ToolCall{
					ID:       "", // Google doesn't include a tool ID
					Type:     "function",
					Function: toString(funcCall["name"]),
					Args:     funcCall["args"],
				})
			}
		}
	}

	// Usage stats if present
	if usage, ok := result["usage_metadata"].(map[string]any); ok {
		res.InputTokens = toInt(usage["prompt_token_count"])
		res.OutputTokens = toInt(usage["candidates_token_count"])
	}

	return res
}
