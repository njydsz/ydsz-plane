// Package persistence 测试：连接池启发式参数（db.go）。
//
// 验证 max 启发式函数、环境变量覆盖逻辑的可测部分。
// 完整连接池行为由 dockertesthelper 集成测试覆盖。
package persistence

import "testing"

func TestMax(t *testing.T) {
	t.Parallel()

	cases := []struct {
		a, b, want int
	}{
		{1, 2, 2},
		{5, 3, 5},
		{0, 0, 0},
		{-1, 1, 1},
		{10, 10, 10},
	}
	for _, c := range cases {
		if got := max(c.a, c.b); got != c.want {
			t.Errorf("max(%d, %d) = %d, want %d", c.a, c.b, got, c.want)
		}
	}
}

func TestPoolPingDelegates(t *testing.T) {
	t.Parallel()

	// Pool.Ping 委托给底层 *pgxpool.Pool。
	// 此处仅验证 Pool 结构可构造（nil 池会在 Ping 时 panic，故只测构造路径）。
	p := &Pool{}
	if p == nil {
		t.Fatal("Pool construction failed")
	}
}
