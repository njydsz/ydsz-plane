#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""为 web/src/views 下 Vue 组件插入组件级 JSDoc（一次性工具）。"""
import io
import os

VIEWS = os.path.join(os.path.dirname(os.path.abspath(__file__)), "..", "..", "web", "src", "views")

# 相对路径（相对于 views 目录）→ 组件描述
DOCS = {
    "auth/LoginView.vue": "登录页 — 邮箱+密码登录，支持限流错误提示与 redirect 回跳。",
    "auth/RegisterView.vue": "注册页 — 创建新账号，成功后自动登录并跳转工作台。",
    "auth/ForgotPasswordView.vue": "忘记密码页 — 提交邮箱触发密码重置邮件。",
    "auth/ResetPasswordView.vue": "重置密码页 — 通过邮件中的 token 设置新密码。",
    "HomeView.vue": "首页 — 简单欢迎页（预留为工作台入口）。",
    "NotFoundView.vue": "404 页面 — 未知路由兜底展示。",
    "project/ProjectListView.vue": "项目列表页 — 展示当前工作空间下的全部项目，支持新建项目。",
    "project/KanbanBoardView.vue": "看板视图 — 按状态分列展示工作项，支持创建/详情跳转。",
    "project/IssueListView.vue": "工作项列表页 — 表格视图展示工作项，支持筛选与分页。",
    "project/IssueDetailView.vue": "工作项详情页 — 展示描述、状态流转、活动日志与工时记录。",
    "project/IssueCreateModal.vue": "创建工作项弹窗 — 类型/标题/优先级/经办人/标签等表单。",
    "project/BurndownChart.vue": "燃尽图组件 — 渲染迭代燃尽曲线（实际 vs 理想线）。",
    "project/DefectPanel.vue": "缺陷面板 — 展示版本日的缺陷列表与修复状态。",
    "project/SprintListView.vue": "迭代列表页 — 展示全部迭代，支持创建/归档/进入详情。",
    "project/SprintPlanningView.vue": "排期规划页 — Backlog 与迭代间拖拽分配工作项。",
    "project/SprintDetailView.vue": "迭代详情页 — 展示进度、工作项列表、燃尽图与复盘。",
    "project/SprintStandupView.vue": "站会模式页 — 按成员/状态聚合展示迭代待办，辅助每日站会。",
    "project/VersionListView.vue": "版本日列表页 — 展示版本日列表，支持创建与状态流转。",
    "project/VersionDetailView.vue": "版本日详情页 — 展示进度、质量门禁、缺陷面板与迭代聚合。",
    "project/VersionReleaseView.vue": "版本日发布页 — 发布前检查清单、发布说明生成与确认发布。",
    "project/DeliveryReportView.vue": "交付报告页 — 展示版本日的交付统计与准出资格。",
    "workspace/WorkspaceListView.vue": "工作空间列表页 — 展示/创建工作空间，进入项目列表。",
    "workspace/CreateWorkspaceModal.vue": "创建工作空间弹窗 — 名称/slug/时区/语言表单。",
    "workspace/InvitePreview.vue": "邀请预览页 — 公开展示邀请信息并接受邀请。",
    "workspace/WorkspaceSettingsView.vue": "工作空间设置页 — 成员管理、邀请、角色变更与空间信息编辑。",
}


def main():
    for rel, desc in DOCS.items():
        path = os.path.join(VIEWS, rel)
        if not os.path.exists(path):
            print(f"[warn] 不存在: {rel}")
            continue
        with io.open(path, encoding="utf-8") as f:
            src = f.read()
        marker = "<script setup lang=\"ts\">"
        if marker not in src:
            print(f"[warn] 未找到 script 标记: {rel}")
            continue
        doc = f"/**\n * {desc}\n */\n"
        # 若紧随 script 标记后已有 /** 注释则跳过
        after = src[src.find(marker) + len(marker):]
        if after.lstrip().startswith("/**"):
            print(f"[skip] 已有注释: {rel}")
            continue
        src = src.replace(marker, marker + "\n" + doc, 1)
        with io.open(path, "w", encoding="utf-8", newline="") as f:
            f.write(src)
        print(f"[ok] {rel}")


if __name__ == "__main__":
    main()
