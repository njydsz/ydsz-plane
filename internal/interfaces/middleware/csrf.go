// middleware/csrf.go —— 会话型 CSRF 防护中间件（HMAC 双提交 Cookie 模式）。
//
// 设计原理
// --------
// 浏览器 SPA 通过 ydsz_access 会话 Cookie 自动携带凭证，这使得它天然面临
// CSRF 攻击：恶意站点可以诱导浏览器向本 API 发起跨站请求，浏览器会自动
// 附带会话 Cookie，从而以受害者身份执行状态变更操作。
//
// 业界主流 SPA CSRF 防御是「双提交 Cookie」(double submit cookie)：
//  1. 服务端下发一个可读（非 HttpOnly）的 CSRF Cookie（X-CSRF-TOKEN）。
//  2. 前端 JS 读取该 Cookie 并在每次状态变更请求中以 X-CSRF-Token 头回传。
//  3. 服务端比对 Cookie 与 Header 中的令牌，一致则放行，否则 403。
//
// 跨站攻击者受同源策略（SOP）限制，无法读取本域的 XSRF Cookie，因而无法
// 构造匹配的 X-CSRF-Token 头；同时自定义头会自动触发 CORS 预检，我们的
// 限制性 CORS 策略（仅允许可信来源）会拦截未授权来源的预检。两重屏障
// 共同确保防御有效（纵深防御）。
//
// 令牌绑定策略
// ------------
// 为避免服务端持久化令牌，我们采用 HMAC-SHA256(AccessToken, JWTSecret)
// 派生 CSRF 令牌。这样做的安全收益：
//  - 每个会话的 CSRF 令牌唯一且不可预测（依赖 JWT Secret 保密）。
//  - 令牌与当前 Access Token 强绑定：Access Token 轮换（刷新）时 CSRF
//    令牌同步变更，旧令牌立即失效，杜绝令牌复用窗口。
//  - 零服务端状态：无需引入 Redis/DB 存储，不影响水平扩展。
//
// 作用域
// ------
// 仅对「使用会话 Cookie 认证的浏览器 SPA」生效：
//  - 状态变更方法（POST/PUT/PATCH/DELETE）+ 存在 ydsz_access Cookie → 强制校验。
//  - 安全方法（GET/HEAD/OPTIONS）→ 跳过校验（仅确保 CSRF Cookie 存在）。
//  - 无会话 Cookie（X-Api-Key / Bearer API 客户端）→ 完全不受影响。
//  - WebSocket 升级请求 → 跳过（握手阶段依赖 origin 校验，见 ws.Hub）。
//
// 实现纯标准库：crypto/hmac、crypto/sha256、crypto/subtle、encoding/hex。
// 参考：OWASP CSRF Prevention Cheat Sheet（Double Submit Cookie 节）。
package middleware

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"net/http"
	"slices"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/njydsz/ydsz-plane/internal/config"
	"github.com/njydsz/ydsz-plane/pkg/errs"
)

const (
	// csrfTokenCookie 是前端可读的 CSRF 双提交令牌 Cookie。
	// 非 HttpOnly —— 前端 JS 必须能读取其值并放入 X-CSRF-Token 头。
	csrfTokenCookie = "X-CSRF-TOKEN"
	// csrfTokenHeader 是前端携带 CSRF 令牌的请求头。
	csrfTokenHeader = "X-CSRF-Token"
	// sessionCookie 是浏览器 SPA 的会话 Access Token Cookie。
	sessionCookie = "ydsz_access"
)

// safeMethods 是无需 CSRF 防护的安全 HTTP 方法（幂等、只读）。
var safeMethods = []string{
	http.MethodGet,
	http.MethodHead,
	http.MethodOptions,
}

