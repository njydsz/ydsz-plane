// Package godmode — SMTP 发送器与遥测上报。
package godmode

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/smtp"
	"strings"
	"time"
)

// MailMessage 邮件内容。
type MailMessage struct {
	From    string
	To      []string
	Subject string
	Body    string
	HTML    string
}

// SendEmail 使用给定 SMTP 配置发送邮件。
//
// 支持：
//   - 明文 / STARTTLS（UseTLS）；
//   - PLAIN 认证（用户名 + 密码）；
//   - 无认证（本地 relay）。
func SendEmail(cfg SMTPConfig, msg MailMessage) error {
	if cfg.Host == "" {
		return fmt.Errorf("smtp: host is required")
	}
	if len(msg.To) == 0 {
		return fmt.Errorf("smtp: no recipients")
	}

	addr := net.JoinHostPort(cfg.Host, itoa(cfg.Port))
	if cfg.Port == 0 {
		addr = net.JoinHostPort(cfg.Host, "25")
	}

	// 构造邮件内容（MIME）
	raw := buildMessage(msg)

	var conn net.Conn
	var err error

	dialer := &net.Dialer{Timeout: 10 * time.Second}

	if cfg.UseTLS {
		conn, err = tls.DialWithDialer(dialer, "tcp", addr, &tls.Config{
			ServerName: cfg.Host,
			MinVersion: tls.VersionTLS12,
		})
	} else {
		conn, err = dialer.Dial("tcp", addr)
	}
	if err != nil {
		return fmt.Errorf("smtp: dial %s: %w", addr, err)
	}
	defer conn.Close()

	// 读写超时
	_ = conn.SetDeadline(time.Now().Add(30 * time.Second))

	c, err := smtp.NewClient(conn, cfg.Host)
	if err != nil {
		return fmt.Errorf("smtp: new client: %w", err)
	}
	defer c.Close()

	// STARTTLS（非隐式 TLS 时尝试升级）
	if !cfg.UseTLS {
		if ok, _ := c.Extension("STARTTLS"); ok {
			if err := c.StartTLS(&tls.Config{ServerName: cfg.Host, MinVersion: tls.VersionTLS12}); err != nil {
				return fmt.Errorf("smtp: starttls: %w", err)
			}
		}
	}

	// 认证
	if cfg.Username != "" {
		auth := smtp.PlainAuth("", cfg.Username, cfg.Password, cfg.Host)
		if ok, _ := c.Extension("AUTH"); ok {
			if err := c.Auth(auth); err != nil {
				return fmt.Errorf("smtp: auth: %w", err)
			}
		}
	}

	from := msg.From
	if from == "" {
		from = cfg.FromEmail
	}
	if err := c.Mail(from); err != nil {
		return fmt.Errorf("smtp: mail from: %w", err)
	}
	for _, to := range msg.To {
		if err := c.Rcpt(to); err != nil {
			return fmt.Errorf("smtp: rcpt to %s: %w", to, err)
		}
	}

	w, err := c.Data()
	if err != nil {
		return fmt.Errorf("smtp: data: %w", err)
	}
	if _, err := w.Write(raw); err != nil {
		return fmt.Errorf("smtp: write: %w", err)
	}
	if err := w.Close(); err != nil {
		return fmt.Errorf("smtp: close data: %w", err)
	}
	return c.Quit()
}

// buildMessage 构造 MIME 邮件。
func buildMessage(msg MailMessage) []byte {
	var sb strings.Builder

	from := msg.From
	if from == "" {
		from = "no-reply@example.com"
	}

	sb.WriteString("From: " + from + "\r\n")
	sb.WriteString("To: " + strings.Join(msg.To, ", ") + "\r\n")
	sb.WriteString("Subject: " + encodeHeader(msg.Subject) + "\r\n")
	sb.WriteString("Date: " + time.Now().Format(time.RFC1123Z) + "\r\n")
	sb.WriteString("MIME-Version: 1.0\r\n")

	if msg.HTML != "" {
		sb.WriteString("Content-Type: text/html; charset=UTF-8\r\n")
		sb.WriteString("Content-Transfer-Encoding: base64\r\n")
		sb.WriteString("\r\n")
		sb.WriteString(encodeBase64(msg.HTML))
	} else {
		sb.WriteString("Content-Type: text/plain; charset=UTF-8\r\n")
		sb.WriteString("Content-Transfer-Encoding: base64\r\n")
		sb.WriteString("\r\n")
		sb.WriteString(encodeBase64(msg.Body))
	}
	return []byte(sb.String())
}

// encodeHeader 编码非 ASCII 邮件头。
func encodeHeader(s string) string {
	if isASCII(s) {
		return s
	}
	return "=?UTF-8?B?" + encodeBase64(s) + "?="
}

func isASCII(s string) bool {
	for _, r := range s {
		if r > 127 {
			return false
		}
	}
	return true
}

// encodeBase64 标准 base64（按行 76 字符折行）。
func encodeBase64(s string) string {
	const chars = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/"
	src := []byte(s)
	var out strings.Builder
	for i := 0; i < len(src); i += 3 {
		var b [3]byte
		n := copy(b[:], src[i:])
		out.WriteByte(chars[b[0]>>2])
		out.WriteByte(chars[(b[0]&0x03)<<4|b[1]>>4])
		if n > 1 {
			out.WriteByte(chars[(b[1]&0x0f)<<2|b[2]>>6])
		} else {
			out.WriteByte('=')
		}
		if n > 2 {
			out.WriteByte(chars[b[2]&0x3f])
		} else {
			out.WriteByte('=')
		}
	}
	// 按行折行
	raw := out.String()
	var folded strings.Builder
	for i := 0; i < len(raw); i += 76 {
		end := i + 76
		if end > len(raw) {
			end = len(raw)
		}
		folded.WriteString(raw[i:end])
		if end < len(raw) {
			folded.WriteString("\r\n")
		}
	}
	folded.WriteString("\r\n")
	return folded.String()
}
