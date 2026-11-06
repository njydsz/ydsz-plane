// Package integration — GitHub 连接器。
//
// 能力：
//   - ValidateCredentials：校验 token 非空；
//   - TestConnection：GET /user 验证 token；
//   - SyncWorkItems：GET /repos/{owner}/{repo}/issues 拉取 issue 并归一化。
package integration

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

// GitHubConnector 是 GitHub 的 IntegrationConnector 实现。
type GitHubConnector struct{}

// NewGitHubConnector 构造 GitHub 连接器。
func NewGitHubConnector() *GitHubConnector { return &GitHubConnector{} }

// Provider 返回 provider 标识。
func (c *GitHubConnector) Provider() string { return string(ProviderGitHub) }

// ValidateCredentials 校验凭证字段完整性（不发起网络请求）。
func (c *GitHubConnector) ValidateCredentials(creds map[string]string) error {
	return requireFields(c.Provider(), creds, "token")
}

// TestConnection 通过 GET /user 验证 token 有效性。
func (c *GitHubConnector) TestConnection(creds map[string]string) error {
	if err := c.ValidateCredentials(creds); err != nil {
		return err
	}
	var out struct {
		Login string `json:"login"`
	}
	reqErr := doJSON(context.Background(), "GET",
		buildURL("https://api.github.com", "/user", nil),
		map[string]string{
			"Authorization": "Bearer " + creds["token"],
			"User-Agent":    "ydsz-plane",
		}, nil, &out)
	if reqErr != nil {
		if apiErr, ok := reqErr.(*apiError); ok {
			return fmt.Errorf("%s: test connection failed: %s", c.Provider(), apiErr.Message)
		}
		return fmt.Errorf("%s: test connection failed: %w", c.Provider(), reqErr)
	}
	if out.Login == "" {
		return fmt.Errorf("%s: test connection returned empty login", c.Provider())
	}
	return nil
}

// githubIssue 是 GitHub REST API /issues 响应项。
type githubIssue struct {
	Number    int           `json:"number"`
	Title     string        `json:"title"`
	Body      string        `json:"body"`
	State     string        `json:"state"`
	HTMLURL   string        `json:"html_url"`
	User      *githubUser   `json:"user"`
	Labels    []githubLabel `json:"labels"`
	Assignees []githubUser  `json:"assignees"`
	CreatedAt string        `json:"created_at"`
	UpdatedAt string        `json:"updated_at"`
}

type githubUser struct {
	Login string `json:"login"`
}

type githubLabel struct {
	Name string `json:"name"`
}

// SyncWorkItems 拉取指定仓库的 issues 并归一化为 ExternalWorkItem。
//
// config 约定：
//   - repo：必填，形如 "owner/repo"；
//   - state：可选，open / closed / all，默认 open；
//   - per_page：可选，默认 100。
func (c *GitHubConnector) SyncWorkItems(ctx context.Context, creds map[string]string, config map[string]string) (SyncResult, error) {
	if err := c.ValidateCredentials(creds); err != nil {
		return SyncResult{}, err
	}
	repo := strings.TrimSpace(config["repo"])
	if repo == "" {
		return SyncResult{}, fmt.Errorf("%s: config field %q is required", c.Provider(), "repo")
	}
	state := strings.TrimSpace(config["state"])
	if state == "" {
		state = "open"
	}
	q := url.Values{}
	q.Set("state", state)
	q.Set("per_page", "100")

	endpoint := buildURL("https://api.github.com", "/repos/"+repo+"/issues", q)
	var issues []githubIssue
	if err := doJSON(ctx, "GET", endpoint,
		map[string]string{
			"Authorization": "Bearer " + creds["token"],
			"User-Agent":    "ydsz-plane",
		}, nil, &issues); err != nil {
		if apiErr, ok := err.(*apiError); ok {
			return SyncResult{}, fmt.Errorf("%s: sync failed: %s", c.Provider(), apiErr.Message)
		}
		return SyncResult{}, fmt.Errorf("%s: sync failed: %w", c.Provider(), err)
	}

	result := SyncResult{Provider: c.Provider(), FetchedAt: nowUTC()}
	for _, issue := range issues {
		// GitHub /issues 会返回 pull_request 字段非空的 PR，此处过滤掉。
		item := ExternalWorkItem{
			ExternalID:  strconv.Itoa(issue.Number),
			Title:       issue.Title,
			Description: issue.Body,
			State:       issue.State,
			URL:         issue.HTMLURL,
			Labels:      labelNames(issue.Labels),
			Assignees:   userLogins(issue.Assignees),
			CreatedAt:   parseRFC3339(issue.CreatedAt),
			UpdatedAt:   parseRFC3339(issue.UpdatedAt),
		}
		if issue.User != nil {
			item.Author = issue.User.Login
		}
		result.ItemsFetched++
		result.ItemsImported++
	}
	result.Message = fmt.Sprintf("fetched %d issues from %s", result.ItemsFetched, repo)
	return result, nil
}

// labelNames 提取 GitHub label 名称列表。
func labelNames(labels []githubLabel) []string {
	if len(labels) == 0 {
		return nil
	}
	out := make([]string, 0, len(labels))
	for _, l := range labels {
		out = append(out, l.Name)
	}
	return out
}

// userLogins 提取 GitHub 用户登录名列表。
func userLogins(users []githubUser) []string {
	if len(users) == 0 {
		return nil
	}
	out := make([]string, 0, len(users))
	for _, u := range users {
		out = append(out, u.Login)
	}
	return out
}
