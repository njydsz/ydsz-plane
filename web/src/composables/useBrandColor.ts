/**
 * 品牌色 — 将工作空间自定义品牌色应用到 CSS 变量。
 *
 * 工作空间设置的品牌色（HEX 格式）会被转换为 oklch 色彩空间，
 * 并覆盖 --brand-* 系列 CSS 变量，实现全局主题色切换。
 */
import { ref, watch, type Ref } from "vue";

/** 默认品牌色（与 tokens.css 中的 --brand-default 一致） */
const DEFAULT_BRAND_COLOR = "#2563eb";

/** 预设品牌色板 */
export const BRAND_COLOR_PRESETS = [
  { name: "经典蓝", value: "#2563eb" },
  { name: "靛蓝", value: "#4f46e5" },
  { name: "紫色", value: "#7c3aed" },
  { name: "粉色", value: "#db2777" },
  { name: "红色", value: "#dc2626" },
  { name: "橙色", value: "#ea580c" },
  { name: "琥珀", value: "#d97706" },
  { name: "绿色", value: "#16a34a" },
  { name: "青色", value: "#0891b2" },
  { name: "石板灰", value: "#475569" },
];

/** HEX 转 RGB */
function hexToRgb(hex: string): { r: number; g: number; b: number } | null {
  const result = /^#?([a-f\d]{2})([a-f\d]{2})([a-f\d]{2})$/i.exec(hex);
  if (!result) return null;
  return {
    r: parseInt(result[1], 16),
    g: parseInt(result[2], 16),
    b: parseInt(result[3], 16),
  };
}

/** RGB 转 oklch 的近似值（用于 CSS 变量覆盖） */
function rgbToOklch(r: number, g: number, b: number): { l: number; c: number; h: number } {
  // 归一化到 [0, 1]
  const rn = r / 255;
  const gn = g / 255;
  const bn = b / 255;

  // sRGB → Linear RGB
  const toLinear = (v: number) => (v <= 0.04045 ? v / 12.92 : Math.pow((v + 0.055) / 1.055, 2.4));
  const rl = toLinear(rn);
  const gl = toLinear(gn);
  const bl = toLinear(bn);

  // Linear RGB → XYZ (D65)
  const x = 0.4124564 * rl + 0.3575761 * gl + 0.1804375 * bl;
  const y = 0.2126729 * rl + 0.7151522 * gl + 0.0721750 * bl;
  const z = 0.0193339 * rl + 0.1191920 * gl + 0.9503041 * bl;

  // XYZ → LMS
  const lms1 = 0.8189330101 * x + 0.3618667424 * y - 0.1288597137 * z;
  const lms2 = 0.0329845436 * x + 0.9293118715 * y + 0.0361456387 * z;
  const lms3 = 0.0482003018 * x + 0.2643662691 * y + 0.6338517070 * z;

  // LMS → LMS' (cube root)
  const cbrt = (v: number) => Math.sign(v) * Math.pow(Math.abs(v), 1 / 3);
  const lms1p = cbrt(lms1);
  const lms2p = cbrt(lms2);
  const lms3p = cbrt(lms3);

  // LMS' → OKLab
  const L = 0.2104542553 * lms1p + 0.7936177850 * lms2p - 0.0040720468 * lms3p;
  const a = 1.9779984951 * lms1p - 2.4285922050 * lms2p + 0.4505937099 * lms3p;
  const bVal = 0.0259040371 * lms1p + 0.7827717662 * lms2p - 0.8086757660 * lms3p;

  // OKLab → OKLCH
  const C = Math.sqrt(a * a + bVal * bVal);
  let h = Math.atan2(bVal, a) * (180 / Math.PI);
  if (h < 0) h += 360;

  return { l: L, c: C, h };
}

/** 生成品牌色阶（100-1200） */
function generateBrandShades(hex: string): Record<string, string> {
  const rgb = hexToRgb(hex);
  if (!rgb) return {};

  const oklch = rgbToOklch(rgb.r, rgb.g, rgb.b);
  const { l, c, h } = oklch;

  // 生成色阶：调整亮度 L，保持色度 C 和色相 H
  const shades: Record<string, string> = {};
  const lightnessSteps: [string, number][] = [
    ["100", Math.min(l + 0.15, 0.985)],
    ["200", Math.min(l + 0.12, 0.97)],
    ["300", Math.min(l + 0.08, 0.94)],
    ["400", Math.min(l + 0.04, 0.90)],
    ["500", Math.min(l + 0.02, 0.84)],
    ["600", l],
    ["700", Math.max(l - 0.05, 0.67)],
    ["900", Math.max(l - 0.15, 0.43)],
    ["1000", Math.max(l - 0.22, 0.34)],
    ["1100", Math.max(l - 0.30, 0.26)],
    ["1200", Math.max(l - 0.36, 0.21)],
  ];

  for (const [key, L] of lightnessSteps) {
    // 高亮度时降低色度，低亮度时适当降低色度
    const chromaAdjust = L > 0.9 ? 0.3 : L > 0.8 ? 0.5 : L < 0.4 ? 0.6 : 0.8;
    const C = Math.min(c * chromaAdjust, 0.15);
    shades[`--brand-${key}`] = `oklch(${L.toFixed(4)} ${C.toFixed(4)} ${h.toFixed(2)})`;
  }

  // brand-default 使用 600 色阶
  shades["--brand-default"] = shades["--brand-600"];

  return shades;
}

/** 应用品牌色到文档根元素 */
export function applyBrandColor(hex: string | null) {
  const root = document.documentElement;
  if (!hex) {
    // 清除自定义品牌色，恢复默认
    clearBrandColor();
    return;
  }

  const shades = generateBrandShades(hex);
  for (const [prop, value] of Object.entries(shades)) {
    root.style.setProperty(prop, value);
  }
}

/** 清除自定义品牌色 */
export function clearBrandColor() {
  const root = document.documentElement;
  const props = [
    "--brand-100", "--brand-200", "--brand-300", "--brand-400",
    "--brand-500", "--brand-600", "--brand-700", "--brand-900",
    "--brand-1000", "--brand-1100", "--brand-1200", "--brand-default",
  ];
  for (const prop of props) {
    root.style.removeProperty(prop);
  }
}

/** 品牌色 composable */
export function useBrandColor(brandColor: Ref<string | undefined | null>) {
  const currentColor = ref(brandColor.value || "");

  watch(
    brandColor,
    (newColor) => {
      currentColor.value = newColor || "";
      if (newColor) {
        applyBrandColor(newColor);
      } else {
        clearBrandColor();
      }
    },
    { immediate: true },
  );

  /** 设置品牌色 */
  function setBrandColor(hex: string) {
    currentColor.value = hex;
    applyBrandColor(hex);
  }

  /** 重置为默认品牌色 */
  function resetBrandColor() {
    currentColor.value = DEFAULT_BRAND_COLOR;
    clearBrandColor();
  }

  return {
    currentColor,
    setBrandColor,
    resetBrandColor,
    presets: BRAND_COLOR_PRESETS,
    defaultColor: DEFAULT_BRAND_COLOR,
  };
}
