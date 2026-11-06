// Package importer — Jira 导入器。
package importer

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// JiraImporter Jira 导入器（支持 Cloud REST API v3 / Data Center）。
type JiraImporter struct {
	httpClient *http.Client
}

// NewJiraImporter 构造导入器。
func NewJiraImporter() *JiraImporter {
	return &JiraImporter{httpClient: &http.Client{Timeout: 30 * time.Second}}
}

// Source 返回来源。
func (j *JiraImporter) Source() Source { return SourceJira }

// Fetch 拉取 Jira Issue。
func (j *JiraImporter) Fetch(ctx context.Context, creds Credentials, opts ImportOptions) ([]WorkItemRecord, error) {
	if err := creds.Validate(SourceJira); err != nil {
		return nil, err
	}
	base := strings.TrimRight(creds.BaseURL, "/")
	jql := fmt.Sprintf("project = %s ORDER BY created ASC", creds.ProjectKey)

	var out []WorkItemRecord
	startAt := 0
	maxResults := 50

	for {
		payload := map[string]any{
			"jql":        jql,
			"startAt":    startAt,
			"maxResults": maxResults,
			"fields": []string{
				"summary", "description", "issuetype", "priority", "status",
				"assignee", "reporter", "labels", "created", "updated", "duedate",
				"parent", "customfield_10016",
			},
		}
		body, _ := json.Marshal(payload)

		reqURL := base + "/rest/api/3/search"
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, reqURL, bytes.NewReader(body))
		if err != nil {
			return nil, err
		}
		req.Header.Set("Content-Type", "application/json")
		setBasicAuth(req, creds.Email, creds.APIToken)

		resp, err := j.httpClient.Do(req)
		if err != nil {
			return nil, err
		}
		data, err := io.ReadAll(io.LimitReader(resp.Body, 16<<20))
		resp.Body.Close()
		if err != nil {
			return nil, err
		}
		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("jira API returned %d: %s", resp.StatusCode, truncate(string(data), 300))
		}

		var result struct {
			Issues []jiraIssue `json:"issues"`
			Total  int         `json:"total"`
		}
		if err := json.Unmarshal(data, &result); err != nil {
			return nil, err
		}

		for _, iss := range result.Issues {
			out = append(out, mapJiraIssue(iss, opts))
		}

		startAt += len(result.Issues)
		if startAt >= result.Total || len(result.Issues) == 0 {
			break
		}
	}
	return out, nil
}

// --- Jira 数据结构 ---

type jiraIssue struct {
	Key    string `json:"key"`
	Fields struct {
		Summary     string `json:"summary"`
		Description any    `json:"description"`
		IssueType   struct {
			Name string `json:"name"`
		} `json:"issuetype"`
		Priority struct {
			Name string `json:"name"`
		} `json:"priority"`
		Status struct {
			Name string `json:"name"`
		} `json:"status"`
		Assignee *struct {
			EmailAddress string `json:"emailAddress"`
			DisplayName  string `json:"displayName"`
		} `json:"assignee"`
		Reporter *struct {
			EmailAddress string `json:"emailAddress"`
		} `json:"reporter"`
		Labels  []string `json:"labels"`
		Created string   `json:"created"`
		Updated string   `json:"updated"`
		DueDate string   `json:"duedate"`
		Parent  *struct {
			Key string `json:"key"`
		} `json:"parent"`
		// customfield_10016 通常是 Story Points（视实例配置而定）
		StoryPoints *float64 `json:"customfield_10016"`
	} `json:"fields"`
}

