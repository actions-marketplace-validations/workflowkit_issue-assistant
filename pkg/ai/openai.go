package ai

import (
	"context"
	"fmt"
	"strings"

	"github.com/canack/issue-assistant/pkg/github"
	"github.com/canack/issue-assistant/pkg/logger"
	"github.com/sashabaranov/go-openai"
)

type OpenAI struct {
	client *openai.Client
}

// TODO: opt processing
// TODO: select model name
// TODO: select temperature
// TODO: select max tokens
// TODO: Adjust max tokens based on the number of files
// TODO: Adjust retries

func newOpenAIService(apiKey string) AIService {
	return &OpenAI{
		client: openai.NewClient(apiKey),
	}
}

func (a *OpenAI) Query(ctx context.Context, question string, files []github.GitHubFile) (answer string, confidence float64, err error) {
	const maxRetries = 3

	prompt := fmt.Sprintf("Analyze codebase and answer the question. Format:\n"+
		"Please provide:\n"+
		"1. A detailed explanation with file references\n"+
		"2. Include relevant Go code examples using ```go tags\n"+
		"3. Keep the response clear and concise\n"+
		"\n"+
		"Codebase content:\n"+
		"%s\n"+
		"\n"+
		"Question:\n"+
		"%s", formatFilesForPrompt(files), question)

	logger.Log.Infof("Analyzing issue: [%s]: %d files", question, len(files))

	var lastErr error
	for attempt := 0; attempt < maxRetries; attempt++ {
		logger.Log.Debugf("making OpenAI request: %d", attempt+1)

		resp, err := a.client.CreateChatCompletion(
			ctx,
			openai.ChatCompletionRequest{
				Model: "chatgpt-4o-latest",
				Messages: []openai.ChatCompletionMessage{
					{
						Role:    openai.ChatMessageRoleUser,
						Content: prompt,
					},
				},
				Temperature: 0.1,
				MaxTokens:   1500,
			},
		)

		if err != nil {
			logger.Log.Warnf("OpenAI request failed attempt: %d: %v", attempt+1, err)
			lastErr = fmt.Errorf("OpenAI API error: %w", err)
			continue
		}

		return resp.Choices[0].Message.Content, 1.0, nil
	}

	logger.Log.Errorf("failed after %d attempts: %v", maxRetries, lastErr)

	return "", 0, fmt.Errorf("failed after %d attempts: %v", maxRetries, lastErr)
}

func formatFilesForPrompt(files []github.GitHubFile) string {
	var result string
	for _, file := range files {
		if !strings.HasSuffix(file.Path, ".go") ||
			strings.Contains(file.Path, "/vendor/") ||
			strings.HasSuffix(file.Path, "_test.go") {
			continue
		}

		result += fmt.Sprintf("File: %s (%s)\n", file.Path, file.FileType)
		result += "Content:\n```go"

		result += fmt.Sprintf("\n%s\n```\n\n", file.Content)
	}
	return result
}
