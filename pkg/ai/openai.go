package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/canack/issue-assistant/pkg/github"
	"github.com/canack/issue-assistant/pkg/logger"
	"github.com/sashabaranov/go-openai"
)

type OpenAI struct {
	client *openai.Client
}

type responseMetadata struct {
	Confidence    float64  `json:"confidence"`
	RelevantFiles []string `json:"relevant_files"`
}

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
		"4. End your response with a confidence score in JSON format:\n"+
		"   {\"confidence\": 0.8, \"relevant_files\": [\"file1.go\"]}\n"+
		"\n"+
		"Note about confidence score:\n"+
		"- 0.0-0.3: Limited context, mostly guessing\n"+
		"- 0.4-0.6: Partial context, moderate confidence\n"+
		"- 0.7-0.9: Good context, high confidence\n"+
		"- 1.0: Complete context, absolute certainty\n"+
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

		answer, metadata, err := parseAIResponse(resp.Choices[0].Message.Content)
		if err != nil {
			logger.Log.Warnf("Failed to parse AI response: %v", err)
			lastErr = fmt.Errorf("response parsing error: %w", err)
			continue
		}

		logger.Log.Debugf("AI Response - Confidence: %.2f, Files: %v", metadata.Confidence, metadata.RelevantFiles)
		return answer, metadata.Confidence, nil
	}

	logger.Log.Errorf("failed after %d attempts: %v", maxRetries, lastErr)
	return "", 0, fmt.Errorf("failed after %d attempts: %v", maxRetries, lastErr)
}

func parseAIResponse(response string) (string, responseMetadata, error) {
	var metadata responseMetadata

	lastBraceIndex := strings.LastIndex(response, "{")
	if lastBraceIndex == -1 {
		return response, metadata, fmt.Errorf("no JSON metadata found")
	}

	answer := strings.TrimSpace(response[:lastBraceIndex])
	jsonPart := response[lastBraceIndex:]

	if err := json.Unmarshal([]byte(jsonPart), &metadata); err != nil {
		return response, metadata, fmt.Errorf("failed to parse metadata: %w", err)
	}

	return answer, metadata, nil
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
