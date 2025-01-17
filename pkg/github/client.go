package github

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/google/go-github/v45/github"
	"golang.org/x/oauth2"
)

type Client struct {
	client *github.Client
	filter FileFilter
}

func NewClient(token string) *Client {
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(context.Background(), ts)
	return &Client{
		client: github.NewClient(tc),
		filter: DefaultFileFilter(),
	}
}

func (c *Client) WithFileFilter(filter FileFilter) *Client {
	c.filter = filter
	return c
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
			if isRelevantFile(*content.Path, c.filter) {
				fileContent, err := c.getFileContent(ctx, owner, repo, *content.Path)
				if err != nil {
					return fmt.Errorf("failed to get file content for %s: %w", *content.Path, err)
				}

				*files = append(*files, GitHubFile{
					Path:    *content.Path,
					Content: fileContent,
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

func isRelevantFile(filename string, filter FileFilter) bool {
	baseName := filepath.Base(filename)

	// Check if file is in allowed files list
	for _, allowedFile := range filter.AllowedFiles {
		if baseName == allowedFile {
			return true
		}
	}

	// Check if file has an allowed extension
	hasAllowedExt := false
	for _, ext := range filter.AllowedExtensions {
		if strings.HasSuffix(filename, ext) {
			hasAllowedExt = true
			break
		}
	}
	if !hasAllowedExt {
		return false
	}

	// Check excluded paths
	for _, path := range filter.ExcludedPaths {
		if strings.Contains(filename, path) {
			return false
		}
	}

	// Check excluded file patterns
	for _, pattern := range filter.ExcludedFiles {
		matched, err := filepath.Match(pattern, baseName)
		if err == nil && matched {
			return false
		}
	}

	return true
}
