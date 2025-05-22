package extractor

import (
	"fmt"

	"github.com/dshills/craftMCP/mcp"
)

func ConvertMCPListTool(tool map[string]any) (mcp.ToolDefinition, error) {
	inputSchema, ok := tool["inputSchema"].(map[string]any)
	if !ok {
		return mcp.ToolDefinition{}, fmt.Errorf("missing or invalid 'inputSchema'")
	}

	properties, ok := inputSchema["properties"].(map[string]any)
	if !ok {
		return mcp.ToolDefinition{}, fmt.Errorf("missing or invalid 'properties'")
	}

	var required []string
	if r, ok := inputSchema["required"].([]any); ok {
		for _, v := range r {
			if s, ok := v.(string); ok {
				required = append(required, s)
			}
		}
	}

	name, _ := tool["name"].(string)
	description, _ := tool["description"].(string)

	return mcp.ToolDefinition{
		Name:        name,
		Description: description,
		Parameters:  properties,
		Required:    required,
	}, nil
}

func ConvertAllToolsFromResult(source string, result map[string]any) ([]mcp.ToolDefinition, error) {
	rawTools, ok := result["tools"].([]any)
	if !ok {
		return nil, fmt.Errorf("'tools' missing or invalid in result")
	}

	var defs []mcp.ToolDefinition
	for _, item := range rawTools {
		if tool, ok := item.(map[string]any); ok {
			def, err := ConvertMCPListTool(tool)
			if err != nil {
				return nil, fmt.Errorf("tool %v conversion error: %w", tool["name"], err)
			}
			def.Source = source
			defs = append(defs, def)
		}
	}
	return defs, nil
}

func toString(val any) string {
	if s, ok := val.(string); ok {
		return s
	}
	return ""
}

func toInt(val any) int {
	switch v := val.(type) {
	case float64:
		return int(v)
	case int:
		return v
	default:
		return 0
	}
}

func toStringFromMap(m map[string]any, key string, subkey string) string {
	if val, ok := m[key]; ok {
		if submap, ok := val.(map[string]any); ok {
			return toString(submap[subkey])
		}
	}
	return ""
}

func toStringFromArray(val any, keys ...string) string {
	arr, ok := val.([]any)
	if !ok || len(arr) == 0 {
		return ""
	}
	obj, ok := arr[0].(map[string]any)
	if !ok {
		return ""
	}
	for _, key := range keys {
		if inner, ok := obj[key].(map[string]any); ok {
			obj = inner
		} else if str, ok := obj[key].(string); ok {
			return str
		} else {
			return ""
		}
	}
	return ""
}

func getArrayFromChoices(val any, keys ...string) []any {
	arr, ok := val.([]any)
	if !ok || len(arr) == 0 {
		return nil
	}
	obj, ok := arr[0].(map[string]any)
	if !ok {
		return nil
	}
	for _, key := range keys {
		if next, ok := obj[key].(map[string]any); ok {
			obj = next
		} else if nextArr, ok := obj[key].([]any); ok {
			return nextArr
		}
	}
	return nil
}

func extractArgs(tc map[string]any) any {
	if fn, ok := tc["function"].(map[string]any); ok {
		return fn["arguments"]
	}
	return nil
}
