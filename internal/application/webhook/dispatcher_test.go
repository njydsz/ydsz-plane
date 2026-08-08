// Package webhook — Webhook 签名与纯函数单元测试。
package webhook

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"testing"
)

// TestComputeSignature 校验 HMAC-SHA256 签名计算正确性。
func TestComputeSignature(t *testing.T) {
	secret := "whsec_test_secret_12345"
	timestamp := int64(1722988800)
	payload := []byte(`{"event":"issue.created","id":42}`)

	// 手动计算预期：HMAC-SHA256(secret, "timestamp.body")
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(fmt.Sprintf("%d.", timestamp)))
	mac.Write(payload)
	expected := hex.EncodeToString(mac.Sum(nil))

	got := ComputeSignature(secret, timestamp, payload)
	if got != expected {
		t.Errorf("ComputeSignature mismatch: got %s want %s", got, expected)
	}
}

// TestVerifySignature 校验签名验证的时序安全比较。
func TestVerifySignature(t *testing.T) {
	secret := "whsec_secret"
	timestamp := int64(1722988800)
	payload := []byte(`{"data":"test"}`)
	sig := ComputeSignature(secret, timestamp, payload)

	cases := []struct {
		name   string
		sig    string
		wantOK bool
	}{
		{"正确签名", sig, true},
		{"错误签名", "invalid_sig", false},
		{"空签名", "", false},
		{"大小写变化", "ABC" + sig[3:], false},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if ok := VerifySignature(secret, timestamp, payload, c.sig); ok != c.wantOK {
				t.Errorf("VerifySignature(%s): got %v want %v", c.name, ok, c.wantOK)
			}
		})
	}
}

// TestSignatureHeader 校验 X-Ydsz-Signature-256 header 格式。
func TestSignatureHeader(t *testing.T) {
	secret := "whsec_secret"
	timestamp := int64(1722988800)
	payload := []byte(`{}`)

	header := SignatureHeader(secret, timestamp, payload)
	expected := "sha256=" + ComputeSignature(secret, timestamp, payload)
	if header != expected {
		t.Errorf("SignatureHeader: got %s want %s", header, expected)
	}
	if len(header) < 7 || header[:7] != "sha256=" {
		t.Errorf("SignatureHeader should start with 'sha256=': %s", header)
	}
}

// TestAllEvents 校验事件全集非空且无重复。
func TestAllEvents(t *testing.T) {
	events := AllEvents()
	if len(events) == 0 {
		t.Error("AllEvents() returned empty slice")
	}
	seen := make(map[string]bool)
	for _, e := range events {
		if seen[e] {
			t.Errorf("duplicate event: %s", e)
		}
		seen[e] = true
	}
	// 验证关键事件存在
	required := []string{
		EventIssueCreated, EventIssueCommentUpdated, EventIssueAttachmentAdded,
		EventProjectCreated, EventProjectMemberAdded,
		EventStateCreated, EventModuleCreated, EventLabelCreated,
		EventSprintIssueAdded,
		EventUserInvited, EventUserRoleUpdated,
		EventVersionDeleted,
		EventIntakeMerged,
		EventSprintCompleted, EventVersionReleased,
	}
	for _, r := range required {
		found := false
		for _, e := range events {
			if e == r {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("required event %s not found in AllEvents()", r)
		}
	}
}
