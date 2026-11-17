package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/Fankhauserli/kube-scout/agent"
	"github.com/Fankhauserli/kube-scout/api"
	tool "github.com/Fankhauserli/kube-scout/tools"
)

func main() {
	ctx := context.Background()

	config, err := loadEnv()
	if err != nil {
		log.Fatalf("Failed to load environment config: %v", err)
	}

	// Set up tools
	toolRegistry := tool.NewToolRegistry()
	registerTools(toolRegistry, config)

	llmClient := api.NewOpenAIClient(config.LLM.Endpoint, config.LLM.ModelName, config.LLM.ApiKeyEnv)

	prompt := config.Agent.PromptTemplate
	if config.Agent.PromptFile != "" {
		promptContent, err := os.ReadFile(config.Agent.PromptFile)
		if err != nil {
			log.Fatalf("Failed to read prompt file: %v", err)
		}
		prompt = string(promptContent)
	}
	// Create agent
	agentInstance := agent.NewAgent(toolRegistry, llmClient, config.Agent.MaxIterations)

	result, err := agentInstance.Run(ctx, prompt)
	if err != nil {
		log.Fatalf("Agent failed: %v", err)
	}

	fmt.Println("Final Result:")
	fmt.Println(result)
}

func registerTools(toolRegistry *tool.ToolRegistry, config *EnvConfig) {
	// Register tools based on the configuration
	disabledTools := make(map[string]bool)
	for _, toolConfig := range config.Tools {
		if toolConfig.Enabled && toolConfig.NewTool.Name != "" {
			// Create a new tool instance based on the configuration
			newTool := tool.NewApiTool(
				toolConfig.NewTool.Name,
				toolConfig.NewTool.Description,
				toolConfig.NewTool.Endpoint,
				toolConfig.NewTool.ApiKeyEnv,
				toolConfig.NewTool.Method,
			)
			toolRegistry.Register(newTool)
		} else if !toolConfig.Enabled && toolConfig.Name != "" {
			disabledTools[toolConfig.Name] = true
		}
	}

	for _, defaultTool := range tool.DefaultTools {
		if !disabledTools[defaultTool.Name()] {
			toolRegistry.Register(defaultTool)
		}
	}
}
