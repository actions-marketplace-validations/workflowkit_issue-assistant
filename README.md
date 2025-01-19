# Issue Assistant

[![Go Report Card](https://goreportcard.com/badge/github.com/workflowkit/issue-assistant)](https://goreportcard.com/report/github.com/workflowkit/issue-assistant)
[![GitHub Actions](https://github.com/workflowkit/issue-assistant/workflows/issue-analyzer/badge.svg)](https://github.com/workflowkit/issue-assistant/actions)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

AI-powered GitHub Issue assistant that provides intelligent responses based on repository content using various AI models (OpenAI GPT-4, Anthropic Claude).

## Features

- 🤖 Automated issue analysis
- 🔍 Deep repository content understanding
- 📝 Markdown-formatted responses
- 🔄 Retry mechanism for reliability
- 📊 Confidence scoring
- 🚀 Docker support
- 🧠 Multiple AI model support (OpenAI, Claude)
- 📋 Customizable response templates

## Requirements

- Go 1.23 or higher
- GitHub Personal Access Token (PAT) with `repo` scope
- AI API Key (OpenAI or Anthropic)
- Docker (optional, for containerized running)

## Environment Variables

| Variable | Description | Required | Default |
|----------|-------------|----------|---------|
| `GITHUB_TOKEN` | GitHub PAT with repo scope | Yes | - |
| `OPENAI_API_KEY` | OpenAI API Key | Yes* | - |
| `CLAUDE_API_KEY` | Anthropic Claude API Key | Yes* | - |
| `AI_TYPE` | AI model type (openai/claude) | Yes | - |

*Either OPENAI_API_KEY or CLAUDE_API_KEY is required based on AI_TYPE

## Project Structure

```
.
├── cmd/
│   └── issue-assistant/
├── internal/
│   └── helper/
├── pkg/
│   ├── ai/
│   ├── github/
│   └── logger/
└── .github/
    └── workflows/
```

## Setup

1. Create required secrets in your repository:
   - `GITHUB_TOKEN`: GitHub token with repo scope (automatically provided by GitHub Actions)
   - `OPENAI_API_KEY`: Your OpenAI API key (if using OpenAI)
   - `CLAUDE_API_KEY`: Your Claude API key (if using Claude)

2. Add the following workflow file to your repository (`.github/workflows/issue-assistant.yml`):

```yaml
name: Issue Assistant
on:
  issues:
    types: [opened]

jobs:
  analyze:
    runs-on: ubuntu-latest
    permissions:
      issues: write
      contents: read
    steps:
      - uses: actions/checkout@v3
        with:
          repository: workflowkit/issue-assistant
          token: ${{ secrets.GITHUB_TOKEN }}
          path: .issue-assistant
      
      - name: Build and run assistant
        working-directory: .issue-assistant
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
          AI_TYPE: "openai"  # or "claude"
          OPENAI_API_KEY: ${{ secrets.OPENAI_API_KEY }}
          CLAUDE_API_KEY: ${{ secrets.CLAUDE_API_KEY }}
        run: |
          docker build -t issue-assistant .
          docker run --rm \
            -e GITHUB_TOKEN \
            -e AI_TYPE \
            -e OPENAI_API_KEY \
            -e CLAUDE_API_KEY \
            -e GITHUB_EVENT_PATH \
            -v $GITHUB_EVENT_PATH:$GITHUB_EVENT_PATH \
            issue-assistant
```

## How It Works

1. When a new issue is opened, the workflow is triggered
2. The assistant reads the repository content
3. OpenAI GPT-4 analyzes the issue and repository content
4. A detailed response is posted as a comment on the issue
5. Confidence scoring ensures high-quality responses

## Development

```bash
# Clone the repository
git clone https://github.com/workflowkit/issue-assistant.git

# Install dependencies
go mod tidy

# Run tests
go test ./...

# Build locally
go build -o bin/issue-assistant

# Build Docker image
docker build -t issue-assistant .
```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request. For major changes, please open an issue first to discuss what you would like to change.

1. Fork the Project
2. Create your Feature Branch (`git checkout -b feature/AmazingFeature`)
3. Commit your Changes (`git commit -m 'Add some AmazingFeature'`)
4. Push to the Branch (`git push origin feature/AmazingFeature`)
5. Open a Pull Request

