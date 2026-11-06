// projectctl — Projects-as-Code CLI。
//
// 能力：
//   - pull  <project_id>：从 API 导出既有项目为 YAML（默认打印到 stdout，-o 写文件）；
//   - apply <file.yaml> [--project-id <id>] [--dry-run]：把 YAML 规范下发到 API（创建或更新项目）；
//   - diff  <file.yaml> --project-id <id>：预览规范与既有项目的差异。
//
// 配置（环境变量）：
//
//	YDSZ_API_BASE   API 基础地址，如 http://localhost:8080（默认）
//	YDSZ_TOKEN      访问令牌（Authorization: Bearer <token>）
//	YDSZ_WORKSPACE  工作空间 ID（也可用 --workspace 覆盖）
//
// 示例：
//
//	YDSZ_API_BASE=http://localhost:8080 YDSZ_TOKEN=xxx YDSZ_WORKSPACE=1 \
//	  projectctl pull 123 -o apollo.yaml
//
//	YDSZ_API_BASE=http://localhost:8080 YDSZ_TOKEN=xxx YDSZ_WORKSPACE=1 \
//	  projectctl apply apollo.yaml
//
// YDSZ_API_BASE=http://localhost:8080 YDSZ_TOKEN=xxx YDSZ_WORKSPACE=1 \
// projectctl diff apollo.yaml --project-id 123
package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

// --- 配置 ---

type config struct {
	apiBase   string
	token     string
	workspace int64
}

func loadConfig(wsOverride int64) config {
	cfg := config{
		apiBase:   strings.TrimRight(getEnv("YDSZ_API_BASE", "http://localhost:8080"), "/"),
		token:     os.Getenv("YDSZ_TOKEN"),
		workspace: wsOverride,
	}
	if cfg.workspace == 0 {
		cfg.workspace, _ = strconv.ParseInt(os.Getenv("YDSZ_WORKSPACE"), 10, 64)
	}
	return cfg
}

func getEnv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

// --- HTTP 客户端 ---

type client struct {
	cfg  config
	http *http.Client
}

func newClient(cfg config) *client {
	return &client{cfg: cfg, http: &http.Client{Timeout: 30 * time.Second}}
}

// do 发起请求并返回响应体。
func (c *client) do(method, path string, body any) ([]byte, int, error) {
	var rdr io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, 0, err
		}
		rdr = bytes.NewReader(b)
	}
	req, err := http.NewRequest(method, c.cfg.apiBase+path, rdr)
	if err != nil {
		return nil, 0, err
	}
	req.Header.Set("Content-Type", "application/json")
	if c.cfg.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.cfg.token)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, err
	}
	return data, resp.StatusCode, nil
}

// --- 命令 ---

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}
	cmd := os.Args[1]
	args := os.Args[2:]

	switch cmd {
	case "pull":
		os.Exit(cmdPull(args))
	case "apply":
		os.Exit(cmdApply(args))
	case "diff":
		os.Exit(cmdDiff(args))
	case "help", "-h", "--help":
		usage()
		os.Exit(0)
	default:
		fmt.Fprintf(os.Stderr, "未知命令: %s\n\n", cmd)
		usage()
		os.Exit(2)
	}
}

func usage() {
	fmt.Fprint(os.Stderr, `projectctl — Projects-as-Code CLI

用法:
  projectctl pull  <project_id> [-o out.yaml] [-w workspace_id]
  projectctl apply <file.yaml> [--project-id id] [--dry-run] [-w workspace_id]
  projectctl diff  <file.yaml> --project-id <id> [-w workspace_id]

环境变量:
  YDSZ_API_BASE    API 基础地址（默认 http://localhost:8080）
  YDSZ_TOKEN       访问令牌（Bearer）
  YDSZ_WORKSPACE   默认工作空间 ID

示例:
  projectctl pull 123 -o apollo.yaml
  projectctl apply apollo.yaml
  projectctl diff apollo.yaml --project-id 123
`)
}

