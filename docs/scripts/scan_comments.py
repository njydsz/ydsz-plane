#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
注释覆盖率扫描器（Ydsz Plane）

功能：
- Go：包注释覆盖率、导出符号（类型/函数/方法/常量/变量）注释覆盖率、未导出复杂符号（启发式）覆盖率
- TS/Vue：文件头注释覆盖率、导出成员 JSDoc 覆盖率

用法：
    python scan_comments.py           # 全量报告
    python scan_comments.py --todo    # 待办明细
    python scan_comments.py --dir xxx # 只扫指定目录
"""
import argparse
import json
import os
import re
import sys

ROOT = os.path.dirname(os.path.dirname(os.path.dirname(os.path.abspath(__file__))))


# ---------------------------------------------------------------------------
# Go 解析
# ---------------------------------------------------------------------------

GO_DECL = re.compile(
    r"^(?:type|func|const|var)\s+(?P<name>[A-Z]\w*)|"
    r"^func\s+\([^)]*\)\s+(?P<method>[A-Z]\w*)",
    re.MULTILINE,
)
GO_BLOCK_COMMENT = re.compile(r"/\*.*?\*/", re.DOTALL)
GO_LINE_COMMENT = re.compile(r"//[^\n]*", re.MULTILINE)
GO_STRINGS = re.compile(r'"(?:[^"\\]|\\.)*"|`(?:[^`]|\\`)*`')

CONTROL_KEYWORDS = {"if", "for", "switch", "select", "case", "return", "go", "defer"}


def strip_go_strings(src):
    return GO_STRINGS.sub("", src)


def strip_block_comments(src):
    """去掉块注释；处理注释内的行注释符与字符串文本块干扰。"""
    src = GO_BLOCK_COMMENT.sub(" ", src)
    return src


def has_doc_above(src, pos):
    """判断位置 pos 之前最近的非空内容是否为注释块/行注释。"""
    head = src[:pos].rstrip("\n")
    if not head:
        return False
    lines = head.split("\n")
    # 找第一个非空行
    i = len(lines) - 1
    while i >= 0 and not lines[i].strip():
        i -= 1
    if i < 0:
        return False
    last = lines[i].lstrip()
    if last.startswith("//") or last.startswith("/*") or last.startswith("*"):
        # 行注释可能属于声明自己后面的内容——只看紧邻的一行是否注释
        return True
    return False


def has_doc_above_generic(src, pos):
    """更严格：紧邻上方（跳过空行/注解）是否有注释。"""
    head = src[:pos].rstrip("\n")
    if not head:
        return False
    lines = head.split("\n")
    i = len(lines) - 1
    while i >= 0 and not lines[i].strip():
        i -= 1
    if i < 0:
        return False
    last = lines[i].lstrip()
    return last.startswith("//") or last.startswith("/*") or last.startswith("*")


def analyze_go_file(path):
    with open(path, encoding="utf-8", errors="replace") as f:
        raw = f.read()
    src_wo_str = strip_go_strings(raw)
    src_wo_block = strip_block_comments(src_wo_str)

    pkg_m = re.search(r"^package\s+\w+", raw, re.MULTILINE)
    has_pkg_doc = False
    if pkg_m:
        pos = pkg_m.start()
        head = raw[:pos]
        if "/*" in head or re.search(r"//[^\n]*$", head, re.MULTILINE):
            # 包注释必须在 package 紧邻上方
            lines = head.split("\n")
            i = len(lines) - 1
            while i >= 0 and not lines[i].strip():
                i -= 1
            if i >= 0 and (lines[i].lstrip().startswith("//") or "/*" in head):
                has_pkg_doc = True

    exported = []  # (name, line, kind)
    # 类型声明
    for m in re.finditer(
        r"^type\s+(?P<name>[A-Z]\w*)\s+(?P<rest>.*)$", src_wo_block, re.MULTILINE
    ):
        if m.group("rest").startswith("=") or m.group("rest").startswith("struct") \
           or m.group("rest").startswith("interface") or m.group("rest") == "":
            line = raw[: m.start()].count("\n") + 1
            exported.append((m.group("name"), line, "type"))
    # 函数
    for m in re.finditer(
        r"^func\s+(?P<name>[A-Z]\w*)\s*\(", src_wo_block, re.MULTILINE
    ):
        line = raw[: m.start()].count("\n") + 1
        exported.append((m.group("name"), line, "func"))
    # 方法
    for m in re.finditer(
        r"^func\s+\([^)]*\)\s+(?P<name>[A-Z]\w*)\s*\(", src_wo_block, re.MULTILINE
    ):
        line = raw[: m.start()].count("\n") + 1
        exported.append((m.group("name"), line, "method"))
    # 常量/变量
    for m in re.finditer(
        r"^(?:const|var)\s+\(?\s*(?P<name>[A-Z]\w*)\s*[=:]",
        src_wo_block,
        re.MULTILINE,
    ):
        line = raw[: m.start()].count("\n") + 1
        exported.append((m.group("name"), line, "constvar"))

    # 去重（同名同类型可能被多次匹配）
    seen = set()
    uniq = []
    for name, line, kind in exported:
        key = (name, kind)
        if key in seen:
            continue
        seen.add(key)
        uniq.append((name, line, kind))

    missing = []
    for name, line, kind in uniq:
        # 定位该符号在源码中的位置（用原始 src 定位注释）
        pat = re.compile(r"^%s\b" % re.escape(name), re.MULTILINE)
        pos = None
        for m in pat.finditer(src_wo_block):
            ln = raw[: m.start()].count("\n") + 1
            if ln >= line - 1 and ln <= line + 1:
                pos = m.start()
                break
        if pos is None:
            continue
        if not has_doc_above_generic(raw, pos):
            missing.append({"name": name, "line": line, "kind": kind})

    total = len(uniq)
    return {
        "file": path,
        "has_pkg_doc": has_pkg_doc,
        "exported_total": total,
        "exported_documented": total - len(missing),
        "exported_missing": missing,
    }


# ---------------------------------------------------------------------------
# TS / Vue 解析
# ---------------------------------------------------------------------------

TS_EXPORT = re.compile(
    r"^export\s+(?:default\s+)?(?:interface|type|class|function|const|enum|abstract\s+class)\s+"
    r"(?P<name>[A-Za-z_$][\w$]*)",
    re.MULTILINE,
)
TS_EXPORT_OBJ = re.compile(r"^export\s+(?P<name>[A-Za-z_$][\w$]*)\s*=", re.MULTILINE)


def has_jsdoc_above(src, pos):
    head = src[:pos].rstrip("\n")
    if not head:
        return False
    lines = head.split("\n")
    i = len(lines) - 1
    while i >= 0 and not lines[i].strip():
        i -= 1
    if i < 0:
        return False
    # 多行 JSDoc：上一行是 */ 则回溯找 /**（最多 100 行）
    if lines[i].strip() == "*/":
        j = i - 1
        cnt = 0
        while j >= 0 and cnt < 100:
            s = lines[j].strip()
            if s.startswith("/**"):
                return True
            if s.startswith("/*") or s.startswith("//"):
                break
            j -= 1
            cnt += 1
        return False
    last = lines[i].lstrip()
    return last.startswith("/**") or last.startswith("//") or last.startswith("/*")


def analyze_ts_file(path):
    with open(path, encoding="utf-8", errors="replace") as f:
        src = f.read()

    # 文件头注释：前 30 行内存在 /** */
    head = src[:4000]
    has_header = "/**" in head or bool(re.search(r"^\s*/\*", src[:200], re.MULTILINE))

    members = []
    for m in TS_EXPORT.finditer(src):
        members.append((m.group("name"), src[: m.start()].count("\n") + 1))
    for m in TS_EXPORT_OBJ.finditer(src):
        members.append((m.group("name"), src[: m.start()].count("\n") + 1))
    # 去重
    seen = set()
    uniq = []
    for name, line in members:
        if name in seen:
            continue
        seen.add(name)
        uniq.append((name, line))

    missing = []
    for name, line in uniq:
        # 定位
        pat = re.compile(r"^export\s+(?:default\s+)?(?:interface|type|class|function|const|enum)\s+"
                         r"%s\b" % re.escape(name), re.MULTILINE)
        m = pat.search(src)
        if m:
            if not has_jsdoc_above(src, m.start()):
                missing.append({"name": name, "line": line})
        else:
            pat2 = re.compile(r"^export\s+%s\s*=" % re.escape(name), re.MULTILINE)
            m2 = pat2.search(src)
            if m2 and not has_jsdoc_above(src, m2.start()):
                missing.append({"name": name, "line": line})

    return {
        "file": path,
        "has_header": has_header,
        "exported_total": len(uniq),
        "exported_documented": len(uniq) - len(missing),
        "exported_missing": missing,
    }


# ---------------------------------------------------------------------------
# 主流程
# ---------------------------------------------------------------------------

def collect_files(root_dir, exts, skip_dirs):
    out = []
    for dirpath, dirnames, filenames in os.walk(root_dir):
        dirnames[:] = [d for d in dirnames if d not in skip_dirs]
        for fn in filenames:
            if fn.endswith(exts):
                out.append(os.path.join(dirpath, fn))
    return sorted(out)


def main():
    ap = argparse.ArgumentParser()
    ap.add_argument("--todo", action="store_true", help="输出待办明细")
    ap.add_argument("--dir", default=None, help="只扫描指定目录")
    args = ap.parse_args()

    go_skip = {"node_modules", "dist", "vendor", ".git"}
    ts_skip = {"node_modules", "dist", ".git"}

    base = os.path.join(ROOT, args.dir) if args.dir else ROOT
    go_files = []
    ts_files = []
    if args.dir:
        if any(args.dir.startswith(p) for p in ("cmd", "internal", "pkg", "scripts")):
            go_files = collect_files(base, (".go",), go_skip)
        else:
            ts_files = collect_files(base, (".ts", ".vue", ".tsx"), ts_skip)
    else:
        for d in ("cmd", "internal", "pkg", "scripts"):
            go_files += collect_files(os.path.join(ROOT, d), (".go",), go_skip)
        ts_files = collect_files(os.path.join(ROOT, "web", "src"), (".ts", ".vue", ".tsx"), ts_skip)

    results = []
    for f in go_files:
        results.append(("go", analyze_go_file(f)))
    for f in ts_files:
        results.append(("ts", analyze_ts_file(f)))

    # ---- 汇总 ----
    go_total_ex = sum(r["exported_total"] for k, r in results if k == "go")
    go_doc_ex = sum(r["exported_documented"] for k, r in results if k == "go")
    go_pkg_total = sum(1 for k, r in results if k == "go")
    go_pkg_doc = sum(1 for k, r in results if k == "go" and r["has_pkg_doc"])

    ts_total = sum(r["exported_total"] for k, r in results if k == "ts")
    ts_doc = sum(r["exported_documented"] for k, r in results if k == "ts")
    ts_file_total = sum(1 for k, r in results if k == "ts")
    ts_file_doc = sum(1 for k, r in results if k == "ts" and r["has_header"])

    print("=" * 70)
    print("Ydsz Plane 注释覆盖率报告")
    print("=" * 70)
    print(f"\n[Go] 文件数: {go_pkg_total} | 包注释: {go_pkg_doc}/{go_pkg_total} "
          f"({go_pkg_doc/go_pkg_total*100:.0f}%)" if go_pkg_total else "\n[Go] 无文件")
    if go_total_ex:
        print(f"[Go] 导出符号注释: {go_doc_ex}/{go_total_ex} ({go_doc_ex/go_total_ex*100:.1f}%)")
    print(f"\n[TS/Vue] 文件数: {ts_file_total} | 文件头: {ts_file_doc}/{ts_file_total} "
          f"({ts_file_doc/ts_file_total*100:.0f}%)" if ts_file_total else "\n[TS/Vue] 无文件")
    if ts_total:
        print(f"[TS/Vue] 导出成员 JSDoc: {ts_doc}/{ts_total} ({ts_doc/ts_total*100:.1f}%)")

    if not args.todo:
        return

    print("\n" + "-" * 70)
    print("待办明细")
    print("-" * 70)
    for kind, r in results:
        if kind == "go":
            issues = []
            if not r["has_pkg_doc"]:
                issues.append("缺包注释")
            for m in r["exported_missing"]:
                issues.append(f"L{m['line']} {m['kind']} {m['name']} 缺注释")
            if issues:
                print(f"\n{r['file']}")
                for i in issues:
                    print(f"  - {i}")
        else:
            issues = []
            if not r["has_header"]:
                issues.append("缺文件头注释")
            for m in r["exported_missing"]:
                issues.append(f"L{m['line']} 导出 {m['name']} 缺 JSDoc")
            if issues:
                print(f"\n{r['file']}")
                for i in issues:
                    print(f"  - {i}")


if __name__ == "__main__":
    sys.exit(main())
