package agent

import (
	"context"

	tool "github.com/Fankhauserli/kube-scout/tools"
)

// Agent represents an AI agent that can use tools to complete tasks
type Agent struct {
	tools               *tool.ToolRegistry
	llmClient           LLMClient
	maxIterations       int
	maxToolResultLength int
}

// LLMClient represents a language model interface
type LLMClient interface {
	// GenerateResponse sends a prompt to the LLM and gets back a response
	GenerateResponse(ctx context.Context, prompt string) (string, error)
}

// ToolCall represents a decision by the agent to call a tool
type ToolCall struct {
	ToolName string `json:"tool_name"`
	Input    string `json:"input"`
	Thought  string `json:"thought"`
}

// ToolResult holds the result of executing a tool
type ToolResult struct {
	ToolName string
	Output   string
	Error    error
}

type ToolCallResponse struct {
	Thought  string `json:"thought"`
	ToolName string `json:"tool_name,omitempty"`
	Input    string `json:"input,omitempty"`
	Error    string `json:"error,omitempty"`
	Output   string `json:"output,omitempty"`
}

func NewAgent(tools *tool.ToolRegistry, llmClient LLMClient, maxIterations int, maxToolResultLength int) *Agent {
	return &Agent{
		tools:               tools,
		llmClient:           llmClient,
		maxIterations:       maxIterations,
		maxToolResultLength: maxToolResultLength,
	}
}
