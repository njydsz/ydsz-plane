#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""为 web/src/api/services 下 5 个 API 文件插入类型 JSDoc 注释（一次性工具）。"""
import io
import os

BASE = os.path.join(os.path.dirname(os.path.abspath(__file__)), "..", "..", "web", "src", "api", "services")

# 每个文件：{ 类型名: 描述 }
DOCS = {
    "auth.ts": {
        "TokenPair": "/** 登录/刷新接口返回的令牌对（与后端 auth.TokenPair 对齐） */",
        "LoginInput": "/** 登录请求参数 */",
        "RegisterInput": "/** 注册请求参数 */",
        "authApi": "/** 认证域 API：登录 / 注册 / 刷新 / 当前用户 / 密码找回与重置 */",
    },
    "issue.ts": {
        "IssueType": "/** 工作项类型：需求 / 任务 / 缺陷 */",
        "IssuePriority": "/** 工作项优先级（降序） */",
        "StateGroup": "/** 状态分组：backlog / started / completed / cancelled */",
        "State": "/** 工作项状态定义（含所属分组与展示色） */",
        "Issue": "/** 工作项（需求/任务/缺陷统一模型），与后端 issue.Issue 对齐 */",
        "IssueActivity": "/** 工作项活动日志条目（谁在何时改了什么字段） */",
        "TimeLog": "/** 工作项工时记录（分钟粒度） */",
        "IssueRelation": "/** 工作项关联关系（如关联/被关联） */",
        "IssueDependency": "/** 工作项依赖关系（前置/后继 + 滞后天数） */",
        "CreateIssueInput": "/** 创建工作项入参 */",
        "UpdateIssueInput": "/** 更新工作项入参（可选字段 + 乐观锁 version） */",
        "ListIssuesParams": "/** 工作项列表查询参数（过滤/搜索/分页） */",
        "issueApi": "/** 工作项域 API：状态 / CRUD / 流转 / 活动 / 工时 / 关联 / 依赖 */",
    },
    "sprint.ts": {
        "SprintStatus": "/** 迭代状态：planned / active / completed */",
        "UnfinishedStrategy": "/** 迭代结束时未完成工作项的处理策略 */",
        "Sprint": "/** 迭代聚合（含可选进度与复盘快照） */",
        "SprintProgress": "/** 迭代实时进度汇总（点数与工作项数） */",
        "ReviewSnapshot": "/** 迭代复盘快照（承诺/完成/加入/移除统计） */",
        "SprintSnapshot": "/** 迭代每日快照记录（燃尽图数据源） */",
        "SnapshotData": "/** 单日快照内容 */",
        "BurndownPoint": "/** 燃尽图单日数据点（含理想线） */",
        "VelocityStats": "/** 迭代速率统计（平均值/中位数/近期迭代） */",
        "SprintVelocity": "/** 单个迭代的速率数据 */",
        "SprintIssueView": "/** 迭代内工作项的视图投影 */",
        "BacklogItem": "/** Backlog 工作项条目（未规划进 active 迭代） */",
        "CreateSprintInput": "/** 创建迭代入参 */",
        "UpdateSprintInput": "/** 更新迭代入参（仅 planned 状态可编辑） */",
        "CompleteSprintInput": "/** 结束迭代入参（未完成工作项处理策略） */",
        "ListSprintsParams": "/** 迭代列表查询参数 */",
        "sprintApi": "/** 迭代域 API：CRUD / 生命周期 / 进度 / 规划 / 燃尽图 / 复盘 / 速率建议 */",
    },
    "version.ts": {
        "VersionStatus": "/** 版本日状态：planning / active / released / archived */",
        "ChecklistItem": "/** 发布检查清单条目 */",
        "SprintProgressRef": "/** 迭代进度摘要（版本日聚合视图用） */",
        "SprintRef": "/** 版本日关联的迭代摘要 */",
        "VersionProgress": "/** 版本日实时进度聚合 */",
        "QualityMetrics": "/** 版本日质量指标（发布准出校验用） */",
        "DeliveryReport": "/** 版本日交付报告（发布前快照） */",
        "Version": "/** 版本日聚合根 */",
        "BugVersionView": "/** 缺陷面板中的缺陷视图投影 */",
        "CreateVersionInput": "/** 创建版本日入参 */",
        "UpdateVersionInput": "/** 更新版本日入参（可选字段 + 乐观锁 version） */",
        "ReleaseVersionInput": "/** 发布版本日入参（草稿覆盖 / 强制清单 / 已知缺陷写入发布说明） */",
        "AddSprintInput": "/** 将迭代关联到版本日的入参 */",
        "ListVersionsParams": "/** 版本日列表查询参数 */",
        "versionApi": "/** 版本日域 API：CRUD / 生命周期 / 进度质量 / 交付报告 / 缺陷面板 / 迭代聚合 */",
    },
    "workspace.ts": {
        "Workspace": "/** 工作空间（含当前用户角色与成员数可选字段） */",
        "Member": "/** 工作空间成员 */",
        "Invitation": "/** 工作空间邀请记录 */",
        "InvitationPreview": "/** 邀请预览（接受前展示，无鉴权） */",
        "Project": "/** 项目（工作空间下二级聚合根） */",
        "workspaceApi": "/** 工作空间域 API：空间 / 成员 / 邀请 / 项目 CRUD */",
    },
}


def insert_doc(src: str, marker: str, doc: str) -> str:
    """在 'export <marker>...' 声明前插入 doc（若该位置上方已存在 /** */ 则跳过）。"""
    idx = src.find(marker)
    if idx == -1:
        return src
    # 检查紧邻上方是否已有 JSDoc
    before = src[:idx].rstrip()
    if before.endswith("*/"):
        return src
    return src[:idx] + doc + "\n" + src[idx:]


def main():
    for fn, docs in DOCS.items():
        path = os.path.join(BASE, fn)
        with io.open(path, encoding="utf-8") as f:
            src = f.read()
        for name, doc in docs.items():
            # export interface X / export type X / export const X
            for kw in ("interface", "type", "const"):
                marker = f"export {kw} {name}"
                if marker in src:
                    src = insert_doc(src, marker, doc)
                    break
            else:
                # export const name = 已处理；否则报告
                if f"export const {name}" not in src:
                    print(f"[warn] {fn}: 未找到 {name}")
        with io.open(path, "w", encoding="utf-8", newline="") as f:
            f.write(src)
        print(f"[ok] {fn}")


if __name__ == "__main__":
    main()