// CSRF 构造 CSRF 防护中间件。
//
// secret 为派生 CSRF 令牌的 HMAC 密钥，推荐传入 JWT Secret。
func CSRF(cfg *config.Config, secret string) gin.HandlerFunc {
	sec := []byte(secret)
	return func(c *gin.Context) {
		// 1. WebSocket 升级请求跳过 CSRF。握手阶段由 ws.Hub 做 origin 校验，
		//    自定义头在部分 WS 客户端中不便设置。
		if isWebSocketUpgrade(c) {
			c.Next()
			return
		}

		// 2. 安全方法跳过状态变更校验，但确保前端持有 CSRF Cookie，
		//    避免 SPA 首次加载时 GET 后紧跟 POST 因缺 Cookie 被 403。
		if slices.Contains(safeMethods, c.Request.Method) {
			ensureCSRFCookie(c, cfg, sec)
			c.Next()
			return
		}

		// 3. 未携带会话 Cookie → 纯 API 客户端（X-Api-Key / Bearer），无 CSRF 风险。
		access, err := c.Cookie(sessionCookie)
		if err != nil || access == "" {
			c.Next()
			return
		}

		// 4. 状态变更 + 会话 Cookie → 强制校验双提交令牌。
		expected := deriveCSRFToken(sec, access)

		// 优先读取自定义头；也接受表单 _csrf 字段作为传统表单兜底。
		presented := c.GetHeader(csrfTokenHeader)
		if presented == "" {
			presented = c.PostForm("_csrf")
		}

		if presented == "" || !constantTimeEqual(presented, expected) {
			respondError(c, errs.ErrCSRFTokenInvalid)
			c.Abort()
			return
		}

		c.Next()
	}
}

// deriveCSRFToken 使用 HMAC-SHA256(secret, accessToken) 派生会话绑定的
// CSRF 令牌（hex 编码）。会话轮换时新 Access Token 会自然生成新令牌，
// 旧令牌立即失效。
func deriveCSRFToken(secret []byte, accessToken string) string {
	mac := hmac.New(sha256.New, secret)
	mac.Write([]byte(accessToken))
	return hex.EncodeToString(mac.Sum(nil))
}

// constantTimeEqual 使用恒定时间比较避免时序侧信道泄露。
func constantTimeEqual(a, b string) bool {
	return subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1
}

// isWebSocketUpgrade 识别 WebSocket 升级请求。
func isWebSocketUpgrade(c *gin.Context) bool {
	conn := strings.ToLower(c.GetHeader("Connection"))
	up := strings.ToLower(c.GetHeader("Upgrade"))
	return strings.Contains(conn, "upgrade") && up == "websocket"
}

// ensureCSRFCookie 确保前端持有 CSRF 双提交 Cookie。
//
// 仅当会话 Cookie 存在且 CSRF Cookie 不存在或已失效时才写入，
// 避免无条件覆盖导致令牌与当前会话不一致。
// SPA 前端应在启动时通过任意 GET 请求触发此逻辑，建立 CSRF Cookie。
func ensureCSRFCookie(c *gin.Context, cfg *config.Config, secret []byte) {
	// 无会话 Cookie 时不写入：会话未建立，写入无意义且浪费带宽。
	access, err := c.Cookie(sessionCookie)
	if err != nil || access == "" {
		return
	}
	// CSRF Cookie 已存在则保留（避免引入不必要的写入与竞态）。
	if existing, err := c.Cookie(csrfTokenCookie); err == nil && existing != "" {
		return
	}
	token := deriveCSRFToken(secret, access)
	writeCSRFCookie(c, cfg, token)
}

// writeCSRFCookie 将 CSRF 令牌写入可前端读取的 Cookie。
func writeCSRFCookie(c *gin.Context, cfg *config.Config, token string) {
	secure := !cfg.IsDev()
	http.SetCookie(c.Writer, &http.Cookie{
		Name: csrfTokenCookie, Value: token, Path: "/",
		// 显式非 HttpOnly：前端 JS 必须可读以完成双提交。
		HttpOnly: false,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
		// 不设 MaxAge → 会话级 Cookie，随浏览器会话消亡而清除；
		// 后端 Access Token 刷新（setAuthCookies）会重写此 Cookie 以轮换。
	})
}

// SetCSRFTokenCookie 公开辅助：在登录/注册/刷新等与会话 Atomic 建立的
// 处理流程中，基于已有 Access Token 写入 CSRF 令牌 Cookie，
// 避免 CSRF Cookie 与 Access Token 不同步。
func SetCSRFTokenCookie(c *gin.Context, cfg *config.Config, secret, accessToken string) {
	writeCSRFCookie(c, cfg, deriveCSRFToken([]byte(secret), accessToken))
}
