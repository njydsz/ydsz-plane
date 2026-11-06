// Package importer — Asana 导入器（REST API）。
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

// AsanaImporter Asana 导入器。
type AsanaImporter struct {
	httpClient *http.Client
}

// NewAsanaImporter 构造导入器。
func NewAsanaImporter() *AsanaImporter {
	return &AsanaImporter{httpClient: &http.Client{Timeout: 30 * time.Second}}
}

// Source 返回来源。
func (a *AsanaImporter) Source() Source { return SourceAsana }

// AsanaAPIBase Asana REST API 基地址。
const AsanaAPIBase = "https://app.asana.com/api/1.0"

// Fetch 拉取 Asana Tasks。
func (a *AsanaImporter) Fetch(ctx context.Context, creds Credentials, opts ImportOptions) ([]WorkItemRecord, error) {
	if err := creds.Validate(SourceAsana); err != nil {
		return nil, err
	}

	// 先取项目下的任务（需要 project_gid 或 workspace 下的全部）
	path := "/tasks?opt_fields=gid,name,notes,resource_subtype,assignee.name,assignee.email,created_by.email,created_at,modified_at,due_on,parent.name,parent.gid,tags.name&limit=100"
	if creds.TeamID != "" {
		path = "/projects/" + creds.TeamID + "/tasks?opt_fields=gid,name,notes,resource_subtype,assignee.name,assignee.email,created_by.email,created_at,modified_at,due_on,parent.name,parent.gid,tags.name&limit=100"
	}

	var out []WorkItemRecord
	for path != "" {
		reqURL := AsanaAPIBase + path
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
		if err != nil {
			return nil, err
		}
		req.Header.Set("Authorization", "Bearer "+creds.APIToken)
		req.Header.Set("Accept", "application/json")

		resp, err := a.httpClient.Do(req)
		if err != nil {
			return nil, err
		}
		data, err := io.ReadAll(io.LimitReader(resp.Body, 16<<20))
		resp.Body.Close()
		if err != nil {
			return nil, err
		}
		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("asana API returned %d: %s", resp.StatusCode, truncate(string(data), 300))
		}

		var result struct {
			Data     []asanaTask `json:"data"`
			NextPage *struct {
				Path string `json:"path"`
			} `json:"next_page"`
		}
		if err := json.Unmarshal(data, &result); err != nil {
			return nil, err
		}

		for _, t := range result.Data {
			out = append(out, mapAsanaTask(t, opts))
		}

		if result.NextPage == nil {
			break
		}
		path = result.NextPage.Path
	}
	return out, nil
}

// --- Asana 数据结构 ---

type asanaTask struct {
	GID          string `json:"gid"`
	Name         string `json:"name"`
	Notes        string `json:"notes"`
	ResourceType string `json:"resource_subtype"`
	Assignee     *struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	} `json:"assignee"`
	CreatedBy *struct {
		Email string `json:"email"`
	} `json:"created_by"`
	CreatedAt  string `json:"created_at"`
	ModifiedAt string `json:"modified_at"`
	DueOn      string `json:"due_on"`
	Parent     *struct {
		GID  string `json:"gid"`
		Name string `json:"name"`
	} `json:"parent"`
	Tags []struct {
		Name string `json:"name"`
	} `json:"tags"`
}

func mapAsanaTask(t asanaTask, opts ImportOptions) WorkItemRecord {
	rec := WorkItemRecord{
		ExternalID:  t.GID,
		Title:       t.Name,
		Description: t.Notes,
		TypeCode:    mapAsanaType(t.ResourceType, opts.DefaultType),
	}
	if t.Assignee != nil {
		rec.AssigneeEmail = t.Assignee.Email
	}
	if t.CreatedBy != nil {
		rec.CreatorEmail = t.CreatedBy.Email
	}
	for _, tag := range t.Tags {
		rec.Labels = append(rec.Labels, tag.Name)
	}
	if t.Parent != nil {
		rec.ParentExternalID = t.Parent.GID
	}
	rec.CreatedAt = parseISOTime(t.CreatedAt)
	rec.UpdatedAt = parseISOTime(t.ModifiedAt)
	if t.DueOn != "" {
		if d, err := time.Parse("2006-01-02", t.DueOn); err == nil {
			rec.DueDate = &d
		}
	}
	return rec
}

// mapAsanaType 映射 Asana 子类型。
func mapAsanaType(subtype, def string) string {
	switch strings.ToLower(subtype) {
	case "default_task", "task":
		return "task"
	case "milestone":
		return "requirement"
	}
	if def != "" {
		return def
	}
	return "task"
}

var _ = bytes.NewReader
