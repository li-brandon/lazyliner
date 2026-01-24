package linear

import (
	"context"
	"fmt"
)

// GetMyIssues returns issues assigned to the current user with pagination support
func (c *Client) GetMyIssues(ctx context.Context, limit int, after string) (IssueConnection, error) {
	if limit <= 0 {
		limit = 50
	}

	query := `
		query MyIssues($limit: Int!, $after: String) {
			viewer {
				assignedIssues(first: $limit, after: $after, orderBy: updatedAt) {
					nodes {
						id
						identifier
						title
						description
						priority
						estimate
						createdAt
						updatedAt
						startedAt
						completedAt
						canceledAt
						dueDate
						branchName
						url
						state {
							id
							name
							color
							type
							position
						}
						assignee {
							id
							name
							displayName
							email
						}
						creator {
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
							icon
							color
						}
						labels {
							nodes {
								id
								name
								color
							}
						}
					}
					pageInfo {
						hasNextPage
						hasPreviousPage
						startCursor
						endCursor
					}
				}
			}
		}
	`

	variables := map[string]interface{}{
		"limit": limit,
	}
	if after != "" {
		variables["after"] = after
	}

	var result struct {
		Viewer struct {
			AssignedIssues struct {
				Nodes    []rawIssue `json:"nodes"`
				PageInfo PageInfo   `json:"pageInfo"`
			} `json:"assignedIssues"`
		} `json:"viewer"`
	}

	if err := c.execute(ctx, query, variables, &result); err != nil {
		return IssueConnection{}, err
	}

	return IssueConnection{
		Nodes:    convertIssues(result.Viewer.AssignedIssues.Nodes),
		PageInfo: result.Viewer.AssignedIssues.PageInfo,
	}, nil
}

// GetIssues returns issues with optional filters and pagination support
func (c *Client) GetIssues(ctx context.Context, filter IssueFilter) (IssueConnection, error) {
	if filter.Limit <= 0 {
		filter.Limit = 50
	}

	query := `
		query Issues($limit: Int!, $filter: IssueFilter, $after: String) {
			issues(first: $limit, after: $after, filter: $filter, orderBy: updatedAt) {
				nodes {
					id
					identifier
					title
					description
					priority
					estimate
					createdAt
					updatedAt
					startedAt
					completedAt
					canceledAt
					dueDate
					branchName
					url
					state {
						id
						name
						color
						type
						position
					}
					assignee {
						id
						name
						displayName
						email
					}
					creator {
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
						icon
						color
					}
					labels {
						nodes {
							id
							name
							color
						}
					}
				}
				pageInfo {
					hasNextPage
					hasPreviousPage
					startCursor
					endCursor
				}
			}
		}
	`

	issueFilter := buildIssueFilter(filter)
	variables := map[string]interface{}{
		"limit":  filter.Limit,
		"filter": issueFilter,
	}
	if filter.After != "" {
		variables["after"] = filter.After
	}

	var result struct {
		Issues struct {
			Nodes    []rawIssue `json:"nodes"`
			PageInfo PageInfo   `json:"pageInfo"`
		} `json:"issues"`
	}

	if err := c.execute(ctx, query, variables, &result); err != nil {
		return IssueConnection{}, err
	}

	return IssueConnection{
		Nodes:    convertIssues(result.Issues.Nodes),
		PageInfo: result.Issues.PageInfo,
	}, nil
}

// GetIssue returns a single issue by ID or identifier
func (c *Client) GetIssue(ctx context.Context, idOrIdentifier string) (*Issue, error) {
	query := `
		query Issue($id: ID!) {
			issue(id: $id) {
				id
				identifier
				title
				description
				priority
				estimate
				createdAt
				updatedAt
				startedAt
				completedAt
				canceledAt
				dueDate
				branchName
				url
				state {
					id
					name
					color
					type
					position
				}
				assignee {
					id
					name
					displayName
					email
				}
				creator {
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
					icon
					color
				}
				labels {
					nodes {
						id
						name
						color
					}
				}
				parent {
					id
					identifier
					title
				}
			}
		}
	`

	variables := map[string]interface{}{
		"id": idOrIdentifier,
	}

	var result struct {
		Issue *rawIssue `json:"issue"`
	}

	if err := c.execute(ctx, query, variables, &result); err != nil {
		return nil, err
	}

	if result.Issue == nil {
		return nil, fmt.Errorf("issue not found: %s", idOrIdentifier)
	}

	issues := convertIssues([]rawIssue{*result.Issue})
	return &issues[0], nil
}

