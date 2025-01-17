package ai

import (
	"context"
	"strings"

	"github.com/canack/issue-assistant/pkg/github"
	"github.com/canack/issue-assistant/pkg/logger"
)

type AIService interface {
	Query(ctx context.Context, question string, files []github.GitHubFile) (answer string, confidence float64, err error)
}

// We do not control AI model type because of every AI service has its own model
// And every single month they are updating their models :)

func NewAIService(aiType AIType, apiKey string) AIService {
	switch aiType {
	case AITypeOpenAI:
		return newOpenAIService(apiKey)
	case AITypeClaude:
		// return newClaudeService(apiKey)
		logger.Log.Fatalf("Claude isn't implemented yet")
	default:
		logger.Log.Fatalf("AI type %s is not supported", aiType)
	}

	return nil
}

type AIType string

const (
	AITypeOpenAI AIType = "openai"
	AITypeClaude AIType = "claude"
)

func ToAIType(s string) AIType {
	switch strings.ToLower(s) {
	case "openai":
		return AITypeOpenAI
	case "claude":
		return AITypeClaude
	default:
		logger.Log.Fatalf("AI type %s is not supported", s)
	}

	return ""
}
