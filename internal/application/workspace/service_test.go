// Package workspace — 工作空间域单元测试。
//
// 覆盖范围：
//  1. sha256Hex 哈希函数
//  2. generateSecureToken 生成格式
//  3. ProjectModuleToggles Value/Scan 序列化
//  4. ProjectModuleAllEnabled 默认值
//  5. driver.Valuer 接口实现校验
package workspace

import (
	"database/sql/driver"
	"testing"
)

// ==================================================================
// sha256Hex
// ==================================================================

func TestSha256Hex_KnownValue(t *testing.T) {
	// echo -n "hello" | sha256sum
	// = 2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824
	got := sha256Hex("hello")
	want := "2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824"
	if got != want {
		t.Errorf("sha256Hex(\"hello\") = %q, want %q", got, want)
	}
}

func TestSha256Hex_Empty(t *testing.T) {
	// echo -n "" | sha256sum
	// = e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855
	got := sha256Hex("")
	want := "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"
	if got != want {
		t.Errorf("sha256Hex(\"\") = %q, want %q", got, want)
	}
}

func TestSha256Hex_Deterministic(t *testing.T) {
	input := "test-invitation-token-123"
	first := sha256Hex(input)
	second := sha256Hex(input)
	if first != second {
		t.Errorf("sha256Hex not deterministic: %q != %q", first, second)
	}
}

func TestSha256Hex_Unique(t *testing.T) {
	a := sha256Hex("token-a")
	b := sha256Hex("token-b")
	if a == b {
		t.Errorf("sha256Hex(\"token-a\") == sha256Hex(\"token-b\"): %q", a)
	}
}

func TestSha256Hex_Length(t *testing.T) {
	got := sha256Hex("any-input")
	if len(got) != 64 {
		t.Errorf("sha256Hex length = %d, want 64", len(got))
	}
}

// ==================================================================
// generateSecureToken
// ==================================================================

func TestGenerateSecureToken_Length(t *testing.T) {
	token := generateSecureToken()
	if len(token) != 64 {
		t.Errorf("generateSecureToken() length = %d, want 64", len(token))
	}
}

func TestGenerateSecureToken_Unique(t *testing.T) {
	seen := make(map[string]bool)
	for i := 0; i < 100; i++ {
		token := generateSecureToken()
		if seen[token] {
			t.Fatalf("generateSecureToken() collision at iteration %d", i)
		}
		seen[token] = true
	}
}

func TestGenerateSecureToken_HexChars(t *testing.T) {
	token := generateSecureToken()
	for _, c := range token {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f')) {
			t.Errorf("generateSecureToken() non-hex char: %q in %q", c, token)
			break
		}
	}
}

// ==================================================================
// ProjectModuleToggles
// ==================================================================

func TestProjectModuleToggles_Value(t *testing.T) {
	m := ProjectModuleToggles{Sprint: true, Version: false, Estimate: true}
	v, err := m.Value()
	if err != nil {
		t.Fatalf("Value: %v", err)
	}
	s, ok := v.([]byte)
	if !ok {
		t.Fatalf("Value returned %T, want []byte", v)
	}
	if len(s) == 0 {
		t.Errorf("Value returned empty bytes")
	}
}

func TestProjectModuleToggles_Value_ZeroDefault(t *testing.T) {
	m := ProjectModuleToggles{}
	v, err := m.Value()
	if err != nil {
		t.Fatalf("Value: %v", err)
	}
	s := v.([]byte)
	if len(s) == 0 {
		t.Errorf("Value returned empty bytes for zero toggles")
	}
}

func TestProjectModuleToggles_Scan_FromBytes(t *testing.T) {
	src := []byte(`{"sprint":true,"version":true,"estimate":false}`)
	var m ProjectModuleToggles
	if err := m.Scan(src); err != nil {
		t.Fatalf("Scan: %v", err)
	}
	if !m.Sprint {
		t.Errorf("Sprint: got false, want true")
	}
	if !m.Version {
		t.Errorf("Version: got false, want true")
	}
	if m.Estimate {
		t.Errorf("Estimate: got true, want false")
	}
}

func TestProjectModuleToggles_Scan_FromString(t *testing.T) {
	var m ProjectModuleToggles
	if err := m.Scan(`{"sprint":false,"version":true,"estimate":true}`); err != nil {
		t.Fatalf("Scan: %v", err)
	}
	if m.Sprint {
		t.Errorf("Sprint: got true, want false")
	}
	if !m.Version {
		t.Errorf("Version: got false, want true")
	}
}

func TestProjectModuleToggles_Scan_Nil(t *testing.T) {
	var m ProjectModuleToggles
	if err := m.Scan(nil); err != nil {
		t.Errorf("Scan(nil) returned error: %v", err)
	}
	if m.Sprint || m.Version || m.Estimate {
		t.Errorf("Scan(nil) should leave struct unchanged, got %+v", m)
	}
}

func TestProjectModuleToggles_Scan_EmptyBytes(t *testing.T) {
	var m ProjectModuleToggles
	if err := m.Scan([]byte{}); err != nil {
		t.Errorf("Scan([]byte{}) returned error: %v", err)
	}
}

func TestProjectModuleToggles_UnsupportedType(t *testing.T) {
	var m ProjectModuleToggles
	err := m.Scan(12345)
	if err == nil {
		t.Error("Scan(int) should return error for unsupported type")
	}
}

func TestProjectModuleToggles_Roundtrip(t *testing.T) {
	original := ProjectModuleToggles{Sprint: true, Version: false, Estimate: true}
	v, err := original.Value()
	if err != nil {
		t.Fatalf("Value: %v", err)
	}
	var decoded ProjectModuleToggles
	if err := decoded.Scan(v); err != nil {
		t.Fatalf("Scan: %v", err)
	}
	if decoded != original {
		t.Errorf("roundtrip mismatch: got %+v, want %+v", decoded, original)
	}
}

// ==================================================================
// ProjectModuleAllEnabled
// ==================================================================

func TestProjectModuleAllEnabled_Default(t *testing.T) {
	m := ProjectModuleAllEnabled()
	if !m.Sprint {
		t.Errorf("ProjectModuleAllEnabled().Sprint = false, want true")
	}
	if !m.Version {
		t.Errorf("ProjectModuleAllEnabled().Version = false, want true")
	}
	if !m.Estimate {
		t.Errorf("ProjectModuleAllEnabled().Estimate = false, want true")
	}
}

// ==================================================================
// driver.Valuer interface check
// ==================================================================

func TestProjectModuleToggles_ImplementsValuer(t *testing.T) {
	var m ProjectModuleToggles
	var _ driver.Valuer = m
	var _ driver.Value = nil
}
