package linear

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	apiURL = "https://api.linear.app/graphql"
)

// Client is a Linear GraphQL API client
type Client struct {
	apiKey     string
	httpClient *http.Client
}

// NewClient creates a new Linear API client
func NewClient(apiKey string) *Client {
	return &Client{
		apiKey: apiKey,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// graphQLRequest represents a GraphQL request
type graphQLRequest struct {
	Query     string                 `json:"query"`
	Variables map[string]interface{} `json:"variables,omitempty"`
}

// graphQLResponse represents a GraphQL response
type graphQLResponse struct {
	Data   json.RawMessage `json:"data"`
	Errors []graphQLError  `json:"errors,omitempty"`
}

type graphQLError struct {
	Message string `json:"message"`
	Path    []any  `json:"path,omitempty"`
}

// execute executes a GraphQL query
func (c *Client) execute(ctx context.Context, query string, variables map[string]interface{}, result interface{}) error {
	reqBody := graphQLRequest{
		Query:     query,
		Variables: variables,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", apiURL, bytes.NewBuffer(jsonBody))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var gqlResp graphQLResponse
	if err := json.Unmarshal(body, &gqlResp); err != nil {
		return fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if len(gqlResp.Errors) > 0 {
		return fmt.Errorf("GraphQL error: %s", gqlResp.Errors[0].Message)
	}

	if result != nil {
		if err := json.Unmarshal(gqlResp.Data, result); err != nil {
			return fmt.Errorf("failed to unmarshal data: %w", err)
		}
	}

	return nil
}

// GetViewer returns the currently authenticated user
func (c *Client) GetViewer(ctx context.Context) (*Viewer, error) {
	query := `
		query Viewer {
			viewer {
				id
				name
				displayName
				email
				active
			}
		}
	`

	var result struct {
		Viewer Viewer `json:"viewer"`
	}

	if err := c.execute(ctx, query, nil, &result); err != nil {
		return nil, err
	}

	return &result.Viewer, nil
}

// GetTeams returns all teams the user has access to
func (c *Client) GetTeams(ctx context.Context) ([]Team, error) {
	query := `
		query Teams {
			teams {
				nodes {
					id
					name
					key
					description
					color
					icon
				}
			}
		}
	`

	var result struct {
		Teams struct {
			Nodes []Team `json:"nodes"`
		} `json:"teams"`
	}

	if err := c.execute(ctx, query, nil, &result); err != nil {
		return nil, err
	}

	return result.Teams.Nodes, nil
}

// GetProjects returns all projects
func (c *Client) GetProjects(ctx context.Context) ([]Project, error) {
	query := `
		query Projects {
			projects(first: 100) {
				nodes {
					id
					name
					description
					icon
					color
					state
					progress
					url
				}
			}
		}
	`

	var result struct {
		Projects struct {
			Nodes []Project `json:"nodes"`
		} `json:"projects"`
	}

	if err := c.execute(ctx, query, nil, &result); err != nil {
		return nil, err
	}

	return result.Projects.Nodes, nil
}

// GetWorkflowStates returns workflow states for a team
func (c *Client) GetWorkflowStates(ctx context.Context, teamID string) ([]WorkflowState, error) {
	query := `
		query WorkflowStates($teamId: ID!) {
			workflowStates(filter: { team: { id: { eq: $teamId } } }) {
				nodes {
					id
					name
					color
					type
					position
				}
			}
		}
	`

	variables := map[string]interface{}{
		"teamId": teamID,
	}

	var result struct {
		WorkflowStates struct {
			Nodes []WorkflowState `json:"nodes"`
		} `json:"workflowStates"`
	}

	if err := c.execute(ctx, query, variables, &result); err != nil {
		return nil, err
	}

	return result.WorkflowStates.Nodes, nil
}

// GetLabels returns all labels for a team
func (c *Client) GetLabels(ctx context.Context, teamID string) ([]Label, error) {
	query := `
		query Labels($teamId: ID!) {
			issueLabels(filter: { team: { id: { eq: $teamId } } }) {
				nodes {
					id
					name
					description
					color
				}
			}
		}
	`

	variables := map[string]interface{}{
		"teamId": teamID,
	}

	var result struct {
		IssueLabels struct {
			Nodes []Label `json:"nodes"`
		} `json:"issueLabels"`
	}

	if err := c.execute(ctx, query, variables, &result); err != nil {
		return nil, err
	}

	return result.IssueLabels.Nodes, nil
}

// GetUsers returns all users in the organization
func (c *Client) GetUsers(ctx context.Context) ([]User, error) {
	query := `
		query Users {
			users {
				nodes {
					id
					name
					displayName
					email
					avatarUrl
					active
				}
			}
		}
	`

	var result struct {
		Users struct {
			Nodes []User `json:"nodes"`
		} `json:"users"`
	}

	if err := c.execute(ctx, query, nil, &result); err != nil {
		return nil, err
	}

	return result.Users.Nodes, nil
}
