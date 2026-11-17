package config

import (
	"fmt"
	"os"

	"go.yaml.in/yaml/v2"
)

type EnvConfig struct {
	LLM     LLMConfig     `yaml:"llm"`
	Agent   AgentConfig   `yaml:"agent"`
	Tools   []ToolConfig  `yaml:"tools"`
	Trigger TriggerConfig `yaml:"trigger"`
}

type LLMConfig struct {
	ModelName string `yaml:"modelName"`
	Endpoint  string `yaml:"endpoint"`
	ApiKeyEnv string `yaml:"apiKeyEnv"`
}

type AgentConfig struct {
	PromptTemplate string `yaml:"promptTemplate"`
	PromptFile     string `yaml:"promptFile"`
	MaxIterations  int    `yaml:"maxIterations"`
}

type ToolConfig struct {
	NewTool NewToolConfig `yaml:"newTool"`
	Enabled bool          `yaml:"enabled"`
	Name    string        `yaml:"name"`
}

type NewToolConfig struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
	Endpoint    string `yaml:"endpoint"`
	ApiKeyEnv   string `yaml:"apiKeyEnv"`
	Method      string `yaml:"method"`
}

type TriggerConfig struct {
	Enabled     bool   `yaml:"enabled"`
	Name        string `yaml:"name"`
	Address     string `yaml:"address"`
	Channel     string `yaml:"channel"`
	UserEnv     string `yaml:"userEnv"`
	PasswordEnv string `yaml:"passwordEnv"`
}

func LoadEnv() (*EnvConfig, error) {
	config := &EnvConfig{}

	file, exists := os.LookupEnv("CONFIG_FILE_LOCATION")
	if !exists {
		file = "config.yaml"
	}

	filecontent, err := os.Open(file)
	if err != nil {
		return nil, fmt.Errorf("failed to open config file: %w", err)
	}
	defer filecontent.Close()

	err = yaml.NewDecoder(filecontent).Decode(config)
	if err != nil {
		return nil, fmt.Errorf("failed to decode config file: %w", err)
	}

	return config, nil
}
