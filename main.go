package main

import (
	"context"
	"os"

	"github.com/canack/issue-assistant/internal/helper"
	"github.com/canack/issue-assistant/pkg/logger"
)

func main() {
	ctx := context.Background()

	logger.SetLogger(logger.ZapLogger)
	logger.Log.Info("starting issue assistant")

	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		logger.Log.Fatal("GITHUB_TOKEN is required")
	}

	openAIKey := os.Getenv("OPENAI_API_KEY")
	if openAIKey == "" {
		logger.Log.Fatal("OPENAI_API_KEY is required")
	}

	aiType := os.Getenv("AI_TYPE")
	if aiType == "" {
		logger.Log.Fatal("AI_TYPE is required")
	}

	eventPath := os.Getenv("GITHUB_EVENT_PATH")
	if eventPath == "" {
		logger.Log.Fatal("GITHUB_EVENT_PATH is required")
	}

	hpr, err := helper.NewHelper(
		helper.WithGitHubClient(token),
		helper.WithAIService(aiType, openAIKey),
		helper.WithGitHubEventPath(eventPath),
	)
	if err != nil {
		logger.Log.Fatalf("failed to create helper: %v", err)
	}

	hpr.Help(ctx)
}
