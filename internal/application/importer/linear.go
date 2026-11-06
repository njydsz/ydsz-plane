// Package importer — Linear 导入器（GraphQL API）。
package importer

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

// LinearImporter Linear 导入器。
type LinearImporter struct {
	httpClient *http.Client
}

// NewLinearImporter 构造导入器。
func NewLinearImporter() *LinearImporter {
	return &LinearImporter{httpClient: &http.Client{Timeout: 30 * time.Second}}
}

// Source 返回来源。
func (l *LinearImporter) Source() Source { return SourceLinear }

// linearGraphQL Linear GraphQL 端点。
const linearGraphQL = "https://api.linear.app/graphql"

// Fetch 拉取 Linear Issue。
func (l *LinearImporter) Fetch(ctx context.Context, creds Credentials, opts ImportOptions) ([]WorkItemRecord, error) {
	if err := creds.Validate(SourceLinear); err != nil {
		return nil, err
	}

	query := `query Issues($teamId: String, $cursor: String) {
		issues(filter: { team: { key: { eq: $teamId } } }, first: 50, after: $cursor) {
			nodes {
				id identifier title description
				type { name }
				priority
				state { name }
				assignee { email name }
				creator { email }
				labels { nodes { name } }
				createdAt updatedAt dueDate
				parent { identifier }
				estimate
			}
			pageInfo { hasNextPage endCursor }
		}
	}`

	variables := map[string]any{}
	if creds.TeamID != "" {
		variables["teamId"] = creds.TeamID
	}

	var out []WorkItemRecord
	cursor := ""

	for {
		if cursor != "" {
			variables["cursor"] = cursor
		}
		body, _ := json.Marshal(map[string]any{"query": query, "variables": variables})

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, linearGraphQL, bytes.NewReader(body))
		if err != nil {
			return nil, err
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", creds.APIToken)

		resp, err := l.httpClient.Do(req)
		if err != nil {
			return nil, err
		}
		data, err := io.ReadAll(io.LimitReader(resp.Body, 16<<20))
		resp.Body.Close()
		if err != nil {
			return nil, err
		}
		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("linear API returned %d: %s", resp.StatusCode, truncate(string(data), 300))
		}

		var result struct {
			Data struct {
				Issues struct {
					Nodes    []linearIssue `json:"nodes"`
					PageInfo struct {
						HasNextPage bool   `json:"hasNextPage"`
						EndCursor   string `json:"endCursor"`
					} `json:"pageInfo"`
				} `json:"issues"`
			} `json:"data"`
			Errors []struct {
				Message string `json:"message"`
			} `json:"errors"`
		}
		if err := json.Unmarshal(data, &result); err != nil {
			return nil, err
		}
		if len(result.Errors) > 0 {
			return nil, fmt.Errorf("linear: %s", result.Errors[0].Message)
		}

		for _, iss := range result.Data.Issues.Nodes {
			out = append(out, mapLinearIssue(iss, opts))
		}

		if !result.Data.Issues.PageInfo.HasNextPage {
			break
		}
		cursor = result.Data.Issues.PageInfo.EndCursor
	}
	return out, nil
}

// --- Linear 数据结构 ---

type linearIssue struct {
	ID          string `json:"id"`
	Identifier  string `json:"identifier"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Type        struct {
		Name string `json:"name"`
	} `json:"type"`
	Priority float64 `json:"priority"`
	State    struct {
		Name string `json:"name"`
	} `json:"state"`
	Assignee *struct {
		Email string `json:"email"`
		Name  string `json:"name"`
	} `json:"assignee"`
	Creator *struct {
		Email string `json:"email"`
	} `json:"creator"`
	Labels struct {
		Nodes []struct {
			Name string `json:"name"`
		} `json:"nodes"`
	} `json:"labels"`
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
	DueDate   string `json:"dueDate"`
	Parent    *struct {
		Identifier string `json:"identifier"`
	} `json:"parent"`
	Estimate *float64 `json:"estimate"`
}

func mapLinearIssue(iss linearIssue, opts ImportOptions) WorkItemRecord {
	rec := WorkItemRecord{
		ExternalID:  iss.Identifier,
		Title:       iss.Title,
		Description: iss.Description,
		TypeCode:    mapLinearType(iss.Type.Name, opts.DefaultType),
		Priority:    mapLinearPriority(iss.Priority),
		StateName:   iss.State.Name,
		Estimate:    iss.Estimate,
	}
	if iss.Assignee != nil {
		rec.AssigneeEmail = iss.Assignee.Email
	}
	if iss.Creator != nil {
		rec.CreatorEmail = iss.Creator.Email
	}
	for _, l := range iss.Labels.Nodes {
		rec.Labels = append(rec.Labels, l.Name)
	}
	if iss.Parent != nil {
		rec.ParentExternalID = iss.Parent.Identifier
	}
	rec.CreatedAt = parseISOTime(iss.CreatedAt)
	rec.UpdatedAt = parseISOTime(iss.UpdatedAt)
	if iss.DueDate != "" {
		if t, err := time.Parse("2006-01-02", iss.DueDate); err == nil {
			rec.DueDate = &t
		}
	}
	return rec
}

func mapLinearType(name, def string) string {
	switch strings.ToLower(name) {
	case "epic", "feature", "story":
		return "requirement"
	case "bug", "defect":
		return "defect"
	case "task", "subtask":
		return "task"
	}
	if def != "" {
		return def
	}
	return "task"
}

// Linear priority: 0 = none, 1 = urgent, 2 = high, 3 = medium, 4 = low
func mapLinearPriority(p float64) string {
	switch int(p) {
	case 1:
		return "urgent"
	case 2:
		return "high"
	case 3:
		return "medium"
	case 4:
		return "low"
	}
	return "none"
}

func parseISOTime(s string) *time.Time {
	if s == "" {
		return nil
	}
	for _, layout := range []string{time.RFC3339, "2006-01-02T15:04:05.000Z"} {
		if t, err := time.Parse(layout, s); err == nil {
			return &t
		}
	}
	return nil
}
