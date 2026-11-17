package agent

import "encoding/json"

func (a *Agent) parseResponse(response string) (*ToolCall, bool) {
	// Try to find JSON in the response
	var jsonStr string

	// Look for JSON object in the response
	startIdx := -1
	for i := 0; i < len(response); i++ {
		if response[i] == '{' {
			startIdx = i
			break
		}
	}

	if startIdx == -1 {
		return nil, false
	}

	// Extract JSON
	braceCount := 0
	endIdx := -1
	for i := startIdx; i < len(response); i++ {
		if response[i] == '{' {
			braceCount++
		} else if response[i] == '}' {
			braceCount--
			if braceCount == 0 {
				endIdx = i + 1
				break
			}
		}
	}

	if endIdx == -1 {
		return nil, false
	}

	jsonStr = response[startIdx:endIdx]

	// Check if this is a final answer
	var finalAnswerCheck map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &finalAnswerCheck); err == nil {
		if _, hasFinalAnswer := finalAnswerCheck["final_answer"]; hasFinalAnswer {
			return nil, true
		}
	}

	var raw struct {
		ToolName string          `json:"tool_name"`
		Input    json.RawMessage `json:"input"`
		Thought  string          `json:"thought"`
	}
	if err := json.Unmarshal([]byte(jsonStr), &raw); err != nil {
		return nil, false
	}

	var inputStr string
	if len(raw.Input) > 0 {
		// Unquote if the LLM sent a JSON string ("get pods -A");
		// otherwise keep the JSON object/array as a raw JSON string.
		if err := json.Unmarshal(raw.Input, &inputStr); err != nil {
			inputStr = string(raw.Input)
		}
	}

	if raw.ToolName == "" || inputStr == "" {
		return nil, false
	}

	return &ToolCall{
		ToolName: raw.ToolName,
		Input:    inputStr,
		Thought:  raw.Thought,
	}, false
}
