package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/dshills/craftMCP/mcpregistry"
	"github.com/dshills/craftMCP/workflow"
)

func main() {
	listTools := false
	workflowPath := ""
	runWorkflow := false
	flag.StringVar(&workflowPath, "workflow", "", "Path to workflow JSON file")
	flag.BoolVar(&listTools, "tools", false, "List all MCP tools")
	flag.BoolVar(&runWorkflow, "run", false, "Run the workflow")
	flag.Parse()

	if workflowPath == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}
	reg := mcpregistry.New()
	ctx := context.Background()

	workflowDef, err := workflow.LoadDefinition(workflowPath)
	if err != nil {
		fmt.Printf("Error loading workflow definition: %v\n", err)
		os.Exit(1)
	}

	for name, mcpSrv := range workflowDef.MCPDefs {
		if mcpSrv.Disabled {
			//log.Printf("MCP Disabled %s\n", name)
			continue
		}
		if mcpSrv.Transport != "html" {
			server, err := mcpregistry.NewStdioServer(name, mcpSrv.Command, mcpSrv.Env, mcpSrv.Args...)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			reg.Add(server)
			log.Printf("MCP Stdio: %s, %v Tools\n", name, len(server.Tools()))
		}
	}

	if listTools {
		for _, tool := range reg.Tools() {
			log.Printf("Tool %s: %s\n", tool.Source, tool.Name)
		}
	}

	if runWorkflow {
		runner := workflow.NewRunner(reg)
		if err := runner.RunDefinition(ctx, workflowDef); err != nil {
			fmt.Printf("Error running workflow: %v\n", err)
			os.Exit(1)
		}
	}

	if err := reg.StopAll(); err != nil {
		fmt.Println(err)
	}
}
