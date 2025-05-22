package workflow

const (
	StepTypeMCPTool     = "mcp_tool"
	StepTypeMCPPrompt   = "mcp_prompt"
	StepTypeMCPResource = "mcp_resource"
	StepTypeLLM         = "llm"
	StepTypeAPI         = "api"
)

type Definition struct {
	MCPDefs map[string]MCPDef `json:"mcp_servers"`
	Steps   []Step            `json:"steps"`
}

type Step struct {
	StepType string  `json:"step_type"`
	Ref      string  `json:"ref"`
	UseRef   string  `json:"use_ref"`
	MCPDef   MCPDef  `json:"mcp_server"`
	LLMStep  LLMStep `json:"llm_step"`
	MCPStep  MCPStep `json:"mcp_step"`
	APIStep  APIStep `json:"api_step"`
}

type MCPDef struct {
	Command   string   `json:"command"`
	Args      []string `json:"args"`
	Transport string   `json:"transport"`
	URL       string   `json:"url"`
	Env       []string `json:"env"`
	Disabled  bool     `json:"disabled"`
}

type APIServer struct {
	Name    string `json:"name"`
	BaseURL string `json:"base_url"`
}

type LLMStep struct {
	Provider  string `json:"provider"`
	Model     string `json:"model"`
	MaxTokens int    `json:"max_tokens"`
	Prompt    string `json:"prompt"`
}

type MCPStep struct {
	Method     string         `json:"method"`
	ServerName string         `json:"server_name"`
	Name       string         `json:"name"`
	Args       map[string]any `json:"args"`
	URI        string         `json:"uri"`
}

type APIStep struct {
	ServerName string         `json:"server_name"`
	Endpoint   string         `json:"endpoint"`
	Body       map[string]any `json:"body"`
}
