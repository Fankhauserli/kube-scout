package agent

import (
	"context"
	"fmt"
	"log"
)

// Run executes the agent loop
func (a *Agent) Run(ctx context.Context, task string) (string, error) {
	conversationHistory := []ToolCallResponse{{Thought: task}}
	iterations := 0

	for iterations < a.maxIterations {
		iterations++
		log.Printf("[INFO] [agent] Iteration %d: Processing task", iterations)

		// Build the prompt with context
		prompt := a.buildPrompt(conversationHistory)

		// if max iterations reached, tell the llm to send a mail
		if iterations >= a.maxIterations-1 {
			prompt += "\nMaximum iterations reached. Please provide a final answer or send an email with the results."
		}

		// Get response from LLM
		response, err := a.llmClient.GenerateResponse(ctx, prompt)
		if err != nil {
			return "", fmt.Errorf("LLM error: %w", err)
		}

		log.Printf("[INFO] [agent] LLM response: %s", response)

		// Try to parse a tool call from the response
		toolCall, finalAnswer, isFinalAnswer := a.parseResponse(response)

		if isFinalAnswer {
			// Agent has provided the final answer
			return finalAnswer, nil
		}

		if toolCall == nil {
			// Couldn't parse a valid tool call
			conversationHistory = append(conversationHistory, ToolCallResponse{
				Thought: response,
				Error:   "Invalid response format. Please use JSON format with tool_name, input, and thought fields.",
			})
			continue
		}

		// Execute the tool
		tool := a.tools.Get(toolCall.ToolName)
		if tool == nil {
			conversationHistory = append(conversationHistory, ToolCallResponse{
				Thought:  toolCall.Thought,
				Input:    toolCall.Input,
				ToolName: toolCall.ToolName,
				Error:    fmt.Sprintf("Tool '%s' not found. Available tools: %v", toolCall.ToolName, a.getToolNames()),
			})
			continue
		}

		toolResult, err := tool.Execute(ctx, toolCall.Input)
		if toolResult != "" && len(toolResult) > a.maxToolResultLength {
			toolResult = toolResult[:a.maxToolResultLength] + "... (truncated)"
		}

		if err != nil {
			log.Printf("[ERROR] [agent] Tool '%s' failed: %v", toolCall.ToolName, err)
			conversationHistory = append(conversationHistory, ToolCallResponse{
				Thought:  toolCall.Thought,
				Input:    toolCall.Input,
				Output:   toolResult,
				ToolName: toolCall.ToolName,
				Error:    err.Error(),
			})
			continue
		}

		log.Printf("[INFO] [agent] Tool '%s' executed successfully", toolCall.ToolName)

		// Add result to conversation history
		conversationHistory = append(conversationHistory, ToolCallResponse{
			Thought:  toolCall.Thought,
			ToolName: toolCall.ToolName,
			Output:   toolResult,
			Input:    toolCall.Input,
		})
	}

	return "", fmt.Errorf("max iterations (%d) reached without completing the task", a.maxIterations)
}
