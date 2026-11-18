package tool

import (
	"context"
	"log"
	"os/exec"
	"strings"
)

type KubectlTool struct{}

func (c *KubectlTool) Name() string {
	return "kubectl"
}

func (c *KubectlTool) Description() string {
	return "Kubernetes command-line tool to interact with the cluster, to use just provide the kubectl command you want to run, e.g. 'get pods -A'"
}

func (c *KubectlTool) Execute(ctx context.Context, input string) (string, error) {
	args := strings.Fields(strings.TrimPrefix(strings.TrimSpace(input), "kubectl "))
	cmd := exec.CommandContext(ctx, "/usr/local/bin/kubectl", args...)

	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("[ERROR] [tool: kubectl] %s", string(output))
		return string(output), err
	}

	log.Printf("[INFO] [tool: kubectl] %s", string(output))
	return string(output), nil
}