// cmdPull 导出项目为 YAML。
func cmdPull(args []string) int {
	fs := flag.NewFlagSet("pull", flag.ContinueOnError)
	out := fs.String("o", "", "输出文件（默认 stdout）")
	ws := fs.Int64("w", 0, "工作空间 ID")
	_ = fs.Parse(args)
	rest := fs.Args()
	if len(rest) < 1 {
		fmt.Fprintln(os.Stderr, "错误: 缺少 project_id")
		return 2
	}
	projectID, err := strconv.ParseInt(rest[0], 10, 64)
	if err != nil || projectID <= 0 {
		fmt.Fprintf(os.Stderr, "错误: 无效的 project_id %q\n", rest[0])
		return 2
	}

	cfg := loadConfig(*ws)
	if cfg.workspace == 0 {
		fmt.Fprintln(os.Stderr, "错误: 未指定工作空间 ID（--workspace 或 YDSZ_WORKSPACE）")
		return 2
	}

	c := newClient(cfg)
	path := fmt.Sprintf("/api/v1/workspaces/%d/project-code/export/%d", cfg.workspace, projectID)
	data, status, err := c.do(http.MethodGet, path, nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "请求失败: %v\n", err)
		return 1
	}
	if status != http.StatusOK {
		fmt.Fprintf(os.Stderr, "导出失败 (HTTP %d): %s\n", status, extractMessage(data))
		return 1
	}

	var resp struct {
		YAML string `json:"yaml"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		fmt.Fprintf(os.Stderr, "解析响应失败: %v\n", err)
		return 1
	}
	if resp.YAML == "" {
		fmt.Fprintln(os.Stderr, "响应中未包含 YAML 内容")
		return 1
	}

	if *out != "" {
		if err := os.WriteFile(*out, []byte(resp.YAML), 0o644); err != nil {
			fmt.Fprintf(os.Stderr, "写文件失败: %v\n", err)
			return 1
		}
		fmt.Printf("已导出到 %s\n", *out)
		return 0
	}
	fmt.Print(resp.YAML)
	if !strings.HasSuffix(resp.YAML, "\n") {
		fmt.Println()
	}
	return 0
}

// cmdApply 应用 YAML 规范。
func cmdApply(args []string) int {
	fs := flag.NewFlagSet("apply", flag.ContinueOnError)
	projectID := fs.Int64("project-id", 0, "目标项目 ID（0=新建）")
	dryRun := fs.Bool("dry-run", false, "仅预览差异，不实际下发")
	ws := fs.Int64("w", 0, "工作空间 ID")
	_ = fs.Parse(args)
	rest := fs.Args()
	if len(rest) < 1 {
		fmt.Fprintln(os.Stderr, "错误: 缺少 YAML 文件路径")
		return 2
	}
	yamlText, err := os.ReadFile(rest[0])
	if err != nil {
		fmt.Fprintf(os.Stderr, "读取文件失败: %v\n", err)
		return 2
	}

	cfg := loadConfig(*ws)
	if cfg.workspace == 0 {
		fmt.Fprintln(os.Stderr, "错误: 未指定工作空间 ID（--workspace 或 YDSZ_WORKSPACE）")
		return 2
	}

	c := newClient(cfg)

	// dry-run：先做 diff。
	if *dryRun {
		if *projectID <= 0 {
			fmt.Fprintln(os.Stderr, "错误: --dry-run 需要 --project-id")
			return 2
		}
		return runDiff(c, cfg.workspace, *projectID, string(yamlText))
	}

	path := fmt.Sprintf("/api/v1/workspaces/%d/project-code/apply", cfg.workspace)
	body := map[string]any{"yaml": string(yamlText)}
	if *projectID > 0 {
		body["project_id"] = *projectID
	}
	data, status, err := c.do(http.MethodPost, path, body)
	if err != nil {
		fmt.Fprintf(os.Stderr, "请求失败: %v\n", err)
		return 1
	}
	if status != http.StatusOK && status != http.StatusCreated {
		fmt.Fprintf(os.Stderr, "apply 失败 (HTTP %d): %s\n", status, extractMessage(data))
		return 1
	}

	var resp struct {
		ProjectID int64  `json:"project_id"`
		Created   bool   `json:"created"`
		Slug      string `json:"project_slug"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		fmt.Fprintf(os.Stderr, "解析响应失败: %v\n", err)
		return 1
	}
	action := "更新"
	if resp.Created {
		action = "创建"
	}
	fmt.Printf("已%s项目 %d（slug=%s）\n", action, resp.ProjectID, resp.Slug)
	return 0
}

