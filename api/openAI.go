package api

import (
	"context"
	"os"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
	"github.com/openai/openai-go/v3/responses"
)

// OpenAIClient implements LLMClient using OpenAI
type OpenAIClient struct {
	model  string
	client *openai.Client
}

func NewOpenAIClient(baseURL, model, apiKeyEnv string) *OpenAIClient {
	openaiClient := openai.NewClient(
		option.WithBaseURL(baseURL),
	)
	if apiKeyEnv != "" {
		apikey := os.Getenv(apiKeyEnv)
		openaiClient = openai.NewClient(
			option.WithAPIKey(apikey),
			option.WithBaseURL(baseURL),
		)
	}
	return &OpenAIClient{
		model:  model,
		client: &openaiClient,
	}
}

func (o *OpenAIClient) GenerateResponse(ctx context.Context, prompt string) (string, error) {
	resp, err := o.client.Responses.New(ctx, responses.ResponseNewParams{
		Input: responses.ResponseNewParamsInputUnion{OfString: openai.String(prompt)},
		Model: o.model,
	})
	if err != nil {
		return "", err
	}

	return resp.OutputText(), err
}
