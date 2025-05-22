package extractor

import (
	"encoding/json"
	"fmt"

	"github.com/dshills/craftMCP/mcp"
)

type OpenAIExtractor struct{}

func (o *OpenAIExtractor) ExtractLLMResult(result map[string]any) mcp.LLMResult {
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

	if toolCalls := getArrayFromChoices(result["choices"], "message", "tool_calls"); len(toolCalls) > 0 {
		for _, tc := range toolCalls {
			if tcMap, ok := tc.(map[string]any); ok {
				res.ToolCalls = append(res.ToolCalls, mcp.ToolCall{
					ID:       toString(tcMap["id"]),
					Type:     toString(tcMap["type"]),
					Function: toStringFromMap(tcMap, "function", "name"),
					Args:     extractArgs(tcMap),
				})
			}
		}
	}

	return res
}

func (o *OpenAIExtractor) ContentFromResult(result map[string]any) (string, error) {
	choices, ok := result["choices"].([]any)
	if !ok || len(choices) == 0 {
		return "", fmt.Errorf("openai: no choices or bad format")
	}
	choice0, ok := choices[0].(map[string]any)
	if !ok {
		return "", fmt.Errorf("openai: invalid choice format")
	}
	msg, ok := choice0["message"].(map[string]any)
	if !ok {
		return "", fmt.Errorf("openai: missing message field")
	}
	content, ok := msg["content"].(string)
	if !ok {
		return "", fmt.Errorf("openai: missing content")
	}
	return content, nil
}

func (o *OpenAIExtractor) ToolCallsFromResult(result map[string]any) []mcp.ToolCall {
	choices, ok := result["choices"].([]any)
	if !ok || len(choices) == 0 {
		return nil
	}
	choice0, ok := choices[0].(map[string]any)
	if !ok {
		return nil
	}
	msg, ok := choice0["message"].(map[string]any)
	if !ok {
		return nil
	}
	rawCalls, ok := msg["tool_calls"].([]any)
	if !ok {
		return nil
	}

	var calls []mcp.ToolCall

	for _, rc := range rawCalls {
		callMap, ok := rc.(map[string]any)
		if !ok {
			continue
		}
		funcMap, ok := callMap["function"].(map[string]any)
		if !ok {
			continue
		}

		// Decode the JSON string in "arguments"
		argsStr, ok := funcMap["arguments"].(string)
		if !ok {
			continue
		}

		var args any
		if err := json.Unmarshal([]byte(argsStr), &args); err != nil {
			continue // skip if arguments can't be parsed
		}

		calls = append(calls, mcp.ToolCall{
			ID:       callMap["id"].(string),
			Type:     callMap["type"].(string),
			Function: funcMap["name"].(string),
			Args:     args,
		})
	}

	return calls
}
