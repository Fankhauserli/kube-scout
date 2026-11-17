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

	// Try to parse as tool call
	var toolCall ToolCall
	if err := json.Unmarshal([]byte(jsonStr), &toolCall); err != nil {
		return nil, false
	}

	if toolCall.ToolName == "" || toolCall.Input == "" {
		return nil, false
	}

	return &toolCall, false
}
