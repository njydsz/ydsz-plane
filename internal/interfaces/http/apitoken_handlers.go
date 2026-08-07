// Package httpapi — 个人 API Token 端点处理函数。
//
// 路由挂载于 /api/v1/me/api-tokens（用户级，与工作空间无关），
// 供 SPA 设置页与脚本/集成管理个人访问令牌。
package httpapi

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/njydsz/ydsz-plane/internal/application/apitoken"
	"github.com/njydsz/ydsz-plane/internal/interfaces/http/dto"
	"github.com/njydsz/ydsz-plane/internal/interfaces/middleware"
	"github.com/njydsz/ydsz-plane/pkg/errs"
)

// listMyApiTokens 返回当前用户的全部活跃令牌。
//
//	@Summary		我的 API 令牌列表
//	@Description	列出当前用户全部未吊销的个人访问令牌（不含明文）
//	@Tags			api-tokens
//	@Produce		json
//	@Success		200	{array}	apitoken.TokenVM
//	@Failure		401	{object}	errs.AppError
//	@Router			/me/api-tokens [get]
func listMyApiTokens(d *Deps) gin.HandlerFunc {
	return func(c *gin.Context) {
		uid := c.GetInt64(middleware.CtxUserID)
		tokens, err := d.ApiTokenSvc.List(c.Request.Context(), uid)
		if err != nil {
			writeError(c, err)
			return
		}
		c.JSON(http.StatusOK, tokens)
	}
}

// createApiToken 创建令牌并一次性返回明文。
//
//	@Summary		创建 API 令牌
//	@Description	创建个人访问令牌，原始值仅在本次响应中返回一次
//	@Tags			api-tokens
//	@Accept			json
//	@Produce		json
//	@Param			body	body		dto.CreateApiTokenRequest	true	"令牌参数"
//	@Success		201		{object}	apitoken.CreatedToken
//	@Failure		401		{object}	errs.AppError
//	@Failure		409		{object}	errs.AppError
//	@Failure		422		{object}	errs.AppError
//	@Router			/me/api-tokens [post]
func createApiToken(d *Deps) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req dto.CreateApiTokenRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			middleware.AbortWithError(c, errs.ErrValidation.WithDetails(fieldDetails(err)...))
			return
		}

		var expiresIn *time.Duration
		if req.ExpiresInSeconds != nil {
			dur := time.Duration(*req.ExpiresInSeconds) * time.Second
			expiresIn = &dur
		}

		created, err := d.ApiTokenSvc.Create(c.Request.Context(), apitoken.CreateInput{
			UserID:    c.GetInt64(middleware.CtxUserID),
			Name:      req.Name,
			Scopes:    req.Scopes,
			ExpiresIn: expiresIn,
		})
		if err != nil {
			writeError(c, err)
			return
		}
		c.JSON(http.StatusCreated, created)
	}
}

// revokeApiToken 吊销指定令牌。
//
//	@Summary		吊销 API 令牌
//	@Description	吊销令牌后该令牌立即失效；只能操作本人令牌
//	@Tags			api-tokens
//	@Param			token_id	path	int	true	"令牌 ID"
//	@Success		204
//	@Failure		401	{object}	errs.AppError
//	@Failure		404	{object}	errs.AppError
//	@Router			/me/api-tokens/{token_id} [delete]
func revokeApiToken(d *Deps) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenID, err := strconv.ParseInt(c.Param("token_id"), 10, 64)
		if err != nil || tokenID <= 0 {
			middleware.AbortWithError(c, errs.ErrValidation.WithDetails(errs.FieldDetail{
				Field: "token_id", Reason: "无效的令牌 ID",
			}))
			return
		}
		if err := d.ApiTokenSvc.Revoke(c.Request.Context(), c.GetInt64(middleware.CtxUserID), tokenID); err != nil {
			writeError(c, err)
			return
		}
		c.Status(http.StatusNoContent)
	}
}
