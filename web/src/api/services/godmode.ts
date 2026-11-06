/**
 * God Mode（实例管理）API —— 实例级设置 / 许可证 / 系统状态 / 用量统计。
 *
 * 后端路由：/api/v1/admin/*
 * 权限：实例管理员。判定方式（任一满足）：
 *   1) 请求头 X-Instance-Admin-Key 与环境变量 YDSZ_INSTANCE_ADMIN_KEY 匹配；
 *   2) 已登录用户 ID 在 YDSZ_INSTANCE_ADMINS 白名单中。
 *
 * 因此本客户端需支持注入 `X-Instance-Admin-Key`。由于 `http` 是全局共享的
 * axios 实例（请求拦截器会统一注入 Bearer / CSRF），此处按请求粒度传入
 * headers，避免污染其他业务的请求头。
 *
 * 注意：后端 `InstanceSettings.smtp.password` 为 `json:"-"`，读取时永远为空串，
 * 写入时仅在传入非空值时覆盖（见 mergeSMTP）。
 */
import { http } from "../client";

/* ------------------------------------------------------------------ */
/* Types                                                              */
/* ------------------------------------------------------------------ */

/** 邮件（SMTP）配置 */
export interface SmtpConfig {
  enabled: boolean;
  host: string;
  port: number;
  username: string;
  /** 读取时后端永不回显；仅写入时用于覆盖 */
  password: string;
  use_tls: boolean;
  from_email: string;
  from_name: string;
}

/** 认证配置 */
export interface AuthConfig {
  allow_email_password: boolean;
  allow_oidc: boolean;
  allow_saml: boolean;
  allow_ldap: boolean;
  force_sso: boolean;
  allowed_email_domains: string[];
  allow_signup: boolean;
  session_minutes: number;
}

/** 安全策略 */
export interface SecurityConfig {
  min_password_length: number;
  force_mfa: boolean;
  max_login_attempts: number;
  lockout_minutes: number;
  audit_retention_days: number;
}

/** 遥测配置 */
export interface TelemetryConfig {
  enabled: boolean;
  endpoint: string;
  instance_id: string;
}

/** 界面配置 */
export interface UiConfig {
  show_version: boolean;
  brand_name: string;
  logo_url: string;
  maintenance_mode: boolean;
  maintenance_message: string;
}

/** 实例设置（GET /admin/settings 返回；password 永远为空串） */
export interface InstanceSettings {
  smtp: SmtpConfig;
  auth: AuthConfig;
  security: SecurityConfig;
  telemetry: TelemetryConfig;
  ui: UiConfig;
  updated_at: string;
  updated_by?: number;
}

/** 许可证信息 */
export interface LicenseInfo {
  edition: string;
  plan: string;
  status: string;
  seats: number;
  used_seats: number;
  issued_at?: string;
  expires_at?: string;
  features: string[];
}

/** 组件健康状态 */
export interface ComponentStatus {
  name: string;
  status: string;
  latency?: string;
  message?: string;
}

/** 系统状态 */
export interface SystemStatus {
  version: string;
  edition: string;
  uptime: string;
  uptime_seconds: number;
  started_at: string;
  environment: string;
  features: Record<string, boolean>;
  components: ComponentStatus[];
}

/** 用量统计 */
export interface UsageStats {
  total_workspaces: number;
  total_users: number;
  active_users_30d: number;
  total_projects: number;
  total_work_items: number;
  total_comments: number;
  storage_used_gb: number;
  api_calls_30d: number;
}

/** 设置更新入参（各分组独立可选，未提供的分组不覆盖） */
export interface UpdateSettingsInput {
  smtp?: Partial<SmtpConfig>;
  auth?: Partial<AuthConfig>;
  security?: Partial<SecurityConfig>;
  telemetry?: Partial<TelemetryConfig>;
  ui?: Partial<UiConfig>;
}

/** 测试 SMTP 返回体 */
export interface TestSmtpResult {
  ok: boolean;
  message: string;
  config?: {
    host: string;
    port: number;
    use_tls: boolean;
    from_email: string;
  };
}

/* ------------------------------------------------------------------ */
/* Helpers                                                            */
/* ------------------------------------------------------------------ */

const wrap = <T>(p: Promise<{ data: T }>) => p.then((r) => r.data);

/** 构造 God Mode 客户端。
 * @param adminKey 可选的实例管理员密钥（X-Instance-Admin-Key）；
 *                 未提供时依赖后端 YDSZ_INSTANCE_ADMINS 白名单鉴权。 */
export function createGodModeApi(adminKey?: string) {
  /** 每个请求注入 X-Instance-Admin-Key（有值时） */
  const authHeaders = () =>
    adminKey ? { "X-Instance-Admin-Key": adminKey } : undefined;

  return {
    /** 读取实例设置 */
    getSettings: () =>
      wrap<InstanceSettings>(http.get("/admin/settings", { headers: authHeaders() })),

    /** 更新实例设置（partial，仅覆盖提供的分组） */
    updateSettings: (input: UpdateSettingsInput) =>
      wrap<InstanceSettings>(
        http.put("/admin/settings", input, { headers: authHeaders() }),
      ),

    /** 重置为默认设置 */
    resetSettings: () =>
      wrap<InstanceSettings>(
        http.post("/admin/settings/reset", undefined, { headers: authHeaders() }),
      ),

    /** 许可证信息 */
    getLicense: () =>
      wrap<LicenseInfo>(http.get("/admin/license", { headers: authHeaders() })),

    /** 系统状态 */
    getStatus: () =>
      wrap<SystemStatus>(http.get("/admin/status", { headers: authHeaders() })),

    /** 用量统计 */
    getUsage: () =>
      wrap<UsageStats>(http.get("/admin/usage", { headers: authHeaders() })),

    /** 测试 SMTP 配置（发送校验邮件到指定收件人） */
    testSmtp: (to: string) =>
      wrap<TestSmtpResult>(
        http.post("/admin/test-smtp", { to }, { headers: authHeaders() }),
      ),
  };
}

export type GodModeApi = ReturnType<typeof createGodModeApi>;

/**
 * 默认客户端：未注入 adminKey，依赖后端白名单鉴权。
 * 如需密钥鉴权：`createGodModeApi(key)`。
 */
export const godModeApi = createGodModeApi();
