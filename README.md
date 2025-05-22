# craftMCP

craftMCP is a command-line tool for orchestrating interactions between large language models (LLMs) and Model Context Protocol (MCP) servers using JSON-defined workflows.

## Features

 - Multi-provider LLM support: OpenAI, Anthropic, Google Gemini, Mistral, Ollama
 - Dynamic tool discovery and invocation via MCP servers
 - JSON-based workflow definitions for chaining LLM calls and tool executions
 - STDIO-based MCP server integration (HTTP transport not yet implemented)

## Prerequisites

 - Go 1.24+ installed
 - (Optional) Node.js and npm for Node-based MCP servers
 - Environment variables for LLM providers:
   - `OPENAI_API_KEY` for OpenAI
   - `ANTHROPIC_API_KEY` for Anthropic
   - `GEMINI_API_KEY` for Google Gemini
   - `MISTRAL_API_KEY` for Mistral
   - `OLLAMA_API_URL` for Ollama (defaults to `http://localhost:11434/api/generate`)
 - External MCP servers available in your `PATH`, for example:
   - `go-mcp` (install with `go install github.com/dshills/go-mcp@latest`)
   - Node-based servers via `npx`, e.g., `npx -y @modelcontextprotocol/server-filesystem`

## Installation

 Clone the repository and build the binary:
 ```bash
 git clone https://github.com/dshills/craftMCP.git
 cd craftMCP
 go build -o craftMCP
 ```

## Usage

### List available MCP tools
 ```bash
 ./craftMCP -workflow scripts/mcp-servers.json -tools
 ```

### Run a workflow
 1. Define a workflow in JSON (see `scripts/workflow.json` for an example).
 2. Run:
    ```bash
    ./craftMCP -workflow path/to/workflow.json -run
    ```

## Workflow Definition

 A workflow JSON has two main sections:

 - `mcp_servers`: Map of server definitions with fields:
   - `command`, `args` (for STDIO transport)
   - `url` (for HTTP transport, not yet implemented)
   - `transport` (default is STDIO)
   - `env` (optional environment variables)
   - `disabled` (boolean)
 - `steps`: Ordered list of steps, each with:
   - `step_type`: `llm`, `mcp_tool`, `mcp_prompt`, `mcp_resource`, or `api`
   - `llm_step`: For LLM calls (`provider`, `model`, `max_tokens`, `prompt`)
   - `mcp_step`: For MCP calls (`server_name`, `method`, `name`, `args`, `uri`)
   - `api_step`: For HTTP API calls (not yet implemented)

 Example:
 ```json
 {
   "mcp_servers": {
     "go-mcp": {
       "command": "go-mcp",
       "args": []
     }
   },
   "steps": [
     {
       "step_type": "llm",
       "llm_step": {
         "provider": "openai",
         "model": "o4-mini",
         "max_tokens": -1,
         "prompt": "What is the weather in Paris?"
       }
     },
     {
       "step_type": "mcp_tool",
       "mcp_step": {
         "server_name": "go-mcp",
         "method": "geocode",
         "name": "geocode",
         "args": { "query": "Paris" }
       }
     }
   ]
 }
 ```

## Project Structure

 - `main.go`: CLI entry point
 - `mcpregistry/`: MCP server registry and STDIO transport
 - `workflow/`: Workflow loader and runner
 - `llm/`: Provider-specific LLM integrations
 - `extractor/`: Parsing and extracting results from LLM responses
 - `scripts/`: Example workflows and server configurations

## Contributing

 Contributions are welcome! Please open issues or pull requests.