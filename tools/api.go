package tool

import (
	"bytes"
	"context"
	"io"
	"log"
	"net/http"
	"os"
)

type ApiTool struct {
	name        string
	description string
	Endpoint    string
	ApiKeyEnv   string
	method      string
}

func (c *ApiTool) Name() string {
	return c.name
}

func (c *ApiTool) Description() string {
	return c.description
}

func (c *ApiTool) Execute(ctx context.Context, input string) (string, error) {
	httpClient := &http.Client{}

	body := bytes.NewReader([]byte(input))

	httpReq, err := http.NewRequestWithContext(ctx, c.method, c.Endpoint, body)
	if err != nil {
		return "", err
	}

	apiKey := ""
	if c.ApiKeyEnv != "" {
		apiKey = os.Getenv(c.ApiKeyEnv)
		if apiKey != "" {
			httpReq.Header.Set("Authorization", "Bearer "+apiKey)
		}
	}

	resp, err := httpClient.Do(httpReq)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// Read the response
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	log.Printf(" [tool: api] %s", string(respBody))
	return string(respBody), nil
}

func NewApiTool(name, description, endpoint, apiKeyEnv, method string) *ApiTool {
	return &ApiTool{
		name:        name,
		description: description,
		Endpoint:    endpoint,
		ApiKeyEnv:   apiKeyEnv,
		method:      method,
	}
}
