package workflow

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/dshills/craftMCP/extractor"
	"github.com/dshills/craftMCP/llm"
	"github.com/dshills/craftMCP/mcp"
	"github.com/dshills/craftMCP/mcpregistry"

	"github.com/google/uuid"
)

type Runner struct {
	Registry *mcpregistry.MCPRegistry
}

func LoadDefinition(path string) (Definition, error) {
	jsonFile, err := os.Open(path)
	if err != nil {
		return Definition{}, fmt.Errorf("failed to open file: %w", err)
	}
	defer func() {
		if err := jsonFile.Close(); err != nil {
			log.Printf("Error closing file: %v", err)
		}
	}()

	var def Definition
	if err := json.NewDecoder(jsonFile).Decode(&def); err != nil {
		return Definition{}, fmt.Errorf("failed to unmarshal JSON: %w", err)
	}
	return def, nil
}

func NewRunner(reg *mcpregistry.MCPRegistry) *Runner {
	return &Runner{
		Registry: reg,
	}
}

func (w *Runner) RunDefinition(ctx context.Context, def Definition) error {
	batchID := uuid.New().String()

	var (
		response *mcp.Response
		err      error
	)
	for i, step := range def.Steps {
		log.Printf("Step %v: %+v", i, step)
		switch step.StepType {
		case StepTypeLLM:
			response, err = w.callLLM(ctx, step, batchID, []mcp.ToolResult{})
		case StepTypeMCPTool, StepTypeMCPPrompt, StepTypeMCPResource:
			response, err = w.callMCP(ctx, step, batchID)
		default:
			return fmt.Errorf("unknown step type %s", step.StepType)
		}
		if err != nil {
			return err
		}
		log.Printf("Step %v: Response: %+v", i, response)
	}

	return nil
}

func (w *Runner) callLLM(ctx context.Context, step Step, id string, toolResults []mcp.ToolResult, context ...string) (*mcp.Response, error) {
	req := StepToLLMRequest(step, w.Registry.Tools(), toolResults, context)
	resp, err := llm.Call(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to call LLM: %w", err)
	}
	var result mcp.LLMResult
	ext := extractor.GetModelExtractor(step.LLMStep.Provider)
	if ext != nil {
		if resultMap, ok := resp.Result.(map[string]any); ok {
			result = ext.ExtractLLMResult(resultMap)
		}
	}
	for _, tc := range result.ToolCalls {
		mcpReq := LLMToolToMCPTool(tc, id)
		mcpResp, err := w.callMCPByTool(ctx, tc.Function, mcpReq)
		if err != nil {
			return nil, fmt.Errorf("failed to call tool %s: %w", tc.Function, err)
		}
		log.Printf("Tool %s response: %+v", tc.Function, mcpResp)
		toolResults = append(toolResults, mcp.ToolResult{ID: tc.ID, Content: mcpResp.Content})
	}
	if len(toolResults) > 0 {
		return w.callLLM(ctx, step, id, toolResults, context...)
	}
	return resp, nil
}

func (w *Runner) callMCP(ctx context.Context, step Step, id string) (*mcp.Response, error) {
	srv, err := w.Registry.Get(step.MCPStep.ServerName)
	if err != nil {
		return nil, fmt.Errorf("unknown MCP server: %s", step.MCPStep.ServerName)
	}

	req := StepToMCPRequest(step, id)
	log.Printf("Calling %s/%s with args: %v", step.MCPStep.ServerName, step.MCPStep.Method, step.MCPStep.Args)
	return srv.Call(ctx, req)
}

func (w *Runner) callMCPByTool(ctx context.Context, toolName string, req mcp.Request) (*mcp.Response, error) {
	srv := w.Registry.GetByTool(toolName)
	if srv == nil {
		return nil, fmt.Errorf("unknown MCP server for tool: %s", toolName)
	}
	log.Printf("Calling MCP %+v", req)
	return srv.Call(ctx, req)
}
