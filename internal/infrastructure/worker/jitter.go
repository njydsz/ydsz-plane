// Package worker — Worker 定时打散策略（Jitter / Stagger）。
//
// 解决多 Worker 实例或同进程内多个定时器在同一时刻并发触发造成的压力峰值问题。
// 例如：每日快照 cron 全部在 00:05:00 同时起线程扫描全表，数据库连接瞬时耗尽。
//
// 打散策略（确定性伪随机）：
//   - 基于规则 ID 或日期 key 计算 0~N 秒的确定性偏移，保证同一 worker 同一天
//     偏移固定（不因重启而重复触发），不同 worker/规则之间错开。
//   - 支持自定义最大偏移量（秒），避免超过执行窗口。
package worker

import (
	"hash/fnv"
	"time"
)

// DeterministicJitter 基于 key 计算一个确定性偏移秒数。
// 同一 key 在任何时刻调用返回值一致（适合"每天同一偏移"语义）。
//
//	maxSeconds: 最大偏移秒数（不含），实际返回值 ∈ [0, maxSeconds)。
func DeterministicJitter(key string, maxSeconds int) time.Duration {
	if maxSeconds <= 0 {
		return 0
	}
	h := fnv.New32a()
	_, _ = h.Write([]byte(key))
	offset := int(h.Sum32()) % maxSeconds
	return time.Duration(offset) * time.Second
}

// IDJitter 基于 int64 ID 计算确定性偏移（适用于 rule_id / project_id 等）。
func IDJitter(id int64, maxSeconds int) time.Duration {
	return DeterministicJitter(int64ToKey(id), maxSeconds)
}

// DayJitter 基于日期 key（格式 "2006-01-02"）计算当日偏移。
// 不同日期偏移不同，同日内固定（避免 worker 重启后偏移漂移）。
func DayJitter(dateKey string, maxSeconds int) time.Duration {
	return DeterministicJitter("day:"+dateKey, maxSeconds)
}

// MinuteJitter 基于 ID + 分钟 key 计算偏移。
// 适用于每分钟都触发但需要错开执行的场景（如自动化 scheduled 规则）。
func MinuteJitter(id int64, minuteKey string, maxSeconds int) time.Duration {
	return DeterministicJitter(int64ToKey(id)+":"+minuteKey, maxSeconds)
}

func int64ToKey(id int64) string {
	// 直接用 binary 编码避免 strconv 分配（微优化）
	b := []byte{byte(id >> 56), byte(id >> 48), byte(id >> 40), byte(id >> 32),
		byte(id >> 24), byte(id >> 16), byte(id >> 8), byte(id)}
	return string(b)
}
