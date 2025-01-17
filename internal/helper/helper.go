package helper

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/canack/issue-assistant/pkg/ai"
	pkggithub "github.com/canack/issue-assistant/pkg/github"
	"github.com/canack/issue-assistant/pkg/logger"
)

// Helper is the main struct that holds the clients and services
type Helper struct {
	githubEventPath string
	githubClient    *pkggithub.Client
	aiService       ai.AIService
}

// Option is a function type that modifies Helper
type Option func(*Helper) error

// NewHelper creates a new Helper instance with the given options
func NewHelper(opts ...Option) (*Helper, error) {
	h := &Helper{}

	for _, opt := range opts {
		if err := opt(h); err != nil {
			return nil, fmt.Errorf("failed to apply option: %w", err)
		}
	}

	if err := h.validate(); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	return h, nil
}

// WithGitHubClient sets the GitHub client
func WithGitHubClient(token string) Option {
	return func(h *Helper) error {
		if token == "" {
			return errors.New("github token cannot be empty")
		}
		h.githubClient = pkggithub.NewClient(token)
		return nil
	}
}

// WithAIService sets the AI service
func WithAIService(aiType string, apiKey string) Option {
	return func(h *Helper) error {
		if aiType == "" {
			return errors.New("ai type cannot be empty")
		}
		if apiKey == "" {
			return errors.New("ai api key cannot be empty")
		}
		h.aiService = ai.NewAIService(ai.ToAIType(aiType), apiKey)
		return nil
	}
}

// WithGitHubEventPath sets the GitHub event path
func WithGitHubEventPath(path string) Option {
	return func(h *Helper) error {
		if path == "" {
			return errors.New("github event path cannot be empty")
		}
		h.githubEventPath = path
		return nil
	}
}

// validate checks if the Helper is properly initialized
func (h *Helper) validate() error {
	if h.githubClient == nil {
		return errors.New("github client is required")
	}
	if h.aiService == nil {
		return errors.New("ai service is required")
	}
	if h.githubEventPath == "" {
		return errors.New("github event path is required")
	}
	return nil
}

// Help processes a GitHub issue event and provides AI-powered assistance
func (h *Helper) Help(ctx context.Context) {
	event, err := h.parseEvent()
	if err != nil {
		logger.Log.Fatalf("failed to parse event: %v", err)
	}

	if event.Action != "opened" {
		logger.Log.Info("event is not a new issue, skipping")
		return
	}

	h.processIssue(ctx, event)
}

// parseEvent reads and parses the GitHub event data
func (h *Helper) parseEvent() (*GitHubEvent, error) {
	eventData, err := os.ReadFile(h.githubEventPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read event data: %w", err)
	}

	var event GitHubEvent
	if err := json.Unmarshal(eventData, &event); err != nil {
		return nil, fmt.Errorf("failed to parse event data: %w", err)
	}

	return &event, nil
}

// processIssue handles the analysis and response for a GitHub issue
func (h *Helper) processIssue(ctx context.Context, event *GitHubEvent) {
	// Get repository content
	files, err := h.githubClient.GetRepositoryContent(ctx, event.Repository.Owner.Login, event.Repository.Name)
	if err != nil {
		logger.Log.Fatalf("failed to get repository content: %v", err)
	}

	// Analyze issue with AI
	answer, _, err := h.aiService.Query(ctx, event.Issue.Body, files)
	if err != nil {
		logger.Log.Errorf("failed to analyze issue: %v", err)
		return
	}

	// Create response comment
	err = h.githubClient.CreateIssueComment(ctx,
		event.Repository.Owner.Login,
		event.Repository.Name,
		event.Issue.Number,
		h.formatComment(answer))
	if err != nil {
		logger.Log.Errorf("failed to create comment: %v", err)
		return
	}

	logger.Log.Info("successfully completed analysis")
}

// formatComment formats the AI response as a GitHub issue comment
func (h *Helper) formatComment(answer string) string {
	return fmt.Sprintf(`🤖 AI Assistant Analysis

%s

---
_This analysis was performed by [Issue Assistant](https://github.com/canack/issue-assistant). If you have any questions, please contact the repository maintainers._`, answer)
}
