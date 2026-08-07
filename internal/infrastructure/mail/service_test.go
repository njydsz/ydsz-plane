package mail

import (
	"strings"
	"testing"
)

func TestBuildMimeMessageMultipartAlternative(t *testing.T) {
	msg := Message{
		To:      "alice@example.com",
		Subject: "Reset",
		Text:    "hello plain",
		HTML:    "<p>hello html</p>",
	}
	body := string(buildMimeMessage("Ydsz <no-reydsz.dev>", msg))
	if !strings.Contains(body, "Content-Type: multipart/alternative") {
		t.Errorf("expected multipart/alternative, got:\n%s", body)
	}
	if !strings.Contains(body, "Content-Type: text/plain") {
		t.Errorf("expected text/plain part")
	}
	if !strings.Contains(body, "Content-Type: text/html") {
		t.Errorf("expected text/html part")
	}
	if !strings.Contains(body, "To: alice@example.com") {
		t.Errorf("To header missing")
	}
	if !strings.Contains(body, "hello plain") || !strings.Contains(body, "<p>hello html</p>") {
		t.Errorf("body content missing")
	}
}

func TestNoopEmailSend(t *testing.T) {
	n := NewNoopService(0)
	if err := n.Send(Message{To: "a@b.c", Subject: "hi"}); err != nil {
		t.Errorf("noop send: %v", err)
	}
}

func TestNoopWithChannel(t *testing.T) {
	n := NewNoopService(1)
	_ = n.Send(Message{To: "a@b.c", Subject: "hi", HTML: "<b>ok</b>"})
	got := <-n.Sent
	if got.Subject != "hi" {
		t.Errorf("noop channel: subject = %q", got.Subject)
	}
}

func TestSmtpWithoutHostErrors(t *testing.T) {
	s := NewSmtpService(SmtpConfig{Host: ""})
	err := s.Send(Message{To: "a@b.c", Subject: "x"})
	if err == nil {
		t.Errorf("expected error when host empty")
	}
}

func TestRenderResetPassword(t *testing.T) {
	m := RenderResetPassword(ResetPasswordData{
		RecipientName: "Alice",
		ResetURL:      "http://x/reset?token=abc",
		TTLMin:        15,
	})
	if !strings.Contains(m.Text, "15") {
		t.Errorf("expected ttl hint in text, got: %q", m.Text)
	}
	if !strings.Contains(m.HTML, "http://x/reset?token=abc") {
		t.Errorf("expected URL in html")
	}
}
