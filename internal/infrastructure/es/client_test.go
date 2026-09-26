// Package es — ES 客户端配置与纯函数测试。
//
// 仅测试不涉及 HTTP 请求的纯逻辑：
//   - DefaultConfig 默认值
//   - NewClient 自动填充默认值
//   - IndexName / 各 Query 构造器
//
// 所有测试不依赖外部 ES 实例。
package es

import (
	"reflect"
	"testing"
	"time"
)

// TestDefaultConfig_Values 验证 DefaultConfig 返回推荐默认值。
func TestDefaultConfig_Values(t *testing.T) {
	cfg := DefaultConfig()

	t.Run("URLs default to localhost", func(t *testing.T) {
		want := []string{"http://127.0.0.1:9200"}
		if !reflect.DeepEqual(cfg.URLs, want) {
			t.Fatalf("URLs: got %v, want %v", cfg.URLs, want)
		}
	})

	t.Run("IndexPrefix default", func(t *testing.T) {
		if cfg.IndexPrefix != "ydsz" {
			t.Fatalf("IndexPrefix: got %q, want %q", cfg.IndexPrefix, "ydsz")
		}
	})

	t.Run("MaxIdleConns default", func(t *testing.T) {
		if cfg.MaxIdleConns != 50 {
			t.Fatalf("MaxIdleConns: got %d, want 50", cfg.MaxIdleConns)
		}
	})

	t.Run("MaxConnsPerHost default", func(t *testing.T) {
		if cfg.MaxConnsPerHost != 10 {
			t.Fatalf("MaxConnsPerHost: got %d, want 10", cfg.MaxConnsPerHost)
		}
	})

	t.Run("Timeout default", func(t *testing.T) {
		if cfg.Timeout != 5*time.Second {
			t.Fatalf("Timeout: got %v, want %v", cfg.Timeout, 5*time.Second)
		}
	})

	t.Run("RetryOnStatus default", func(t *testing.T) {
		want := []int{502, 503, 504}
		if !reflect.DeepEqual(cfg.RetryOnStatus, want) {
			t.Fatalf("RetryOnStatus: got %v, want %v", cfg.RetryOnStatus, want)
		}
	})

	t.Run("MaxRetries default", func(t *testing.T) {
		if cfg.MaxRetries != 2 {
			t.Fatalf("MaxRetries: got %d, want 2", cfg.MaxRetries)
		}
	})
}

// TestNewClient_FillsDefaults 验证 NewClient 对零值字段自动填充默认值。
// 不发起任何网络请求。
func TestNewClient_FillsDefaults(t *testing.T) {
	tests := []struct {
		name   string
		input  Config
		errMsg string
	}{
		{
			name:   "empty URLs returns error",
			input:  Config{},
			errMsg: "at least one URL is required",
		},
		{
			name:   "URLs provided with minimal config succeeds",
			input:  Config{URLs: []string{"http://localhost:9200"}},
			errMsg: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, err := NewClient(tt.input)
			if tt.errMsg != "" {
				if err == nil {
					t.Fatalf("expected error containing %q, got nil", tt.errMsg)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			defer c.Close()

			// NewClient 默认健康和批处理
			if !c.IsHealthy() {
				t.Fatal("expected client to be healthy on init")
			}
			if c.ReindexBatchSize != 1000 {
				t.Fatalf("ReindexBatchSize: got %d, want 1000", c.ReindexBatchSize)
			}
			if c.ReindexWorkers < 1 {
				t.Fatalf("ReindexWorkers: got %d, want >= 1", c.ReindexWorkers)
			}
		})
	}
}

// TestIndexName 验证索引名前缀拼接逻辑。
func TestIndexName(t *testing.T) {
	tests := []struct {
		prefix string
		name   string
		want   string
	}{
		{"ydsz", "issues", "ydsz_issues"},
		{"ydsz", "sprints", "ydsz_sprints"},
		{"test", "foo", "test_foo"},
		{"", "bar", "_bar"},
	}

	for _, tt := range tests {
		t.Run(tt.prefix+"_"+tt.name, func(t *testing.T) {
			c := &Client{cfg: Config{IndexPrefix: tt.prefix}}
			got := c.IndexName(tt.name)
			if got != tt.want {
				t.Fatalf("IndexName(%q): got %q, want %q", tt.name, got, tt.want)
			}
		})
	}
}

// TestMatchQuery 验证 match 查询结构体输出。
func TestMatchQuery(t *testing.T) {
	q := MatchQuery("title", "hello world")

	matchMap, ok := q["match"].(map[string]any)
	if !ok {
		t.Fatal("expected 'match' key in result")
	}
	if matchMap["title"] != "hello world" {
		t.Fatalf("expected title=hello world, got %v", matchMap["title"])
	}
}

// TestRangeQuery 验证范围查询构造器边界条件。
func TestRangeQuery(t *testing.T) {
	tests := []struct {
		name     string
		field    string
		gte      any
		lte      any
		gt       any
		lt       any
		wantKeys []string // expected keys in rangeMap
	}{
		{
			name:     "only gte",
			field:    "price",
			gte:      10,
			wantKeys: []string{"gte"},
		},
		{
			name:     "gte and lte",
			field:    "price",
			gte:      10,
			lte:      100,
			wantKeys: []string{"gte", "lte"},
		},
		{
			name:     "gt and lt",
			field:    "score",
			gt:       0,
			lt:       100,
			wantKeys: []string{"gt", "lt"},
		},
		{
			name:     "all bounds",
			field:    "amount",
			gte:      1,
			lte:      99,
			gt:       2,
			lt:       98,
			wantKeys: []string{"gte", "lte", "gt", "lt"},
		},
		{
			name:     "no bounds",
			field:    "amount",
			wantKeys: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q := RangeQuery(tt.field, tt.gte, tt.lte, tt.gt, tt.lt)
			rangeMap := q["range"].(map[string]any)[tt.field].(map[string]any)

			if len(rangeMap) != len(tt.wantKeys) {
				t.Fatalf("expected %d keys in range map, got %d: %v",
					len(tt.wantKeys), len(rangeMap), rangeMap)
			}
			for _, key := range tt.wantKeys {
				if _, ok := rangeMap[key]; !ok {
					t.Fatalf("expected key %q in range map, not found", key)
				}
			}
		})
	}
}