// SearchIssues searches for issues by text
func (c *Client) SearchIssues(ctx context.Context, query string, limit int) ([]Issue, error) {
	if limit <= 0 {
		limit = 20
	}

	gqlQuery := `
		query SearchIssues($query: String!, $limit: Int!) {
			searchIssues(query: $query, first: $limit) {
				nodes {
					id
					identifier
					title
					description
					priority
					createdAt
					updatedAt
					url
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
				}
			}
		}
	`

	variables := map[string]interface{}{
		"query": query,
		"limit": limit,
	}

	var result struct {
		SearchIssues struct {
			Nodes []rawIssue `json:"nodes"`
		} `json:"searchIssues"`
	}

	if err := c.execute(ctx, gqlQuery, variables, &result); err != nil {
		return nil, err
	}

	return convertIssues(result.SearchIssues.Nodes), nil
}

// rawIssue is the raw issue structure from the API with labels as connection
type rawIssue struct {
	Issue
	Labels struct {
		Nodes []Label `json:"nodes"`
	} `json:"labels"`
}

// convertIssues converts raw issues to the Issue type
func convertIssues(raw []rawIssue) []Issue {
	issues := make([]Issue, len(raw))
	for i, r := range raw {
		issues[i] = r.Issue
		issues[i].Labels = r.Labels.Nodes
	}
	return issues
}

// GetProjectIssues returns issues for a specific project with pagination support.
// By default excludes completed/canceled issues unless includeCompleted is true.
func (c *Client) GetProjectIssues(ctx context.Context, projectID string, limit int, includeCompleted bool, after string) (IssueConnection, error) {
	if limit <= 0 {
		limit = 100
	}

	query := `
		query ProjectIssues($limit: Int!, $filter: IssueFilter, $after: String) {
			issues(first: $limit, after: $after, filter: $filter, orderBy: updatedAt) {
				nodes {
					id
					identifier
					title
					description
					priority
					estimate
					createdAt
					updatedAt
					startedAt
					completedAt
					canceledAt
					dueDate
					branchName
					url
					state {
						id
						name
						color
						type
						position
					}
					assignee {
						id
						name
						displayName
						email
					}
					creator {
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
						icon
						color
					}
					labels {
						nodes {
							id
							name
							color
						}
					}
				}
				pageInfo {
					hasNextPage
					hasPreviousPage
					startCursor
					endCursor
				}
			}
		}
	`

	filter := map[string]interface{}{
		"project": map[string]interface{}{
			"id": map[string]interface{}{"eq": projectID},
		},
	}

	if !includeCompleted {
		filter["state"] = map[string]interface{}{
			"type": map[string]interface{}{
				"nin": []string{"completed", "canceled"},
			},
		}
	}

	variables := map[string]interface{}{
		"limit":  limit,
		"filter": filter,
	}
	if after != "" {
		variables["after"] = after
	}

	var result struct {
		Issues struct {
			Nodes    []rawIssue `json:"nodes"`
			PageInfo PageInfo   `json:"pageInfo"`
		} `json:"issues"`
	}

	if err := c.execute(ctx, query, variables, &result); err != nil {
		return IssueConnection{}, err
	}

	return IssueConnection{
		Nodes:    convertIssues(result.Issues.Nodes),
		PageInfo: result.Issues.PageInfo,
	}, nil
}

func buildIssueFilter(filter IssueFilter) map[string]interface{} {
	f := make(map[string]interface{})

	if filter.TeamID != "" {
		f["team"] = map[string]interface{}{
			"id": map[string]interface{}{"eq": filter.TeamID},
		}
	}

	if filter.ProjectID != "" {
		f["project"] = map[string]interface{}{
			"id": map[string]interface{}{"eq": filter.ProjectID},
		}
	}

	if filter.AssigneeID != "" {
		f["assignee"] = map[string]interface{}{
			"id": map[string]interface{}{"eq": filter.AssigneeID},
		}
	}

	if filter.StateType != "" {
		f["state"] = map[string]interface{}{
			"type": map[string]interface{}{"eq": filter.StateType},
		}
	}

	return f
}
