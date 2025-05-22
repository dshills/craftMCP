package llm

import (
	"context"
	"fmt"

	"github.com/dshills/craftMCP/mcp"
)

func Call(ctx context.Context, req mcp.LLMRequest) (*mcp.Response, error) {
	switch req.Provider {
	case "openai":
		return callOpenAI(ctx, req)
	case "anthropic":
		return callAnthropic(ctx, req)
	case "google":
		return callGoogle(ctx, req)
	case "mistral":
		return callMistral(ctx, req)
	case "ollama":
		return callOllama(ctx, req)
	default:
		return nil, fmt.Errorf("unknown provider: %s", req.Provider)
	}
}

func promptWithContext(req mcp.LLMRequest) string {
	if len(req.Context) == 0 {
		return req.Prompt
	}
	context := ""
	for _, c := range req.Context {
		context += c + "\n"
	}
	return context + req.Prompt
}
