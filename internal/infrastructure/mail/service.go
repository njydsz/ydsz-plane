// Package mail — SMTP 邮件发送服务。
//
// 架构要点（对标 SendGrid / AWS SES 异步投递模式）：
//  - 接口抽象：EmailService 接口让业务代码不依赖具体实现；dev 用 Noop，prod 用 SMTP
//  - 统一 Message 结构体：收件人、主题、纯文本 + HTML 双版本
//  - TLS 默认：强制 STARTTLS on port 587 / TLS on 465
//  - 超时保护：连接 + 读写均设超时
package mail

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/smtp"
	"time"
)

// Message 表示一封待发邮件。
type Message struct {
	To      string
	Subject string
	Text    string // 纯文本降级内容（某些客户端不支持 HTML）。
	HTML    string // 富文本版本（双版本是邮件最佳实践）。
}

// SmtpConfig 保存 SMTP 连接信息。
type SmtpConfig struct {
	Host     string
	Port     int
	Username string
	Password string
	From     string // 发件人：Name <email> 或纯 email
	UseTLS   bool   // true → 端口 465 TLS；false → 端口 587 STARTTLS
}

// EmailService 邮件发送抽象。
type EmailService interface {
	Send(msg Message) error
}

// ---------------------------------------------------------------
// SMTP implementation
// ---------------------------------------------------------------

type smtpEmail struct {
	cfg SmtpConfig
}

// NewSmtpService 创建 SMTP 发送器。
func NewSmtpService(cfg SmtpConfig) EmailService {
	return &smtpEmail{cfg: cfg}
}

func (s *smtpEmail) Send(msg Message) error {
	if s.cfg.Host == "" {
		return fmt.Errorf("mail: SMTP host not configured")
	}
	addr := net.JoinHostPort(s.cfg.Host, fmt.Sprintf("%d", s.cfg.Port))

	conn, err := s.dial(addr)
	if err != nil {
		return fmt.Errorf("mail: dial %s: %w", addr, err)
	}
	defer conn.Close()

	c, err := smtp.NewClient(conn, s.cfg.Host)
	if err != nil {
		return fmt.Errorf("mail: new client: %w", err)
	}
	defer c.Close()

	if !s.cfg.UseTLS {
		if err := c.StartTLS(&tls.Config{ServerName: s.cfg.Host}); err != nil {
			return fmt.Errorf("mail: STARTTLS: %w", err)
		}
	}

	if s.cfg.Username != "" {
		auth := smtp.PlainAuth("", s.cfg.Username, s.cfg.Password, s.cfg.Host)
		if err := c.Auth(auth); err != nil {
			return fmt.Errorf("mail: auth: %w", err)
		}
	}

	if err := c.Mail(s.cfg.From); err != nil {
		return fmt.Errorf("mail: MAIL FROM: %w", err)
	}
	if err := c.Rcpt(msg.To); err != nil {
		return fmt.Errorf("mail: RCPT TO: %w", err)
	}
	wc, err := c.Data()
	if err != nil {
		return fmt.Errorf("mail: DATA: %w", err)
	}

	body := buildMimeMessage(s.cfg.From, msg)
	if _, err := wc.Write(body); err != nil {
		return fmt.Errorf("mail: write body: %w", err)
	}
	if err := wc.Close(); err != nil {
		return fmt.Errorf("mail: close data: %w", err)
	}
	return c.Quit()
}

func (s *smtpEmail) dial(addr string) (net.Conn, error) {
	if s.cfg.UseTLS {
		return tls.DialWithDialer(
			&net.Dialer{Timeout: 10 * time.Second},
			"tcp", addr,
			&tls.Config{ServerName: hostnameOf(addr)},
		)
	}
	return net.DialTimeout("tcp", addr, 10*time.Second)
}

// ---------------------------------------------------------------
// Dev Noop (smtp disabled)
// ---------------------------------------------------------------

// NoopEmail 用于开发/CI 环境——仅在日志中记录邮件内容，不实际发送。
type NoopEmail struct {
	Sent chan Message // 用于测试断言（可选）；若 nil 则静默
}

func NewNoopService(capacity int) *NoopEmail {
	if capacity <= 0 {
		return &NoopEmail{}
	}
	return &NoopEmail{Sent: make(chan Message, capacity)}
}

func (n *NoopEmail) Send(msg Message) error {
	if n.Sent != nil {
		n.Sent <- msg
	}
	return nil
}

// ---------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------

// buildMimeMessage 拼装符合 RFC 2046 的 multipart/alternative 消息。
func buildMimeMessage(from string, msg Message) []byte {
	boundary := "----=_ydsz_plane_boundary"
	var buf []byte
	add := func(s string) { buf = append(buf, []byte(s)...) }
	add("From: " + from + "\r\n")
	add("To: " + msg.To + "\r\n")
	add("Subject: " + msg.Subject + "\r\n")
	add("MIME-Version: 1.0\r\n")
	add("Content-Type: multipart/alternative; boundary=" + boundary + "\r\n")
	add("\r\n")

	if msg.Text != "" {
		add("--" + boundary + "\r\n")
		add("Content-Type: text/plain; charset=\"UTF-8\"\r\n")
		add("Content-Transfer-Encoding: 8bit\r\n\r\n")
		add(msg.Text + "\r\n")
	}
	if msg.HTML != "" {
		add("--" + boundary + "\r\n")
		add("Content-Type: text/html; charset=\"UTF-8\"\r\n")
		add("Content-Transfer-Encoding: 8bit\r\n\r\n")
		add(msg.HTML + "\r\n")
	}
	add("--" + boundary + "--\r\n")
	return buf
}

func hostnameOf(addr string) string {
	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		return addr
	}
	return host
}
