// Package mq — RabbitMQ 常量、纯函数与校验逻辑测试。
//
// 不涉及 AMQP 网络连接，仅测试：
//   - Exchange 与队列常量值
//   - RoutingKey / queueName 字符串拼接
//   - retryBackoff 计算
//   - redactURL 凭证隐藏逻辑
//   - Task 零值检测
package mq

import (
	"strings"
	"testing"
	"time"
)

// TestExchangeConstants 验证 exchange 常量符合架构约定。
func TestExchangeConstants(t *testing.T) {
	tests := []struct {
		name  string
		got   string
		want  string
	}{
		{"EventExchange", EventExchange, "plane.events"},
		{"TaskExchange", TaskExchange, "plane.tasks"},
		{"DeadLetterExchange", DeadLetterExchange, "plane.dlx"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.want {
				t.Fatalf("%s = %q, want %q", tt.name, tt.got, tt.want)
			}
		})
	}
}

// TestDefaultConstants 验证其他包级常量默认值。
func TestDefaultConstants(t *testing.T) {
	if DefaultConsumerTag != "plane-consumer" {
		t.Fatalf("DefaultConsumerTag = %q, want %q", DefaultConsumerTag, "plane-consumer")
	}
	if MaxReconnectAttempts != 10 {
		t.Fatalf("MaxReconnectAttempts = %d, want 10", MaxReconnectAttempts)
	}
}

// TestRoutingKey 验证路由键拼接格式。
func TestRoutingKey(t *testing.T) {
	tests := []struct {
		aggregate string
		eventType string
		want      string
	}{
		{"issue", "created", "plane.events.issue.created"},
		{"workspace", "member_added", "plane.events.workspace.member_added"},
		{"sprint", "started", "plane.events.sprint.started"},
		{"", "orphan", "plane.events..orphan"}, // 边界：空 aggregate 仍可拼接
	}

	for _, tt := range tests {
		t.Run(tt.aggregate+"."+tt.eventType, func(t *testing.T) {
			got := RoutingKey(tt.aggregate, tt.eventType)
			if got != tt.want {
				t.Fatalf("RoutingKey(%q, %q) = %q, want %q",
					tt.aggregate, tt.eventType, got, tt.want)
			}
		})
	}
}

// TestTask_RoutingKey 验证 Task 实例路由键。
func TestTask_RoutingKey(t *testing.T) {
	tests := []struct {
		taskType string
		want     string
	}{
		{"send_email", "task.send_email"},
		{"generate_report", "task.generate_report"},
		{"", "task."}, // 边界：空类型
	}

	for _, tt := range tests {
		t.Run(tt.taskType, func(t *testing.T) {
			task := Task{Type: tt.taskType}
			got := task.RoutingKey()
			if got != tt.want {
				t.Fatalf("Task{%q}.RoutingKey() = %q, want %q",
					tt.taskType, got, tt.want)
			}
		})
	}
}

// TestTask_IsZero 验证 Task 零值检测。
func TestTask_IsZero(t *testing.T) {
	tests := []struct {
		name string
		task Task
		want bool
	}{
		{"empty Task", Task{}, true},
		{"Type set", Task{Type: "email"}, false},
		{"only ID set (no Type)", Task{ID: "foo"}, true},
		{"only Priority", Task{Priority: 5}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.task.IsZero(); got != tt.want {
				t.Fatalf("IsZero() = %v, want %v (task=%+v)", got, tt.want, tt.task)
			}
		})
	}
}

// TestQueueName 验证 queueName 辅助函数。
func TestQueueName(t *testing.T) {
	tests := []struct {
		taskType string
		want     string
	}{
		{"email", "task.email"},
		{"digest", "task.digest"},
		{"", "task."},
	}

	for _, tt := range tests {
		t.Run(tt.taskType, func(t *testing.T) {
			got := queueName(tt.taskType)
			if got != tt.want {
				t.Fatalf("queueName(%q) = %q, want %q", tt.taskType, got, tt.want)
			}
		})
	}
}

// TestRetryBackoff 验证指数退避计算与封顶逻辑。
// 公式: d = time.Duration(1<<uint(attempt)) * time.Second; if d > 30s → d = 30s
func TestRetryBackoff(t *testing.T) {
	tests := []struct {
		attempt int
		want    time.Duration
	}{
		{0, 1 * time.Second},   // 1<<0 = 1
		{1, 2 * time.Second},   // 1<<1 = 2
		{2, 4 * time.Second},   // 1<<2 = 4
		{3, 8 * time.Second},   // 1<<3 = 8
		{4, 16 * time.Second},  // 1<<4 = 16
		{5, 30 * time.Second},  // 1<<5 = 32 > 30s → cap at 30s
		{6, 30 * time.Second},  // 1<<6 = 64 > 30s → cap
		{10, 30 * time.Second}, // 远大于 30s → cap
		{30, 30 * time.Second}, // int overflow territory but still caps
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			got := retryBackoff(tt.attempt)
			if got != tt.want {
				t.Fatalf("retryBackoff(%d) = %v, want %v", tt.attempt, got, tt.want)
			}
		})
	}
}

// TestRedactedURL 验证 AMQP URL 凭证隐藏逻辑。
func TestRedactedURL(t *testing.T) {
	// 验证不包含原始密码（密码被隐藏或不在输出中）
	t.Run("password not visible in output", func(t *testing.T) {
		got := RedactedURL("amqp://admin:secretpassword@rabbitmq:5672/")
		// amqp091 ParseURI 不解析密码字段，输出不含原始密码
		if strings.Contains(got, "secretpassword") {
			t.Fatalf("password leaked: %q", got)
		}
		// 输出应包含用户和主机信息
		if !strings.Contains(got, "admin") {
			t.Fatalf("expected username in output, got %q", got)
		}
		if !strings.Contains(got, "rabbitmq") {
			t.Fatalf("expected host in output, got %q", got)
		}
	})

	// 无密码 URL 不含 ***
	t.Run("no password → no ***", func(t *testing.T) {
		got := RedactedURL("amqp://guest@rabbitmq:5672/vhost")
		if strings.Contains(got, "***") {
			t.Fatalf("unexpected *** in output without password: %q", got)
		}
		if !strings.Contains(got, "guest") {
			t.Fatalf("username missing: %q", got)
		}
	})

	// URL with credentials: check user@host:port preserved
	t.Run("user@host:port preserved", func(t *testing.T) {
		got := RedactedURL("amqp://myuser:mypass@myhost:5671/somevhost")
		if !strings.Contains(got, "myuser") {
			t.Fatalf("expected username in output, got %q", got)
		}
		if !strings.Contains(got, "myhost") {
			t.Fatalf("expected host in output, got %q", got)
		}
		if !strings.Contains(got, "5671") {
			t.Fatalf("expected port in output, got %q", got)
		}
		// 确保不含原始密码
		if strings.Contains(got, "mypass") {
			t.Fatalf("password leaked: %q", got)
		}
	})

	// 不可解析 URL
	t.Run("unparseable URL", func(t *testing.T) {
		got := RedactedURL("://:::invalid")
		if got != "(unparseable)" {
			t.Fatalf("expected '(unparseable)', got %q", got)
		}
	})
}

// TestRedactedURL_DoesNotLeakPassword 模糊测试：确保密码不出现在输出中。
func TestRedactedURL_DoesNotLeakPassword(t *testing.T) {
	password := "SuperSecret123!"
	input := "amqp://user:" + password + "@host:5672/"
	got := RedactedURL(input)

	if strings.Contains(got, password) {
		t.Fatalf("RedactedURL leaked password: %q", got)
	}
}