// cmdDiff 预览差异。
func cmdDiff(args []string) int {
	fs := flag.NewFlagSet("diff", flag.ContinueOnError)
	projectID := fs.Int64("project-id", 0, "目标项目 ID")
	ws := fs.Int64("w", 0, "工作空间 ID")
	_ = fs.Parse(args)
	rest := fs.Args()
	if len(rest) < 1 {
		fmt.Fprintln(os.Stderr, "错误: 缺少 YAML 文件路径")
		return 2
	}
	if *projectID <= 0 {
		fmt.Fprintln(os.Stderr, "错误: diff 需要 --project-id")
		return 2
	}
	yamlText, err := os.ReadFile(rest[0])
	if err != nil {
		fmt.Fprintf(os.Stderr, "读取文件失败: %v\n", err)
		return 2
	}
	cfg := loadConfig(*ws)
	if cfg.workspace == 0 {
		fmt.Fprintln(os.Stderr, "错误: 未指定工作空间 ID（--workspace 或 YDSZ_WORKSPACE）")
		return 2
	}
	return runDiff(newClient(cfg), cfg.workspace, *projectID, string(yamlText))
}

// runDiff 调用 diff 端点并打印结果。
func runDiff(c *client, workspaceID, projectID int64, yamlText string) int {
	path := fmt.Sprintf("/api/v1/workspaces/%d/project-code/diff", workspaceID)
	data, status, err := c.do(http.MethodPost, path, map[string]any{
		"project_id": projectID,
		"yaml":       yamlText,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "请求失败: %v\n", err)
		return 1
	}
	if status != http.StatusOK {
		fmt.Fprintf(os.Stderr, "diff 失败 (HTTP %d): %s\n", status, extractMessage(data))
		return 1
	}

	var resp struct {
		HasChanges  bool `json:"has_changes"`
		AddCount    int  `json:"add_count"`
		UpdateCount int  `json:"update_count"`
		RemoveCount int  `json:"remove_count"`
		Items       []struct {
			Kind     string `json:"kind"`
			Resource string `json:"resource"`
			Name     string `json:"name"`
			Field    string `json:"field"`
			Old      any    `json:"old"`
			New      any    `json:"new"`
		} `json:"items"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		fmt.Fprintf(os.Stderr, "解析响应失败: %v\n", err)
		return 1
	}

	if !resp.HasChanges {
		fmt.Println("无差异：既有项目与规范一致。")
		return 0
	}

	fmt.Printf("差异（+%d ~%d -%d）：\n", resp.AddCount, resp.UpdateCount, resp.RemoveCount)
	for _, it := range resp.Items {
		switch it.Kind {
		case "add":
			fmt.Printf("  + [%s] %s\n", it.Resource, it.Name)
		case "remove":
			fmt.Printf("  - [%s] %s\n", it.Resource, it.Name)
		case "update":
			if it.Field != "" {
				fmt.Printf("  ~ [%s] %s.%s: %v → %v\n", it.Resource, it.Name, it.Field, it.Old, it.New)
			} else {
				fmt.Printf("  ~ [%s] %s: %v → %v\n", it.Resource, it.Name, it.Old, it.New)
			}
		}
	}
	return 0
}

// extractMessage 从错误响应中提取 message 字段。
func extractMessage(data []byte) string {
	var env struct {
		Error struct {
			Message string `json:"message"`
			Code    string `json:"code"`
		} `json:"error"`
		Message string `json:"message"`
	}
	if err := json.Unmarshal(data, &env); err == nil {
		if env.Error.Message != "" {
			return fmt.Sprintf("[%s] %s", env.Error.Code, env.Error.Message)
		}
		if env.Message != "" {
			return env.Message
		}
	}
	return strings.TrimSpace(string(data))
}