// mapJiraIssue 将 Jira Issue 映射为统一模型。
func mapJiraIssue(iss jiraIssue, opts ImportOptions) WorkItemRecord {
	rec := WorkItemRecord{
		ExternalID:       iss.Key,
		Title:            iss.Fields.Summary,
		TypeCode:         mapJiraType(iss.Fields.IssueType.Name, opts.DefaultType),
		Priority:         mapJiraPriority(iss.Fields.Priority.Name),
		StateName:        iss.Fields.Status.Name,
		Labels:           iss.Fields.Labels,
		Estimate:         iss.Fields.StoryPoints,
		ParentExternalID: "",
	}
	if iss.Fields.Assignee != nil {
		rec.AssigneeEmail = iss.Fields.Assignee.EmailAddress
	}
	if iss.Fields.Reporter != nil {
		rec.CreatorEmail = iss.Fields.Reporter.EmailAddress
	}
	if iss.Fields.Parent != nil {
		rec.ParentExternalID = iss.Fields.Parent.Key
	}
	rec.Description = flattenJiraDescription(iss.Fields.Description)
	rec.CreatedAt = parseJiraTime(iss.Fields.Created)
	rec.UpdatedAt = parseJiraTime(iss.Fields.Updated)
	if iss.Fields.DueDate != "" {
		if t, err := time.Parse("2006-01-02", iss.Fields.DueDate); err == nil {
			rec.DueDate = &t
		}
	}
	return rec
}

// flattenJiraDescription 将 Jira ADF（Atlassian Document Format）拍平为纯文本。
func flattenJiraDescription(v any) string {
	switch d := v.(type) {
	case string:
		return d
	case nil:
		return ""
	}
	// ADF 结构化文档：递归提取 text 字段
	var sb strings.Builder
	extractADFText(v, &sb)
	if sb.Len() == 0 {
		// 回退为 JSON
		b, _ := json.Marshal(v)
		return string(b)
	}
	return sb.String()
}

func extractADFText(v any, sb *strings.Builder) {
	switch t := v.(type) {
	case map[string]any:
		if txt, ok := t["text"].(string); ok {
			sb.WriteString(txt)
		}
		if content, ok := t["content"].([]any); ok {
			for _, c := range content {
				extractADFText(c, sb)
			}
			sb.WriteString("\n")
		}
	case []any:
		for _, c := range t {
			extractADFText(c, sb)
		}
	}
}

func mapJiraType(name, def string) string {
	switch strings.ToLower(name) {
	case "story", "epic", "feature", "需求":
		return "requirement"
	case "bug", "defect", "缺陷":
		return "defect"
	case "task", "sub-task", "subtask", "任务":
		return "task"
	}
	if def != "" {
		return def
	}
	return "task"
}

func mapJiraPriority(name string) string {
	switch strings.ToLower(name) {
	case "highest", "urgent", "p0":
		return "urgent"
	case "high", "p1":
		return "high"
	case "low", "lowest", "p3", "p4":
		return "low"
	}
	return "medium"
}

func parseJiraTime(s string) *time.Time {
	if s == "" {
		return nil
	}
	for _, layout := range []string{time.RFC3339, "2006-01-02T15:04:05.000-0700", "2006-01-02"} {
		if t, err := time.Parse(layout, s); err == nil {
			return &t
		}
	}
	return nil
}

// --- helpers ---

func setBasicAuth(req *http.Request, email, token string) {
	req.Header.Set("Authorization", "Basic "+base64Encode(email+":"+token))
}

func base64Encode(s string) string {
	const chars = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/"
	src := []byte(s)
	var out strings.Builder
	for i := 0; i < len(src); i += 3 {
		var b [3]byte
		n := copy(b[:], src[i:])
		out.WriteByte(chars[b[0]>>2])
		out.WriteByte(chars[(b[0]&0x03)<<4|b[1]>>4])
		if n > 1 {
			out.WriteByte(chars[(b[1]&0x0f)<<2|b[2]>>6])
		} else {
			out.WriteByte('=')
		}
		if n > 2 {
			out.WriteByte(chars[b[2]&0x3f])
		} else {
			out.WriteByte('=')
		}
	}
	return out.String()
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n]
}

var _ = url.Values{}
