package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/sashabaranov/go-openai"
	"github.com/workflowkit/issue-assistant/pkg/github"
	"github.com/workflowkit/issue-assistant/pkg/logger"
)

type AIResponse struct {
	Answer        string   `json:"answer"`
	Confidence    float64  `json:"confidence"`
	RelevantFiles []string `json:"relevant_files"`
}

type OpenAI struct {
	client *openai.Client
}

func newOpenAIService(apiKey string) AIService {
	return &OpenAI{
		client: openai.NewClient(apiKey),
	}
}

// makeRequest is a helper function to make OpenAI API requests with retries
func (a *OpenAI) makeRequest(ctx context.Context, systemPrompt, userPrompt string) (string, error) {
	const maxRetries = 3
	var lastErr error
	baseTemperature := float32(0.1)

	for attempt := 0; attempt < maxRetries; attempt++ {
		logger.Log.Debugf("making OpenAI request: %d with temperature: %.2f", attempt+1, baseTemperature)

		resp, err := a.client.CreateChatCompletion(
			ctx,
			openai.ChatCompletionRequest{
				Model: openai.GPT4oMini,
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
				Temperature: baseTemperature,
				MaxTokens:   2000,
			},
		)

		if err != nil {
			logger.Log.Warnf("OpenAI request failed attempt: %d: %v", attempt+1, err)
			lastErr = fmt.Errorf("OpenAI API error: %w", err)
			baseTemperature -= 0.02
			continue
		}

		content := resp.Choices[0].Message.Content

		// Remove any markdown code block wrapping if exists
		content = strings.TrimSpace(content)
		content = strings.TrimPrefix(content, "```json")
		content = strings.TrimPrefix(content, "```")
		content = strings.TrimSuffix(content, "```")
		content = strings.TrimSpace(content)

		// Try to parse as generic JSON first
		var jsonContent interface{}
		if err := json.Unmarshal([]byte(content), &jsonContent); err != nil {
			logger.Log.Warnf("Invalid JSON response: %v", err)
			lastErr = fmt.Errorf("invalid JSON response: %w", err)
			baseTemperature -= 0.02
			continue
		}

		return content, nil
	}

	return "", fmt.Errorf("failed after %d attempts: %w", maxRetries, lastErr)
}

func (a *OpenAI) AnalyzeCode(ctx context.Context, question string, files []github.GitHubFile) (answer string, confidence float64, err error) {
	systemPrompt := `You are a specialized AI code assistant with expertise in analyzing codebases and providing technical explanations.

Your core responsibilities:
1. Analyze code thoroughly and provide accurate, well-structured explanations
2. Focus on practical, implementation-focused responses
3. Always include relevant code examples and file references
4. Maintain a professional and educational tone
5. Ensure responses are complete and well-organized
6. MUST return valid JSON with properly escaped strings (\\n for newlines, \" for quotes)

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
		"8. MUST escape all newlines with \\n and quotes with \\\n"+
		"\n"+
		"Available Files:\n"+
		"%s\n"+
		"\n"+
		"Question:\n"+
		"%s", formatFilesForPrompt(files), question)

	logger.Log.Infof("Analyzing code: [%s] with %d files", question, len(files))

	content, err := a.makeRequest(ctx, systemPrompt, userPrompt)
	if err != nil {
		return "", 0, err
	}

	var aiResp AIResponse
	if err := json.Unmarshal([]byte(content), &aiResp); err != nil {
		return "", 0, fmt.Errorf("failed to parse AI response: %w", err)
	}

	return aiResp.Answer, aiResp.Confidence, nil
}

func (a *OpenAI) AnalyzeLabels(ctx context.Context, title, body string, availableLabels string) (github.LabelAnalysis, error) {
	systemPrompt := `You are an AI assistant specialized in analyzing GitHub issues and suggesting appropriate labels.

Your task is to:
1. Analyze the issue title and body
2. Consider the available labels and their descriptions
3. Suggest relevant labels with confidence scores
4. Provide brief explanations for your suggestions

Guidelines:
- Only suggest labels that are highly relevant
- Consider both technical and non-technical aspects
- Be conservative with confidence scores
- Focus on the main topics and themes of the issue`

	userPrompt := fmt.Sprintf("Analyze the issue and suggest appropriate labels. Provide your response in the following JSON format (DO NOT wrap in code blocks):\n"+
		"{\n"+
		"  \"suggestedLabels\": {\n"+
		"    \"label-name\": 0.95,\n"+
		"    \"another-label\": 0.85\n"+
		"  },\n"+
		"  \"explanation\": \"Brief explanation of why these labels were chosen\"\n"+
		"}\n\n"+
		"Confidence Score Guide:\n"+
		"- 0.0-0.3: Weak relevance\n"+
		"- 0.4-0.6: Moderate relevance\n"+
		"- 0.7-0.9: Strong relevance\n"+
		"- 1.0: Perfect match\n\n"+
		"Issue Title: %s\nIssue Body:\n%s\n\nAvailable Labels:\n%s",
		title, body, availableLabels)

	logger.Log.Infof("Analyzing labels for issue: [%s]", title)

	content, err := a.makeRequest(ctx, systemPrompt, userPrompt)
	if err != nil {
		return github.LabelAnalysis{}, err
	}

	var analysis github.LabelAnalysis
	if err := json.Unmarshal([]byte(content), &analysis); err != nil {
		return github.LabelAnalysis{}, fmt.Errorf("failed to parse label analysis: %w", err)
	}

	return analysis, nil
}

func formatFilesForPrompt(files []github.GitHubFile) string {
	var result string
	for _, file := range files {
		result += fmt.Sprintf("File: %s\nContent:\n%s\n\n", file.Path, file.Content)
	}
	return result
}
