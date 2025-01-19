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

## Quick Start

1. Add this workflow to your repository (`.github/workflows/issue-assistant.yml`):

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
      - uses: workflowkit/issue-assistant@v1.0.0
        with:
          github_token: ${{ secrets.GITHUB_TOKEN }}
          ai_type: "openai"  # or "claude"
          openai_api_key: ${{ secrets.OPENAI_API_KEY }}
          # claude_api_key: ${{ secrets.CLAUDE_API_KEY }}  # if using claude
          enable_comment: "true"   # required: at least one feature must be enabled
          enable_label: "false"    # required: at least one feature must be enabled
```

2. Add required secrets to your repository:
   - `OPENAI_API_KEY` (if using OpenAI)
   - `CLAUDE_API_KEY` (if using Claude)

That's it! Now when someone opens an issue:
- AI will analyze the issue content
- AI will analyze your repository code
- AI will post a helpful response as a comment
- Optionally, AI can suggest labels

## Configuration Options

| Option | Description | Required | Default |
|--------|-------------|----------|---------|
| `github_token` | GitHub token (automatically provided) | Yes | - |
| `ai_type` | AI model to use (`openai` or `claude`) | Yes | - |
| `openai_api_key` | OpenAI API Key | Yes* | - |
| `claude_api_key` | Claude API Key | Yes* | - |
| `enable_comment` | Enable AI comments on issues | Yes** | false |
| `enable_label` | Enable AI label suggestions | Yes** | false |

*Either `openai_api_key` or `claude_api_key` is required based on `ai_type`
**At least one feature (`enable_comment` or `enable_label`) must be enabled

## Advanced Usage

### Using with OpenAI:
```yaml
- uses: workflowkit/issue-assistant@v1.0.0
  with:
    github_token: ${{ secrets.GITHUB_TOKEN }}
    ai_type: "openai"
    openai_api_key: ${{ secrets.OPENAI_API_KEY }}
    enable_comment: "true"
```

### Using with Claude:
```yaml
- uses: workflowkit/issue-assistant@v1.0.0
  with:
    github_token: ${{ secrets.GITHUB_TOKEN }}
    ai_type: "claude"
    claude_api_key: ${{ secrets.CLAUDE_API_KEY }}
    enable_label: "true"
```

### Enable All Features:
```yaml
- uses: workflowkit/issue-assistant@v1.0.0
  with:
    github_token: ${{ secrets.GITHUB_TOKEN }}
    ai_type: "openai"
    openai_api_key: ${{ secrets.OPENAI_API_KEY }}
    enable_comment: "true"
    enable_label: "true"
```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request. For major changes, please open an issue first to discuss what you would like to change.

1. Fork the Project
2. Create your Feature Branch (`git checkout -b feature/AmazingFeature`)
3. Commit your Changes (`git commit -m 'Add some AmazingFeature'`)
4. Push to the Branch (`git push origin feature/AmazingFeature`)
5. Open a Pull Request

