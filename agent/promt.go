package agent

import (
	"encoding/json"
	"fmt"
)

func (a *Agent) buildPrompt(history []ToolCallResponse) string {
	toolDescriptions := a.getToolDescriptions()

	prompt := `Tools:
` + toolDescriptions + `
Respond ONLY with JSON:
Tool call: {"tool_name":"...","input":"string or JSON object","thought":"..."}
Final answer: {"final_answer":"..."}

History:
`
	for i, entry := range history {
		jsonEntry, err := json.Marshal(entry)
		if err != nil {
			prompt += fmt.Sprintf("%d. Error marshaling entry\n", i+1)
		} else {
			prompt += fmt.Sprintf("%d. %s\n", i+1, jsonEntry)
		}
	}

	prompt += "\nNext action:"

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
