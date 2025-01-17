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

type AIResponse struct {
	Answer        string   `json:"answer"`
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

	systemPrompt := `You are a specialized AI code assistant with expertise in analyzing codebases and providing technical explanations.

Your core responsibilities:
1. Analyze code thoroughly and provide accurate, well-structured explanations
2. Focus on practical, implementation-focused responses
3. Always include relevant code examples and file references
4. Maintain a professional and educational tone
5. Ensure responses are complete and well-organized

When analyzing code:
- Start with a high-level overview
- Break down complex concepts into clear sections
- Provide concrete examples for each explanation
- Reference specific files and code sections
- Include practical use cases and best practices

Your responses should be:
- Technical yet accessible
- Well-structured with clear sections
- Supported by code examples
- Focused on practical implementation
- Complete and self-contained`

	userPrompt := fmt.Sprintf("Analyze the codebase and provide a response in the following JSON format (DO NOT wrap the response in code blocks):\n"+
		"{\n"+
		"  \"answer\": \"Your detailed explanation here. Structure your answer as follows:\\n\\n"+
		"1. Start with a brief overview (2-3 sentences)\\n"+
		"2. Break down the explanation into clear sections using markdown headers (###)\\n"+
		"3. For each section:\\n"+
		"   - Provide a clear explanation\\n"+
		"   - Include relevant code examples\\n"+
		"   - Explain when and how to use the feature\\n"+
		"4. Add relevant code references\\n"+
		"5. Include practical examples and use cases\\n\\n"+
		"Use proper markdown formatting for better readability.\",\n"+
		"  \"confidence\": 0.8,\n"+
		"  \"relevant_files\": [\"path/to/file.ext\"]\n"+
		"}\n"+
		"\n"+
		"Response Requirements:\n"+
		"1. Make explanations comprehensive yet concise\n"+
		"2. Use markdown headers (###) to organize content\n"+
		"3. Include code examples with proper markdown code blocks\n"+
		"4. Reference specific files and line numbers when relevant\n"+
		"5. Provide practical usage examples\n"+
		"6. Ensure the response is complete (no truncated sentences or examples)\n"+
		"7. Return ONLY the JSON response, do not wrap it in markdown code blocks\n"+
		"\n"+
		"Confidence Score Guide:\n"+
		"- 0.0-0.3: Limited context or understanding\n"+
		"- 0.4-0.6: Partial context, moderate understanding\n"+
		"- 0.7-0.9: Good context, clear understanding\n"+
		"- 1.0: Complete context, full understanding\n"+
		"\n"+
		"Available Files:\n"+
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
				Model: "gpt-4o-mini",
				Messages: []openai.ChatCompletionMessage{
					{
						Role:    openai.ChatMessageRoleSystem,
						Content: systemPrompt,
					},
					{
						Role:    openai.ChatMessageRoleUser,
						Content: userPrompt,
					},
				},
				Temperature: 0.1,
				MaxTokens:   2000,
			},
		)

		if err != nil {
			logger.Log.Warnf("OpenAI request failed attempt: %d: %v", attempt+1, err)
			lastErr = fmt.Errorf("OpenAI API error: %w", err)
			continue
		}

		var aiResp AIResponse
		content := resp.Choices[0].Message.Content

		// Remove any markdown code block wrapping if exists
		content = strings.TrimPrefix(content, "```json")
		content = strings.TrimPrefix(content, "```")
		content = strings.TrimSuffix(content, "```")
		content = strings.TrimSpace(content)

		if err := json.Unmarshal([]byte(content), &aiResp); err != nil {
			logger.Log.Warnf("Failed to parse AI response [%s]: %v", content, err)
			lastErr = fmt.Errorf("response parsing error: %w", err)
			continue
		}

		logger.Log.Debugf("AI Response - Confidence: %.2f, Files: %v", aiResp.Confidence, aiResp.RelevantFiles)
		return aiResp.Answer, aiResp.Confidence, nil
	}

	logger.Log.Errorf("failed after %d attempts: %v", maxRetries, lastErr)
	return "", 0, fmt.Errorf("failed after %d attempts: %v", maxRetries, lastErr)
}

func formatFilesForPrompt(files []github.GitHubFile) string {
	var result string
	for _, file := range files {
		result += fmt.Sprintf("File: %s\nContent:\n%s\n\n", file.Path, file.Content)
	}
	return result
}
