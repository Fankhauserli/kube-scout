package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Fankhauserli/kube-scout/agent"
	"github.com/Fankhauserli/kube-scout/api"
	"github.com/Fankhauserli/kube-scout/config"
	"github.com/Fankhauserli/kube-scout/event"
	tool "github.com/Fankhauserli/kube-scout/tools"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	config, err := config.LoadEnv()
	if err != nil {
		log.Fatalf("[ERROR] [main] Failed to load environment config: %v", err)
	}

	// Set up tools
	toolRegistry := tool.NewToolRegistry()
	registerTools(toolRegistry, config)

	llmClient := api.NewLLMClient(config.LLM)

	prompt := getPrompt(config)
	// Create agent
	agentInstance := agent.NewAgent(toolRegistry, llmClient, config.Agent.MaxIterations, config.Agent.MaxToolResultLength)

	if config.Trigger.Enabled {
		eventConn, err := event.NewRabbitMQEventQueue(config.Trigger.Address, config.Trigger.Channel, config.Trigger.UserEnv, config.Trigger.PasswordEnv)
		if err != nil {
			log.Fatalf("[ERROR] [main] Failed to connect to RabbitMQ event queue: %v", err)
		}
		defer eventConn.Close()

		err = eventConn.Consume(ctx, func(ctx context.Context, message string) error {
			log.Printf("[INFO] [main] Received a message: %s", message)
			runCtx, cancel := context.WithTimeout(ctx, 5*time.Minute)
			result, err := agentInstance.Run(runCtx, prompt+"\n\nEvent: "+message)
			cancel()

			if err != nil {
				log.Printf("[ERROR] [main] Error running agent: %v", err)
				return err
			}
			log.Printf("[INFO] [main] Final result: %s", result)
			return nil
		})
		if err != nil {
			log.Fatalf("[ERROR] [main] Failed to consume messages from RabbitMQ: %v", err)
		}
	} else {
		result, err := agentInstance.Run(ctx, prompt)
		if err != nil {
			log.Fatalf("[ERROR] [main] Agent failed: %v", err)
		}

		log.Printf("[INFO] [main] Final result: %s", result)
	}
}

func getPrompt(config *config.EnvConfig) string {
	prompt := config.Agent.PromptTemplate
	if config.Agent.PromptFile != "" {
		promptContent, err := os.ReadFile(config.Agent.PromptFile)
		if err != nil {
			log.Fatalf("[ERROR] [main] Failed to read prompt file: %v", err)
		}
		prompt = string(promptContent)
	}
	return prompt
}

func registerTools(toolRegistry *tool.ToolRegistry, config *config.EnvConfig) {
	// Register tools based on the configuration
	disabledTools := make(map[string]bool)
	for _, toolConfig := range config.Tools {
		if toolConfig.Enabled && toolConfig.Name == "mail" {
			// Register the mail tool
			mailTool := &tool.MailTool{
				Addr:     toolConfig.Configuration["addr"],
				MailFrom: toolConfig.Configuration["mailFrom"],
				MailTo:   toolConfig.Configuration["mailTo"],
			}
			toolRegistry.Register(mailTool)
		} else if toolConfig.Enabled && toolConfig.NewTool.Name != "" {
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
