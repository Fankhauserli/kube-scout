package agent

import "fmt"

func (a *Agent) buildPrompt(history []string) string {
	toolDescriptions := a.getToolDescriptions()

	prompt := `You are an AI agent that can use tools to complete tasks.

Available tools:
` + toolDescriptions + `

Instructions:
1. If you need information, use the appropriate tool
2. After using a tool, wait for the result and use it to continue
3. When you have enough information to answer the original question, provide the final answer
4. Always respond in JSON format with this structure:
{
  "tool_name": "name_of_tool",
  "input": "string or JSON object for the tool",
  "thought": "Your reasoning"
}
5. When providing the final answer, use this format:
{
  "final_answer": "Your complete answer here"
}

Conversation history:
`
	for i, entry := range history {
		prompt += fmt.Sprintf("%d. %s\n", i+1, entry)
	}

	prompt += "\nWhat should be your next action?"

	return prompt
}

func (a *Agent) getToolDescriptions() string {
	var descriptions string
	for _, tool := range a.tools.List() {
		descriptions += fmt.Sprintf("- %s: %s\n", tool.Name(), tool.Description())
	}
	return descriptions
}

func (a *Agent) getToolNames() []string {
	var names []string
	for _, tool := range a.tools.List() {
		names = append(names, tool.Name())
	}
	return names
}
