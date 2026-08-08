/**
 * 轻量 Markdown 渲染器 — 将 Markdown 文本安全地渲染为 HTML。
 *
 * 项目未引入 markdown-it / marked 等重依赖，这里提供一个
 * 自包含、默认 HTML 转义（防 XSS）的最小实现，覆盖常见语法：
 *   标题、粗体/斜体/删除线、行内代码、代码块、链接、图片、
 *   无序/有序列表、引用、分隔线、表格、段落与换行。
 *
 * 注意：这是「够用」级别实现，不追求完整 CommonMark 规范。
 */
export function renderMarkdown(md: string): string {
  if (!md) return "";
  return inline(blocks(escapeHtml(md)));
}

/** 按块级语法切分处理（标题/引用/列表/表格/代码块/分隔线/段落）。 */
function blocks(src: string): string {
  const lines = src.split(/\r?\n/);
  const out: string[] = [];
  let i = 0;

  while (i < lines.length) {
    const line = lines[i];

    // 空行
    if (/^\s*$/.test(line)) {
      i++;
      continue;
    }

    // 代码块（```lang ... ```）
    const fence = line.match(/^`{3,}\s*([\w+-]*)\s*$/);
    if (fence) {
      const lang = fence[1] || "";
      const buf: string[] = [];
      i++;
      while (i < lines.length && !/^`{3,}\s*$/.test(lines[i])) {
        buf.push(lines[i]);
        i++;
      }
      i++; // 跳过结束围栏
      out.push(`<pre><code${lang ? ` class="lang-${lang}"` : ""}>${buf.join("\n")}</code></pre>`);
      continue;
    }

    // 分隔线
    if (/^\s*(---+|\*\*\*+|___+)\s*$/.test(line)) {
      out.push("<hr />");
      i++;
      continue;
    }

    // ATX 标题
    const h = line.match(/^(#{1,6})\s+(.*)$/);
    if (h) {
      const level = h[1].length;
      out.push(`<h${level}>${inline(h[2].trim())}</h${level}>`);
      i++;
      continue;
    }

    // 引用（逐行累积）
    if (/^\s*>\s?/.test(line)) {
      const buf: string[] = [];
      while (i < lines.length && /^\s*>\s?/.test(lines[i])) {
        buf.push(lines[i].replace(/^\s*>\s?/, ""));
        i++;
      }
      out.push(`<blockquote>${blocks(buf.join("\n"))}</blockquote>`);
      continue;
    }

    // 无序列表（逐项累积）
    if (/^\s*[-*+]\s+/.test(line)) {
      const items: string[] = [];
      while (i < lines.length) {
        const m = lines[i].match(/^\s*[-*+]\s+(.*)$/);
        if (m) {
          items.push(`<li>${inline(m[1].trim())}</li>`);
          i++;
        } else if (/^\s*$/.test(lines[i])) {
          break;
        } else {
          break;
        }
      }
      out.push(`<ul>${items.join("")}</ul>`);
      continue;
    }

    // 有序列表（逐项累积）
    if (/^\s*\d+[.)]\s+/.test(line)) {
      const items: string[] = [];
      while (i < lines.length) {
        const m = lines[i].match(/^\s*\d+[.)]\s+(.*)$/);
        if (m) {
          items.push(`<li>${inline(m[1].trim())}</li>`);
          i++;
        } else if (/^\s*$/.test(lines[i])) {
          break;
        } else {
          break;
        }
      }
      out.push(`<ol>${items.join("")}</ol>`);
      continue;
    }

    // 表格（表头 + 分隔行 + 数据行）
    if (/^\s*\|/.test(line) && i + 1 < lines.length && /^\s*\|?[\s:|-]+\|?\s*$/.test(lines[i + 1]) && lines[i + 1].includes("-")) {
      const parseRow = (s: string): string[] =>
        s
          .trim()
          .replace(/^\|/, "")
          .replace(/\|$/, "")
          .split("|")
          .map((c) => c.trim());

      const header = parseRow(line);
      i += 2; // 跳过表头与分隔行
      const rows: string[][] = [];
      while (i < lines.length && /^\s*\|/.test(lines[i])) {
        rows.push(parseRow(lines[i]));
        i++;
      }
      const ths = header.map((c) => `<th>${inline(c)}</th>`).join("");
      const trs = rows
        .map((r) => `<tr>${r.map((c) => `<td>${inline(c)}</td>`).join("")}</tr>`)
        .join("");
      out.push(`<table><thead><tr>${ths}</tr></thead><tbody>${trs}</tbody></table>`);
      continue;
    }

    // 普通段落：累积连续非空行
    const buf: string[] = [];
    while (i < lines.length && !/^\s*$/.test(lines[i])) {
      if (
        /^(#{1,6})\s+/.test(lines[i]) ||
        /^\s*(---+|\*\*\*+|___+)\s*$/.test(lines[i]) ||
        /^\s*>\s?/.test(lines[i]) ||
        /^\s*[-*+]\s+/.test(lines[i]) ||
        /^\s*\d+[.)]\s+/.test(lines[i]) ||
        /^`{3,}/.test(lines[i])
      ) {
        break;
      }
      buf.push(lines[i]);
      i++;
    }
    out.push(`<p>${inline(buf.join("<br />"))}</p>`);
  }

  return out.join("\n");
}

/** 行内语法：行内代码 / 图片 / 链接 / 粗体 / 斜体 / 删除线。 */
function inline(src: string): string {
  let s = src;

  // 行内代码 `code`
  s = s.replace(/`([^`]+)`/g, (_, code: string) => `<code>${code}</code>`);

  // 图片 ![alt](url)
  s = s.replace(/!\[([^\]]*)\]\(([^)\s]+)(?:\s+"[^"]*")?\)/g, (_m, alt: string, url: string) => {
    const href = sanitizeUrl(url);
    return `<img src="${href}" alt="${escapeAttr(alt)}" />`;
  });

  // 链接 [text](url)
  s = s.replace(
    /\[([^\]]+)\]\(([^)\s]+)(?:\s+"[^"]*")?\)/g,
    (_m, text: string, url: string) => {
      const href = sanitizeUrl(url);
      return `<a href="${href}" target="_blank" rel="noopener noreferrer">${text}</a>`;
    },
  );

  // 粗体 **text** 或 __text__
  s = s.replace(/\*\*([^*]+)\*\*|__([^_]+)__/g, (_m, a: string, b: string) => `<strong>${a ?? b}</strong>`);

  // 删除线 ~~text~~
  s = s.replace(/~~([^~]+)~~/g, "<del>$1</del>");

  // 斜体 *text* 或 _text_
  s = s.replace(/\*([^*]+)\*/g, "<em>$1</em>");
  s = s.replace(/_([^_]+)_/g, "<em>$1</em>");

  return s;
}

/** 仅允许 http/https/相对地址，防止 javascript: 等危险协议。 */
function sanitizeUrl(url: string): string {
  const trimmed = url.trim();
  if (/^(https?:|mailto:|#)/i.test(trimmed) || trimmed.startsWith("/")) {
    return trimmed;
  }
  return "#";
}

function escapeAttr(v: string): string {
  return v.replace(/&/g, "&amp;").replace(/"/g, "&quot;").replace(/</g, "&lt;").replace(/>/g, "&gt;");
}

function escapeHtml(s: string): string {
  return s.replace(/&/g, "&amp;").replace(/</g, "&lt;").replace(/>/g, "&gt;").replace(/"/g, "&quot;");
}
