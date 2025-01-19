package helper

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/workflowkit/issue-assistant/pkg/ai"
	pkggithub "github.com/workflowkit/issue-assistant/pkg/github"
	"github.com/workflowkit/issue-assistant/pkg/logger"
)

// Feature represents an AI assistant feature
type Feature string

const (
	FeatureComment Feature = "comment" // AI analysis comments
	FeatureLabel   Feature = "label"   // Label suggestions
)

// Helper is the main struct that holds the clients and services
type Helper struct {
	githubEventPath string
	githubClient    *pkggithub.Client
	aiService       ai.AIService
	features        []Feature
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

// WithFeatures sets the enabled features
func WithFeatures(features []Feature) Option {
	return func(h *Helper) error {
		for _, f := range features {
			switch f {
			case FeatureComment, FeatureLabel:
				h.features = append(h.features, f)
			default:
				return fmt.Errorf("unknown feature: %s", f)
			}
		}
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
	// Process each enabled feature
	for _, feature := range h.features {
		switch feature {
		case FeatureComment:
			h.processComment(ctx, event)
		case FeatureLabel:
			h.processLabels(ctx, event)
		}
	}

	logger.Log.Info("completed issue processing")
}

// processComment handles AI analysis comment feature
func (h *Helper) processComment(ctx context.Context, event *GitHubEvent) {
	files, err := h.githubClient.GetRepositoryContent(ctx, event.Repository.Owner.Login, event.Repository.Name)
	if err != nil {
		logger.Log.Errorf("failed to get repository content: %v", err)
		return
	}

	answer, _, err := h.aiService.AnalyzeCode(ctx, event.Issue.Body, files)
	if err != nil {
		logger.Log.Errorf("failed to analyze issue: %v", err)
		return
	}

	err = h.githubClient.CreateIssueComment(ctx,
		event.Repository.Owner.Login,
		event.Repository.Name,
		event.Issue.Number,
		h.formatComment(answer))
	if err != nil {
		logger.Log.Errorf("failed to create comment: %v", err)
		return
	}

	logger.Log.Info("successfully added AI analysis comment")
}

// processLabels handles label analysis feature
func (h *Helper) processLabels(ctx context.Context, event *GitHubEvent) {
	// Get repository labels
	labelInfo, labels, err := h.githubClient.GetLabelsForAIAnalysis(ctx,
		event.Repository.Owner.Login,
		event.Repository.Name)
	if err != nil {
		logger.Log.Errorf("failed to get repository labels: %v", err)
		return
	}

	if len(labels) == 0 {
		logger.Log.Info("no labels found in repository")
		return
	}

	// Query AI for label suggestions
	analysis, err := h.aiService.AnalyzeLabels(ctx, event.Issue.Title, event.Issue.Body, labelInfo)
	if err != nil {
		logger.Log.Errorf("failed to analyze labels: %v", err)
		return
	}

	// Filter labels with high confidence (> 0.7)
	var suggestedLabels []string
	for label, confidence := range analysis.SuggestedLabels {
		if confidence >= 0.7 {
			suggestedLabels = append(suggestedLabels, label)
		}
	}

	if len(suggestedLabels) == 0 {
		logger.Log.Info("no high-confidence label suggestions found")
		return
	}

	// Add suggested labels to the issue
	if err := h.githubClient.AddLabelsToIssue(ctx,
		event.Repository.Owner.Login,
		event.Repository.Name,
		event.Issue.Number,
		suggestedLabels); err != nil {
		logger.Log.Errorf("failed to add labels to issue: %v", err)
		return
	}

	// Add explanation as a comment
	comment := h.formatLabelExplanation(suggestedLabels, analysis.Explanation)
	if err := h.githubClient.CreateIssueComment(ctx,
		event.Repository.Owner.Login,
		event.Repository.Name,
		event.Issue.Number,
		comment); err != nil {
		logger.Log.Errorf("failed to add label explanation comment: %v", err)
		return
	}

	logger.Log.Info("successfully added labels and explanation comment")
}

// formatLabelExplanation formats the label explanation as a GitHub issue comment
func (h *Helper) formatLabelExplanation(labels []string, explanation string) string {
	return fmt.Sprintf(`🏷️ AI Label Analysis

I've added the following labels to this issue:
%s

**Explanation:**
%s

---
_This label analysis was performed by [Issue Assistant](https://github.com/workflowkit/issue-assistant). If you have any questions, please contact the repository maintainers._`,
		formatLabelList(labels),
		explanation,
	)
}

// formatLabelList formats a list of labels as a bullet point list
func formatLabelList(labels []string) string {
	var result string
	for _, label := range labels {
		result += fmt.Sprintf("- `%s`\n", label)
	}
	return result
}

// formatComment formats the AI response as a GitHub issue comment
func (h *Helper) formatComment(answer string) string {
	return fmt.Sprintf(`🤖 AI Assistant Analysis

%s

---
_This analysis was performed by [Issue Assistant](https://github.com/workflowkit/issue-assistant). If you have any questions, please contact the repository maintainers._`, answer)
}
