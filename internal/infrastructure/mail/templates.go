// Package mail — 邮件模板。
//
// 模板写死在代码里（无外部依赖），满足 MVP 阶段的「密码重置」+「工作空间邀请」场景。
// 后续如需富模板可替换为 html/template 或嵌入 gotmpl 文件。
package mail

// ResetPasswordData 用于密码重置邮件。
type ResetPasswordData struct {
	RecipientName string
	ResetURL      string // 一次性重置链接
	TTLMin        int    // token 有效期（分钟），UI 提示用
}

// RenderResetPassword 返回主题 + Text + HTML。
func RenderResetPassword(d ResetPasswordData) Message {
	subject := "重置您的 Ydsz Plane 密码"
	text := "您好 " + d.RecipientName + ",\r\n\r\n" +
		"我们收到了重置您 Ydsz Plane 密码的请求。请点击下面的链接：\r\n\r\n" +
		d.ResetURL + "\r\n\r\n" +
		"链接将在 " + itoa(d.TTLMin) + " 分钟后过期。如非本人操作请忽略此邮件。\r\n\r\n" +
		"— Ydsz Plane 团队"
	html := `<p>您好 ` + d.RecipientName + `，</p>` +
		`<p>我们收到了重置您 Ydsz Plane 密码的请求。点击下方按钮完成操作：</p>` +
		`<p><a href="` + d.ResetURL + `" style="display:inline-block;padding:10px 20px;background:#2563eb;color:#fff;text-decoration:none;border-radius:6px;">重置密码</a></p>` +
		`<p style="color:#666">如按钮无法点击，请复制以下链接到浏览器：<br>` + d.ResetURL + `</p>` +
		`<p style="color:#999;font-size:12px">链接将在 ` + itoa(d.TTLMin) + ` 分钟后过期。如非本人操作请忽略此邮件。</p>` +
		`<p>— Ydsz Plane 团队</p>`
	return Message{To: d.RecipientName, Subject: subject, Text: text, HTML: html}
}

// InviteData 用于工作空间邀请邮件。
type InviteData struct {
	InviteeName   string
	InviterName   string
	WorkspaceName string
	AcceptURL     string
}

// RenderInvitation 返回邀请邮件。
func RenderInvitation(d InviteData) Message {
	subject := d.InviterName + " 邀请您加入「" + d.WorkspaceName + "」"
	text := "您好 " + d.InviteeName + ",\r\n\r\n" +
		d.InviterName + " 邀请您加入工作空间「" + d.WorkspaceName + "」。" +
		"请点击链接接受邀请：\r\n\r\n" +
		d.AcceptURL + "\r\n\r\n" +
		"— Ydsz Plane 团队"
	html := `<p>您好 ` + d.InviteeName + `，</p>` +
		`<p>` + d.InviterName + ` 邀请您加入工作空间「<strong>` + d.WorkspaceName + `</strong>」。</p>` +
		`<p><a href="` + d.AcceptURL + `" style="display:inline-block;padding:10px 20px;background:#2563eb;color:#fff;text-decoration:none;border-radius:6px;">加入工作空间</a></p>` +
		`<p style="color:#999;font-size:12px">— Ydsz Plane 团队</p>`
	// InviteeName 字段同时占位 To；调用方需覆写 To
	return Message{To: d.InviteeName, Subject: subject, Text: text, HTML: html}
}

// itoa 避免 imported strconv 的小整数转换。
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var neg bool
	if n < 0 {
		neg = true
		n = -n
	}
	var b [20]byte
	i := len(b)
	for n > 0 {
		i--
		b[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		b[i] = '-'
	}
	return string(b[i:])
}
