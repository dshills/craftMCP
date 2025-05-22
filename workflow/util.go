package workflow

import "github.com/dshills/craftMCP/mcp"

func StepToMCPRequest(step Step, id string) mcp.Request {
	req := mcp.Request{JSONRPC: "2.0", ID: id, Method: step.MCPStep.Method}
	switch step.StepType {
	case StepTypeMCPTool, StepTypeMCPPrompt:
		params := make(map[string]any)
		params["name"] = step.MCPStep.Name
		params["arguments"] = step.MCPStep.Args
		req.Params = params
		return req

	case StepTypeMCPResource:
		params := make(map[string]any)
		params["uri"] = step.MCPStep.URI
		req.Params = params
		return req
	}

	req.Method = step.MCPStep.Method
	req.Params = step.MCPStep.Args
	return req
}

func StepToLLMRequest(step Step, tools []mcp.ToolDefinition, toolResults []mcp.ToolResult, context []string) mcp.LLMRequest {
	return mcp.LLMRequest{
		Provider:    step.LLMStep.Provider,
		Model:       step.LLMStep.Model,
		MaxTokens:   step.LLMStep.MaxTokens,
		Prompt:      step.LLMStep.Prompt,
		Tools:       tools,
		Context:     context,
		ToolResults: toolResults,
	}
}

func LLMToolToMCPTool(tc mcp.ToolCall, id string) mcp.Request {
	req := mcp.Request{
		JSONRPC: "2.0",
		ID:      id,
		Method:  "tools/call",
	}
	params := make(map[string]any)
	params["name"] = tc.Function
	params["arguments"] = tc.Args
	req.Params = params
	return req
}
