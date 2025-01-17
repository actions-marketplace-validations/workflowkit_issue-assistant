package github

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/go-github/v45/github"
	"golang.org/x/oauth2"
)

type Client struct {
	client *github.Client
}

func NewClient(token string) *Client {
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(context.Background(), ts)
	ghClient := github.NewClient(tc)

	return &Client{
		client: ghClient,
	}
}

func (c *Client) GetRepositoryContent(ctx context.Context, owner, repo string) ([]GitHubFile, error) {
	var files []GitHubFile

	err := c.traverseContent(ctx, owner, repo, "", &files)
	if err != nil {
		return nil, fmt.Errorf("failed to traverse repository: %w", err)
	}

	return files, nil
}

func (c *Client) CreateIssueComment(ctx context.Context, owner, repo string, issueNumber int, comment string) error {
	_, _, err := c.client.Issues.CreateComment(ctx, owner, repo, issueNumber, &github.IssueComment{
		Body: github.String(comment),
	})

	return err
}

func (c *Client) traverseContent(ctx context.Context, owner, repo, path string, files *[]GitHubFile) error {
	_, directoryContent, _, err := c.client.Repositories.GetContents(ctx, owner, repo, path, nil)
	if err != nil {
		return fmt.Errorf("failed to get contents: %w", err)
	}

	for _, content := range directoryContent {
		if content == nil || content.Type == nil || content.Name == nil || content.Path == nil {
			continue // Skip invalid entries
		}

		switch *content.Type {
		case "file":
			if isRelevantFile(*content.Name) {
				fileContent, err := c.getFileContent(ctx, owner, repo, *content.Path)
				if err != nil {
					return fmt.Errorf("failed to get file content for %s: %w", *content.Path, err)
				}

				*files = append(*files, GitHubFile{
					Path:     *content.Path,
					Content:  fileContent,
					FileType: getFileType(*content.Name),
				})
			}
		case "dir":
			if err := c.traverseContent(ctx, owner, repo, *content.Path, files); err != nil {
				return fmt.Errorf("failed to traverse directory %s: %w", *content.Path, err)
			}
		}
	}

	return nil
}

func (c *Client) getFileContent(ctx context.Context, owner, repo, path string) (string, error) {
	fileContent, _, _, err := c.client.Repositories.GetContents(ctx, owner, repo, path, nil)
	if err != nil {
		return "", fmt.Errorf("failed to get file contents: %w", err)
	}

	if fileContent == nil {
		return "", fmt.Errorf("no content found for file: %s", path)
	}

	content, err := fileContent.GetContent()
	if err != nil {
		return "", fmt.Errorf("failed to decode content: %w", err)
	}

	return content, nil
}

func isRelevantFile(filename string) bool {
	relevantExtensions := []string{
		".go", ".js", ".ts", ".py", ".java", ".rb", ".php",
		".md", ".txt", ".yaml", ".yml", ".json",
	}

	for _, ext := range relevantExtensions {
		if strings.HasSuffix(filename, ext) {
			return true
		}
	}
	return false
}

func getFileType(filename string) string {
	if strings.HasSuffix(filename, ".md") {
		return "documentation"
	}
	if strings.HasSuffix(filename, ".go") {
		return "source"
	}
	if strings.HasSuffix(filename, ".yaml") || strings.HasSuffix(filename, ".yml") {
		return "configuration"
	}
	return "other"
}
