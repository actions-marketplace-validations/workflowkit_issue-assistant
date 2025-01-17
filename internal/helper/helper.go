package helper

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/canack/issue-assistant/pkg/ai"
	"github.com/canack/issue-assistant/pkg/logger"

	pkggithub "github.com/canack/issue-assistant/pkg/github"
)

type Helper struct {
	githubToken     string
	githubEventPath string
	aiType          string
	aiAPIKey        string
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

// WithGitHubToken sets the GitHub token
func WithGitHubToken(token string) Option {
	return func(h *Helper) error {
		if token == "" {
			return errors.New("github token cannot be empty")
		}
		h.githubToken = token
		return nil
	}
}

// WithAIConfig sets the AI configuration
func WithAIConfig(aiType, apiKey string) Option {
	return func(h *Helper) error {
		if aiType == "" {
			return errors.New("ai type cannot be empty")
		}
		if apiKey == "" {
			return errors.New("ai api key cannot be empty")
		}
		h.aiType = aiType
		h.aiAPIKey = apiKey
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

func (h *Helper) validate() error {
	if h.githubToken == "" {
		return errors.New("github token is required")
	}
	if h.aiType == "" {
		return errors.New("ai type is required")
	}
	if h.aiAPIKey == "" {
		return errors.New("ai api key is required")
	}
	if h.githubEventPath == "" {
		return errors.New("github event path is required")
	}
	return nil
}

func (h *Helper) Help(ctx context.Context) {
	logger.Log.Info("initializing clients")
	client := pkggithub.NewClient(h.githubToken)
	analyzer := ai.NewAIService(ai.ToAIType(h.aiType), h.aiAPIKey)

	eventData, err := os.ReadFile(h.githubEventPath)
	if err != nil {
		logger.Log.Fatalf("failed to read event data: %v", err)
	}

	logger.Log.Info("processing GitHub event")
	var event struct {
		Action string `json:"action"`
		Issue  struct {
			Number int    `json:"number"`
			Title  string `json:"title"`
			Body   string `json:"body"`
		} `json:"issue"`
		Repository struct {
			Owner struct {
				Login string `json:"login"`
			} `json:"owner"`
			Name string `json:"name"`
		} `json:"repository"`
	}

	if err := json.Unmarshal(eventData, &event); err != nil {
		logger.Log.Fatalf("failed to parse event data: %v", err)
	}

	if event.Action != "opened" {
		logger.Log.Info("event is not a new issue, skipping")
		return
	}

	files, err := client.GetRepositoryContent(ctx, event.Repository.Owner.Login, event.Repository.Name)
	if err != nil {
		logger.Log.Fatalf("failed to get repository content: %v", err)
	}

	answer, _, err := analyzer.Query(ctx, event.Issue.Body, files)
	if err != nil {
		logger.Log.Errorf("failed to analyze issue: %v", err)
	}

	err = client.CreateIssueComment(ctx,
		event.Repository.Owner.Login, event.Repository.Name, event.Issue.Number, formatComment(answer))
	if err != nil {
		logger.Log.Errorf("failed to create comment: %v", err)
	}

	logger.Log.Info("successfully completed analysis")
}

func formatComment(answer string) string {
	return fmt.Sprintf(`🤖 AI Assistant Analysis

%s

---
_This analysis was performed by [Issue Analyzer](https://github.com/canack/issue-assistant). If you have any questions, please contact the repository maintainers._`, answer)
}
