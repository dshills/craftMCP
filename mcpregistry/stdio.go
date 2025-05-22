package mcpregistry

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/dshills/craftMCP/mcp"
)

var debug = false

func NewStdioServer(name, command string, env []string, args ...string) (mcp.MCPServer, error) {
	srv := _mcpServer{
		name:      name,
		command:   command,
		env:       env,
		args:      args,
		transport: "STDIO",
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	err := srv.startStdioServer(ctx)
	if err != nil {
		return nil, fmt.Errorf("error starting server %s: %v", name, err)
	}

	initCTX, cancelInit := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelInit()

	err = srv.initialize(initCTX)
	if err != nil {
		return nil, fmt.Errorf("error initializing server %s: %v", name, err)
	}

	toolCtx, cancelTool := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelTool()

	err = srv.getToolList(toolCtx)
	if err != nil {
		return nil, fmt.Errorf("error getting tool list: %v", err)
	}
	return &srv, nil
}

func (s *_mcpServer) startStdioServer(ctx context.Context) error {
	s.m.Lock()
	defer s.m.Unlock()
	if s.isRunning {
		return nil
	}

	s.cmd = exec.CommandContext(ctx, s.command, s.args...)
	s.cmd.Env = append(os.Environ(), s.env...)
	var err error

	s.stderr, err = s.cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("getting stderr: %w", err)
	}

	s.stdin, err = s.cmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("getting stdin: %w", err)
	}

	s.stdout, err = s.cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("getting stdout: %w", err)
	}

	if debug {
		go s.listenStderr()
	}

	if err := s.cmd.Start(); err != nil {
		return fmt.Errorf("starting cmd: %w", err)
	}

	s.isRunning = true
	return nil
}

// nolint
func (s *_mcpServer) listenStderr() {
	scanner := bufio.NewScanner(s.stderr)
	for scanner.Scan() {
		log.Printf("[%s stderr] %s", s.name, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return
	}
}

func (s *_mcpServer) stopStdioServer() error {
	s.m.Lock()
	defer s.m.Unlock()

	if !s.isRunning {
		return nil
	}

	var errs []string

	if s.stdin != nil {
		if err := s.stdin.Close(); err != nil {
			errs = append(errs, fmt.Sprintf("closing stdin: %v", err))
		}
	}

	// Try to gracefully wait for the process to exit
	if s.cmd.ProcessState == nil || !s.cmd.ProcessState.Exited() {
		if err := s.cmd.Wait(); err != nil {
			// If process was killed, don't treat as hard error
			if exitErr, ok := err.(*exec.ExitError); !ok || exitErr.ExitCode() != -1 {
				errs = append(errs, fmt.Sprintf("waiting for process: %v", err))
			}
		}
	}

	s.isRunning = false

	if len(errs) > 0 {
		return fmt.Errorf("shutdown errors: %s", strings.Join(errs, "; "))
	}

	return nil
}

func (s *_mcpServer) stdioCall(ctx context.Context, req mcp.Request) (*mcp.Response, error) {
	// Write request
	enc := json.NewEncoder(s.stdin)
	if err := enc.Encode(req); err != nil {
		return nil, fmt.Errorf("encoding request: %w", err)
	}

	// Read response with context timeout
	resultChan := make(chan mcp.Response, 1)
	errChan := make(chan error, 1)

	go s.listenStdin(resultChan, errChan)

	select {
	case <-ctx.Done():
		return nil, fmt.Errorf("context done: %w", ctx.Err())
	case err := <-errChan:
		return nil, err
	case resp := <-resultChan:
		return &resp, nil
	}
}

func (s *_mcpServer) listenStdin(resultChan chan mcp.Response, errChan chan error) {
	defer close(resultChan)
	defer close(errChan)

	decoder := json.NewDecoder(s.stdout)
	var resp mcp.Response
	if err := decoder.Decode(&resp); err != nil {
		errChan <- fmt.Errorf("decoding response: %w", err)
		return
	}
	if resp.Error != nil {
		errChan <- fmt.Errorf("response error: %s, code: %v", resp.Error.Message, resp.Error.Code)
		return
	}
	resultChan <- resp
}
