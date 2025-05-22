package extractor

import (
	"strings"

	"github.com/dshills/craftMCP/mcp"
)

type Extractor interface {
	ContentFromResult(result map[string]any) (string, error)
	ToolCallsFromResult(result map[string]any) []mcp.ToolCall
	ExtractLLMResult(result map[string]any) mcp.LLMResult
}

func GetModelExtractor(provider string) Extractor {
	switch strings.ToLower(provider) {
	case "openai":
		return &OpenAIExtractor{}
	case "anthropic":
		return &AnthropicExtractor{}
	case "google":
		return &GoogleExtractor{}
	case "mistral":
		return &MistralExtractor{}
	case "ollama":
		return &OllamaExtractor{}
	default:
		return nil
	}
}

func safeString(v any) string {
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}
