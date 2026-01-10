package linear

import (
	"context"
)

// CreateIssue creates a new issue
func (c *Client) CreateIssue(ctx context.Context, input IssueCreateInput) (*Issue, error) {
	query := `
		mutation CreateIssue($input: IssueCreateInput!) {
			issueCreate(input: $input) {
				success
				issue {
					id
					identifier
					title
					description
					priority
					createdAt
					updatedAt
					url
					branchName
					state {
						id
						name
						color
						type
					}
					assignee {
						id
						name
						displayName
					}
					team {
						id
						name
						key
					}
					project {
						id
						name
					}
					labels {
						nodes {
							id
							name
							color
						}
					}
				}
			}
		}
	`

	variables := map[string]interface{}{
		"input": input,
	}

	var result struct {
		IssueCreate struct {
			Success bool      `json:"success"`
			Issue   *rawIssue `json:"issue"`
		} `json:"issueCreate"`
	}

	if err := c.execute(ctx, query, variables, &result); err != nil {
		return nil, err
	}

	if result.IssueCreate.Issue == nil {
		return nil, nil
	}

	issues := convertIssues([]rawIssue{*result.IssueCreate.Issue})
	return &issues[0], nil
}

// UpdateIssue updates an existing issue
func (c *Client) UpdateIssue(ctx context.Context, issueID string, input IssueUpdateInput) (*Issue, error) {
	query := `
		mutation UpdateIssue($id: ID!, $input: IssueUpdateInput!) {
			issueUpdate(id: $id, input: $input) {
				success
				issue {
					id
					identifier
					title
					description
					priority
					createdAt
					updatedAt
					url
					branchName
					state {
						id
						name
						color
						type
					}
					assignee {
						id
						name
						displayName
					}
					team {
						id
						name
						key
					}
					project {
						id
						name
					}
					labels {
						nodes {
							id
							name
							color
						}
					}
				}
			}
		}
	`

	variables := map[string]interface{}{
		"id":    issueID,
		"input": input,
	}

	var result struct {
		IssueUpdate struct {
			Success bool      `json:"success"`
			Issue   *rawIssue `json:"issue"`
		} `json:"issueUpdate"`
	}

	if err := c.execute(ctx, query, variables, &result); err != nil {
		return nil, err
	}

	if result.IssueUpdate.Issue == nil {
		return nil, nil
	}

	issues := convertIssues([]rawIssue{*result.IssueUpdate.Issue})
	return &issues[0], nil
}

// UpdateIssueState updates only the state of an issue
func (c *Client) UpdateIssueState(ctx context.Context, issueID string, stateID string) (*Issue, error) {
	return c.UpdateIssue(ctx, issueID, IssueUpdateInput{
		StateID: &stateID,
	})
}

// UpdateIssueAssignee updates only the assignee of an issue
func (c *Client) UpdateIssueAssignee(ctx context.Context, issueID string, assigneeID *string) (*Issue, error) {
	return c.UpdateIssue(ctx, issueID, IssueUpdateInput{
		AssigneeID: assigneeID,
	})
}

// UpdateIssuePriority updates only the priority of an issue
func (c *Client) UpdateIssuePriority(ctx context.Context, issueID string, priority int) (*Issue, error) {
	return c.UpdateIssue(ctx, issueID, IssueUpdateInput{
		Priority: &priority,
	})
}

// UpdateIssueLabels updates the labels of an issue
func (c *Client) UpdateIssueLabels(ctx context.Context, issueID string, labelIDs []string) (*Issue, error) {
	return c.UpdateIssue(ctx, issueID, IssueUpdateInput{
		LabelIDs: labelIDs,
	})
}

// AddIssueLabel adds a label to an issue
func (c *Client) AddIssueLabel(ctx context.Context, issueID string, labelID string) error {
	query := `
		mutation AddIssueLabel($issueId: ID!, $labelId: ID!) {
			issueAddLabel(id: $issueId, labelId: $labelId) {
				success
			}
		}
	`

	variables := map[string]interface{}{
		"issueId": issueID,
		"labelId": labelID,
	}

	var result struct {
		IssueAddLabel struct {
			Success bool `json:"success"`
		} `json:"issueAddLabel"`
	}

	return c.execute(ctx, query, variables, &result)
}

// RemoveIssueLabel removes a label from an issue
func (c *Client) RemoveIssueLabel(ctx context.Context, issueID string, labelID string) error {
	query := `
		mutation RemoveIssueLabel($issueId: ID!, $labelId: ID!) {
			issueRemoveLabel(id: $issueId, labelId: $labelId) {
				success
			}
		}
	`

	variables := map[string]interface{}{
		"issueId": issueID,
		"labelId": labelID,
	}

	var result struct {
		IssueRemoveLabel struct {
			Success bool `json:"success"`
		} `json:"issueRemoveLabel"`
	}

	return c.execute(ctx, query, variables, &result)
}

// DeleteIssue moves an issue to trash
func (c *Client) DeleteIssue(ctx context.Context, issueID string) error {
	query := `
		mutation DeleteIssue($issueId: String!) {
			issueDelete(id: $issueId) {
				success
			}
		}
	`

	variables := map[string]interface{}{
		"issueId": issueID,
	}

	var result struct {
		IssueDelete struct {
			Success bool `json:"success"`
		} `json:"issueDelete"`
	}

	return c.execute(ctx, query, variables, &result)
}
