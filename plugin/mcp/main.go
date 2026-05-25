// cem-mcp — Model Context Protocol server that exposes cem as three tools:
//
//	think(prompt)  →  cem "prompt"
//	write(prompt)  →  cem -w "prompt"
//	pair(prompt)   →  cem -p "prompt"
//
// MCP hosts (Claude Desktop, MCP-compatible IDEs, Antigravity if/when it adds
// MCP support) can register this server in their config and gain a "second
// model opinion" inside any conversation.
//
// Protocol: line-delimited JSON-RPC 2.0 over stdio (per MCP spec).
//
// This is a minimal hand-rolled implementation — no external dependencies —
// so it builds anywhere with just Go. Once the ecosystem stabilises we may
// migrate to modelcontextprotocol/go-sdk.
package main

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

// ─── JSON-RPC framing ────────────────────────────────────────────────────────

type rpcReq struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id,omitempty"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

type rpcResp struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id,omitempty"`
	Result  any             `json:"result,omitempty"`
	Error   *rpcErr         `json:"error,omitempty"`
}

type rpcErr struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

// ─── MCP-specific shapes ─────────────────────────────────────────────────────

type initializeResult struct {
	ProtocolVersion string      `json:"protocolVersion"`
	ServerInfo      serverInfo  `json:"serverInfo"`
	Capabilities    capabilities `json:"capabilities"`
}

type serverInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type capabilities struct {
	Tools toolsCapability `json:"tools"`
}

type toolsCapability struct {
	ListChanged bool `json:"listChanged"`
}

type tool struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	InputSchema map[string]any `json:"inputSchema"`
}

type listToolsResult struct {
	Tools []tool `json:"tools"`
}

type callToolParams struct {
	Name      string         `json:"name"`
	Arguments map[string]any `json:"arguments"`
}

type contentBlock struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type callToolResult struct {
	Content []contentBlock `json:"content"`
	IsError bool           `json:"isError,omitempty"`
}

// ─── Server state ───────────────────────────────────────────────────────────

var cemBin = func() string {
	if v := os.Getenv("CEM_BIN"); v != "" {
		return v
	}
	return "cem"
}()

const protocolVersion = "2024-11-05" // current MCP spec at time of writing

var schema = map[string]any{
	"type": "object",
	"properties": map[string]any{
		"prompt": map[string]any{
			"type":        "string",
			"description": "The user prompt to send to cem.",
		},
	},
	"required": []string{"prompt"},
}

var tools = []tool{
	{Name: "think", Description: "Ask cem's thinker AI a question. Single-shot reasoning.", InputSchema: schema},
	{Name: "write", Description: "Ask cem's writer AI to produce code.", InputSchema: schema},
	{Name: "pair",  Description: "Run cem in pair mode (thinker → writer).", InputSchema: schema},
}

// ─── Dispatch ────────────────────────────────────────────────────────────────

func handle(req *rpcReq) *rpcResp {
	resp := &rpcResp{JSONRPC: "2.0", ID: req.ID}
	switch req.Method {
	case "initialize":
		resp.Result = initializeResult{
			ProtocolVersion: protocolVersion,
			ServerInfo:      serverInfo{Name: "cem-mcp", Version: "0.1.0"},
			Capabilities:    capabilities{Tools: toolsCapability{ListChanged: false}},
		}
	case "tools/list":
		resp.Result = listToolsResult{Tools: tools}
	case "tools/call":
		var p callToolParams
		if err := json.Unmarshal(req.Params, &p); err != nil {
			resp.Error = &rpcErr{Code: -32602, Message: "invalid params: " + err.Error()}
			return resp
		}
		out, err := runCem(p.Name, p.Arguments)
		if err != nil {
			resp.Result = callToolResult{
				Content: []contentBlock{{Type: "text", Text: out + "\n" + err.Error()}},
				IsError: true,
			}
		} else {
			resp.Result = callToolResult{Content: []contentBlock{{Type: "text", Text: out}}}
		}
	case "notifications/initialized", "ping":
		// Spec: no response needed for notifications; ping should return empty.
		if req.ID == nil {
			return nil
		}
		resp.Result = map[string]any{}
	default:
		resp.Error = &rpcErr{Code: -32601, Message: "method not found: " + req.Method}
	}
	return resp
}

func runCem(name string, args map[string]any) (string, error) {
	prompt, _ := args["prompt"].(string)
	if strings.TrimSpace(prompt) == "" {
		return "", errors.New("prompt is required")
	}
	cmdArgs := []string{}
	switch name {
	case "think":
		// no flag
	case "write":
		cmdArgs = append(cmdArgs, "-w")
	case "pair":
		cmdArgs = append(cmdArgs, "-p")
	default:
		return "", fmt.Errorf("unknown tool: %s", name)
	}
	cmdArgs = append(cmdArgs, prompt)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	cmd := exec.CommandContext(ctx, cemBin, cmdArgs...)
	out, err := cmd.CombinedOutput()
	return string(out), err
}

// ─── stdio loop ──────────────────────────────────────────────────────────────

func main() {
	reader := bufio.NewReader(os.Stdin)
	writer := bufio.NewWriter(os.Stdout)
	defer writer.Flush()

	for {
		line, err := reader.ReadBytes('\n')
		if err != nil {
			if errors.Is(err, io.EOF) {
				return
			}
			fmt.Fprintln(os.Stderr, "read error:", err)
			return
		}
		var req rpcReq
		if err := json.Unmarshal(line, &req); err != nil {
			fmt.Fprintln(os.Stderr, "json parse error:", err)
			continue
		}
		resp := handle(&req)
		if resp == nil {
			continue // notification
		}
		b, _ := json.Marshal(resp)
		writer.Write(b)
		writer.WriteByte('\n')
		writer.Flush()
	}
}
