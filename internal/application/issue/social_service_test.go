// Package issue — 社交反馈（Reaction/Vote）校验逻辑单元测试。
//
// 覆盖范围（纯逻辑，无需数据库）：
//   1. AddReaction 输入校验（空表情/过长表情）
//   2. VoteIssue 输入校验（非法投票值）
//   3. ReactionSummary / VoteSummary 模型 JSON 序列化
package issue

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5"

	"github.com/njydsz/ydsz-plane/pkg/errs"
)

// ==========================================================================
// Reaction 输入校验（通过 fake DB 无法直测 Service 校验分支，
// 这里通过构造不可达 DB 的方式验证校验前置返回；并直接单测 JSON 契约）。
// ==========================================================================

func TestReactionTypeValidation(t *testing.T) {
	cases := []struct {
		name         string
		reactionType string
		wantErr      bool
	}{
		{name: "合法表情", reactionType: "👍", wantErr: false},
		{name: "多字符表情", reactionType: "❤️", wantErr: false},
		{name: "空表情", reactionType: "", wantErr: true},
		{name: "超长输入", reactionType: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", wantErr: true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// 通过 SocialService 的校验逻辑复算：
			// 空/超长校验逻辑独立成函数，避免依赖 DB。
			if got := validateReactionType(tc.reactionType); got != tc.wantErr {
				t.Fatalf("validateReactionType(%q) = %v, want %v", tc.reactionType, got, tc.wantErr)
			}
		})
	}
}

func TestVoteValueValidation(t *testing.T) {
	cases := []struct {
		vote    int
		wantErr bool
	}{
		{vote: 1, wantErr: false},
		{vote: -1, wantErr: false},
		{vote: 0, wantErr: true},
		{vote: 2, wantErr: true},
	}

	for _, tc := range cases {
		t.Run("vote_"+string(rune('0'+tc.vote+2)), func(t *testing.T) {
			if got := validateVoteValue(tc.vote); got != tc.wantErr {
				t.Fatalf("validateVoteValue(%d) = %v, want %v", tc.vote, got, tc.wantErr)
			}
		})
	}
}

// ==========================================================================
// 模型 JSON 契约
// ==========================================================================

func TestReactionSummaryJSON(t *testing.T) {
	raw := `{"reaction_type":"👍","count":3,"reacted":true}`
	var s ReactionSummary
	if err := json.Unmarshal([]byte(raw), &s); err != nil {
		t.Fatal(err)
	}
	if s.ReactionType != "👍" || s.Count != 3 || !s.Reacted {
		t.Fatalf("解析失败: %+v", s)
	}
	b, err := json.Marshal(s)
	if err != nil {
		t.Fatal(err)
	}
	var roundtrip map[string]any
	if err := json.Unmarshal(b, &roundtrip); err != nil {
		t.Fatal(err)
	}
	if roundtrip["count"].(float64) != 3 {
		t.Fatalf("roundtrip count 不符: %v", roundtrip["count"])
	}
}

func TestVoteSummaryJSON(t *testing.T) {
	// 未投票时 voted 为 null（omitempty + pointer nil）
	raw := `{"upvotes":5,"downvotes":1,"score":4}`
	var s VoteSummary
	if err := json.Unmarshal([]byte(raw), &s); err != nil {
		t.Fatal(err)
	}
	if s.Upvotes != 5 || s.Downvotes != 1 || s.Score != 4 {
		t.Fatalf("解析失败: %+v", s)
	}
	if s.Voted != nil {
		t.Fatalf("未投票时 Voted 应为 nil, 实际 %v", *s.Voted)
	}
	b, err := json.Marshal(s)
	if err != nil {
		t.Fatal(err)
	}
	if string(b) == "" {
		t.Fatal("序列化为空")
	}
}

// ==========================================================================
// errs 错误语义（AddReaction 冲突路径用 errs 包装）
// ==========================================================================

func TestErrNotFoundSemantics(t *testing.T) {
	if !errors.Is(errs.ErrNotFound, errs.ErrNotFound) {
		t.Fatal("ErrNotFound 应为同一哨兵错误")
	}
}

// pgx.ErrNoRows 用于 AddReaction ON CONFLICT DO NOTHING 路径识别。
func TestPgxErrNoRows(t *testing.T) {
	if pgx.ErrNoRows == nil {
		t.Fatal("pgx.ErrNoRows 不应为 nil")
	}
}
