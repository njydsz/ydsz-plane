/**
 * 主题管理 — 亮/暗模式切换与持久化。
 *
 * 对标 Plane 的主题体系：tokens.css 已定义 [data-theme="light"/"dark"] 两套完整
 * 语义色板，本模块负责切换、持久化与跟随系统偏好。
 *
 * 优先级：用户显式选择 > 系统偏好 (prefers-color-scheme) > 默认 light。
 */
import { ref } from "vue";

const STORAGE_KEY = "ydsz:theme";

export type ThemeMode = "light" | "dark" | "system";

/** 当前实际生效的主题 */
const current = ref<"light" | "dark">("light");

/** 当前用户偏好模式（light/dark/system） */
const mode = ref<ThemeMode>(loadMode());

function loadMode(): ThemeMode {
  try {
    const raw = localStorage.getItem(STORAGE_KEY);
    if (raw === "light" || raw === "dark" || raw === "system") return raw;
  } catch { /* ignore */ }
  return "system";
}

function systemPrefersDark(): boolean {
  return (
    typeof window !== "undefined" &&
    window.matchMedia &&
    window.matchMedia("(prefers-color-scheme: dark)").matches
  );
}

function apply(theme: "light" | "dark") {
  current.value = theme;
  document.documentElement.setAttribute("data-theme", theme);
}

function persist(m: ThemeMode) {
  try {
    localStorage.setItem(STORAGE_KEY, m);
  } catch { /* ignore */ }
}

/** 解析最终主题并应用到 <html> */
export function resolveTheme(m: ThemeMode = mode.value): "light" | "dark" {
  const t = m === "system" ? (systemPrefersDark() ? "dark" : "light") : m;
  apply(t);
  return t;
}

/** 切换偏好模式（内部应用；可选持久化由调用方决定） */
export function setThemeMode(m: ThemeMode, persistPref = true): "light" | "dark" {
  mode.value = m;
  if (persistPref) persist(m);
  return resolveTheme(m);
}

/** 获取当前偏好模式 */
export function getThemeMode(): ThemeMode {
  return mode.value;
}

/** 当前生效主题（reactive，可直接在组件中 watch） */
export function useTheme() {
  return current;
}

/** 监听系统偏好变化（system 模式下自动跟随） */
export function watchSystemTheme() {
  if (typeof window === "undefined" || !window.matchMedia) return;
  const mq = window.matchMedia("(prefers-color-scheme: dark)");
  const handler = () => {
    if (mode.value === "system") resolveTheme("system");
  };
  mq.addEventListener("change", handler);
}

/** 应用启动初始化 */
export function initTheme() {
  resolveTheme(mode.value);
  watchSystemTheme();
}
