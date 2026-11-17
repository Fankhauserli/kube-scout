package api

type LLMClient interface {
	GenerateResponse(prompt string) (string, error)
}
