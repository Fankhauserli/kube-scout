package api

import (
	"context"

	"github.com/Fankhauserli/kube-scout/config"
)

type LLMClient interface {
	GenerateResponse(ctx context.Context, prompt string) (string, error)
}

func NewLLMClient(cfg config.LLMConfig) LLMClient {
	switch cfg.Type {
	case "openai":
		return NewOpenAIClient(cfg.Endpoint, cfg.ModelName, cfg.ApiKeyEnv)
	case "ollama":
		return NewOllamaClient(cfg.Endpoint, cfg.ModelName)
	default:
		panic("unsupported LLM type: " + cfg.Type)
	}
}
