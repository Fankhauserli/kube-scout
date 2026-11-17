package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/Fankhauserli/kube-scout/agent"
	"github.com/Fankhauserli/kube-scout/api"
	"github.com/Fankhauserli/kube-scout/config"
	"github.com/Fankhauserli/kube-scout/event"
	tool "github.com/Fankhauserli/kube-scout/tools"
)

func main() {
	ctx := context.Background()

	config, err := config.LoadEnv()
	if err != nil {
		log.Fatalf("Failed to load environment config: %v", err)
	}

	// Set up tools
	toolRegistry := tool.NewToolRegistry()
	registerTools(toolRegistry, config)

	llmClient := api.NewOpenAIClient(config.LLM.Endpoint, config.LLM.ModelName, config.LLM.ApiKeyEnv)

	prompt := getPrompt(config)
	// Create agent
	agentInstance := agent.NewAgent(toolRegistry, llmClient, config.Agent.MaxIterations)

	if config.Trigger.Enabled {
		eventConn, err := event.NewRabbitMQEventQueue(config.Trigger.Address, config.Trigger.Channel, config.Trigger.UserEnv, config.Trigger.PasswordEnv)
		if err != nil {
			log.Fatalf("Failed to connect to RabbitMQ event queue: %v", err)
		}
		defer eventConn.Close()

		msgs, err := eventConn.Channel.Consume(
			eventConn.Queue.Name, // queue
			"",                   // consumer
			true,                 // auto-ack
			false,                // exclusive
			false,                // no-local
			false,                // no-wait
			nil,                  // args
		)

		var forever chan struct{}

		go func() {
			for d := range msgs {
				result, err := agentInstance.Run(ctx, prompt+"\n\nEvent: "+string(d.Body))
				if err != nil {
					log.Printf("Error running agent: %v", err)
				}
				log.Printf(" [x] %s", result)
			}
		}()
		<-forever
	} else {
		result, err := agentInstance.Run(ctx, prompt)
		if err != nil {
			log.Fatalf("Agent failed: %v", err)
		}

		fmt.Println("Final Result:")
		fmt.Println(result)
	}
}

func getPrompt(config *config.EnvConfig) string {
	prompt := config.Agent.PromptTemplate
	if config.Agent.PromptFile != "" {
		promptContent, err := os.ReadFile(config.Agent.PromptFile)
		if err != nil {
			log.Fatalf("Failed to read prompt file: %v", err)
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
				Addr:     toolConfig.Configuration["addr"].(string),
				MailFrom: toolConfig.Configuration["mailFrom"].(string),
				MailTo:   toolConfig.Configuration["mailTo"].(string),
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
