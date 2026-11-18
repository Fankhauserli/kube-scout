package agent

import (
	"encoding/json"
	"strings"
)

func (a *Agent) parseResponse(response string) (*ToolCall, string, bool) {
	startIdx := strings.IndexByte(response, '{')
	if startIdx == -1 {
		return nil, "", false
	}

	var raw struct {
		FinalAnswer string          `json:"final_answer"`
		ToolName    string          `json:"tool_name"`
		Input       json.RawMessage `json:"input"`
		Thought     string          `json:"thought"`
	}

	// json.NewDecoder reads one complete JSON object and ignores any trailing text (like ```)
	dec := json.NewDecoder(strings.NewReader(response[startIdx:]))
	if err := dec.Decode(&raw); err != nil {
		return nil, "", false
	}

	if raw.FinalAnswer != "" {
		return nil, raw.FinalAnswer, true
	}

	var inputStr string
	if len(raw.Input) > 0 {
		if err := json.Unmarshal(raw.Input, &inputStr); err != nil {
			inputStr = string(raw.Input)
		}
	}

	if raw.ToolName == "" || inputStr == "" {
		return nil, "", false
	}

	return &ToolCall{
		ToolName: raw.ToolName,
		Input:    inputStr,
		Thought:  raw.Thought,
	}, "", false
}
