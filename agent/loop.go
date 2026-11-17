package agent

import (
	"context"
	"fmt"
	"log"
)

// Run executes the agent loop
func (a *Agent) Run(ctx context.Context, task string) (string, error) {
	conversationHistory := []string{task}
	iterations := 0

	for iterations < a.maxIterations {
		iterations++
		log.Printf("Iteration %d: Processing task", iterations)

		// Build the prompt with context
		prompt := a.buildPrompt(conversationHistory)

		// if max iterations reached, tell the llm to send a mail
		if iterations >= a.maxIterations {
			prompt += "\nMaximum iterations reached. Please provide a final answer or send an email with the results."
		}

		// Get response from LLM
		response, err := a.llmClient.GenerateResponse(ctx, prompt)
		if err != nil {
			return "", fmt.Errorf("LLM error: %w", err)
		}

		log.Printf("LLM Response: %s", response)

		// Try to parse a tool call from the response
		toolCall, isFinalAnswer := a.parseResponse(response)

		if isFinalAnswer {
			// Agent has provided the final answer
			return response, nil
		}

		if toolCall == nil {
			// Couldn't parse a valid tool call
			conversationHistory = append(conversationHistory, "Invalid response format. Please use JSON format with tool_name, input, and thought fields.")
			continue
		}

		// Execute the tool
		tool := a.tools.Get(toolCall.ToolName)
		if tool == nil {
			conversationHistory = append(conversationHistory, fmt.Sprintf("Tool '%s' not found. Available tools: %v", toolCall.ToolName, a.getToolNames()))
			continue
		}

		toolResult, err := tool.Execute(ctx, toolCall.Input)
		if err != nil {
			log.Printf("Tool '%s' failed: %v", toolCall.ToolName, err)
			conversationHistory = append(conversationHistory, fmt.Sprintf("Tool execution failed: %v", err))
			continue
		}

		log.Printf("Tool '%s' executed successfully", toolCall.ToolName)

		if len(toolResult) > a.maxToolResultLength {
			toolResult = toolResult[:a.maxToolResultLength] + "... (truncated)"
		}

		// Add result to conversation history
		resultMessage := fmt.Sprintf("Tool '%s' returned: %s", toolCall.ToolName, toolResult)
		conversationHistory = append(conversationHistory, resultMessage)
	}

	return "max iterations reached without completing the task", nil
}
