package ai

import (
	"context"
	"fmt"

	"github.com/brandonli/lazyliner/internal/config"
)

// GenerateIssueInput contains the prompt and context for generating an issue
type GenerateIssueInput struct {
	Prompt          string   // User's natural language prompt
	AvailableLabels []string // Available labels to suggest from
}

// GenerateIssueOutput contains the AI-generated issue content
type GenerateIssueOutput struct {
	Title             string   // Generated issue title
	Description       string   // Generated issue description (markdown)
	SuggestedLabels   []string // Suggested labels from available options
	SuggestedPriority int      // Suggested priority (0-4: none, urgent, high, medium, low)
}

// Provider defines the interface for AI providers
type Provider interface {
	// GenerateIssue generates issue content from a natural language prompt
	GenerateIssue(ctx context.Context, input GenerateIssueInput) (*GenerateIssueOutput, error)
	// Name returns the provider name
	Name() string
}

// NewProvider creates a new AI provider based on configuration
func NewProvider(cfg config.AIConfig) (Provider, error) {
	switch cfg.Provider {
	case "openai":
		if cfg.OpenAI.APIKey == "" {
			return nil, fmt.Errorf("OpenAI API key not configured")
		}
		return NewOpenAIProvider(cfg.OpenAI.APIKey, cfg.OpenAI.Model), nil
	case "anthropic":
		if cfg.Anthropic.APIKey == "" {
			return nil, fmt.Errorf("Anthropic API key not configured")
		}
		return NewAnthropicProvider(cfg.Anthropic.APIKey, cfg.Anthropic.Model), nil
	case "ollama":
		return NewOllamaProvider(cfg.Ollama.Host, cfg.Ollama.Model), nil
	default:
		return nil, fmt.Errorf("unknown AI provider: %s", cfg.Provider)
	}
}

// issuePrompt is the system prompt used to generate issues
const issuePrompt = `You are an expert at creating well-structured issue tickets for software development. 
Given a user's natural language description, generate a clear and actionable issue.

Guidelines:
- Title should be concise (max 80 chars), action-oriented, and describe the task
- Description should be in markdown format with:
  - A brief summary paragraph
  - Implementation details or steps if applicable
  - Acceptance criteria as a checklist
- Suggest appropriate labels from the available options
- Suggest a priority level (1=Urgent, 2=High, 3=Medium, 4=Low, 0=None)

Respond in JSON format:
{
  "title": "Issue title here",
  "description": "Markdown description here",
  "suggestedLabels": ["label1", "label2"],
  "suggestedPriority": 3
}`
