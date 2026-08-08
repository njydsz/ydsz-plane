-- 0002: 文档模块增强 — 添加文档模板与公开分享链接表
--
-- 变更内容:
--   1. 创建 page_templates 表（项目级 / 工作空间级文档模板）
--   2. 创建 page_shares 表（文档公开分享链接，可选密码与过期时间）
--   3. 内置 4 条开箱即用模板（需求文档 / 技术方案 / 会议纪要 / 复盘）

BEGIN;

-- 1. 文档模板表
-- project_id = 0 表示工作空间级模板（跨项目共享）
CREATE TABLE IF NOT EXISTS page_templates (
    id              BIGSERIAL PRIMARY KEY,
    workspace_id    BIGINT NOT NULL,
    project_id      BIGINT NOT NULL DEFAULT 0,
    name            TEXT NOT NULL,
    description     TEXT,
    content_html    TEXT,
    category        TEXT,
    created_by      BIGINT NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT fk_page_templates_workspace FOREIGN KEY (workspace_id) REFERENCES workspaces(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_page_templates_workspace ON page_templates(workspace_id, project_id);

COMMENT ON TABLE page_templates IS '文档模板表（项目级 / 工作空间级）';
COMMENT ON COLUMN page_templates.project_id IS '项目 FK；0 表示工作空间级模板';
COMMENT ON COLUMN page_templates.content_html '模板的富文本 HTML 内容';
COMMENT ON COLUMN page_templates.category IS '模板分类（requirement / technical / meeting / review）';

-- 2. 文档公开分享链接表
CREATE TABLE IF NOT EXISTS page_shares (
    id              BIGSERIAL PRIMARY KEY,
    page_id         BIGINT NOT NULL,
    workspace_id    BIGINT NOT NULL,
    project_id      BIGINT NOT NULL,
    token           TEXT NOT NULL UNIQUE DEFAULT md5(random()::text || clock_timestamp()::text),
    is_active       BOOLEAN NOT NULL DEFAULT true,
    password_hash   TEXT,
    expires_at      TIMESTAMPTZ,
    created_by      BIGINT NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT fk_page_shares_page FOREIGN KEY (page_id) REFERENCES pages(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_page_shares_page ON page_shares(page_id);
CREATE INDEX IF NOT EXISTS idx_page_shares_token ON page_shares(token);

COMMENT ON TABLE page_shares IS '文档公开分享链接表（token 访问，可选密码与过期时间）';
COMMENT ON COLUMN page_shares.token IS '分享令牌（md5 随机生成，URL 中唯一标识）';
COMMENT ON COLUMN page_shares.password_hash IS '可选访问密码（bcrypt 哈希；NULL 表示无需密码）';
COMMENT ON COLUMN page_shares.expires_at IS '过期时间（NULL 表示永不过期）';

-- 3. 内置模板（workspace_id = 0、project_id = 0 表示全局内置模板）
INSERT INTO page_templates (workspace_id, project_id, name, description, content_html, category, created_by)
VALUES
(0, 0, '需求文档模板', '标准产品需求文档结构：背景、目标、功能说明、验收标准',
 '<h2>一、背景</h2><p>描述业务背景与问题。</p><h2>二、目标</h2><p>列出本需求要达成的核心目标。</p><h2>三、功能说明</h2><p>详细描述功能点与交互逻辑。</p><h2>四、验收标准</h2><ol><li>验收条件 1</li><li>验收条件 2</li></ol>',
 'requirement', 0),

(0, 0, '技术方案模板', '技术设计方案结构：概述、架构设计、接口设计、风险评估',
 '<h2>一、概述</h2><p>方案背景与范围。</p><h2>二、架构设计</h2><p>系统架构图与核心模块说明。</p><h2>三、接口设计</h2><p>API 定义、请求/响应格式。</p><h2>四、风险评估</h2><p>潜在风险与缓解措施。</p>',
 'technical', 0),

(0, 0, '会议纪要模板', '会议纪要标准结构：参会人、议题、决议、待办',
 '<h2>会议信息</h2><ul><li>时间：</li><li>地点/链接：</li><li>参会人：</li></ul><h2>议题讨论</h2><p>逐项记录讨论要点。</p><h2>决议事项</h2><ol><li>决议 1</li><li>决议 2</li></ol><h2>待办项</h2><ul><li>【负责人】任务描述 — 截止日期</li></ul>',
 'meeting', 0),

(0, 0, '复盘模板', '项目复盘结构：回顾目标、评估结果、分析原因、总结行动',
 '<h>、回顾目标</h2><p>当初的目标是什么？</p><h2>二、评估结果</h2><p>实际达成的结果与偏差。</p><h2>三、分析原因</h2><p>成功因素与失败根因。</p><h2>四、行动计划</h2><ol><li>继续保持：...</li><li>开始做：...</li><li>停止做：...</li></ol>',
 'review', 0);

COMMIT;