// TestBoolQuery 验证 bool 查询构造器各子句。
func TestBoolQuery(t *testing.T) {
	t.Run("empty inputs produces empty bool", func(t *testing.T) {
		q := BoolQuery(nil, nil, nil, nil)
		bq := q["bool"].(map[string]any)
		if len(bq) != 0 {
			t.Fatalf("expected empty bool map, got %v", bq)
		}
	})

	t.Run("must and filter only", func(t *testing.T) {
		must := []any{"a", "b"}
		filter := []any{"c"}
		q := BoolQuery(must, filter, nil, nil)
		bq := q["bool"].(map[string]any)

		if _, ok := bq["must"]; !ok {
			t.Fatal("expected 'must' key")
		}
		if _, ok := bq["filter"]; !ok {
			t.Fatal("expected 'filter' key")
		}
		if _, ok := bq["should"]; ok {
			t.Fatal("unexpected 'should' key")
		}
	})

	t.Run("all clauses", func(t *testing.T) {
		q := BoolQuery(
			[]any{"must1"},
			[]any{"filter1"},
			[]any{"should1"},
			[]any{"must_not1"},
		)
		bq := q["bool"].(map[string]any)

		for _, key := range []string{"must", "filter", "should", "must_not"} {
			if _, ok := bq[key]; !ok {
				t.Fatalf("expected %q key in bool map", key)
			}
		}
	})
}

// TestHighlightConfig 验证高亮配置默认 tag 与 fragment 参数。
func TestHighlightConfig(t *testing.T) {
	t.Run("empty fields", func(t *testing.T) {
		hc := HighlightConfig()
		fields := hc["fields"].(map[string]any)
		if len(fields) != 0 {
			t.Fatalf("expected 0 fields, got %d", len(fields))
		}
	})

	t.Run("single field with defaults", func(t *testing.T) {
		hc := HighlightConfig("title")
		fields := hc["fields"].(map[string]any)
		titleCfg, ok := fields["title"].(map[string]any)
		if !ok {
			t.Fatal("expected 'title' field config")
		}
		if titleCfg["fragment_size"] != 150 {
			t.Fatalf("fragment_size: got %v, want 150", titleCfg["fragment_size"])
		}
		if titleCfg["number_of_fragments"] != 3 {
			t.Fatalf("number_of_fragments: got %v, want 3", titleCfg["number_of_fragments"])
		}

		preTags, ok := titleCfg["pre_tags"].([]string)
		if !ok || len(preTags) != 1 || preTags[0] != "<mark>" {
			t.Fatalf("pre_tags: got %v, want [<mark>]", titleCfg["pre_tags"])
		}

		postTags, ok := titleCfg["post_tags"].([]string)
		if !ok || len(postTags) != 1 || postTags[0] != "</mark>" {
			t.Fatalf("post_tags: got %v, want [</mark>]", titleCfg["post_tags"])
		}
	})

	t.Run("multiple fields", func(t *testing.T) {
		hc := HighlightConfig("title", "content", "description")
		fields := hc["fields"].(map[string]any)
		if len(fields) != 3 {
			t.Fatalf("expected 3 fields, got %d", len(fields))
		}
	})
}

// TestIssueMapping 验证 issues 索引 mapping 结构完整性。
func TestIssueMapping(t *testing.T) {
	m := IssueMapping()

	// settings
	settings, ok := m["settings"].(map[string]any)
	if !ok {
		t.Fatal("expected 'settings' key")
	}
	indexSettings, ok := settings["index"].(map[string]any)
	if !ok {
		t.Fatal("expected 'settings.index' key")
	}
	if indexSettings["number_of_shards"] != 3 {
		t.Fatalf("number_of_shards: got %v, want 3", indexSettings["number_of_shards"])
	}
	if indexSettings["number_of_replicas"] != 1 {
		t.Fatalf("number_of_replicas: got %v, want 1", indexSettings["number_of_replicas"])
	}

	// mappings.properties
	props := m["mappings"].(map[string]any)["properties"].(map[string]any)
	required := []string{"workspace_id", "doc_type", "title", "content", "priority", "assignee_ids", "created_at"}
	for _, field := range required {
		if _, ok := props[field]; !ok {
			t.Fatalf("missing required field %q in issue mapping", field)
		}
	}
}

// TestSprintMapping 验证 sprints 索引 mapping 结构。
func TestSprintMapping(t *testing.T) {
	m := SprintMapping()
	props := m["mappings"].(map[string]any)["properties"].(map[string]any)

	required := []string{"workspace_id", "status", "start_date", "end_date", "created_at"}
	for _, field := range required {
		if _, ok := props[field]; !ok {
			t.Fatalf("missing required field %q in sprint mapping", field)
		}
	}

	indexSettings := m["settings"].(map[string]any)["index"].(map[string]any)
	if indexSettings["number_of_shards"] != 1 {
		t.Fatalf("number_of_shards: got %v, want 1", indexSettings["number_of_shards"])
	}
}
