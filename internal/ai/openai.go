package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type OpenAIProvider struct {
	apiKey     string
	model      string
	httpClient *http.Client
}

func NewOpenAIProvider(apiKey, model string) *OpenAIProvider {
	if model == "" {
		model = "gpt-4"
	}
	return &OpenAIProvider{
		apiKey: apiKey,
		model:  model,
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

func (p *OpenAIProvider) Name() string {
	return "openai"
}

func (p *OpenAIProvider) GenerateIssue(ctx context.Context, input GenerateIssueInput) (*GenerateIssueOutput, error) {
	labelsContext := ""
	if len(input.AvailableLabels) > 0 {
		labelsJSON, _ := json.Marshal(input.AvailableLabels)
		labelsContext = fmt.Sprintf("\n\nAvailable labels to choose from: %s", string(labelsJSON))
	}

	reqBody := openAIRequest{
		Model: p.model,
		Messages: []openAIMessage{
			{Role: "system", Content: issuePrompt + labelsContext},
			{Role: "user", Content: input.Prompt},
		},
		Temperature: 0.7,
		ResponseFormat: &openAIResponseFormat{
			Type: "json_object",
		},
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.openai.com/v1/chat/completions", bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+p.apiKey)

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("OpenAI API error (status %d): %s", resp.StatusCode, string(body))
	}

	var result openAIResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if len(result.Choices) == 0 {
		return nil, fmt.Errorf("no response from OpenAI")
	}

	return parseIssueResponse(result.Choices[0].Message.Content)
}

type openAIRequest struct {
	Model          string                `json:"model"`
	Messages       []openAIMessage       `json:"messages"`
	Temperature    float64               `json:"temperature"`
	ResponseFormat *openAIResponseFormat `json:"response_format,omitempty"`
}

type openAIMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type openAIResponseFormat struct {
	Type string `json:"type"`
}

type openAIResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

func parseIssueResponse(content string) (*GenerateIssueOutput, error) {
	var output struct {
		Title             string   `json:"title"`
		Description       string   `json:"description"`
		SuggestedLabels   []string `json:"suggestedLabels"`
		SuggestedPriority int      `json:"suggestedPriority"`
	}

	if err := json.Unmarshal([]byte(content), &output); err != nil {
		return nil, fmt.Errorf("failed to parse AI response: %w", err)
	}

	return &GenerateIssueOutput{
		Title:             output.Title,
		Description:       output.Description,
		SuggestedLabels:   output.SuggestedLabels,
		SuggestedPriority: output.SuggestedPriority,
	}, nil
}
