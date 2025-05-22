package llm

import (
	"bytes"
	"encoding/json"
	"net/http"
)

func EncodeJSON(v any) *bytes.Reader {
	data, _ := json.Marshal(v)
	return bytes.NewReader(data)
}

func DecodeJSON(r *http.Response) (map[string]any, error) {
	var out map[string]any
	err := json.NewDecoder(r.Body).Decode(&out)
	return out, err
}

func GetStr(m map[string]any, key, def string) string {
	if v, ok := m[key].(string); ok {
		return v
	}
	return def
}

func GetInt(m map[string]any, key string, def int) int {
	v, ok := m[key]
	if !ok {
		return def
	}
	switch val := v.(type) {
	case float64:
		return int(val)
	case float32:
		return int(val)
	case int:
		return val
	case int64:
		return int(val)
	default:
		return def
	}
}

func GetMap(m map[string]any, key string) map[string]any {
	if v, ok := m[key]; ok {
		if result, ok := v.(map[string]any); ok {
			return result
		}
	}
	return map[string]any{}
}

func GetBool(m map[string]any, key string, def bool) bool {
	if v, ok := m[key]; ok {
		if b, ok := v.(bool); ok {
			return b
		}
	}
	return def
}

func GetSlice(m map[string]any, key string) []any {
	if v, ok := m[key]; ok {
		if s, ok := v.([]any); ok {
			return s
		}
	}
	return nil
}

func GetStringSlice(m map[string]any, key string) []string {
	if v, ok := m[key]; ok {
		if raw, ok := v.([]any); ok {
			out := make([]string, 0, len(raw))
			for _, item := range raw {
				if s, ok := item.(string); ok {
					out = append(out, s)
				}
			}
			return out
		}
	}
	return nil
}
