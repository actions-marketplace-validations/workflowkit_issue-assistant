package main

import (
	"context"
	"os"

	"github.com/workflowkit/issue-assistant/internal/helper"
	"github.com/workflowkit/issue-assistant/pkg/logger"
)

func main() {
	ctx := context.Background()

	logger.SetLogger(logger.ZapLogger)
	logger.Log.Info("starting issue assistant")

	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		logger.Log.Fatal("GITHUB_TOKEN is required")
	}

	aiType := os.Getenv("AI_TYPE")
	if aiType == "" {
		logger.Log.Fatal("AI_TYPE is required")
	}

	var apiKey string
	if aiType == "openai" {
		openAIKey := os.Getenv("OPENAI_API_KEY")
		if openAIKey == "" {
			logger.Log.Fatal("OPENAI_API_KEY is required when using OpenAI")
		}
		apiKey = openAIKey
	} else if aiType == "claude" {
		claudeKey := os.Getenv("CLAUDE_API_KEY")
		if claudeKey == "" {
			logger.Log.Fatal("CLAUDE_API_KEY is required when using Claude")
		}
		apiKey = claudeKey
	} else {
		logger.Log.Fatal("AI_TYPE must be either 'openai' or 'claude'")
	}

	eventPath := os.Getenv("GITHUB_EVENT_PATH")
	if eventPath == "" {
		logger.Log.Fatal("GITHUB_EVENT_PATH is required")
	}

	// Convert boolean flags to feature array
	var features []helper.Feature
	if os.Getenv("ENABLE_COMMENT") == "true" {
		features = append(features, helper.FeatureComment)
	}
	if os.Getenv("ENABLE_LABEL") == "true" {
		features = append(features, helper.FeatureLabel)
	}

	if len(features) == 0 {
		logger.Log.Fatal("at least one feature must be enabled")
	}

	hpr, err := helper.NewHelper(
		helper.WithGitHubClient(token),
		helper.WithAIService(aiType, apiKey),
		helper.WithGitHubEventPath(eventPath),
		helper.WithFeatures(features),
	)
	if err != nil {
		logger.Log.Fatalf("failed to create helper: %v", err)
	}

	hpr.Help(ctx)
}
