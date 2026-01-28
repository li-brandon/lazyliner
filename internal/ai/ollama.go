package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type OllamaProvider struct {
	host       string
	model      string
	httpClient *http.Client
}

func NewOllamaProvider(host, model string) *OllamaProvider {
	if host == "" {
		host = "http://localhost:11434"
	}
	if model == "" {
		model = "llama2"
	}
	return &OllamaProvider{
		host:  strings.TrimSuffix(host, "/"),
		model: model,
		httpClient: &http.Client{
			Timeout: 120 * time.Second,
		},
	}
}

func (p *OllamaProvider) Name() string {
	return "ollama"
}

func (p *OllamaProvider) GenerateIssue(ctx context.Context, input GenerateIssueInput) (*GenerateIssueOutput, error) {
	labelsContext := ""
	if len(input.AvailableLabels) > 0 {
		labelsJSON, _ := json.Marshal(input.AvailableLabels)
		labelsContext = fmt.Sprintf("\n\nAvailable labels to choose from: %s", string(labelsJSON))
	}

	prompt := issuePrompt + labelsContext + "\n\nUser request: " + input.Prompt + "\n\nRespond with valid JSON only:"

	reqBody := ollamaRequest{
		Model:  p.model,
		Prompt: prompt,
		Stream: false,
		Format: "json",
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", p.host+"/api/generate", bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

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
		return nil, fmt.Errorf("Ollama API error (status %d): %s", resp.StatusCode, string(body))
	}

	var result ollamaResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if result.Response == "" {
		return nil, fmt.Errorf("no response from Ollama")
	}

	return parseIssueResponse(result.Response)
}

type ollamaRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	Stream bool   `json:"stream"`
	Format string `json:"format"`
}

type ollamaResponse struct {
	Response string `json:"response"`
}
