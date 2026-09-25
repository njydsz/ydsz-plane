-- P0-3: Refresh Token 轮换与家族吊销
--
-- 为每次登录签发唯一的 refresh_token 家族 UUID，Refresh 时将旧家族标记为 revoked(reason=rotation)，
-- 签发新家族。若发现家族已被撤销（如检测到重放复用），则拒绝此次刷新并返回安全吊销信号。
--
-- Redis 缓存暂不做（TODO: 使用 Redis Hash 缓存家族状态实现 O(1) 查询，key=rtf:{family_id} TTL=refresh_ttl）

CREATE TABLE IF NOT EXISTS refresh_token_families (
    family_id      TEXT PRIMARY KEY,
    user_id        BIGINT NOT NULL,
    provider_id    BIGINT NULL,
    issued_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    expires_at     TIMESTAMPTZ NOT NULL,
    revoked_at     TIMESTAMPTZ NULL,
    revoked_reason TEXT NULL
);

CREATE INDEX IF NOT EXISTS idx_rtf_user ON refresh_token_families(user_id, revoked_at);

COMMENT ON TABLE refresh_token_families IS 'refresh token 家族：每次登录产生唯一家族，Refresh 时旧家族 revoked(rotation) 后签发新家族；重放使用已撤销家族时触发安全告警';
COMMENT ON COLUMN refresh_token_families.family_id IS '每次登录生成的唯一 UUID，嵌入 refresh JWT 的 fid claim';
COMMENT ON COLUMN refresh_token_families.user_id IS '持有者用户 ID';
COMMENT ON COLUMN refresh_token_families.provider_id IS 'SSO Provider 场景关联（密码登录为 NULL）';
COMMENT ON COLUMN refresh_token_families.revoked_at IS 'NULL = 有效；非空 = 已吊销';
COMMENT ON COLUMN refresh_token_families.revoked_reason IS 'revoke 原因: rotation(轮换) / logout(登出) / security(安全吊销)';
