# **<font style="color:rgb(31,78,121);">1. 概述</font>**
本文档是Ydsz Plane 项目管理工具的最终完整产品需求文档 (PRD V1.0)，融合了产品架构框架、需求/任务/缺陷属性竞品对比、以及效率增强功能的详细需求描述。文档对标云效、TAPD、ONES、PingCode、Jira 等主流产品，结合 Ydsz Plane 项目现有代码实现，形成从顶层架构到字段级规格再到功能级交互说明的完整需求层次。

在产品经理（提需求）、技术经理（提任务）、测试经理（提缺陷）的典型团队协作模式下，需求、任务、缺陷三类实体通过 WBS 层级结构化解耦关联，实现从宏观规划到微观执行的完整追踪。

## **<font style="color:rgb(46,117,182);">1.1 产品定位</font>**
Ydsz Plane 致力于打造面向中国软件开发团队的开源、自托管项目管理平台，深度融合云效/TAPD/ONES 在需求管理、任务管理、缺陷管理、迭代协同方面的本土化实践，补齐项目仪表盘、个人工作台等效率增强能力，成为真正符合国内软件研发现状的「开源 PM SaaS 底座」。

Ydsz Plane 是一款开源的现代项目管理工具，提供需求池、迭代、缺陷、模块、版本等核心能力。定位为 Jira 的轻量级替代品，面向中小型敏捷开发团队。

## **<font style="color:rgb(46,117,182);">1.2 核心架构</font>**
工作空间 (Workspace) ⟶ 项目 (Project) ⟶ 需求 (Requirement) / 任务 (Task) / 缺陷 (Defect)

• 需求、任务、缺陷为三类独立的业务实体，各自独立建表、独立追踪（不存在统一的需求/任务/缺陷单表模型）

• 三类实体通过 parent 字段实现 WBS 层级：需求→子需求，任务→子任务，缺陷→子缺陷

• 版本 (Version) = 1~N 迭代 (Sprint) 的里程碑容器

• 模块 (Module) 是需求/任务/缺陷的归档属性，非独立管理对象

# **<font style="color:rgb(31,78,121);">2. 竞品对标</font>**
下表对标一线竞品。● 已支持 / △ 部分 / ○ 不支持 / ✕ 不适用

| **<font style="color:rgb(255,255,255);">能力域</font>** | **<font style="color:rgb(255,255,255);">云效</font>** | **<font style="color:rgb(255,255,255);">TAPD</font>** | **<font style="color:rgb(255,255,255);">ONES</font>** | **<font style="color:rgb(255,255,255);">PingCode</font>** | **<font style="color:rgb(255,255,255);">Jira</font>** | **<font style="color:rgb(255,255,255);">Ydsz Plane</font>** |
| --- | --- | --- | --- | --- | --- | --- |
| <font style="color:rgb(51,51,51);">需求管理</font><font style="color:rgb(51,51,51);"> (含子需求)</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> |
| <font style="color:rgb(51,51,51);">任务管理</font><font style="color:rgb(51,51,51);"> (含子任务)</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> |
| <font style="color:rgb(51,51,51);">缺陷管理</font><font style="color:rgb(51,51,51);"> (含子缺陷)</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> |
| <font style="color:rgb(51,51,51);">模块</font><font style="color:rgb(51,51,51);"> (=需求/任务/缺陷归档属性)</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">△ 修正中</font> |
| <font style="color:rgb(51,51,51);">迭代</font><font style="color:rgb(51,51,51);"> (Sprint)</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> |
| <font style="color:rgb(51,51,51);">版本</font><font style="color:rgb(51,51,51);"> (=1~N 迭代)</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">○</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">△ 规划中</font> |
| <font style="color:rgb(51,51,51);">项目仪表盘</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">△ 增强中</font> |
| <font style="color:rgb(51,51,51);">个人工作台</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">△ 增强中</font> |
| <font style="color:rgb(51,51,51);">收件箱</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> |
| <font style="color:rgb(51,51,51);">效能度量</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">○</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">○</font> | <font style="color:rgb(51,51,51);">△ 增强中</font> |
| <font style="color:rgb(51,51,51);">开源自托管</font><font style="color:rgb(51,51,51);"> / 信创</font> | <font style="color:rgb(51,51,51);">✕</font> | <font style="color:rgb(51,51,51);">✕</font> | <font style="color:rgb(51,51,51);">△</font> | <font style="color:rgb(51,51,51);">△</font> | <font style="color:rgb(51,51,51);">✕</font> | <font style="color:rgb(51,51,51);">● 核心差异</font> |


# **<font style="color:rgb(31,78,121);">3. 项目结构与核心模型</font>**
## **<font style="color:rgb(46,117,182);">3.1 系统层级 (Workspace > Project)</font>**
| **<font style="color:rgb(255,255,255);">层</font>** | **<font style="color:rgb(255,255,255);">对象</font>** | **<font style="color:rgb(255,255,255);">类型</font>** | **<font style="color:rgb(255,255,255);">说明</font>** | **<font style="color:rgb(255,255,255);">主责角色</font>** |
| --- | --- | --- | --- | --- |
| <font style="color:rgb(51,51,51);">L0</font> | <font style="color:rgb(51,51,51);">工作空间</font> | <font style="color:rgb(51,51,51);">组织</font> | <font style="color:rgb(51,51,51);">企业</font><font style="color:rgb(51,51,51);">/团队级容器，隔离数据与成员</font> | <font style="color:rgb(51,51,51);">IT 管理员</font> |
| <font style="color:rgb(51,51,51);">L1</font> | <font style="color:rgb(51,51,51);">项目</font> | <font style="color:rgb(51,51,51);">项目</font> | <font style="color:rgb(51,51,51);">具体产品的研发或交付项目</font> | <font style="color:rgb(51,51,51);">项目经理</font><font style="color:rgb(51,51,51);"> / Scrum Master</font> |
| <font style="color:rgb(51,51,51);">L2a</font> | <font style="color:rgb(51,51,51);">版本</font> | <font style="color:rgb(51,51,51);">里程碑</font> | <font style="color:rgb(51,51,51);">产品发版里程碑，含</font><font style="color:rgb(51,51,51);"> 1~N 个迭代</font> | <font style="color:rgb(51,51,51);">产品经理</font><font style="color:rgb(51,51,51);"> / 发布经理</font> |
| <font style="color:rgb(51,51,51);">L2b</font> | <font style="color:rgb(51,51,51);">迭代</font> | <font style="color:rgb(51,51,51);">时间盒</font> | <font style="color:rgb(51,51,51);">固定周期开发单元</font><font style="color:rgb(51,51,51);"> (Sprint)</font> | <font style="color:rgb(51,51,51);">Scrum Master / Tech Lead</font> |
| <font style="color:rgb(51,51,51);">L3</font> | <font style="color:rgb(51,51,51);">需求</font> | <font style="color:rgb(51,51,51);">需求/任务/缺陷</font> | <font style="color:rgb(51,51,51);">产品经理提出，可拆分子需求</font> | <font style="color:rgb(51,51,51);">产品经理</font> |
| <font style="color:rgb(51,51,51);">L3</font> | <font style="color:rgb(51,51,51);">任务</font> | <font style="color:rgb(51,51,51);">需求/任务/缺陷</font> | <font style="color:rgb(51,51,51);">技术经理提出，可拆分子任务</font> | <font style="color:rgb(51,51,51);">技术经理</font> |
| <font style="color:rgb(51,51,51);">L3</font> | <font style="color:rgb(51,51,51);">缺陷</font> | <font style="color:rgb(51,51,51);">需求/任务/缺陷</font> | <font style="color:rgb(51,51,51);">由需求</font><font style="color:rgb(51,51,51);">/任务产生，可拆分子缺陷</font> | <font style="color:rgb(51,51,51);">测试经理</font><font style="color:rgb(51,51,51);"> / QA</font> |
| <font style="color:rgb(51,51,51);">属性</font> | <font style="color:rgb(51,51,51);">模块</font> | <font style="color:rgb(51,51,51);">归档属性</font> | <font style="color:rgb(51,51,51);">需求</font><font style="color:rgb(51,51,51);">/任务/缺陷的分类维度</font> | <font style="color:rgb(51,51,51);">产品负责人</font><font style="color:rgb(51,51,51);"> / 架构师</font> |


## **<font style="color:rgb(46,117,182);">3.2 版本-迭代关系模型</font>**
一个版本（里程碑）的达成，可能涉及 1 个或者多个迭代周期。Version 是面向市场的发版节奏，迭代是面向团队的开发节奏。

• 版本 v2.6（6月30日发版）包含 Sprint-12（5月）、Sprint-13（6月上）、Sprint-14（6月下）

• 版本 v2.7（9月30日发版）包含 Sprint-15~Sprint-18 共 4 个迭代

• 版本具有独立的交付目标、准出标准、检查清单和 Release Notes

• 迭代是版本的组成单元，迭代完成度汇总到版本进度

## **<font style="color:rgb(46,117,182);">3.3 需求-任务-缺陷联动模型（核心）</font>**
核心原则：

• 需求由产品经理提出，可拆解为多个子需求（需求≠ 任务的来源）

• 任务由技术经理提出，可拆分为多个子任务

• 需求和任务都有可能产生缺陷（不是需求→任务的线性关系）

• 缺陷由测试人员提交，可关联到触发它的需求或任务

| **<font style="color:rgb(255,255,255);">关系</font>** | **<font style="color:rgb(255,255,255);">说明</font>** | **<font style="color:rgb(255,255,255);">创建者</font>** | **<font style="color:rgb(255,255,255);">触发条件</font>** |
| --- | --- | --- | --- |
| <font style="color:rgb(51,51,51);">需求</font><font style="color:rgb(51,51,51);">→ 子需求</font> | <font style="color:rgb(51,51,51);">需求可拆分为多个子需求（</font><font style="color:rgb(51,51,51);">Epic→Feature→Story）</font> | <font style="color:rgb(51,51,51);">产品经理</font> | <font style="color:rgb(51,51,51);">需求评审、范围拆解</font> |
| <font style="color:rgb(51,51,51);">任务</font><font style="color:rgb(51,51,51);">→ 子任务</font> | <font style="color:rgb(51,51,51);">任务可拆分为多个子任务（</font><font style="color:rgb(51,51,51);">WBS 工作分解结构）</font> | <font style="color:rgb(51,51,51);">技术经理</font> | <font style="color:rgb(51,51,51);">迭代规划、技术方案拆解</font> |
| <font style="color:rgb(51,51,51);">需求</font><font style="color:rgb(51,51,51);">→ 缺陷</font> | <font style="color:rgb(51,51,51);">测试需求时发现缺陷，缺陷关联到该需求</font> | <font style="color:rgb(51,51,51);">测试经理</font> | <font style="color:rgb(51,51,51);">需求准出验证、系统测试</font> |
| <font style="color:rgb(51,51,51);">任务</font><font style="color:rgb(51,51,51);">→ 缺陷</font> | <font style="color:rgb(51,51,51);">开发</font><font style="color:rgb(51,51,51);">/测试任务过程中发现缺陷</font> | <font style="color:rgb(51,51,51);">开发</font><font style="color:rgb(51,51,51);">/测试</font> | <font style="color:rgb(51,51,51);">单元测试、联调、回归测试</font> |
| <font style="color:rgb(51,51,51);">需求</font><font style="color:rgb(51,51,51);">↔</font><font style="color:rgb(51,51,51);"> 任务</font> | <font style="color:rgb(51,51,51);">关联关系（非从属）：任务实现需求，需求指导任务</font> | <font style="color:rgb(51,51,51);">—</font> | <font style="color:rgb(51,51,51);">规划阶段人工关联</font> |




**关键规则：**

• ❌ 需求拆解不产生任务 — 需求 → 子需求，任务由技术经理独立创建

• ✅ 需求和任务都可能产生缺陷 — 需求测试发现缺陷、任务开发测试发现缺陷

• ✅ 缺陷可关联到需求或任务 — 追溯缺陷来源，评估交付质量

• ✅ 三级 WBS 层级 — 子需求/子任务/子缺陷均支持三级层级

## **<font style="color:rgb(46,117,182);">3.4 模块 = 需求/任务/缺陷归档属性</font>**
模块不是单独的对象，而是需求、任务、缺陷的一个归档属性（标签/分类维度）。对标 ONES 模块、云效模块。

• 模块列表在工作空间/项目维度维护：支付模块、订单模块、用户模块、消息模块

• 模块作为必填或选填字段出现在需求/任务/缺陷的创建表单中

• 需求/任务/缺陷可按模块过滤、分组、统计

• 模块不独立存在，不能脱离需求/任务/缺陷被单独访问

• 模块可关联模块负责人（Owner），但不具备独立的权限体系

• 模块可标记目标交付版本

## **<font style="color:rgb(46,117,182);">3.5 WBS 三层级结构</font>**
| **<font style="color:rgb(255,255,255);">需求/任务/缺陷类型</font>** | **<font style="color:rgb(255,255,255);">第一层</font>** | **<font style="color:rgb(255,255,255);">第二层</font>** | **<font style="color:rgb(255,255,255);">第三层</font>** |
| --- | --- | --- | --- |
| <font style="color:rgb(51,51,51);">需求</font> | <font style="color:rgb(51,51,51);">Epic (史诗)</font> | <font style="color:rgb(51,51,51);">Feature (特性)</font> | <font style="color:rgb(51,51,51);">Story (用户故事)</font> |
| <font style="color:rgb(51,51,51);">任务</font> | <font style="color:rgb(51,51,51);">主任务</font> | <font style="color:rgb(51,51,51);">子任务</font> | <font style="color:rgb(51,51,51);">子子任务</font> |
| <font style="color:rgb(51,51,51);">缺陷</font> | <font style="color:rgb(51,51,51);">主缺陷</font> | <font style="color:rgb(51,51,51);">子缺陷</font> | <font style="color:rgb(51,51,51);">子子缺陷</font> |


  


# **<font style="color:rgb(31,78,121);">4. 需求/任务/缺陷属性详细对比（</font>****<font style="color:rgb(31,78,121);">Ydsz Plane</font>****<font style="color:rgb(31,78,121);"> vs 竞品）</font>**
本章从Ydsz Plane 项目代码模型出发，对比各竞品的需求/任务/缺陷属性设计。

## **<font style="color:rgb(46,117,182);">4.1 需求 (Requirement) 属性对比</font>**
需求由产品经理提出，描述产品需要实现的功能或解决的问题。需求可拆分为子需求（Epic→Feature→Story 模式）。

| **<font style="color:rgb(255,255,255);">属性名</font>** | **<font style="color:rgb(255,255,255);">Ydsz Plane</font>** | **<font style="color:rgb(255,255,255);">Jira</font>** | **<font style="color:rgb(255,255,255);">云效</font>** | **<font style="color:rgb(255,255,255);">TAPD</font>** | **<font style="color:rgb(255,255,255);">ONES/PingCode</font>** |
| --- | --- | --- | --- | --- | --- |
| <font style="color:rgb(51,51,51);">标题</font><font style="color:rgb(51,51,51);"> (Title)</font> | <font style="color:rgb(51,51,51);">name</font> | <font style="color:rgb(51,51,51);">Summary</font> | <font style="color:rgb(51,51,51);">标题</font> | <font style="color:rgb(51,51,51);">标题</font> | <font style="color:rgb(51,51,51);">标题</font> |
| <font style="color:rgb(51,51,51);">描述</font><font style="color:rgb(51,51,51);"> (Description)</font> | <font style="color:rgb(51,51,51);">description_json/html</font> | <font style="color:rgb(51,51,51);">Description</font> | <font style="color:rgb(51,51,51);">描述</font> | <font style="color:rgb(51,51,51);">描述</font> | <font style="color:rgb(51,51,51);">描述</font> |
| <font style="color:rgb(51,51,51);">需求类型</font><font style="color:rgb(51,51,51);"> (Type)</font> | <font style="color:rgb(51,51,51);">IssueType FK</font> | <font style="color:rgb(51,51,51);">Epic/Story/</font><font style="color:rgb(51,51,51);">Defect</font> | <font style="color:rgb(51,51,51);">需求类型</font> | <font style="color:rgb(51,51,51);">需求类型</font> | <font style="color:rgb(51,51,51);">需求</font><font style="color:rgb(51,51,51);">/子需求/用户故事</font> |
| <font style="color:rgb(51,51,51);">状态</font><font style="color:rgb(51,51,51);"> (State)</font> | <font style="color:rgb(51,51,51);">state FK (State 模型)</font> | <font style="color:rgb(51,51,51);">Status</font> | <font style="color:rgb(51,51,51);">状态</font> | <font style="color:rgb(51,51,51);">状态</font> | <font style="color:rgb(51,51,51);">状态</font> |
| <font style="color:rgb(51,51,51);">优先级</font><font style="color:rgb(51,51,51);"> (Priority)</font> | <font style="color:rgb(51,51,51);">urgent/high/medium/low/none (5级)</font> | <font style="color:rgb(51,51,51);">Highest/High/Medium/Low/Lowest (5级)</font> | <font style="color:rgb(51,51,51);">紧急</font><font style="color:rgb(51,51,51);">/高/中/低</font> | <font style="color:rgb(51,51,51);">紧急</font><font style="color:rgb(51,51,51);">/高/中/低/无关紧要</font> | <font style="color:rgb(51,51,51);">紧急</font><font style="color:rgb(51,51,51);">/高/中/低 (5级)</font> |
| <font style="color:rgb(51,51,51);">严重程度</font><font style="color:rgb(51,51,51);"> (Severity)</font> | <font style="color:rgb(51,51,51);">— (需补充)</font> | <font style="color:rgb(51,51,51);">S1/S2/S3/S4/S5 (5级)</font> | <font style="color:rgb(51,51,51);">建议补充</font> | <font style="color:rgb(51,51,51);">建议补充</font> | <font style="color:rgb(51,51,51);">致命</font><font style="color:rgb(51,51,51);">/严重/一般/轻微/建议</font> |
| <font style="color:rgb(51,51,51);">负责人</font><font style="color:rgb(51,51,51);"> (Assignee)</font> | <font style="color:rgb(51,51,51);">assignees (M2M)</font> | <font style="color:rgb(51,51,51);">Assignee</font> | <font style="color:rgb(51,51,51);">负责人</font> | <font style="color:rgb(51,51,51);">负责人</font> | <font style="color:rgb(51,51,51);">负责人</font> |
| <font style="color:rgb(51,51,51);">创建人</font><font style="color:rgb(51,51,51);"> (Reporter)</font> | <font style="color:rgb(51,51,51);">created_by</font> | <font style="color:rgb(51,51,51);">Reporter</font> | <font style="color:rgb(51,51,51);">创建人</font> | <font style="color:rgb(51,51,51);">创建人</font> | <font style="color:rgb(51,51,51);">创建人</font> |
| <font style="color:rgb(51,51,51);">参与人</font><font style="color:rgb(51,51,51);"> (Participants)</font> | <font style="color:rgb(51,51,51);">— (建议补充)</font> | <font style="color:rgb(51,51,51);">Watchers</font> | <font style="color:rgb(51,51,51);">参与人</font> | <font style="color:rgb(51,51,51);">参与人</font> | <font style="color:rgb(51,51,51);">参与人</font> |
| <font style="color:rgb(51,51,51);">抄送人</font><font style="color:rgb(51,51,51);"> (CC)</font> | <font style="color:rgb(51,51,51);">— (建议补充)</font> | <font style="color:rgb(51,51,51);">—</font> | <font style="color:rgb(51,51,51);">抄送人</font> | <font style="color:rgb(51,51,51);">—</font> | <font style="color:rgb(51,51,51);">—</font> |
| <font style="color:rgb(51,51,51);">开始日期</font><font style="color:rgb(51,51,51);"> (Start Date)</font> | <font style="color:rgb(51,51,51);">start_date</font> | <font style="color:rgb(51,51,51);">— (需配置)</font> | <font style="color:rgb(51,51,51);">开始时间</font> | <font style="color:rgb(51,51,51);">开始日期</font> | <font style="color:rgb(51,51,51);">开始日期</font> |
| <font style="color:rgb(51,51,51);">截止日期</font><font style="color:rgb(51,51,51);"> (Due Date)</font> | <font style="color:rgb(51,51,51);">target_date</font> | <font style="color:rgb(51,51,51);">Due Date</font> | <font style="color:rgb(51,51,51);">截止时间</font> | <font style="color:rgb(51,51,51);">截止日期</font> | <font style="color:rgb(51,51,51);">截止日期</font> |
| <font style="color:rgb(51,51,51);">迭代</font><font style="color:rgb(51,51,51);"> (</font><font style="color:rgb(51,51,51);">Sprint</font><font style="color:rgb(51,51,51);">)</font> | <font style="color:rgb(51,51,51);">s</font><font style="color:rgb(51,51,51);">print</font><font style="color:rgb(51,51,51);"> </font><font style="color:rgb(51,51,51);">(关联)</font> | <font style="color:rgb(51,51,51);">Sprint</font> | <font style="color:rgb(51,51,51);">迭代</font> | <font style="color:rgb(51,51,51);">迭代</font> | <font style="color:rgb(51,51,51);">迭代</font> |
| <font style="color:rgb(51,51,51);">版本</font><font style="color:rgb(51,51,51);"> (Version)</font> | <font style="color:rgb(51,51,51);">— (建议补充)</font> | <font style="color:rgb(51,51,51);">Fix Version</font> | <font style="color:rgb(51,51,51);">版本</font> | <font style="color:rgb(51,51,51);">版本</font> | <font style="color:rgb(51,51,51);">版本</font> |
| <font style="color:rgb(51,51,51);">模块</font><font style="color:rgb(51,51,51);"> (Module)</font> | <font style="color:rgb(51,51,51);">module_issues M2M (归档属性)</font> | <font style="color:rgb(51,51,51);">Components</font> | <font style="color:rgb(51,51,51);">模块</font><font style="color:rgb(51,51,51);">/分类</font> | <font style="color:rgb(51,51,51);">模块</font> | <font style="color:rgb(51,51,51);">模块</font> |
| <font style="color:rgb(51,51,51);">标签</font><font style="color:rgb(51,51,51);"> (Labels)</font> | <font style="color:rgb(51,51,51);">labels M2M</font> | <font style="color:rgb(51,51,51);">Labels</font> | <font style="color:rgb(51,51,51);">标签</font> | <font style="color:rgb(51,51,51);">标签</font> | <font style="color:rgb(51,51,51);">标签</font> |
| <font style="color:rgb(51,51,51);">父需求</font><font style="color:rgb(51,51,51);"> (Parent)</font> | <font style="color:rgb(51,51,51);">parent FK (WBS)</font> | <font style="color:rgb(51,51,51);">Epic Link</font> | <font style="color:rgb(51,51,51);">父需求</font> | <font style="color:rgb(51,51,51);">父需求</font> | <font style="color:rgb(51,51,51);">父需求</font> |
| <font style="color:rgb(51,51,51);">子需求</font><font style="color:rgb(51,51,51);"> (Sub-issues)</font> | <font style="color:rgb(51,51,51);">Sub-issue count</font> | <font style="color:rgb(51,51,51);">Sub-task count</font> | <font style="color:rgb(51,51,51);">子需求树</font> | <font style="color:rgb(51,51,51);">子需求树</font> | <font style="color:rgb(51,51,51);">子需求</font><font style="color:rgb(51,51,51);">/子任务/孙任务</font> |
| <font style="color:rgb(51,51,51);">故事点</font><font style="color:rgb(51,51,51);"> (Story Points)</font> | <font style="color:rgb(51,51,51);">point (0-12)</font> | <font style="color:rgb(51,51,51);">Story Points</font> | <font style="color:rgb(51,51,51);">规模</font> | <font style="color:rgb(51,51,51);">规模</font> | <font style="color:rgb(51,51,51);">故事点数</font> |
| <font style="color:rgb(51,51,51);">估算工时</font><font style="color:rgb(51,51,51);"> (Estimate)</font> | <font style="color:rgb(51,51,51);">estimate_point FK</font> | <font style="color:rgb(51,51,51);">Estimation</font> | <font style="color:rgb(51,51,51);">预计工作量</font> | <font style="color:rgb(51,51,51);">预计工作量</font> | <font style="color:rgb(51,51,51);">预计工作量</font> |
| <font style="color:rgb(51,51,51);">完成时间</font><font style="color:rgb(51,51,51);"> (Completed)</font> | <font style="color:rgb(51,51,51);">completed_at</font> | <font style="color:rgb(51,51,51);">Resolved</font> | <font style="color:rgb(51,51,51);">完成时间</font> | <font style="color:rgb(51,51,51);">关闭时间</font> | <font style="color:rgb(51,51,51);">完成时间</font> |
| <font style="color:rgb(51,51,51);">附件</font><font style="color:rgb(51,51,51);"> (Attachments)</font> | <font style="color:rgb(51,51,51);">IssueAttachment</font> | <font style="color:rgb(51,51,51);">Attachment</font> | <font style="color:rgb(51,51,51);">附件</font> | <font style="color:rgb(51,51,51);">附件</font> | <font style="color:rgb(51,51,51);">附件</font> |
| <font style="color:rgb(51,51,51);">关联需求/任务/缺陷</font><font style="color:rgb(51,51,51);"> (Relations)</font> | <font style="color:rgb(51,51,51);">IssueRelation (6种)</font> | <font style="color:rgb(51,51,51);">Linked Issues</font> | <font style="color:rgb(51,51,51);">关联</font> | <font style="color:rgb(51,51,51);">关联需求</font><font style="color:rgb(51,51,51);">/缺陷</font> | <font style="color:rgb(51,51,51);">关联需求</font><font style="color:rgb(51,51,51);">/缺陷</font> |
| <font style="color:rgb(51,51,51);">来源</font><font style="color:rgb(51,51,51);"> (Source)</font> | <font style="color:rgb(51,51,51);">— (建议补充)</font> | <font style="color:rgb(51,51,51);">Custom Field</font> | <font style="color:rgb(51,51,51);">客户反馈</font><font style="color:rgb(51,51,51);">/内部</font> | <font style="color:rgb(51,51,51);">来源</font> | <font style="color:rgb(51,51,51);">来源</font> |
| <font style="color:rgb(51,51,51);">进度</font><font style="color:rgb(51,51,51);"> (%)</font> | <font style="color:rgb(51,51,51);">自动</font><font style="color:rgb(51,51,51);"> (子需求完成率)</font> | <font style="color:rgb(51,51,51);">自动</font> | <font style="color:rgb(51,51,51);">进度</font> | <font style="color:rgb(51,51,51);">进度</font> | <font style="color:rgb(51,51,51);">进度</font> |
| <font style="color:rgb(51,51,51);">是否草稿</font> | <font style="color:rgb(51,51,51);">is_draft</font> | <font style="color:rgb(51,51,51);">—</font> | <font style="color:rgb(51,51,51);">草稿</font> | <font style="color:rgb(51,51,51);">草稿</font> | <font style="color:rgb(51,51,51);">草稿</font> |




注：Ydsz Plane 的模块定位为归档属性而非独立对象，因此使用 M2M 关联而非外键。

### **<font style="color:rgb(51,51,51);">竞品属性详细分析</font>**
• Jira: 支持完整的 Epic Hierarchy（Epic > Story > Sub-task），自定义字段极其丰富，支持 Labels、Components、Versions、Sprints 多维度组织。

• 云效: 强调需求全生命周期管理，支持需求关联测试用例、代码、设计文档，独有的「参与人+抄送人」机制。

• ONES/PingCode: 提供需求来源字段（客户反馈/线上/内部），支持需求进度自动计算和父子需求关联的完整追溯链。

### **<font style="color:rgb(51,51,51);">Ydsz Plane</font>****<font style="color:rgb(51,51,51);"> </font>****<font style="color:rgb(51,51,51);">需补充的属性建议</font>**
1. 严重程度 (Severity)：1~5 级，用于区分业务影响（致命/严重/一般/轻微/建议）

2. 参与人/抄送人：支持多人关注，接收通知

3. 需求来源：区分客户反馈、内部需求、竞品需求等

4. 修复/影响版本：关联版本，便于追溯需求在哪个版本发布

5. 进度自动计算：基于子需求完成状态自动更新

## **<font style="color:rgb(46,117,182);">4.2 任务 (Task) 属性对比</font>**
任务由技术经理提出，描述研发侧需要完成的具体任务。任务可拆分为子任务。

| **<font style="color:rgb(255,255,255);">属性名</font>** | **<font style="color:rgb(255,255,255);">Ydsz Plane</font>** | **<font style="color:rgb(255,255,255);">Jira</font>** | **<font style="color:rgb(255,255,255);">云效</font>** | **<font style="color:rgb(255,255,255);">TAPD</font>** | **<font style="color:rgb(255,255,255);">ONES/PingCode</font>** |
| --- | --- | --- | --- | --- | --- |
| <font style="color:rgb(51,51,51);">标题</font><font style="color:rgb(51,51,51);"> (Title)</font> | <font style="color:rgb(51,51,51);">name</font> | <font style="color:rgb(51,51,51);">Summary</font> | <font style="color:rgb(51,51,51);">标题</font> | <font style="color:rgb(51,51,51);">标题</font> | <font style="color:rgb(51,51,51);">标题</font> |
| <font style="color:rgb(51,51,51);">描述</font><font style="color:rgb(51,51,51);"> (Description)</font> | <font style="color:rgb(51,51,51);">description_json/html</font> | <font style="color:rgb(51,51,51);">Description</font> | <font style="color:rgb(51,51,51);">描述</font> | <font style="color:rgb(51,51,51);">描述</font> | <font style="color:rgb(51,51,51);">描述</font> |
| <font style="color:rgb(51,51,51);">任务类型</font><font style="color:rgb(51,51,51);"> (Type)</font> | <font style="color:rgb(51,51,51);">IssueType FK</font> | <font style="color:rgb(51,51,51);">Task</font> | <font style="color:rgb(51,51,51);">任务类型</font> | <font style="color:rgb(51,51,51);">任务类型</font> | <font style="color:rgb(51,51,51);">任务</font><font style="color:rgb(51,51,51);">/子任务</font> |
| <font style="color:rgb(51,51,51);">状态</font><font style="color:rgb(51,51,51);"> (State)</font> | <font style="color:rgb(51,51,51);">state FK (State 模型)</font> | <font style="color:rgb(51,51,51);">Status</font> | <font style="color:rgb(51,51,51);">状态</font> | <font style="color:rgb(51,51,51);">状态</font> | <font style="color:rgb(51,51,51);">状态</font> |
| <font style="color:rgb(51,51,51);">优先级</font><font style="color:rgb(51,51,51);"> (Priority)</font> | <font style="color:rgb(51,51,51);">urgent/high/medium/low/none (5级)</font> | <font style="color:rgb(51,51,51);">Highest/High/Medium/Low/Lowest</font> | <font style="color:rgb(51,51,51);">紧急</font><font style="color:rgb(51,51,51);">/高/中/低</font> | <font style="color:rgb(51,51,51);">紧急</font><font style="color:rgb(51,51,51);">/高/中/低</font> | <font style="color:rgb(51,51,51);">紧急</font><font style="color:rgb(51,51,51);">/高/中/低</font> |
| <font style="color:rgb(51,51,51);">负责人</font><font style="color:rgb(51,51,51);"> (Assignee)</font> | <font style="color:rgb(51,51,51);">assignees (M2M)</font> | <font style="color:rgb(51,51,51);">Assignee</font> | <font style="color:rgb(51,51,51);">负责人</font> | <font style="color:rgb(51,51,51);">负责人</font> | <font style="color:rgb(51,51,51);">负责人</font> |
| <font style="color:rgb(51,51,51);">创建人</font><font style="color:rgb(51,51,51);"> (Reporter)</font> | <font style="color:rgb(51,51,51);">created_by</font> | <font style="color:rgb(51,51,51);">Reporter</font> | <font style="color:rgb(51,51,51);">创建人</font> | <font style="color:rgb(51,51,51);">创建人</font> | <font style="color:rgb(51,51,51);">创建人</font> |
| <font style="color:rgb(51,51,51);">参与人</font><font style="color:rgb(51,51,51);"> (Members)</font> | <font style="color:rgb(51,51,51);">— (建议补充)</font> | <font style="color:rgb(51,51,51);">Watchers</font> | <font style="color:rgb(51,51,51);">参与人</font> | <font style="color:rgb(51,51,51);">—</font> | <font style="color:rgb(51,51,51);">参与人</font> |
| <font style="color:rgb(51,51,51);">开始日期</font><font style="color:rgb(51,51,51);"> (Start Date)</font> | <font style="color:rgb(51,51,51);">start_date</font> | <font style="color:rgb(51,51,51);">—</font> | <font style="color:rgb(51,51,51);">开始时间</font> | <font style="color:rgb(51,51,51);">开始日期</font> | <font style="color:rgb(51,51,51);">开始日期</font> |
| <font style="color:rgb(51,51,51);">截止日期</font><font style="color:rgb(51,51,51);"> (Due Date)</font> | <font style="color:rgb(51,51,51);">target_date</font> | <font style="color:rgb(51,51,51);">Due Date</font> | <font style="color:rgb(51,51,51);">截止时间</font> | <font style="color:rgb(51,51,51);">截止日期</font> | <font style="color:rgb(51,51,51);">截止日期</font> |
| <font style="color:rgb(51,51,51);">迭代</font><font style="color:rgb(51,51,51);"> (</font><font style="color:rgb(51,51,51);">Sprint</font><font style="color:rgb(51,51,51);">)</font> | <font style="color:rgb(51,51,51);">sprint</font><font style="color:rgb(51,51,51);"> </font><font style="color:rgb(51,51,51);">(关联)</font> | <font style="color:rgb(51,51,51);">Sprint</font> | <font style="color:rgb(51,51,51);">迭代</font> | <font style="color:rgb(51,51,51);">迭代</font> | <font style="color:rgb(51,51,51);">迭代</font> |
| <font style="color:rgb(51,51,51);">模块</font><font style="color:rgb(51,51,51);"> (Module)</font> | <font style="color:rgb(51,51,51);">module_issues M2M</font> | <font style="color:rgb(51,51,51);">Components</font> | <font style="color:rgb(51,51,51);">模块</font> | <font style="color:rgb(51,51,51);">模块</font> | <font style="color:rgb(51,51,51);">模块</font> |
| <font style="color:rgb(51,51,51);">任务分类</font><font style="color:rgb(51,51,51);"> (Category)</font> | <font style="color:rgb(51,51,51);">— (建议补充)</font> | <font style="color:rgb(51,51,51);">Custom Field</font> | <font style="color:rgb(51,51,51);">分类</font> | <font style="color:rgb(51,51,51);">分类</font> | <font style="color:rgb(51,51,51);">分类（前端</font><font style="color:rgb(51,51,51);">/后端/测试）</font> |
| <font style="color:rgb(51,51,51);">标签</font><font style="color:rgb(51,51,51);"> (Labels)</font> | <font style="color:rgb(51,51,51);">labels M2M</font> | <font style="color:rgb(51,51,51);">Labels</font> | <font style="color:rgb(51,51,51);">标签</font> | <font style="color:rgb(51,51,51);">标签</font> | <font style="color:rgb(51,51,51);">标签</font> |
| <font style="color:rgb(51,51,51);">父任务</font><font style="color:rgb(51,51,51);"> (Parent)</font> | <font style="color:rgb(51,51,51);">parent FK (WBS)</font> | <font style="color:rgb(51,51,51);">Parent Link</font> | <font style="color:rgb(51,51,51);">父任务</font> | <font style="color:rgb(51,51,51);">父任务</font> | <font style="color:rgb(51,51,51);">父任务</font> |
| <font style="color:rgb(51,51,51);">子任务</font><font style="color:rgb(51,51,51);"> (Sub-tasks)</font> | <font style="color:rgb(51,51,51);">Sub-issue count</font> | <font style="color:rgb(51,51,51);">Sub-task count</font> | <font style="color:rgb(51,51,51);">子任务树</font> | <font style="color:rgb(51,51,51);">子任务树</font> | <font style="color:rgb(51,51,51);">子任务</font><font style="color:rgb(51,51,51);">/孙任务</font> |
| <font style="color:rgb(51,51,51);">故事点</font><font style="color:rgb(51,51,51);"> (Story Points)</font> | <font style="color:rgb(51,51,51);">point (0-12)</font> | <font style="color:rgb(51,51,51);">Story Points</font> | <font style="color:rgb(51,51,51);">规模</font> | <font style="color:rgb(51,51,51);">规模</font> | <font style="color:rgb(51,51,51);">故事点数</font> |
| <font style="color:rgb(51,51,51);">估算工时</font><font style="color:rgb(51,51,51);"> (Estimate)</font> | <font style="color:rgb(51,51,51);">estimate_point FK</font> | <font style="color:rgb(51,51,51);">Estimation</font> | <font style="color:rgb(51,51,51);">预计工作量</font> | <font style="color:rgb(51,51,51);">预计工作量</font> | <font style="color:rgb(51,51,51);">预计工作量</font> |
| <font style="color:rgb(51,51,51);">实际工作量</font><font style="color:rgb(51,51,51);"> (Actual)</font> | <font style="color:rgb(51,51,51);">— (建议补充)</font> | <font style="color:rgb(51,51,51);">Time Tracking</font> | <font style="color:rgb(51,51,51);">实际工作量</font> | <font style="color:rgb(51,51,51);">实际工作量</font> | <font style="color:rgb(51,51,51);">实际工作量</font> |
| <font style="color:rgb(51,51,51);">完成时间</font><font style="color:rgb(51,51,51);"> (Completed)</font> | <font style="color:rgb(51,51,51);">completed_at</font> | <font style="color:rgb(51,51,51);">Resolved</font> | <font style="color:rgb(51,51,51);">完成时间</font> | <font style="color:rgb(51,51,51);">关闭时间</font> | <font style="color:rgb(51,51,51);">完成时间</font> |
| <font style="color:rgb(51,51,51);">延期原因</font><font style="color:rgb(51,51,51);"> (Delay Reason)</font> | <font style="color:rgb(51,51,51);">— (建议补充)</font> | <font style="color:rgb(51,51,51);">Custom Field</font> | <font style="color:rgb(51,51,51);">—</font> | <font style="color:rgb(51,51,51);">—</font> | <font style="color:rgb(51,51,51);">延期原因</font> |
| <font style="color:rgb(51,51,51);">附件</font><font style="color:rgb(51,51,51);"> (Attachments)</font> | <font style="color:rgb(51,51,51);">IssueAttachment</font> | <font style="color:rgb(51,51,51);">Attachment</font> | <font style="color:rgb(51,51,51);">附件</font> | <font style="color:rgb(51,51,51);">附件</font> | <font style="color:rgb(51,51,51);">附件</font> |
| <font style="color:rgb(51,51,51);">关联需求/任务/缺陷</font><font style="color:rgb(51,51,51);"> (Relations)</font> | <font style="color:rgb(51,51,51);">IssueRelation (6种)</font> | <font style="color:rgb(51,51,51);">Linked Issues</font> | <font style="color:rgb(51,51,51);">关联</font> | <font style="color:rgb(51,51,51);">关联需求</font><font style="color:rgb(51,51,51);">/缺陷</font> | <font style="color:rgb(51,51,51);">关联需求</font><font style="color:rgb(51,51,51);">/缺陷</font> |
| <font style="color:rgb(51,51,51);">关联需求</font><font style="color:rgb(51,51,51);"> (Related Req)</font> | <font style="color:rgb(51,51,51);">IssueRelation (implemented_by)</font> | <font style="color:rgb(51,51,51);">Linked Issues</font> | <font style="color:rgb(51,51,51);">关联需求</font> | <font style="color:rgb(51,51,51);">关联需求</font> | <font style="color:rgb(51,51,51);">关联需求</font> |
| <font style="color:rgb(51,51,51);">进度</font><font style="color:rgb(51,51,51);"> (%)</font> | <font style="color:rgb(51,51,51);">自动</font><font style="color:rgb(51,51,51);"> (子任务完成率)</font> | <font style="color:rgb(51,51,51);">自动</font> | <font style="color:rgb(51,51,51);">进度</font> | <font style="color:rgb(51,51,51);">进度</font> | <font style="color:rgb(51,51,51);">进度</font> |


### **<font style="color:rgb(51,51,51);">竞品属性详细分析</font>**
• Jira: 内置 Time Tracking 字段（Original Estimate, Remaining Estimate, Time Spent），可精确追踪任务工时偏差。

• TAPD: 任务支持「抄送」功能，可设置「关联需求 ID」字段进行跨类型关联。任务优先级默认「中」。

• ONES: 支持任务分类（前端/后端/测试/UI），以及「延期原因」字段，用于事后分析任务延迟根因。

### **<font style="color:rgb(51,51,51);">Ydsz Plane</font>****<font style="color:rgb(51,51,51);"> </font>****<font style="color:rgb(51,51,51);">需补充的属性建议</font>**
6. 实际工作量 (Actual Work)：支持手动记录或自动计时

7. 任务分类：区分开发任务、测试任务、文档任务等

8. 延期原因：任务未按计划完成时的分类枚举（需求变更/资源不足/技术阻塞/其他）

## **<font style="color:rgb(46,117,182);">4.3 缺陷 (Defect) 属性对比</font>**
缺陷由测试经理提出，记录软件中发现的异常或功能偏差。缺陷可由需求或任务关联产生，也可独立提交。

| **<font style="color:rgb(255,255,255);">属性名</font>** | **<font style="color:rgb(255,255,255);">Ydsz Plane</font>** | **<font style="color:rgb(255,255,255);">Jira</font>** | **<font style="color:rgb(255,255,255);">云效</font>** | **<font style="color:rgb(255,255,255);">TAPD</font>** | **<font style="color:rgb(255,255,255);">ONES/PingCode</font>** |
| --- | --- | --- | --- | --- | --- |
| <font style="color:rgb(51,51,51);">标题</font><font style="color:rgb(51,51,51);"> (Title)</font> | <font style="color:rgb(51,51,51);">name</font> | <font style="color:rgb(51,51,51);">Summary</font> | <font style="color:rgb(51,51,51);">标题</font> | <font style="color:rgb(51,51,51);">标题</font> | <font style="color:rgb(51,51,51);">标题</font> |
| <font style="color:rgb(51,51,51);">描述</font><font style="color:rgb(51,51,51);"> (Description)</font> | <font style="color:rgb(51,51,51);">description_json/html</font> | <font style="color:rgb(51,51,51);">Description</font> | <font style="color:rgb(51,51,51);">描述</font> | <font style="color:rgb(51,51,51);">描述</font> | <font style="color:rgb(51,51,51);">描述</font> |
| <font style="color:rgb(51,51,51);">缺陷类型</font><font style="color:rgb(51,51,51);"> (Type)</font> | <font style="color:rgb(51,51,51);">IssueType FK</font> | <font style="color:rgb(51,51,51);">Defect</font> | <font style="color:rgb(51,51,51);">缺陷类型</font> | <font style="color:rgb(51,51,51);">缺陷类型</font> | <font style="color:rgb(51,51,51);">缺陷</font><font style="color:rgb(51,51,51);">/子缺陷</font> |
| <font style="color:rgb(51,51,51);">状态</font><font style="color:rgb(51,51,51);"> (State)</font> | <font style="color:rgb(51,51,51);">state FK (State 模型)</font> | <font style="color:rgb(51,51,51);">Status</font> | <font style="color:rgb(51,51,51);">状态</font> | <font style="color:rgb(51,51,51);">状态</font> | <font style="color:rgb(51,51,51);">状态</font> |
| <font style="color:rgb(51,51,51);">严重程度</font><font style="color:rgb(51,51,51);"> (Severity)</font> | <font style="color:rgb(51,51,51);">— (需补充)</font> | <font style="color:rgb(51,51,51);">Severity (Blocker/Critical/Major/Trivial/Minor)</font> | <font style="color:rgb(51,51,51);">建议补充</font> | <font style="color:rgb(51,51,51);">致命</font><font style="color:rgb(51,51,51);">/严重/一般/提示/建议 (5级)</font> | <font style="color:rgb(51,51,51);">致命</font><font style="color:rgb(51,51,51);">/严重/一般/轻微/建议 (5级)</font> |
| <font style="color:rgb(51,51,51);">优先级</font><font style="color:rgb(51,51,51);"> (Priority)</font> | <font style="color:rgb(51,51,51);">urgent/high/medium/low/none</font> | <font style="color:rgb(51,51,51);">Priority (Highest/High/Medium/Low)</font> | <font style="color:rgb(51,51,51);">紧急</font><font style="color:rgb(51,51,51);">/高/中/低</font> | <font style="color:rgb(51,51,51);">紧急</font><font style="color:rgb(51,51,51);">/高/中/低/无关紧要 (5级)</font> | <font style="color:rgb(51,51,51);">紧急</font><font style="color:rgb(51,51,51);">/高/中/低</font> |
| <font style="color:rgb(51,51,51);">缺陷发现阶段</font> | <font style="color:rgb(51,51,51);">— (建议补充)</font> | <font style="color:rgb(51,51,51);">Custom Field</font> | <font style="color:rgb(51,51,51);">建议补充</font> | <font style="color:rgb(51,51,51);">发现阶段</font> | <font style="color:rgb(51,51,51);">发现阶段（测试</font><font style="color:rgb(51,51,51);">/线上/内部）</font> |
| <font style="color:rgb(51,51,51);">发现版本</font><font style="color:rgb(51,51,51);"> (Found Ver)</font> | <font style="color:rgb(51,51,51);">— (建议补充)</font> | <font style="color:rgb(51,51,51);">Affected Version</font> | <font style="color:rgb(51,51,51);">建议补充</font> | <font style="color:rgb(51,51,51);">发现版本</font> | <font style="color:rgb(51,51,51);">发现版本</font> |
| <font style="color:rgb(51,51,51);">修复版本</font><font style="color:rgb(51,51,51);"> (Fix Ver)</font> | <font style="color:rgb(51,51,51);">— (建议补充)</font> | <font style="color:rgb(51,51,51);">Fix Version</font> | <font style="color:rgb(51,51,51);">—</font> | <font style="color:rgb(51,51,51);">修复版本</font> | <font style="color:rgb(51,51,51);">修复版本</font> |
| <font style="color:rgb(51,51,51);">负责人</font><font style="color:rgb(51,51,51);"> (Assignee)</font> | <font style="color:rgb(51,51,51);">assignees (M2M)</font> | <font style="color:rgb(51,51,51);">Assignee</font> | <font style="color:rgb(51,51,51);">处理人</font> | <font style="color:rgb(51,51,51);">处理人</font> | <font style="color:rgb(51,51,51);">负责人</font><font style="color:rgb(51,51,51);">/解决人</font> |
| <font style="color:rgb(51,51,51);">创建人</font><font style="color:rgb(51,51,51);"> (Reporter)</font> | <font style="color:rgb(51,51,51);">created_by</font> | <font style="color:rgb(51,51,51);">Reporter</font> | <font style="color:rgb(51,51,51);">创建人</font> | <font style="color:rgb(51,51,51);">创建人</font> | <font style="color:rgb(51,51,51);">创建人</font> |
| <font style="color:rgb(51,51,51);">验证人</font><font style="color:rgb(51,51,51);"> (Verifier)</font> | <font style="color:rgb(51,51,51);">— (建议补充)</font> | <font style="color:rgb(51,51,51);">—</font> | <font style="color:rgb(51,51,51);">验证人</font> | <font style="color:rgb(51,51,51);">验证人</font> | <font style="color:rgb(51,51,51);">验证人</font> |
| <font style="color:rgb(51,51,51);">复现步骤</font> | <font style="color:rgb(51,51,51);">— (建议补充)</font> | <font style="color:rgb(51,51,51);">Custom Field</font> | <font style="color:rgb(51,51,51);">建议补充</font> | <font style="color:rgb(51,51,51);">缺陷描述</font> | <font style="color:rgb(51,51,51);">缺陷描述</font> |
| <font style="color:rgb(51,51,51);">期望结果</font> | <font style="color:rgb(51,51,51);">— (建议补充)</font> | <font style="color:rgb(51,51,51);">—</font> | <font style="color:rgb(51,51,51);">—</font> | <font style="color:rgb(51,51,51);">期望结果</font> | <font style="color:rgb(51,51,51);">期望结果</font> |
| <font style="color:rgb(51,51,51);">实际结果</font> | <font style="color:rgb(51,51,51);">— (建议补充)</font> | <font style="color:rgb(51,51,51);">—</font> | <font style="color:rgb(51,51,51);">—</font> | <font style="color:rgb(51,51,51);">实际结果</font> | <font style="color:rgb(51,51,51);">实际结果</font> |
| <font style="color:rgb(51,51,51);">环境信息</font><font style="color:rgb(51,51,51);"> (Env)</font> | <font style="color:rgb(51,51,51);">— (建议补充)</font> | <font style="color:rgb(51,51,51);">Environment</font> | <font style="color:rgb(51,51,51);">建议补充</font> | <font style="color:rgb(51,51,51);">测试环境</font> | <font style="color:rgb(51,51,51);">环境</font><font style="color:rgb(51,51,51);">/浏览器版本</font> |
| <font style="color:rgb(51,51,51);">根因</font><font style="color:rgb(51,51,51);"> (Root Cause)</font> | <font style="color:rgb(51,51,51);">— (建议补充)</font> | <font style="color:rgb(51,51,51);">Custom Field</font> | <font style="color:rgb(51,51,51);">—</font> | <font style="color:rgb(51,51,51);">根因</font> | <font style="color:rgb(51,51,51);">根因分类</font> |
| <font style="color:rgb(51,51,51);">解决方案</font><font style="color:rgb(51,51,51);"> (Resolution)</font> | <font style="color:rgb(51,51,51);">— (建议补充)</font> | <font style="color:rgb(51,51,51);">Resolution</font> | <font style="color:rgb(51,51,51);">解决方案</font> | <font style="color:rgb(51,51,51);">解决方案</font> | <font style="color:rgb(51,51,51);">解决方案</font> |
| <font style="color:rgb(51,51,51);">开始日期</font> | <font style="color:rgb(51,51,51);">start_date</font> | <font style="color:rgb(51,51,51);">—</font> | <font style="color:rgb(51,51,51);">开始时间</font> | <font style="color:rgb(51,51,51);">开始日期</font> | <font style="color:rgb(51,51,51);">开始日期</font> |
| <font style="color:rgb(51,51,51);">截止日期</font> | <font style="color:rgb(51,51,51);">target_date</font> | <font style="color:rgb(51,51,51);">Due Date</font> | <font style="color:rgb(51,51,51);">截止时间</font> | <font style="color:rgb(51,51,51);">截止日期</font> | <font style="color:rgb(51,51,51);">截止日期</font> |
| <font style="color:rgb(51,51,51);">迭代</font><font style="color:rgb(51,51,51);"> (</font><font style="color:rgb(51,51,51);">Sprint</font><font style="color:rgb(51,51,51);">)</font> | <font style="color:rgb(51,51,51);">sprint</font><font style="color:rgb(51,51,51);"> </font><font style="color:rgb(51,51,51);">(关联)</font> | <font style="color:rgb(51,51,51);">Sprint</font> | <font style="color:rgb(51,51,51);">迭代</font> | <font style="color:rgb(51,51,51);">迭代</font> | <font style="color:rgb(51,51,51);">迭代</font> |
| <font style="color:rgb(51,51,51);">模块</font><font style="color:rgb(51,51,51);"> (Module)</font> | <font style="color:rgb(51,51,51);">module_issues M2M</font> | <font style="color:rgb(51,51,51);">Components</font> | <font style="color:rgb(51,51,51);">模块</font> | <font style="color:rgb(51,51,51);">模块</font> | <font style="color:rgb(51,51,51);">模块</font> |
| <font style="color:rgb(51,51,51);">标签</font><font style="color:rgb(51,51,51);"> (Labels)</font> | <font style="color:rgb(51,51,51);">labels M2M</font> | <font style="color:rgb(51,51,51);">Labels</font> | <font style="color:rgb(51,51,51);">标签</font> | <font style="color:rgb(51,51,51);">标签</font> | <font style="color:rgb(51,51,51);">标签</font> |
| <font style="color:rgb(51,51,51);">父缺陷</font><font style="color:rgb(51,51,51);"> (Parent)</font> | <font style="color:rgb(51,51,51);">parent FK (WBS)</font> | <font style="color:rgb(51,51,51);">Parent Link</font> | <font style="color:rgb(51,51,51);">父缺陷</font> | <font style="color:rgb(51,51,51);">父缺陷</font> | <font style="color:rgb(51,51,51);">父缺陷</font> |
| <font style="color:rgb(51,51,51);">子缺陷</font><font style="color:rgb(51,51,51);"> (Sub-</font><font style="color:rgb(51,51,51);">defect</font><font style="color:rgb(51,51,51);">s)</font> | <font style="color:rgb(51,51,51);">Sub-issue count</font> | <font style="color:rgb(51,51,51);">Sub-task count</font> | <font style="color:rgb(51,51,51);">子缺陷</font> | <font style="color:rgb(51,51,51);">子缺陷</font> | <font style="color:rgb(51,51,51);">子缺陷</font><font style="color:rgb(51,51,51);">/孙缺陷</font> |
| <font style="color:rgb(51,51,51);">关联需求</font><font style="color:rgb(51,51,51);"> (Related Req)</font> | <font style="color:rgb(51,51,51);">IssueRelation</font> | <font style="color:rgb(51,51,51);">Linked Issues</font> | <font style="color:rgb(51,51,51);">关联需求</font> | <font style="color:rgb(51,51,51);">关联需求</font> | <font style="color:rgb(51,51,51);">关联需求</font> |
| <font style="color:rgb(51,51,51);">关联任务</font><font style="color:rgb(51,51,51);"> (Related Task)</font> | <font style="color:rgb(51,51,51);">IssueRelation</font> | <font style="color:rgb(51,51,51);">Linked Issues</font> | <font style="color:rgb(51,51,51);">关联任务</font> | <font style="color:rgb(51,51,51);">关联任务</font> | <font style="color:rgb(51,51,51);">关联任务</font> |
| <font style="color:rgb(51,51,51);">完成时间</font><font style="color:rgb(51,51,51);"> (Completed)</font> | <font style="color:rgb(51,51,51);">completed_at</font> | <font style="color:rgb(51,51,51);">Resolved</font> | <font style="color:rgb(51,51,51);">关闭时间</font> | <font style="color:rgb(51,51,51);">关闭时间</font> | <font style="color:rgb(51,51,51);">关闭时间</font> |
| <font style="color:rgb(51,51,51);">附件</font><font style="color:rgb(51,51,51);">/截图 (Attachments)</font> | <font style="color:rgb(51,51,51);">IssueAttachment</font> | <font style="color:rgb(51,51,51);">Attachment</font> | <font style="color:rgb(51,51,51);">附件</font> | <font style="color:rgb(51,51,51);">截图</font> | <font style="color:rgb(51,51,51);">附件</font><font style="color:rgb(51,51,51);">/截图</font> |


### **<font style="color:rgb(51,51,51);">严重程度级别对比（跨竞品）</font>**
| **<font style="color:rgb(255,255,255);">级别</font>** | **<font style="color:rgb(255,255,255);">TAPD/云效</font>** | **<font style="color:rgb(255,255,255);">ONES/PingCode</font>** | **<font style="color:rgb(255,255,255);">Jira</font>** | **<font style="color:rgb(255,255,255);">建议</font>****<font style="color:rgb(255,255,255);">Ydsz Plane</font>****<font style="color:rgb(255,255,255);"> </font>****<font style="color:rgb(255,255,255);">标准</font>** |
| --- | --- | --- | --- | --- |
| <font style="color:rgb(51,51,51);">1级</font> | <font style="color:rgb(51,51,51);">致命</font><font style="color:rgb(51,51,51);"> (Block)</font> | <font style="color:rgb(51,51,51);">致命</font><font style="color:rgb(51,51,51);"> (Blocker)</font> | <font style="color:rgb(51,51,51);">Blocker</font> | <font style="color:rgb(51,51,51);">致命</font> |
| <font style="color:rgb(51,51,51);">2级</font> | <font style="color:rgb(51,51,51);">严重</font><font style="color:rgb(51,51,51);"> (Critical)</font> | <font style="color:rgb(51,51,51);">严重</font><font style="color:rgb(51,51,51);"> (Critical)</font> | <font style="color:rgb(51,51,51);">Critical</font> | <font style="color:rgb(51,51,51);">严重</font> |
| <font style="color:rgb(51,51,51);">3级</font> | <font style="color:rgb(51,51,51);">一般</font><font style="color:rgb(51,51,51);"> (Major)</font> | <font style="color:rgb(51,51,51);">一般</font><font style="color:rgb(51,51,51);"> (Major)</font> | <font style="color:rgb(51,51,51);">Major</font> | <font style="color:rgb(51,51,51);">一般</font> |
| <font style="color:rgb(51,51,51);">4级</font> | <font style="color:rgb(51,51,51);">提示</font><font style="color:rgb(51,51,51);"> (Minor)</font> | <font style="color:rgb(51,51,51);">轻微</font><font style="color:rgb(51,51,51);"> (Minor)</font> | <font style="color:rgb(51,51,51);">Minor</font> | <font style="color:rgb(51,51,51);">提示</font> |
| <font style="color:rgb(51,51,51);">5级</font> | <font style="color:rgb(51,51,51);">建议</font><font style="color:rgb(51,51,51);"> (Suggestion)</font> | <font style="color:rgb(51,51,51);">建议</font><font style="color:rgb(51,51,51);"> (Trivial)</font> | <font style="color:rgb(51,51,51);">Trivial</font> | <font style="color:rgb(51,51,51);">建议</font> |


### **<font style="color:rgb(51,51,51);">竞品缺陷字段详细说明</font>**
• TAPD: 区分「优先级」和「严重程度」两个独立维度。支持「发现阶段」「缺陷类型」「发现版本」等自定义字段，支持父子缺陷嵌套。

• Jira: Environment 字段记录运行环境（浏览器/OS/AWS region 等）。Resolution 字段记录解决方式（Fixed/Cannot Reproduce/Won't Fix/Duplicate 等）。

• ONES: 支持「根因」分类（需求问题/技术问题/环境问题/数据问题），便于缺陷度量分析。关联需求后自动联动需求状态。

### **<font style="color:rgb(51,51,51);">Ydsz Plane</font>****<font style="color:rgb(51,51,51);"> </font>****<font style="color:rgb(51,51,51);">需补充的属性建议（优先级排序）</font>**
9. 严重程度 (Severity)：5级（致命/严重/一般/提示/建议），必填字段，用于缺陷分类统计

10. 复现步骤 (Steps to Reproduce)：富文本字段 + 期望/实际结果，便于开发定位

11. 发现阶段 (Found Phase)：枚举（单元测试/集成测试/UAT/线上/客户反馈），必填

12. 发现/修复版本 (Found/Fix Version)：关联版本字段，唯一必填

13. 环境信息 (Environment)：可选文本，记录测试环境/浏览器/设备信息

14. 根因 (Root Cause)：解决后必填，用于事后度量分析

15. 验证人 (Verifier)：回归验证的负责人

  


# **<font style="color:rgb(31,78,121);">5. 版本与迭代属性对比</font>**
## **<font style="color:rgb(46,117,182);">5.1 版本 (Version) 属性</font>**
| **<font style="color:rgb(255,255,255);">属性名</font>** | **<font style="color:rgb(255,255,255);">Ydsz Plane</font>** | **<font style="color:rgb(255,255,255);">Jira (Fix Version)</font>** | **<font style="color:rgb(255,255,255);">云效</font>****<font style="color:rgb(255,255,255);"> (发布计划)</font>** | **<font style="color:rgb(255,255,255);">ONES (规划)</font>** |
| --- | --- | --- | --- | --- |
| <font style="color:rgb(51,51,51);">名称</font> | <font style="color:rgb(51,51,51);">— (需新建)</font> | <font style="color:rgb(51,51,51);">Version Name</font> | <font style="color:rgb(51,51,51);">发布名称</font> | <font style="color:rgb(51,51,51);">版本名称</font> |
| <font style="color:rgb(51,51,51);">描述</font> | <font style="color:rgb(51,51,51);">—</font> | <font style="color:rgb(51,51,51);">Description</font> | <font style="color:rgb(51,51,51);">描述</font> | <font style="color:rgb(51,51,51);">描述</font> |
| <font style="color:rgb(51,51,51);">开始日期</font> | <font style="color:rgb(51,51,51);">—</font> | <font style="color:rgb(51,51,51);">Release Date</font> | <font style="color:rgb(51,51,51);">开始日期</font> | <font style="color:rgb(51,51,51);">开始日期</font> |
| <font style="color:rgb(51,51,51);">发布日期</font> | <font style="color:rgb(51,51,51);">—</font> | <font style="color:rgb(51,51,51);">Released?</font> | <font style="color:rgb(51,51,51);">发布日期</font> | <font style="color:rgb(51,51,51);">发布日期</font> |
| <font style="color:rgb(51,51,51);">迭代列表</font> | <font style="color:rgb(51,51,51);">— (关联)</font> | <font style="color:rgb(51,51,51);">—</font> | <font style="color:rgb(51,51,51);">—</font> | <font style="color:rgb(51,51,51);">—</font> |
| <font style="color:rgb(51,51,51);">状态</font> | <font style="color:rgb(51,51,51);">—</font> | <font style="color:rgb(51,51,51);">Released/Unreleased</font> | <font style="color:rgb(51,51,51);">未发布</font><font style="color:rgb(51,51,51);">/已发布/归档</font> | <font style="color:rgb(51,51,51);">未开始</font><font style="color:rgb(51,51,51);">/进行中/已发布</font> |
| <font style="color:rgb(51,51,51);">进度</font> | <font style="color:rgb(51,51,51);">—</font> | <font style="color:rgb(51,51,51);">自动统计</font> | <font style="color:rgb(51,51,51);">自动统计</font> | <font style="color:rgb(51,51,51);">自动统计</font> |


## **<font style="color:rgb(46,117,182);">5.2 迭代 (</font>****<font style="color:rgb(46,117,182);">Sprint</font>****<font style="color:rgb(46,117,182);">) 属性</font>**
| **<font style="color:rgb(255,255,255);">属性名</font>** | **<font style="color:rgb(255,255,255);">Ydsz Plane</font>** | **<font style="color:rgb(255,255,255);">Jira (Sprint)</font>** | **<font style="color:rgb(255,255,255);">云效</font>****<font style="color:rgb(255,255,255);"> (迭代)</font>** | **<font style="color:rgb(255,255,255);">ONES (迭代)</font>** | **<font style="color:rgb(255,255,255);">TAPD (迭代)</font>** |
| --- | --- | --- | --- | --- | --- |
| <font style="color:rgb(51,51,51);">名称</font> | <font style="color:rgb(51,51,51);">name</font> | <font style="color:rgb(51,51,51);">Sprint Name</font> | <font style="color:rgb(51,51,51);">迭代名称</font> | <font style="color:rgb(51,51,51);">迭代名称</font> | <font style="color:rgb(51,51,51);">名称</font> |
| <font style="color:rgb(51,51,51);">描述</font> | <font style="color:rgb(51,51,51);">description</font> | <font style="color:rgb(51,51,51);">Goal</font> | <font style="color:rgb(51,51,51);">描述</font> | <font style="color:rgb(51,51,51);">描述</font> | <font style="color:rgb(51,51,51);">描述</font> |
| <font style="color:rgb(51,51,51);">开始日期</font> | <font style="color:rgb(51,51,51);">start_date</font> | <font style="color:rgb(51,51,51);">Start Date</font> | <font style="color:rgb(51,51,51);">开始时间</font> | <font style="color:rgb(51,51,51);">开始日期</font> | <font style="color:rgb(51,51,51);">开始时间</font> |
| <font style="color:rgb(51,51,51);">结束日期</font> | <font style="color:rgb(51,51,51);">end_date</font> | <font style="color:rgb(51,51,51);">End Date</font> | <font style="color:rgb(51,51,51);">结束时间</font> | <font style="color:rgb(51,51,51);">结束日期</font> | <font style="color:rgb(51,51,51);">结束时间</font> |
| <font style="color:rgb(51,51,51);">负责人</font> | <font style="color:rgb(51,51,51);">owned_by</font> | <font style="color:rgb(51,51,51);">—</font> | <font style="color:rgb(51,51,51);">负责人</font> | <font style="color:rgb(51,51,51);">负责人</font> | <font style="color:rgb(51,51,51);">负责人</font> |
| <font style="color:rgb(51,51,51);">状态</font> | <font style="color:rgb(51,51,51);">— (active/closed)</font> | <font style="color:rgb(51,51,51);">Active/Closed</font> | <font style="color:rgb(51,51,51);">计划中</font><font style="color:rgb(51,51,51);">/进行中/已完成</font> | <font style="color:rgb(51,51,51);">未开始</font><font style="color:rgb(51,51,51);">/进行中/已完成</font> | <font style="color:rgb(51,51,51);">进行中</font><font style="color:rgb(51,51,51);">/已完成</font> |
| <font style="color:rgb(51,51,51);">进度快照</font> | <font style="color:rgb(51,51,51);">progress_snapshot</font> | <font style="color:rgb(51,51,51);">—</font> | <font style="color:rgb(51,51,51);">—</font> | <font style="color:rgb(51,51,51);">完成进度</font><font style="color:rgb(51,51,51);">%</font> | <font style="color:rgb(51,51,51);">—</font> |
| <font style="color:rgb(51,51,51);">版本</font> | <font style="color:rgb(51,51,51);">version (FK)</font> | <font style="color:rgb(51,51,51);">—</font> | <font style="color:rgb(51,51,51);">归属版本</font> | <font style="color:rgb(51,51,51);">归属版本</font> | <font style="color:rgb(51,51,51);">—</font> |
| <font style="color:rgb(51,51,51);">需求/任务/缺陷类型过滤</font> | <font style="color:rgb(51,51,51);">display_filters.type</font> | <font style="color:rgb(51,51,51);">—</font> | <font style="color:rgb(51,51,51);">—</font> | <font style="color:rgb(51,51,51);">—</font> | <font style="color:rgb(51,51,51);">—</font> |


## **<font style="color:rgb(46,117,182);">5.3 版本模型设计建议</font>**
• 建议在现有Sprint 模型之上新增 Version 模型，一个 Version 包含 1~N 个 Sprints。

• Version 的状态流转：规划中 ⟶ 开发中 ⟶ 已发布 ⟶ 已归档（可选）。

• 一个需求/任务可以在多个 Sprints 中实现，但只能归属一个 Version（首次发布版本）。

• 缺陷必须有「发现版本」和「修复版本」两个字段，用于缺陷度量分析。

# **<font style="color:rgb(31,78,121);">6. 状态机属性设计</font>**
## **<font style="color:rgb(46,117,182);">6.1 </font>****<font style="color:rgb(46,117,182);">Ydsz Plane</font>****<font style="color:rgb(46,117,182);"> State 模型现有设计</font>**
| **<font style="color:rgb(255,255,255);">属性名</font>** | **<font style="color:rgb(255,255,255);">类型</font>****<font style="color:rgb(255,255,255);">/选项</font>** | **<font style="color:rgb(255,255,255);">说明</font>** |
| --- | --- | --- |
| <font style="color:rgb(51,51,51);">name</font> | <font style="color:rgb(51,51,51);">CharField</font> | <font style="color:rgb(51,51,51);">状态名称（如</font><font style="color:rgb(51,51,51);"> 'In Progress'）</font> |
| <font style="color:rgb(51,51,51);">description</font> | <font style="color:rgb(51,51,51);">TextField</font> | <font style="color:rgb(51,51,51);">状态描述</font> |
| <font style="color:rgb(51,51,51);">color</font> | <font style="color:rgb(51,51,51);">CharField</font> | <font style="color:rgb(51,51,51);">状态颜色（</font><font style="color:rgb(51,51,51);">UI 展示）</font> |
| <font style="color:rgb(51,51,51);">sequence</font> | <font style="color:rgb(51,51,51);">IntegerField</font> | <font style="color:rgb(51,51,51);">排序权重</font> |
| <font style="color:rgb(51,51,51);">group</font> | <font style="color:rgb(51,51,51);">Choices: backlog/unstarted/started/completed/cancelled/triage</font> | <font style="color:rgb(51,51,51);">状态分组（决定统计口径）</font> |
| <font style="color:rgb(51,51,51);">is_triage</font> | <font style="color:rgb(51,51,51);">BooleanField</font> | <font style="color:rgb(51,51,51);">是否为</font><font style="color:rgb(51,51,51);"> Triage（分诊态，默认隐藏）</font> |
| <font style="color:rgb(51,51,51);">default</font> | <font style="color:rgb(51,51,51);">BooleanField</font> | <font style="color:rgb(51,51,51);">新建需求/任务/缺陷的默认状态</font> |
| <font style="color:rgb(51,51,51);">slug</font> | <font style="color:rgb(51,51,51);">CharField</font> | <font style="color:rgb(51,51,51);">机器名（用于</font><font style="color:rgb(51,51,51);"> API/过滤）</font> |


## **<font style="color:rgb(46,117,182);">6.2 竞品状态设计对比</font>**
| **<font style="color:rgb(255,255,255);">产品</font>** | **<font style="color:rgb(255,255,255);">状态设计</font>** | **<font style="color:rgb(255,255,255);">支持程度</font>** |
| --- | --- | --- |
| <font style="color:rgb(51,51,51);">Ydsz Plane</font> | <font style="color:rgb(51,51,51);">6 状态组，自定义状态，每项目独立</font> | <font style="color:rgb(51,51,51);">★★★★★ 高灵活度</font> |
| <font style="color:rgb(51,51,51);">Jira</font> | <font style="color:rgb(51,51,51);">每项目独立配置</font><font style="color:rgb(51,51,51);">+工作流+转换条件</font> | <font style="color:rgb(51,51,51);">★★★★★ 极高（但复杂）</font> |
| <font style="color:rgb(51,51,51);">云效</font> | <font style="color:rgb(51,51,51);">预定义模板</font><font style="color:rgb(51,51,51);">+可自定义状态+状态组</font> | <font style="color:rgb(51,51,51);">★★★★ 高</font> |
| <font style="color:rgb(51,51,51);">TAPD</font> | <font style="color:rgb(51,51,51);">状态模板</font><font style="color:rgb(51,51,51);">+工作流+字段控制</font> | <font style="color:rgb(51,51,51);">★★★★ 中高</font> |
| <font style="color:rgb(51,51,51);">ONES</font> | <font style="color:rgb(51,51,51);">状态机</font><font style="color:rgb(51,51,51);">+状态流转规则+自动化</font> | <font style="color:rgb(51,51,51);">★★★★ 中高</font> |
| <font style="color:rgb(51,51,51);">PingCode</font> | <font style="color:rgb(51,51,51);">状态设计</font><font style="color:rgb(51,51,51);">+状态组+触发器</font> | <font style="color:rgb(51,51,51);">★★★★ 中高</font> |


## **<font style="color:rgb(46,117,182);">6.3 默认状态推荐</font>**
• Product Backlog: Backlog 状态组，需求进入需求池等待评审

• 开发流程: Todo ⟶ In Progress ⟶ In Review ⟶ Done

• 缺陷流程: Open ⟶ In Progress ⟶ Resolved ⟶ Verified / Reopened

• 其他可选状态: Cancelled, Triage, Blocked, Duplicate

# **<font style="color:rgb(31,78,121);">7. 属性补充优先级总结</font>**
| **<font style="color:rgb(255,255,255);">优先级</font>** | **<font style="color:rgb(255,255,255);">属性</font>** | **<font style="color:rgb(255,255,255);">影响范围</font>** | **<font style="color:rgb(255,255,255);">实施建议</font>** |
| --- | --- | --- | --- |
| <font style="color:rgb(51,51,51);">P0 必补</font> | <font style="color:rgb(51,51,51);">严重程度</font><font style="color:rgb(51,51,51);"> (Severity)</font> | <font style="color:rgb(51,51,51);">缺陷</font> | <font style="color:rgb(51,51,51);">新增必填枚举字段，</font><font style="color:rgb(51,51,51);">5级别</font> |
| <font style="color:rgb(51,51,51);">P0 必补</font> | <font style="color:rgb(51,51,51);">复现步骤</font><font style="color:rgb(51,51,51);"> + 期望/实际结果</font> | <font style="color:rgb(51,51,51);">缺陷</font> | <font style="color:rgb(51,51,51);">新增富文本字段（</font><font style="color:rgb(51,51,51);">description 复用）</font> |
| <font style="color:rgb(51,51,51);">P0 必补</font> | <font style="color:rgb(51,51,51);">发现阶段</font><font style="color:rgb(51,51,51);"> (Found Phase)</font> | <font style="color:rgb(51,51,51);">缺陷</font> | <font style="color:rgb(51,51,51);">新增必填枚举字段</font> |
| <font style="color:rgb(51,51,51);">P0 必补</font> | <font style="color:rgb(51,51,51);">发现版本</font><font style="color:rgb(51,51,51);"> + 修复版本</font> | <font style="color:rgb(51,51,51);">缺陷</font> | <font style="color:rgb(51,51,51);">新增</font><font style="color:rgb(51,51,51);"> Version 模型 + FK，1:N</font> |
| <font style="color:rgb(51,51,51);">P0 必补</font> | <font style="color:rgb(51,51,51);">版本</font><font style="color:rgb(51,51,51);"> (Version) 模型</font> | <font style="color:rgb(51,51,51);">需求</font><font style="color:rgb(51,51,51);">/任务/缺陷</font> | <font style="color:rgb(51,51,51);">新建模型</font><font style="color:rgb(51,51,51);"> + 跨 </font><font style="color:rgb(51,51,51);">Sprint</font><font style="color:rgb(51,51,51);"> </font><font style="color:rgb(51,51,51);">关联</font> |
| <font style="color:rgb(51,51,51);">P1 应补</font> | <font style="color:rgb(51,51,51);">参与人</font><font style="color:rgb(51,51,51);">/抄送人 (Watchers)</font> | <font style="color:rgb(51,51,51);">需求</font><font style="color:rgb(51,51,51);">/任务</font> | <font style="color:rgb(51,51,51);">M2M 关联 User，支持通知</font> |
| <font style="color:rgb(51,51,51);">P1 应补</font> | <font style="color:rgb(51,51,51);">任务分类</font><font style="color:rgb(51,51,51);"> (Category)</font> | <font style="color:rgb(51,51,51);">任务</font> | <font style="color:rgb(51,51,51);">新增可选枚举字段</font> |
| <font style="color:rgb(51,51,51);">P1 应补</font> | <font style="color:rgb(51,51,51);">需求来源</font><font style="color:rgb(51,51,51);"> (Source)</font> | <font style="color:rgb(51,51,51);">需求</font> | <font style="color:rgb(51,51,51);">新增可选枚举字段</font> |
| <font style="color:rgb(51,51,51);">P1 应补</font> | <font style="color:rgb(51,51,51);">进度自动计算</font><font style="color:rgb(51,51,51);"> (%)</font> | <font style="color:rgb(51,51,51);">需求</font><font style="color:rgb(51,51,51);">/任务</font> | <font style="color:rgb(51,51,51);">基于子需求/子任务完成率自动更新</font> |
| <font style="color:rgb(51,51,51);">P1 应补</font> | <font style="color:rgb(51,51,51);">根因分类</font><font style="color:rgb(51,51,51);"> (Root Cause)</font> | <font style="color:rgb(51,51,51);">缺陷</font> | <font style="color:rgb(51,51,51);">解决缺陷时必填</font> |
| <font style="color:rgb(51,51,51);">P2 建议</font> | <font style="color:rgb(51,51,51);">环境信息</font><font style="color:rgb(51,51,51);"> (Environment)</font> | <font style="color:rgb(51,51,51);">缺陷</font> | <font style="color:rgb(51,51,51);">新增可选文本字段</font> |
| <font style="color:rgb(51,51,51);">P2 建议</font> | <font style="color:rgb(51,51,51);">验证人</font><font style="color:rgb(51,51,51);"> (Verifier)</font> | <font style="color:rgb(51,51,51);">缺陷</font> | <font style="color:rgb(51,51,51);">新增可选</font><font style="color:rgb(51,51,51);"> FK 到 User</font> |
| <font style="color:rgb(51,51,51);">P2 建议</font> | <font style="color:rgb(51,51,51);">延期原因</font><font style="color:rgb(51,51,51);"> (Delay Reason)</font> | <font style="color:rgb(51,51,51);">任务</font> | <font style="color:rgb(51,51,51);">延期时填写，可选枚举</font> |
| <font style="color:rgb(51,51,51);">P2 建议</font> | <font style="color:rgb(51,51,51);">实际工作量</font><font style="color:rgb(51,51,51);"> (Actual)</font> | <font style="color:rgb(51,51,51);">任务</font> | <font style="color:rgb(51,51,51);">记录实际耗时，用于效能度量</font> |


  


# **<font style="color:rgb(31,78,121);">8.</font>****<font style="color:rgb(31,78,121);"> </font>****<font style="color:rgb(31,78,121);">核心业务功能</font>**
**核心业务功能需求细化**

本章对Ydsz Plane 项目 PRD V7.0 中定义的 10 个核心业务功能模块逐一展开详细的功能需求描述。每个模块包含功能概述、竞品对标、用户故事、功能需求详述（P0/P1/P2 分级）、交互流程和数据模型展望。  


# **<font style="color:rgb(31,78,121);">8.1 工作空间管理 (Workspace)</font>**
工作空间是Ydsz Plane 系统中的顶级容器，代表一个企业或团队级组织单元。每个工作空间拥有独立的数据隔离、成员体系、权限模型和配置，是租户隔离的基本单位。对标 Jira Workspace、云效企业、TAPD 公司。

## **<font style="color:rgb(46,117,182);">1.1 功能概述</font>**
工作空间是组织级容器，承载所有项目、成员、配置和数据的隔离边界。一个工作空间对应一个团队或企业，空间内的数据和成员与其他空间完全隔离。

## **<font style="color:rgb(46,117,182);">1.2 竞品对标</font>**
| **<font style="color:rgb(255,255,255);">功能点</font>** | **<font style="color:rgb(255,255,255);">Jira</font>** | **<font style="color:rgb(255,255,255);">云效企业</font>** | **<font style="color:rgb(255,255,255);">TAPD 公司</font>** | **<font style="color:rgb(255,255,255);">ONES 团队</font>** | **<font style="color:rgb(255,255,255);">Ydsz Plane</font>****<font style="color:rgb(255,255,255);"> </font>****<font style="color:rgb(255,255,255);">现有</font>** |
| --- | --- | --- | --- | --- | --- |
| <font style="color:rgb(51,51,51);">创建</font><font style="color:rgb(51,51,51);">/编辑空间</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> |
| <font style="color:rgb(51,51,51);">Slug 唯一校验</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> |
| <font style="color:rgb(51,51,51);">成员邀请</font><font style="color:rgb(51,51,51);">/审核</font> | <font style="color:rgb(51,51,51);">● 邮箱+链接</font> | <font style="color:rgb(51,51,51);">● 邮箱+链接</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> |
| <font style="color:rgb(51,51,51);">多级管理员</font> | <font style="color:rgb(51,51,51);">● Owner/Admin</font> | <font style="color:rgb(51,51,51);">● 多级</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">● Owner/Admin/Member/Guest</font> |
| <font style="color:rgb(51,51,51);">API Token 管理</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> |
| <font style="color:rgb(51,51,51);">空间级权限配置</font> | <font style="color:rgb(51,51,51);">● 细粒度</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> |
| <font style="color:rgb(51,51,51);">SSO/SAML 集成</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">○ 规划中</font> |
| <font style="color:rgb(51,51,51);">多語言</font><font style="color:rgb(51,51,51);">/时区</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">● 中英</font> | <font style="color:rgb(51,51,51);">● 中英</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> |
| <font style="color:rgb(51,51,51);">数据归档</font><font style="color:rgb(51,51,51);">/导出</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">○ 规划中</font> |
| <font style="color:rgb(51,51,51);">品牌定制</font><font style="color:rgb(51,51,51);">(Logo/配色)</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">○ 规划中</font> |


## **<font style="color:rgb(46,117,182);">1.3 用户故事</font>**
• 作为 IT 管理员，我希望创建独立的工作空间来隔离不同部门的数据。

• 作为团队负责人，我希望邀请成员加入工作空间并分配合适的角色。

• 作为安全管理员，我希望配置 SSO 登录策略以确保企业账号安全。

• 作为工作空间 Owner，我希望自定义空间 Logo 和主题色以体现团队品牌。

## **<font style="color:rgb(46,117,182);">1.4 功能需求详述</font>**
### **<font style="color:rgb(51,51,51);">1.4.1 空间生命周期管理</font>**
| **<font style="color:rgb(255,255,255);">功能</font>** | **<font style="color:rgb(255,255,255);">详细描述</font>** | **<font style="color:rgb(255,255,255);">验收标准</font>** | **<font style="color:rgb(255,255,255);">优先级</font>** |
| --- | --- | --- | --- |
| <font style="color:rgb(51,51,51);">创建空间</font> | <font style="color:rgb(51,51,51);">支持输入空间名称、描述、唯一</font><font style="color:rgb(51,51,51);"> Slug、时区、语言、Logo；Slug 全局唯一校验</font> | <font style="color:rgb(51,51,51);">创建成功返回空间详情，</font><font style="color:rgb(51,51,51);">Slug 冲突提示</font> | <font style="color:rgb(51,51,51);">P0</font> |
| <font style="color:rgb(51,51,51);">编辑空间</font> | <font style="color:rgb(51,51,51);">Owner 可修改空间基本信息（名称、描述、时区、语言、Logo、主题色）</font> | <font style="color:rgb(51,51,51);">修改后立即生效，前端实时更新</font> | <font style="color:rgb(51,51,51);">P0</font> |
| <font style="color:rgb(51,51,51);">归档空间</font> | <font style="color:rgb(51,51,51);">Owner 可将非活跃空间归档，归档后只读不可新建项目，可恢复</font> | <font style="color:rgb(51,51,51);">归档后项目列表只读，归档操作需二次确认</font> | <font style="color:rgb(51,51,51);">P1</font> |
| <font style="color:rgb(51,51,51);">删除空间</font> | <font style="color:rgb(51,51,51);">Owner 可删除空空间（无项目），有项目时需先删除所有项目</font> | <font style="color:rgb(51,51,51);">删除操作需二次确认并输入空间名称校验</font> | <font style="color:rgb(51,51,51);">P2</font> |
| <font style="color:rgb(51,51,51);">空间列表</font> | <font style="color:rgb(51,51,51);">用户可查看自己加入的所有工作空间，支持搜索和排序</font> | <font style="color:rgb(51,51,51);">显示空间</font><font style="color:rgb(51,51,51);"> Logo、名称、成员数、项目数</font> | <font style="color:rgb(51,51,51);">P0</font> |


### **<font style="color:rgb(51,51,51);">1.4.2 成员管理</font>**
| **<font style="color:rgb(255,255,255);">功能</font>** | **<font style="color:rgb(255,255,255);">详细描述</font>** | **<font style="color:rgb(255,255,255);">验收标准</font>** | **<font style="color:rgb(255,255,255);">优先级</font>** |
| --- | --- | --- | --- |
| <font style="color:rgb(51,51,51);">邀请成员</font> | <font style="color:rgb(51,51,51);">通过邮箱链接邀请，支持批量邀请；邀请码有效期</font><font style="color:rgb(51,51,51);"> 7 天，可撤销</font> | <font style="color:rgb(51,51,51);">被邀请人收到邮件，点击链接注册</font><font style="color:rgb(51,51,51);">/加入</font> | <font style="color:rgb(51,51,51);">P0</font> |
| <font style="color:rgb(51,51,51);">审核成员</font> | <font style="color:rgb(51,51,51);">支持开放（免审核）和邀请制（需管理员审核）两种模式</font> | <font style="color:rgb(51,51,51);">审核模式需管理员批准后才能加入</font> | <font style="color:rgb(51,51,51);">P1</font> |
| <font style="color:rgb(51,51,51);">角色管理</font> | <font style="color:rgb(51,51,51);">4 级角色：Owner/Admin/Member/Guest，每级有独立权限集</font> | <font style="color:rgb(51,51,51);">角色切换后权限实时生效</font> | <font style="color:rgb(51,51,51);">P0</font> |
| <font style="color:rgb(51,51,51);">移除成员</font> | <font style="color:rgb(51,51,51);">Owner/Admin 可移除成员，移除后该空间内需求/任务/缺陷保留但不再有权限</font> | <font style="color:rgb(51,51,51);">移除操作需二次确认</font> | <font style="color:rgb(51,51,51);">P1</font> |
| <font style="color:rgb(51,51,51);">成员列表</font> | <font style="color:rgb(51,51,51);">展示成员姓名、邮箱、角色、加入时间、最后活跃时间</font> | <font style="color:rgb(51,51,51);">支持按角色</font><font style="color:rgb(51,51,51);">/加入时间筛选排序</font> | <font style="color:rgb(51,51,51);">P0</font> |
| <font style="color:rgb(51,51,51);">成员批量导入</font> | <font style="color:rgb(51,51,51);">通过</font><font style="color:rgb(51,51,51);"> CSV 批量导入成员（姓名、邮箱、邮箱已存在则关联）</font> | <font style="color:rgb(51,51,51);">显示导入结果（成功</font><font style="color:rgb(51,51,51);">/失败/已存在统计）</font> | <font style="color:rgb(51,51,51);">P2</font> |


### **<font style="color:rgb(51,51,51);">1.4.3 权限体系</font>**
| **<font style="color:rgb(255,255,255);">权限点</font>** | **<font style="color:rgb(255,255,255);">Owner</font>** | **<font style="color:rgb(255,255,255);">Admin</font>** | **<font style="color:rgb(255,255,255);">Member</font>** | **<font style="color:rgb(255,255,255);">Guest</font>** |
| --- | --- | --- | --- | --- |
| <font style="color:rgb(51,51,51);">空间设置</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">—</font> | <font style="color:rgb(51,51,51);">—</font> |
| <font style="color:rgb(51,51,51);">成员管理</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">—</font> | <font style="color:rgb(51,51,51);">—</font> |
| <font style="color:rgb(51,51,51);">创建</font><font style="color:rgb(51,51,51);">/删除项目</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">—</font> | <font style="color:rgb(51,51,51);">—</font> |
| <font style="color:rgb(51,51,51);">创建</font><font style="color:rgb(51,51,51);">/编辑项目内需求/任务/缺陷</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">—</font> |
| <font style="color:rgb(51,51,51);">查看项目</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> |
| <font style="color:rgb(51,51,51);">评论</font><font style="color:rgb(51,51,51);">/提及</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">○ 待配置</font> |
| <font style="color:rgb(51,51,51);">管理</font><font style="color:rgb(51,51,51);"> API Token</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">—</font> | <font style="color:rgb(51,51,51);">—</font> |
| <font style="color:rgb(51,51,51);">归档</font><font style="color:rgb(51,51,51);">/删除空间</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">—</font> | <font style="color:rgb(51,51,51);">—</font> | <font style="color:rgb(51,51,51);">—</font> |


## **<font style="color:rgb(46,117,182);">1.5 交互流程</font>**
16. 创建空间：点击「新建工作空间」→ 填写名称/Slug/时区 → 上传 Logo → 提交 → 自动成为 Owner

17. 邀请成员：进入空间设置→ 成员管理 → 点击「邀请成员」→ 输入邮箱（支持多个）→ 选择角色 → 发送邀请

18. 权限变更：成员列表→ 点击成员行 → 选择新角色 → 保存 → 强制刷新该成员会话

## **<font style="color:rgb(46,117,182);">1.6 数据模型设计</font>**
| **<font style="color:rgb(255,255,255);">📋</font>****<font style="color:rgb(255,255,255);"> workspace</font>** | | |
| --- | --- | --- |
| **<font style="color:rgb(255,255,255);">字段名</font>** | **<font style="color:rgb(255,255,255);">类型</font>****<font style="color:rgb(255,255,255);">/约束</font>** | **<font style="color:rgb(255,255,255);">说明</font>** |
| id | **PK (auto)** | 主键自增ID |
| name | VARCHAR(255) | 名称/标识 |
| slug | VARCHAR(255) | 名称/标识 |
| logo_url | TEXT |  |
| timezone | TEXT |  |
| language | TEXT |  |
| owner | **FK → owner** | 外键关联 |
| status | active/archived | active/archived |
| created_at | TIMESTAMP | 时间戳 |
| updated_at | TIMESTAMP | 时间戳 |
| **<font style="color:rgb(255,255,255);">📋</font>****<font style="color:rgb(255,255,255);"> workspace_member</font>** | | |
| **<font style="color:rgb(255,255,255);">字段名</font>** | **<font style="color:rgb(255,255,255);">类型</font>****<font style="color:rgb(255,255,255);">/约束</font>** | **<font style="color:rgb(255,255,255);">说明</font>** |
| workspace | **FK → workspace** | 外键关联 |
| user | **FK → user** | 外键关联 |
| role | owner/admin/member/guest | owner/admin/member/guest |
| joined_at | TIMESTAMP | 时间戳 |
| **<font style="color:rgb(255,255,255);">📋</font>****<font style="color:rgb(255,255,255);"> api_token</font>** | | |
| **<font style="color:rgb(255,255,255);">字段名</font>** | **<font style="color:rgb(255,255,255);">类型</font>****<font style="color:rgb(255,255,255);">/约束</font>** | **<font style="color:rgb(255,255,255);">说明</font>** |
| workspace | **FK → workspace** | 外键关联 |
| user | **FK → user** | 外键关联 |
| name | VARCHAR(255) | 名称/标识 |
| token_hash | TEXT |  |
| scopes | JSON | JSON对象 |
| expires_at | TIMESTAMP | 时间戳 |
| last_used_at | TIMESTAMP | 时间戳 |


  


# **<font style="color:rgb(31,78,121);">8.2 项目管理 (Project)</font>**
项目是Ydsz Plane 系统中最核心的工作容器，承载需求/任务/缺陷/迭代/版本等。每个项目拥有独立的工作流、模块、类别配置。对标 Jira Project、云效项目、TAPD 项目。

## **<font style="color:rgb(46,117,182);">2.1 功能概述</font>**
项目是具体产品的研发或交付单元，在工作空间下创建。项目包含工作流、模块、迭代、版本等配置，以及需求/任务/缺陷/文档等。

## **<font style="color:rgb(46,117,182);">2.2 竞品对标</font>**
| **<font style="color:rgb(255,255,255);">功能点</font>** | **<font style="color:rgb(255,255,255);">Jira Project</font>** | **<font style="color:rgb(255,255,255);">云效项目</font>** | **<font style="color:rgb(255,255,255);">TAPD 项目</font>** | **<font style="color:rgb(255,255,255);">ONES 项目</font>** | **<font style="color:rgb(255,255,255);">Ydsz Plane</font>****<font style="color:rgb(255,255,255);"> </font>****<font style="color:rgb(255,255,255);">现有</font>** |
| --- | --- | --- | --- | --- | --- |
| <font style="color:rgb(51,51,51);">项目</font><font style="color:rgb(51,51,51);"> CRUD</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> |
| <font style="color:rgb(51,51,51);">Identifier 自动生成</font> | <font style="color:rgb(51,51,51);">● 唯一 Key</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> |
| <font style="color:rgb(51,51,51);">网络类型</font><font style="color:rgb(51,51,51);">(公开/私有)</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> |
| <font style="color:rgb(51,51,51);">模块化管理开关</font> | <font style="color:rgb(51,51,51);">● Modules</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">● 可配置</font> |
| <font style="color:rgb(51,51,51);">项目模板</font> | <font style="color:rgb(51,51,51);">● 多模板</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">● 敏捷/瀑布</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">● 规划中</font> |
| <font style="color:rgb(51,51,51);">自动归档</font><font style="color:rgb(51,51,51);">/关闭</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> |
| <font style="color:rgb(51,51,51);">项目复制</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">○</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">○ 规划中</font> |
| <font style="color:rgb(51,51,51);">项目分组</font><font style="color:rgb(51,51,51);">/分类</font> | <font style="color:rgb(51,51,51);">● Category</font> | <font style="color:rgb(51,51,51);">● 项目集</font> | <font style="color:rgb(51,51,51);">● 项目集</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">○ 规划中</font> |
| <font style="color:rgb(51,51,51);">Emoji/图标</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">○</font> | <font style="color:rgb(51,51,51);">○</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> |
| <font style="color:rgb(51,51,51);">封面图片</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">○</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> |


## **<font style="color:rgb(46,117,182);">2.3 用户故事</font>**
• 作为项目经理，我希望为每个产品/业务线创建独立的项目来隔离需求和工作流。

• 作为 Scrum Master，我希望复用已有的项目模板快速启动新项目。

• 作为团队负责人，我希望控制项目的网络类型（私有/公开）来管理可见性。

• 作为 PMO，我希望查看多个项目的汇总进度来评估资源分配。

## **<font style="color:rgb(46,117,182);">2.4 功能需求详述</font>**
### **<font style="color:rgb(51,51,51);">2.4.1 项目生命周期</font>**
| **<font style="color:rgb(255,255,255);">功能</font>** | **<font style="color:rgb(255,255,255);">详细描述</font>** | **<font style="color:rgb(255,255,255);">验收标准</font>** | **<font style="color:rgb(255,255,255);">优先级</font>** |
| --- | --- | --- | --- |
| <font style="color:rgb(51,51,51);">创建项目</font> | <font style="color:rgb(51,51,51);">支持输入名称、描述、网络类型、选择模板、上传封面</font><font style="color:rgb(51,51,51);">/Emoji，系统自动生成唯一 Identifier</font> | <font style="color:rgb(51,51,51);">创建成功返回项目详情页，</font><font style="color:rgb(51,51,51);">Identifier 唯一</font> | <font style="color:rgb(51,51,51);">P0</font> |
| <font style="color:rgb(51,51,51);">编辑项目</font> | <font style="color:rgb(51,51,51);">Owner/Admin 可修改项目基本信息、网络类型、封面、功能模块开关</font> | <font style="color:rgb(51,51,51);">修改后立即生效</font> | <font style="color:rgb(51,51,51);">P0</font> |
| <font style="color:rgb(51,51,51);">归档项目</font> | <font style="color:rgb(51,51,51);">非活跃项目可归档，归档后只读不可新建</font><font style="color:rgb(51,51,51);">/编辑需求/任务/缺陷，可恢复</font> | <font style="color:rgb(51,51,51);">归档后需求/任务/缺陷列表只读</font> | <font style="color:rgb(51,51,51);">P1</font> |
| <font style="color:rgb(51,51,51);">删除项目</font> | <font style="color:rgb(51,51,51);">仅无任何需求/任务/缺陷时可删除，需二次确认</font> | <font style="color:rgb(51,51,51);">删除后无法恢复</font> | <font style="color:rgb(51,51,51);">P2</font> |
| <font style="color:rgb(51,51,51);">项目列表</font> | <font style="color:rgb(51,51,51);">按最近活跃</font><font style="color:rgb(51,51,51);">/名称排序，支持搜索和筛选（归档状态）</font> | <font style="color:rgb(51,51,51);">显示封面、名称、</font><font style="color:rgb(51,51,51);">Identifier、成员数</font> | <font style="color:rgb(51,51,51);">P0</font> |
| <font style="color:rgb(51,51,51);">项目模板</font> | <font style="color:rgb(51,51,51);">提供敏捷</font><font style="color:rgb(51,51,51);">/瀑布/通用模板，模板预设工作流和模块</font> | <font style="color:rgb(51,51,51);">基于模板创建项目时自动复制配置</font> | <font style="color:rgb(51,51,51);">P1</font> |


### **<font style="color:rgb(51,51,51);">2.4.2 功能模块开关</font>**
19. Intake（收件箱）：启用后用户可提交 Intake 提报，由管理员转正进入需求/任务/缺陷。

20. Sprint（迭代）：启用后可创建迭代管理 Sprint。

21. Version（版本）：启用后可创建版本管理发版。

22. Estimate（估算）：启用后可配置故事点/工时估算体系。

23. 每个模块可独立开关，关闭后隐藏相关菜单和功能。

## **<font style="color:rgb(46,117,182);">2.5 交互流程</font>**
24. 创建项目：工作空间→「新建项目」→ 填写名称/Identifier/网络类型 → 选择模板 → 提交

25. 设置网络类型：项目设置→「网络类型」→ 选择 私有/公开/内部 → 保存

26. 归档项目：项目设置→「归档项目」→ 确认 → 项目状态变为 archived

## **<font style="color:rgb(46,117,182);">2.6 数据模型设计</font>**
| **<font style="color:rgb(255,255,255);">📋</font>****<font style="color:rgb(255,255,255);"> project</font>** | | |
| --- | --- | --- |
| **<font style="color:rgb(255,255,255);">字段名</font>** | **<font style="color:rgb(255,255,255);">类型</font>****<font style="color:rgb(255,255,255);">/约束</font>** | **<font style="color:rgb(255,255,255);">说明</font>** |
| id | **PK (auto)** | 主键自增ID |
| workspace | **FK → workspace** | 外键关联 |
| name | VARCHAR(255) | 名称/标识 |
| identifier | TEXT |  |
| description | VARCHAR(255) | 名称/标识 |
| network | private/public/internal | private/public/internal |
| status | active/archived | active/archived |
| lead | **FK → lead** | 外键关联 |
| logo_props | TEXT |  |
| created_at | TIMESTAMP | 时间戳 |
| **<font style="color:rgb(255,255,255);">📋</font>****<font style="color:rgb(255,255,255);"> project_member</font>** | | |
| **<font style="color:rgb(255,255,255);">字段名</font>** | **<font style="color:rgb(255,255,255);">类型</font>****<font style="color:rgb(255,255,255);">/约束</font>** | **<font style="color:rgb(255,255,255);">说明</font>** |
| project | **FK → project** | 外键关联 |
| user | **FK → user** | 外键关联 |
| role | TEXT |  |
| joined_at | TIMESTAMP | 时间戳 |
| **<font style="color:rgb(255,255,255);">📋</font>****<font style="color:rgb(255,255,255);"> project_module</font>** | | |
| **<font style="color:rgb(255,255,255);">字段名</font>** | **<font style="color:rgb(255,255,255);">类型</font>****<font style="color:rgb(255,255,255);">/约束</font>** | **<font style="color:rgb(255,255,255);">说明</font>** |
| project | **FK → project** | 外键关联 |
| module | intake/sprint/version/estimate | intake/sprint/version/estimate |
| enabled | boolean | boolean |


  


# **<font style="color:rgb(31,78,121);">8.3 版本管理 (Version)</font>**
版本是产品发版的里程碑容器，包含 1~N 个迭代周期。Version 是面向市场的发版节奏，迭代是面向团队的开发节奏。对标云效发布计划、ONES 版本、Jira Fix Version。

## **<font style="color:rgb(46,117,182);">3.1 功能概述</font>**
版本是产品发版的里程碑，具有独立的交付目标、准出标准、检查清单和 Release Notes。版本包含 1~N 个迭代，支持进度聚合、变更日志和交付报告。

## **<font style="color:rgb(46,117,182);">3.2 竞品对标</font>**
| **<font style="color:rgb(255,255,255);">功能点</font>** | **<font style="color:rgb(255,255,255);">云效发布计划</font>** | **<font style="color:rgb(255,255,255);">ONES 版本</font>** | **<font style="color:rgb(255,255,255);">Jira Fix Version</font>** | **<font style="color:rgb(255,255,255);">TAPD 发布</font>** | **<font style="color:rgb(255,255,255);">Ydsz Plane</font>****<font style="color:rgb(255,255,255);"> </font>****<font style="color:rgb(255,255,255);">规划</font>** |
| --- | --- | --- | --- | --- | --- |
| <font style="color:rgb(51,51,51);">版本创建</font><font style="color:rgb(51,51,51);">/编辑</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">△ 规划中</font> |
| <font style="color:rgb(51,51,51);">多迭代聚合</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">○</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">△</font> |
| <font style="color:rgb(51,51,51);">进度聚合视图</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">△</font> |
| <font style="color:rgb(51,51,51);">变更日志自动生成</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">○</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">△</font> |
| <font style="color:rgb(51,51,51);">发布检查清单</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">○</font> | <font style="color:rgb(51,51,51);">○</font> | <font style="color:rgb(51,51,51);">○</font> |
| <font style="color:rgb(51,51,51);">准出率指标</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">○</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">○</font> |
| <font style="color:rgb(51,51,51);">发布模板</font> | <font style="color:rgb(51,51,51);">● 常规/热修复/大版本</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">○</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">○</font> |
| <font style="color:rgb(51,51,51);">版本号语义化校验</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">○</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">○</font> |


## **<font style="color:rgb(46,117,182);">3.3 用户故事</font>**
• 作为产品经理，我希望创建版本来管理产品发版的里程碑和交付目标。

• 作为发布经理，我希望查看版本的聚合进度来评估是否按期发布。

• 作为 PMO，我希望对比多个版本的交付质量和准时率。

## **<font style="color:rgb(46,117,182);">3.4 功能需求详述</font>**
### **<font style="color:rgb(51,51,51);">3.4.1 版本生命周期</font>**
| **<font style="color:rgb(255,255,255);">功能</font>** | **<font style="color:rgb(255,255,255);">详细描述</font>** | **<font style="color:rgb(255,255,255);">验收标准</font>** | **<font style="color:rgb(255,255,255);">优先级</font>** |
| --- | --- | --- | --- |
| <font style="color:rgb(51,51,51);">创建版本</font> | <font style="color:rgb(51,51,51);">输入版本号（语义化校验）、名称、描述、发布日期、负责人、模板类型</font> | <font style="color:rgb(51,51,51);">创建成功，版本号格式符合</font><font style="color:rgb(51,51,51);"> semver</font> | <font style="color:rgb(51,51,51);">P0</font> |
| <font style="color:rgb(51,51,51);">编辑版本</font> | <font style="color:rgb(51,51,51);">修改基本信息、发布日期、负责人、描述</font> | <font style="color:rgb(51,51,51);">修改后关联迭代自动更新</font> | <font style="color:rgb(51,51,51);">P0</font> |
| <font style="color:rgb(51,51,51);">发布版本</font> | <font style="color:rgb(51,51,51);">流转状态为「已发布」，自动生成</font><font style="color:rgb(51,51,51);"> Release Notes 和交付报告</font> | <font style="color:rgb(51,51,51);">已发布后只读，不可删除</font> | <font style="color:rgb(51,51,51);">P0</font> |
| <font style="color:rgb(51,51,51);">归档版本</font> | <font style="color:rgb(51,51,51);">旧版本可归档，归档后隐藏于主视图但可搜索查看</font> | <font style="color:rgb(51,51,51);">归档后视图默认隐藏</font> | <font style="color:rgb(51,51,51);">P1</font> |
| <font style="color:rgb(51,51,51);">版本列表</font> | <font style="color:rgb(51,51,51);">按发布日期倒序，支持按状态筛选</font> | <font style="color:rgb(51,51,51);">显示版本号、名称、发布日期、进度</font><font style="color:rgb(51,51,51);">%</font> | <font style="color:rgb(51,51,51);">P0</font> |


### **<font style="color:rgb(51,51,51);">3.4.2 聚合视图与报告</font>**
27. 进度聚合：自动汇总所有关联迭代的需求/任务/缺陷完成度，显示完成故事点/总故事点。

28. 变更日志：自动汇总版本需求变更记录，生成结构化 Release Notes（可按模板定制）。

29. 交付报告：发布时自动生成，包含缺陷数、通过率、准出率、7日留存、30日留存等指标。

30. 检查清单：发布前必填的验证项（如文档完备性、代码冻结、回归通过等）。

## **<font style="color:rgb(46,117,182);">3.5 交互流程</font>**
31. 创建版本：项目→ 版本模块 →「新建版本」→ 填写版本号/发布日期/目标 → 选择关联迭代 → 提交

32. 发布版本：版本详情→「发布」→ 确认检查清单 → 生成 Release Notes → 流转状态

33. 查看进度：项目仪表盘→ 版本卡片 → 点击版本号 → 聚合视图

## **<font style="color:rgb(46,117,182);">3.6 数据模型设计</font>**
| **<font style="color:rgb(255,255,255);">📋</font>****<font style="color:rgb(255,255,255);"> version</font>** | | |
| --- | --- | --- |
| **<font style="color:rgb(255,255,255);">字段名</font>** | **<font style="color:rgb(255,255,255);">类型</font>****<font style="color:rgb(255,255,255);">/约束</font>** | **<font style="color:rgb(255,255,255);">说明</font>** |
| id | **PK (auto)** | 主键自增ID |
| project | **FK → project** | 外键关联 |
| name | VARCHAR(255) | 名称/标识 |
| version_number | TEXT |  |
| description | VARCHAR(255) | 名称/标识 |
| status | planning/active/released/archived | planning/active/released/archived |
| release_date | TIMESTAMP | 时间戳 |
| owner | **FK → owner** | 外键关联 |
| check_list | JSON | JSON对象 |
| progress_snapshot | JSON | JSON对象 |
| created_at | TIMESTAMP | 时间戳 |
| **<font style="color:rgb(255,255,255);">📋</font>****<font style="color:rgb(255,255,255);"> version_</font>****<font style="color:rgb(255,255,255);">sprint</font>** | | |
| **<font style="color:rgb(255,255,255);">字段名</font>** | **<font style="color:rgb(255,255,255);">类型</font>****<font style="color:rgb(255,255,255);">/约束</font>** | **<font style="color:rgb(255,255,255);">说明</font>** |
| version | **FK → version** | 外键关联 |
| sprint | **FK → ****sprint** | 外键关联 |


  


# **<font style="color:rgb(31,78,121);">8.4 迭代管理 (</font>****<font style="color:rgb(31,78,121);">Sprint</font>****<font style="color:rgb(31,78,121);">)</font>**
迭代是固定周期的开发单元（通常 1~4 周），团队在迭代中承诺完成一组需求/任务/缺陷。对标 Jira Sprint、云效迭代、ONES Sprint。

## **<font style="color:rgb(46,117,182);">4.1 功能概述</font>**
迭代是敏捷开发的基本时间盒，团队在每个迭代中承诺完成一组需求/任务/缺陷。迭代包含规划、执行、复盘三个阶段，支持容量规划、站会、燃尽图和速率分析。

## **<font style="color:rgb(46,117,182);">4.2 竞品对标</font>**
| **<font style="color:rgb(255,255,255);">功能点</font>** | **<font style="color:rgb(255,255,255);">Jira Sprint</font>** | **<font style="color:rgb(255,255,255);">云效迭代</font>** | **<font style="color:rgb(255,255,255);">ONES Sprint</font>** | **<font style="color:rgb(255,255,255);">TAPD 迭代</font>** | **<font style="color:rgb(255,255,255);">Ydsz Plane</font>****<font style="color:rgb(255,255,255);"> </font>****<font style="color:rgb(255,255,255);">现有</font>** |
| --- | --- | --- | --- | --- | --- |
| <font style="color:rgb(51,51,51);">迭代</font><font style="color:rgb(51,51,51);"> CRUD</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> |
| <font style="color:rgb(51,51,51);">容量规划</font><font style="color:rgb(51,51,51);">(人天)</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">○</font> |
| <font style="color:rgb(51,51,51);">迭代目标</font> | <font style="color:rgb(51,51,51);">● Goal</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> |
| <font style="color:rgb(51,51,51);">拖拽规划</font><font style="color:rgb(51,51,51);">(Backlog→Sprint)</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> |
| <font style="color:rgb(51,51,51);">快速开始</font><font style="color:rgb(51,51,51);">/强制结束</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> |
| <font style="color:rgb(51,51,51);">迭代快照留存</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">● 自动</font> |
| <font style="color:rgb(51,51,51);">燃尽图</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> |
| <font style="color:rgb(51,51,51);">燃起图</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> |
| <font style="color:rgb(51,51,51);">累积流图</font><font style="color:rgb(51,51,51);"> (CFD)</font> | <font style="color:rgb(51,51,51);">● 插件</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">○ 规划中</font> |
| <font style="color:rgb(51,51,51);">迭代速率趋势</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> |
| <font style="color:rgb(51,51,51);">未完成任务自动处理</font> | <font style="color:rgb(51,51,51);">● 下一 Sprint</font> | <font style="color:rgb(51,51,51);">● 退回 Backlog</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> |
| <font style="color:rgb(51,51,51);">站会集成</font> | <font style="color:rgb(51,51,51);">○ 第三方</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">○ 规划中</font> |


## **<font style="color:rgb(46,117,182);">4.3 用户故事</font>**
• 作为 Scrum Master，我希望在迭代规划时从 Backlog 拖拽需求/任务/缺陷进入 Sprint。

• 作为团队成员，我希望能看到今天的迭代进度和我的剩余工作。

• 作为项目经理，我希望能分析团队的历史速率来指导未来迭代规划。

## **<font style="color:rgb(46,117,182);">4.4 功能需求详述</font>**
### **<font style="color:rgb(51,51,51);">4.4.1 迭代规划</font>**
34. 创建迭代：输入名称、描述、起止日期、目标、容量（可选，基于团队人天）。

35. 规划 Backlog 入 Sprint：从需求池拖拽需求/任务/缺陷进入迭代，自动汇总故事点数。

36. 容量校验：设置容量后，拖拽时显示饱和度（总故事点/容量），超额高亮警告。

37. 迭代冲突校验：同一需求/任务/缺陷不可同时存在于两个活跃迭代中（配置化开关）。

38. 团队速率建议：基于近 3~6 个 Sprint 的平均速率，推荐本次迭代承诺范围。

### **<font style="color:rgb(51,51,51);">4.4.2 迭代执行</font>**
39. 每日进度：迭代卡片显示已完成故事点/总故事点百分比。

40. 站会模式：迭代详情一键进入站会视图，按「昨日完成/今日计划/阻塞」分组展示。

41. 迭代快照：每日 00:00 自动保存迭代状态快照，用于燃尽图回溯。

42. 中途加项：迭代执行中允许 Admin 加入紧急需求，标记为「中途加入」并统计对速率的影响。

### **<font style="color:rgb(51,51,51);">4.4.3 迭代复盘</font>**
43. 强制结束：迭代到期自动提醒，Scrum Master 可强制结束（未完成任务退回 Backlog 或推入下一迭代）。

44. 自动复盘：结束时自动生成复盘数据（承诺故事点/完成故事点/完成率/加入/移除)。

45. 速率计算：本次完成故事点自动计入团队速率趋势图。

46. 迭代复盘文档：基于模板自动生成复盘报告，可指定评审人。

## **<font style="color:rgb(46,117,182);">4.5 交互流程</font>**
47. 迭代启动：项目→ Sprint 模块→「新建迭代」→ 填写名称/起止日期/目标 → 从 Backlog 拖拽需求/任务/缺陷 → 启动

48. 站会模式：迭代详情→ 点击「站会」→ 进入站会视图 → 成员更新状态

49. 迭代结束：迭代详情→「结束迭代」→ 选择未完成任务处理方式 → 生成复盘报告

## **<font style="color:rgb(46,117,182);">4.6 数据模型设计</font>**
| **<font style="color:rgb(255,255,255);">📋</font>****<font style="color:rgb(255,255,255);"> </font>****<font style="color:rgb(255,255,255);">sprint</font>** | | |
| --- | --- | --- |
| **<font style="color:rgb(255,255,255);">字段名</font>** | **<font style="color:rgb(255,255,255);">类型</font>****<font style="color:rgb(255,255,255);">/约束</font>** | **<font style="color:rgb(255,255,255);">说明</font>** |
| id | **PK (auto)** | 主键自增ID |
| project | **FK → project** | 外键关联 |
| name | VARCHAR(255) | 名称/标识 |
| description | VARCHAR(255) | 名称/标识 |
| start_date | TIMESTAMP | 时间戳 |
| end_date | TIMESTAMP | 时间戳 |
| status | active/completed | active/completed |
| capacity | TEXT |  |
| goal | TEXT |  |
| owner | **FK → owner** | 外键关联 |
| version | **FK → version** | 外键关联 |
| viewport | JSON | JSON对象 |
| **<font style="color:rgb(255,255,255);">📋</font>****<font style="color:rgb(255,255,255);"> </font>****<font style="color:rgb(255,255,255);">sprint</font>****<font style="color:rgb(255,255,255);">_issue</font>** | | |
| **<font style="color:rgb(255,255,255);">字段名</font>** | **<font style="color:rgb(255,255,255);">类型</font>****<font style="color:rgb(255,255,255);">/约束</font>** | **<font style="color:rgb(255,255,255);">说明</font>** |
| sprint | **FK → ****sprint** | 外键关联 |
| issue | **FK → issue** | 外键关联 |
| **<font style="color:rgb(255,255,255);">📋</font>****<font style="color:rgb(255,255,255);"> </font>****<font style="color:rgb(255,255,255);">sprint</font>****<font style="color:rgb(255,255,255);">_snapshot</font>** | | |
| **<font style="color:rgb(255,255,255);">字段名</font>** | **<font style="color:rgb(255,255,255);">类型</font>****<font style="color:rgb(255,255,255);">/约束</font>** | **<font style="color:rgb(255,255,255);">说明</font>** |
| sprint | **FK → ****sprint** | 外键关联 |
| snapshot_date | TIMESTAMP | 时间戳 |
| data | JSON | JSON对象 |


  


# **<font style="color:rgb(31,78,121);">8.5 需求管理 (Requirement)</font>**
需求由产品经理提出，描述产品需要实现的功能或解决的问题。需求可拆分为子需求（Epic→Feature→Story 模式），需求不产生任务，需求测试时可能产生缺陷。对标 Jira Epic/Story、云效需求、ONES 需求。

## **<font style="color:rgb(46,117,182);">8.5 功能概述</font>**
需求管理是面向产品经理的产品待办事项池管理模块，支持需求的创建、编辑、拆分、评审和跟踪闭环。对标 Jira Issues, ONES 需求池, 云效需求。

## **<font style="color:rgb(46,117,182);">5.1 竞品对标</font>**
| **<font style="color:rgb(255,255,255);">功能点</font>** | **<font style="color:rgb(255,255,255);">Jira</font>** | **<font style="color:rgb(255,255,255);">云效</font>** | **<font style="color:rgb(255,255,255);">ONES</font>** | **<font style="color:rgb(255,255,255);">TAPD</font>** | **<font style="color:rgb(255,255,255);">Ydsz Plane</font>****<font style="color:rgb(255,255,255);"> </font>****<font style="color:rgb(255,255,255);">现有</font>** |
| --- | --- | --- | --- | --- | --- |
| <font style="color:rgb(51,51,51);">需求</font><font style="color:rgb(51,51,51);"> CRUD</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> |
| <font style="color:rgb(51,51,51);">Epic→Feature→Story 三层 WBS</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> |
| <font style="color:rgb(51,51,51);">多子类型</font><font style="color:rgb(51,51,51);"> (Story/Epic/Feature)</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> |
| <font style="color:rgb(51,51,51);">优先级矩阵</font><font style="color:rgb(51,51,51);">(价值/复杂度)</font> | <font style="color:rgb(51,51,51);">○ 插件</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">○</font> |
| <font style="color:rgb(51,51,51);">需求评审工作流</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">○ 规划中</font> |
| <font style="color:rgb(51,51,51);">需求来源标记</font> | <font style="color:rgb(51,51,51);">○ 自定义字段</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">○</font> |
| <font style="color:rgb(51,51,51);">关联缺陷一键创建</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> |
| <font style="color:rgb(51,51,51);">需求进度自动汇总</font> | <font style="color:rgb(51,51,51);">● 子需求/子任务</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> |
| <font style="color:rgb(51,51,51);">需求复制</font><font style="color:rgb(51,51,51);">/移动</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> |
| <font style="color:rgb(51,51,51);">需求模板</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">○ 规划中</font> |


## **<font style="color:rgb(46,117,182);">5.2 用户故事</font>**
• 作为产品经理，我需要创建和编辑需求，为需求分配优先级和模块。

• 作为产品经理，我希望将大的 Epic 拆分为多个 Feature 和 Story 来细化需求。

• 作为需求评审者，我希望能对需求进行评审打分和状态流转。

• 作为测试人员，我希望从需求详情一键创建关联缺陷。

• 作为团队成员，我希望能看到需求的历史变更和关联的需求/任务/缺陷。

## **<font style="color:rgb(46,117,182);">5.3 功能需求详述</font>**
### **<font style="color:rgb(51,51,51);">5.3.1 需求 CRUD</font>**
| **<font style="color:rgb(255,255,255);">功能</font>** | **<font style="color:rgb(255,255,255);">详细描述</font>** | **<font style="color:rgb(255,255,255);">验收标准</font>** | **<font style="color:rgb(255,255,255);">优先级</font>** |
| --- | --- | --- | --- |
| <font style="color:rgb(51,51,51);">创建需求</font> | <font style="color:rgb(51,51,51);">填写标题、描述、优先级、模块、指派人、标签、开始</font><font style="color:rgb(51,51,51);">/截止日期、故事点</font> | <font style="color:rgb(51,51,51);">标题默认必填，其他字段根据项目配置可选</font><font style="color:rgb(51,51,51);">/必填</font> | <font style="color:rgb(51,51,51);">P0</font> |
| <font style="color:rgb(51,51,51);">编辑需求</font> | <font style="color:rgb(51,51,51);">修改所有字段，支持富文本描述（图片</font><font style="color:rgb(51,51,51);">/表格/链接），变更记录到活动日志</font> | <font style="color:rgb(51,51,51);">编辑后活动日志同步更新</font> | <font style="color:rgb(51,51,51);">P0</font> |
| <font style="color:rgb(51,51,51);">删除需求</font> | <font style="color:rgb(51,51,51);">软删除，有子需求时需级联删除或转移父级，需二次确认</font> | <font style="color:rgb(51,51,51);">删除后进入回收站可恢复</font> | <font style="color:rgb(51,51,51);">P1</font> |
| <font style="color:rgb(51,51,51);">复制需求</font> | <font style="color:rgb(51,51,51);">复制当前需求（可连同子需求一起复制），新项目支持跨项目复制</font> | <font style="color:rgb(51,51,51);">复制后</font><font style="color:rgb(51,51,51);"> Identifier 重新生成</font> | <font style="color:rgb(51,51,51);">P1</font> |
| <font style="color:rgb(51,51,51);">移动需求</font> | <font style="color:rgb(51,51,51);">在同一项目内移动到其他模块</font><font style="color:rgb(51,51,51);">/父需求，跨项目移动需求/任务/缺陷</font> | <font style="color:rgb(51,51,51);">移动后关联关系保留</font> | <font style="color:rgb(51,51,51);">P1</font> |
| <font style="color:rgb(51,51,51);">归档需求</font> | <font style="color:rgb(51,51,51);">已关闭且长期无变化的需求自动归档，归档后隐藏于主视图</font> | <font style="color:rgb(51,51,51);">可手动取消归档</font> | <font style="color:rgb(51,51,51);">P2</font> |


### **<font style="color:rgb(51,51,51);">5.3.2 WBS 层级管理</font>**
50. 创建子需求：在需求详情页点击「添加子需求」，支持三级层级（Epic→Feature→Story）。

51. 层级校验：同一需求不可同时作为自己和自己的父需求（防止循环依赖）。

52. 层级展示：支持树形视图和扁平化视图切换，树形视图支持折叠/展开。

53. 层级进度：父需求进度 = sum(子需求完成故事点) / sum(子需求总故事点)。

54. 层级删除：删除父需求时提示是否级联删除子需求或转移子需求至其他父级。

### **<font style="color:rgb(51,51,51);">5.3.3 需求评审工作流</font>**
55. 草稿→ 评审中：PM 填写完毕后提交评审，通知评审人。

56. 评审中→ 已采纳/已拒绝：评审人可打分、评论，达标后流转为「已采纳」。

57. 已采纳→ 已开发：团队承诺开发后流转，自动通知团队。

58. 已开发→ 已验证：开发测试通过后流转，PM 确认发布。

59. 评审人可配置，支持多人评审，评审模板可自定义评审维度。

## **<font style="color:rgb(46,117,182);">5.4 交互流程</font>**
60. 创建需求：项目→ 需求模块 →「新建需求」→ 填写标题/描述/优先级 → 选择父需求（可选）→ 保存

61. 拆解 WBS：需求详情 →「添加子需求」→ 填写子需求信息 → 选择层级 → 保存

62. 提关联缺陷：需求详情→「创建关联缺陷」→ 系统自动关联当前需求 ID → 保存

## **<font style="color:rgb(46,117,182);">5.5 数据模型设计</font>**
> 注：本节数据模型为早期草稿，实际已按需求/任务/缺陷三独立业务表拆分，字段与关系表（relation/activity 等）定义以《Ydsz Plane 数据库表设计》为准。
| **<font style="color:rgb(255,255,255);">📋</font>****<font style="color:rgb(255,255,255);"> requirement</font>** | | |
| --- | --- | --- |
| **<font style="color:rgb(255,255,255);">字段名</font>** | **<font style="color:rgb(255,255,255);">类型</font>****<font style="color:rgb(255,255,255);">/约束</font>** | **<font style="color:rgb(255,255,255);">说明</font>** |
| id | **PK (auto)** | 主键自增ID |
| project | **FK → project** | 外键关联 |
| name | VARCHAR(255) | 名称/标识 |
| description_json/html/stripped | TEXT |  |
| state | **FK → state** | 外键关联 |
| priority | TEXT |  |
| type | **FK:requirement** | FK:requirement |
| parent | **FK:sub** | FK:sub |
| module | **FK → module** | 外键关联 |
| version | **FK → version** | 外键关联 |
| sprint | **FK → ****sprint** | 外键关联 |
| estimate_point | **FK → estimate_point** | 外键关联 |
| assignees | M2M → assignees | 多对多关联 |
| labels | M2M → labels | 多对多关联 |
| sequence_id | **FK → sequence** | 外键关联sequence |
| is_draft | BOOLEAN | 布尔标志 |
| created_at | TIMESTAMP | 时间戳 |
| updated_at | TIMESTAMP | 时间戳 |
| **<font style="color:rgb(255,255,255);">📋</font>****<font style="color:rgb(255,255,255);"> issue_relation</font>** | | |
| **<font style="color:rgb(255,255,255);">字段名</font>** | **<font style="color:rgb(255,255,255);">类型</font>****<font style="color:rgb(255,255,255);">/约束</font>** | **<font style="color:rgb(255,255,255);">说明</font>** |
| issue | **FK:source** | FK:source |
| related_issue | **FK:target** | FK:target |
| relation_type | duplicate/relates_to/blocked_by/start_before/finish_before/implemented_by | duplicate/relates_to/blocked_by/start_before/finish_before/implemented_by |
| **<font style="color:rgb(255,255,255);">📋</font>****<font style="color:rgb(255,255,255);"> issue_activity</font>** | | |
| **<font style="color:rgb(255,255,255);">字段名</font>** | **<font style="color:rgb(255,255,255);">类型</font>****<font style="color:rgb(255,255,255);">/约束</font>** | **<font style="color:rgb(255,255,255);">说明</font>** |
| issue | **FK → issue** | 外键关联 |
| verb | TEXT |  |
| field | TEXT |  |
| old_value | TEXT |  |
| new_value | TEXT |  |
| created_by | **FK → created_by** | 外键关联 |
| created_at | TIMESTAMP | 时间戳 |


  


# **<font style="color:rgb(31,78,121);">8.6 任务管理 (Task)</font>**
任务由技术经理提出，描述研发侧需要完成的具体任务。任务可拆分为子任务（主任务→子任务→子子任务）。任务与需求平行存在，不产生任务分解关系。对标 Jira Task、云效任务、ONES 任务。

## **<font style="color:rgb(46,117,182);">8.6 功能概述</font>**
任务管理是面向研发工程师的工作分解执行模块，支持任务创建、WBS 子任务拆分、工时管理、依赖管理。对标 Jira Task, ONES 任务。

## **<font style="color:rgb(46,117,182);">6.1 竞品对标</font>**
| **<font style="color:rgb(255,255,255);">功能点</font>** | **<font style="color:rgb(255,255,255);">Jira Task</font>** | **<font style="color:rgb(255,255,255);">云效任务</font>** | **<font style="color:rgb(255,255,255);">ONES 任务</font>** | **<font style="color:rgb(255,255,255);">TAPD 任务</font>** | **<font style="color:rgb(255,255,255);">Ydsz Plane</font>****<font style="color:rgb(255,255,255);"> </font>****<font style="color:rgb(255,255,255);">现有</font>** |
| --- | --- | --- | --- | --- | --- |
| <font style="color:rgb(51,51,51);">任务</font><font style="color:rgb(51,51,51);"> CRUD</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> |
| <font style="color:rgb(51,51,51);">三级</font><font style="color:rgb(51,51,51);"> WBS 子任务</font> | <font style="color:rgb(51,51,51);">● Sub-task</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> |
| <font style="color:rgb(51,51,51);">任务工时</font><font style="color:rgb(51,51,51);">(预估/实际/剩余)</font> | <font style="color:rgb(51,51,51);">● 内置</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">○ 规划中</font> |
| <font style="color:rgb(51,51,51);">任务依赖</font><font style="color:rgb(51,51,51);">(FS/SS/FF/SF)</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">○ 规划中</font> |
| <font style="color:rgb(51,51,51);">任务分类</font> | <font style="color:rgb(51,51,51);">○ 自定义</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">○</font> |
| <font style="color:rgb(51,51,51);">关关联需求</font> | <font style="color:rgb(51,51,51);">● 链接</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> |
| <font style="color:rgb(51,51,51);">关联代码提交</font> | <font style="color:rgb(51,51,51);">● 智能提交</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">○ 规划中</font> |
| <font style="color:rgb(51,51,51);">关联缺陷</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> |
| <font style="color:rgb(51,51,51);">工作量热力图成员视图</font> | <font style="color:rgb(51,51,51);">○ 插件</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">○ 规划中</font> |
| <font style="color:rgb(51,51,51);">跨项目复制</font><font style="color:rgb(51,51,51);">/移动</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> |


## **<font style="color:rgb(46,117,182);">6.2 用户故事</font>**
• 作为技术经理，我希望能创建和编辑任务，分配给团队成员。

• 作为技术经理，我希望将大任务拆解为多个子任务来细化工作分解。

• 作为开发工程师，我希望能记录我的实际工时来追踪投入。

• 作为团队成员，我希望能关联任务到对应的需求来追溯来源。

• 作为团队负责人，我希望能看到团队成员的工作量分布是否均衡。

## **<font style="color:rgb(46,117,182);">6.3 功能需求详述</font>**
### **<font style="color:rgb(51,51,51);">6.3.1 任务 CRUD</font>**
| **<font style="color:rgb(255,255,255);">功能</font>** | **<font style="color:rgb(255,255,255);">详细描述</font>** | **<font style="color:rgb(255,255,255);">验收标准</font>** | **<font style="color:rgb(255,255,255);">优先级</font>** |
| --- | --- | --- | --- |
| <font style="color:rgb(51,51,51);">创建任务</font> | <font style="color:rgb(51,51,51);">填写标题、描述、优先级、指派人、标签、故事点、开始</font><font style="color:rgb(51,51,51);">/截止日期</font> | <font style="color:rgb(51,51,51);">标题必填，其他字段根据项目配置可选</font><font style="color:rgb(51,51,51);">/必填</font> | <font style="color:rgb(51,51,51);">P0</font> |
| <font style="color:rgb(51,51,51);">编辑任务</font> | <font style="color:rgb(51,51,51);">修改所有字段，记录变更到活动日志，支持富文本描述</font> | <font style="color:rgb(51,51,51);">编辑后日志同步更新</font> | <font style="color:rgb(51,51,51);">P0</font> |
| <font style="color:rgb(51,51,51);">删除任务</font> | <font style="color:rgb(51,51,51);">软删除，有子任务时提示级联删除或转移</font> | <font style="color:rgb(51,51,51);">删除后进入回收站可恢复</font> | <font style="color:rgb(51,51,51);">P1</font> |
| <font style="color:rgb(51,51,51);">批量创建</font> | <font style="color:rgb(51,51,51);">支持批量创建任务（如从</font><font style="color:rgb(51,51,51);"> Excel 导入模板）</font> | <font style="color:rgb(51,51,51);">显示导入统计（成功</font><font style="color:rgb(51,51,51);">/失败/跳过）</font> | <font style="color:rgb(51,51,51);">P1</font> |
| <font style="color:rgb(51,51,51);">批量操作</font> | <font style="color:rgb(51,51,51);">批量更新状态、指派人、标签、迭代</font> | <font style="color:rgb(51,51,51);">原子操作，部分失败时回滚或提示详情</font> | <font style="color:rgb(51,51,51);">P1</font> |


### **<font style="color:rgb(51,51,51);">6.3.2 工时管理</font>**
63. 预估工时：创建任务时填写预估故事点或人时。

64. 实际工时：开发过程中记录实际投入时间（支持手动记录或计时器）。

65. 剩余工时：手动更新系统自动计算（剩余 = 预估 - 实际），也可由开发者手动修正。

66. 工时日志：记录每个成员在任务上的时间投入，支持按日/周汇总。

67. 工时偏差分析：任务完成时自动对比预估/实际，计入效能报告。

### **<font style="color:rgb(51,51,51);">6.3.3 任务依赖</font>**
68. FS (Finish-to-Start)：前置任务完成后才能开始后续任务（默认）。

69. SS (Start-to-Start)：前置任务开始时才能开始后续任务。

70. FF (Finish-to-Finish)：前置任务完成时后续任务也必须完成。

71. SF (Start-to-Finish)：前置任务开始时后续任务必须完成（较少用）。

72. 甘特图展示：依赖关系在甘特图上以箭头展示。

73. 循环依赖校验：系统检测并阻止循环依赖。

## **<font style="color:rgb(46,117,182);">6.4 交互流程</font>**
74. 创建任务：项目→ 任务模块 →「新建任务」→ 填写标题/描述/工时 → 保存

75. 记录工时：任务详情→「记录工时」→ 填写日期/工时/描述 → 保存

76. 设置依赖：任务详情→「添加依赖」→ 选择前置任务 → 选择依赖类型 → 保存

## **<font style="color:rgb(46,117,182);">6.5 数据模型设计</font>**
> 注：本节数据模型为早期草稿，实际已按需求/任务/缺陷三独立业务表拆分，字段与关系表（dependency/relation/activity 等）定义以《Ydsz Plane 数据库表设计》为准。
| **<font style="color:rgb(255,255,255);">📋</font>****<font style="color:rgb(255,255,255);"> task</font>** | | |
| --- | --- | --- |
| **<font style="color:rgb(255,255,255);">字段名</font>** | **<font style="color:rgb(255,255,255);">类型</font>****<font style="color:rgb(255,255,255);">/约束</font>** | **<font style="color:rgb(255,255,255);">说明</font>** |
| id | **PK (auto)** | 主键自增ID |
| project | **FK → project** | 外键关联 |
| name | VARCHAR(255) | 名称/标识 |
| description_json/html/stripped | TEXT |  |
| state | **FK → state** | 外键关联 |
| priority | TEXT |  |
| type | **FK:task** | FK:task |
| parent | **FK:sub** | FK:sub |
| estimate_point | TEXT |  |
| actual_effort | TEXT |  |
| remaining_effort | TEXT |  |
| module | **FK → module** | 外键关联 |
| sprint | **FK → ****sprint** | 外键关联 |
| assignees | M2M → assignees | 多对多关联 |
| labels | M2M → labels | 多对多关联 |
| sequence_id | **FK → sequence** | 外键关联sequence |
| **<font style="color:rgb(255,255,255);">📋</font>****<font style="color:rgb(255,255,255);"> issue_dependency</font>** | | |
| **<font style="color:rgb(255,255,255);">字段名</font>** | **<font style="color:rgb(255,255,255);">类型</font>****<font style="color:rgb(255,255,255);">/约束</font>** | **<font style="color:rgb(255,255,255);">说明</font>** |
| issue | **FK:dependent** | FK:dependent |
| depends_on | **FK → depends_on** | 外键关联 |
| dependency_type | FS/SS/FF/SF | FS/SS/FF/SF |
| lag_days | TEXT |  |
| **<font style="color:rgb(255,255,255);">📋</font>****<font style="color:rgb(255,255,255);"> time_log</font>** | | |
| **<font style="color:rgb(255,255,255);">字段名</font>** | **<font style="color:rgb(255,255,255);">类型</font>****<font style="color:rgb(255,255,255);">/约束</font>** | **<font style="color:rgb(255,255,255);">说明</font>** |
| issue | **FK → issue** | 外键关联 |
| user | **FK → user** | 外键关联 |
| date | TEXT |  |
| hours | TEXT |  |
| description | VARCHAR(255) | 名称/标识 |
| created_at | TIMESTAMP | 时间戳 |


  


# **<font style="color:rgb(31,78,121);">8.7 缺陷管理 (Defect)</font>**
缺陷由测试经理提交，记录软件中发现的异常或功能偏差。缺陷可由需求或任务关联产生，也可独立提交（线上反馈、客户投诉等）。缺陷可拆分为子缺陷。对标 Jira Defect、云效缺陷、ONES 缺陷、TAPD Defect。

## **<font style="color:rgb(46,117,182);">8.7 功能概述</font>**
缺陷管理是面向 QA/测试工程师的质量保障模块，支持缺陷提交、跟踪、状态流转、根因分析。对标 Jira Defect, ONES 缺陷, 云效缺陷。

## **<font style="color:rgb(46,117,182);">7.1 竞品对标</font>**
| **<font style="color:rgb(255,255,255);">功能点</font>** | **<font style="color:rgb(255,255,255);">Jira </font>****<font style="color:rgb(255,255,255);">Defect</font>** | **<font style="color:rgb(255,255,255);">云效缺陷</font>** | **<font style="color:rgb(255,255,255);">ONES 缺陷</font>** | **<font style="color:rgb(255,255,255);">TAPD </font>****<font style="color:rgb(255,255,255);">Defect</font>** | **<font style="color:rgb(255,255,255);">Ydsz Plane</font>****<font style="color:rgb(255,255,255);"> </font>****<font style="color:rgb(255,255,255);">现有</font>** |
| --- | --- | --- | --- | --- | --- |
| <font style="color:rgb(51,51,51);">缺陷</font><font style="color:rgb(51,51,51);"> CRUD</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> |
| <font style="color:rgb(51,51,51);">严重程度</font><font style="color:rgb(51,51,51);"> 5 级</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">○ 需补充</font> |
| <font style="color:rgb(51,51,51);">优先级</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> |
| <font style="color:rgb(51,51,51);">缺陷类型</font> | <font style="color:rgb(51,51,51);">○ 自定义</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">○</font> |
| <font style="color:rgb(51,51,51);">发现阶段</font> | <font style="color:rgb(51,51,51);">○ 自定义</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">○</font> |
| <font style="color:rgb(51,51,51);">发现</font><font style="color:rgb(51,51,51);">/修复版本</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">○ 规划中</font> |
| <font style="color:rgb(51,51,51);">复现步骤</font><font style="color:rgb(51,51,51);">+期望/实际结果</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">○</font> |
| <font style="color:rgb(51,51,51);">一键提</font><font style="color:rgb(51,51,51);">Defect</font><font style="color:rgb(51,51,51);">(关联需求/任务)</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> |
| <font style="color:rgb(51,51,51);">缺陷状态机</font> | <font style="color:rgb(51,51,51);">● 可配置</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">○ 规划中</font> |
| <font style="color:rgb(51,51,51);">缺陷龄分析</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">○</font> |
| <font style="color:rgb(51,51,51);">根因分类</font> | <font style="color:rgb(51,51,51);">○ 自定义</font> | <font style="color:rgb(51,51,51);">○</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">○</font> |
| <font style="color:rgb(51,51,51);">缺陷模板库</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">○ 规划中</font> |


## **<font style="color:rgb(46,117,182);">7.2 用户故事</font>**
• 作为测试经理，我希望能创建缺陷并指定严重程度、发现阶段和关联版本。

• 作为测试人员，我希望从需求详情一键创建关联缺陷。

• 作为开发工程师，我希望能看到缺陷的复现步骤和期望结果来定位问题。

• 作为测试经理，我希望能分析缺陷的龄分布来识别滞留Defect。

• 作为团队负责人，我希望能按模块和严重程度查看缺陷分布。

## **<font style="color:rgb(46,117,182);">7.3 功能需求详述</font>**
### **<font style="color:rgb(51,51,51);">7.3.1 缺陷 CRUD</font>**
| **<font style="color:rgb(255,255,255);">功能</font>** | **<font style="color:rgb(255,255,255);">详细描述</font>** | **<font style="color:rgb(255,255,255);">验收标准</font>** | **<font style="color:rgb(255,255,255);">优先级</font>** |
| --- | --- | --- | --- |
| <font style="color:rgb(51,51,51);">创建缺陷</font> | <font style="color:rgb(51,51,51);">填写标题、描述、严重程度、优先级、模块、指派人、标签、发现版本、缺陷类型、环境信息</font> | <font style="color:rgb(51,51,51);">标题和严重程度必填</font> | <font style="color:rgb(51,51,51);">P0</font> |
| <font style="color:rgb(51,51,51);">编辑缺陷</font> | <font style="color:rgb(51,51,51);">修改所有字段，记录变更到活动日志</font> | <font style="color:rgb(51,51,51);">编辑后日志同步更新</font> | <font style="color:rgb(51,51,51);">P0</font> |
| <font style="color:rgb(51,51,51);">批量状态更新</font> | <font style="color:rgb(51,51,51);">勾选多个缺陷批量流转状态</font> | <font style="color:rgb(51,51,51);">原子操作，部分失败时提示</font> | <font style="color:rgb(51,51,51);">P1</font> |
| <font style="color:rgb(51,51,51);">缺陷删除</font> | <font style="color:rgb(51,51,51);">仅</font><font style="color:rgb(51,51,51);"> Admin 可硬删除，需二次确认</font> | <font style="color:rgb(51,51,51);">删除后进入回收站</font> | <font style="color:rgb(51,51,51);">P2</font> |
| <font style="color:rgb(51,51,51);">缺陷导出</font> | <font style="color:rgb(51,51,51);">按过滤器导出缺陷列表为</font><font style="color:rgb(51,51,51);"> CSV/Excel</font> | <font style="color:rgb(51,51,51);">导出内容与当前视图一致</font> | <font style="color:rgb(51,51,51);">P1</font> |


### **<font style="color:rgb(51,51,51);">7.3.2 缺陷状态机（推荐）</font>**
77. 推荐状态流转：新建→ 确认 → 处理中 → 修复 → 待验 → 关闭/拒绝/重新打开

78. 状态转换规则：

• → 确认：测试经理确认缺陷有效，分配负责人

• 确认→ 处理中：开发工程师确认并处理

• 处理中→ 修复：开发完成修复，进入待验证

• 修复→ 待验：等待测试验证

• 待验→ 关闭：测试验证通过

• 待验→ 重新打开：测试验证失败，需重新处理

• 任何状态→ 拒绝：缺陷无效（如重复/非缺陷/无法复现）

79. 状态流转可配置：管理员可自定义状态和转换规则。

### **<font style="color:rgb(51,51,51);">7.3.3 缺陷分析报表</font>**
80. 按模块分布：各模块缺陷数量/密度（每故事点缺陷数）。

81. 按严重程度分布：1~5 级缺陷数量和占比。

82. 按发现阶段分布：单元测试/集成测试/UAT/线上/客户反馈。

83. 缺陷龄分析：各状态滞留时长，超过阈值自动告警（如超过 7 天未处理标红）。

84. 缺陷趋势：每日新增/关闭/累计数量趋势（燃尽/燃起）。

85. 根因分布：需求问题/技术问题/环境问题/数据问题占比。

## **<font style="color:rgb(46,117,182);">7.4 交互流程</font>**
86. 创建缺陷（独立）：项目→ 缺陷模块 →「新建缺陷」→ 填写信息 → 保存

87. 创建缺陷（关联）：需求详情/任务详情 →「创建关联缺陷」→ 系统自动关联 → 补充缺陷信息 → 保存

88. 流转状态：缺陷详情→ 点击状态按钮 → 选择下一状态 → 填写流转说明 → 确认

## **<font style="color:rgb(46,117,182);">7.5 数据模型设计</font>**
> 注：本节数据模型为早期草稿，实际已按需求/任务/缺陷三独立业务表拆分，字段与关系表（template/relation/activity 等）定义以《Ydsz Plane 数据库表设计》为准。
| **<font style="color:rgb(255,255,255);">📋</font>****<font style="color:rgb(255,255,255);"> defect</font>** | | |
| --- | --- | --- |
| **<font style="color:rgb(255,255,255);">字段名</font>** | **<font style="color:rgb(255,255,255);">类型</font>****<font style="color:rgb(255,255,255);">/约束</font>** | **<font style="color:rgb(255,255,255);">说明</font>** |
| id | **PK (auto)** | 主键自增ID |
| project | **FK → project** | 外键关联 |
| name | VARCHAR(255) | 名称/标识 |
| description_json/html/stripped | TEXT |  |
| state | **FK → state** | 外键关联 |
| priority | TEXT |  |
| severity | 1-5 | 1-5 |
| type | **FK:****defect** | FK:defect |
| defect_type | TEXT |  |
| found_phase | TEXT |  |
| found_version | **FK → found_version** | 外键关联 |
| fix_version | **FK → fix_version** | 外键关联 |
| parent | **FK:sub** | FK:sub |
| module | **FK → module** | 外键关联 |
| sprint | **FK → ****sprint** | 外键关联 |
| environment | JSON | JSON对象 |
| root_cause_category | TEXT |  |
| assignees | M2M → assignees | 多对多关联 |
| labels | M2M → labels | 多对多关联 |
| verifier | **FK → verifier** | 外键关联 |
| sequence_id | **FK → sequence** | 外键关联sequence |
| **<font style="color:rgb(255,255,255);">📋</font>****<font style="color:rgb(255,255,255);"> issue_template</font>** | | |
| **<font style="color:rgb(255,255,255);">字段名</font>** | **<font style="color:rgb(255,255,255);">类型</font>****<font style="color:rgb(255,255,255);">/约束</font>** | **<font style="color:rgb(255,255,255);">说明</font>** |
| project | **FK → project** | 外键关联 |
| name | VARCHAR(255) | 名称/标识 |
| content | JSON | JSON对象 |
| defect_type | TEXT |  |


  


# **<font style="color:rgb(31,78,121);">8.8 项目文档归档管理 (Documentation)</font>**
项目文档归档提供轻量级文档管理能力，聚焦交付物归档而非实时协作编辑。文档可与需求/任务/缺陷/迭代/版本关联，支持版本管理和评审。对标 Confluence Page（简化版）、云效文档。

## **<font style="color:rgb(46,117,182);">8.8 功能概述</font>**
项目文档归档是面向项目团队的轻量级文档管理模块，支持文档创建、版本管理、需求/任务/缺陷关联。对标 Confluence Page (简化版), ONES 文档。

## **<font style="color:rgb(46,117,182);">10.1 竞品对标</font>**
| **<font style="color:rgb(255,255,255);">功能点</font>** | **<font style="color:rgb(255,255,255);">Confluence</font>** | **<font style="color:rgb(255,255,255);">云效文档</font>** | **<font style="color:rgb(255,255,255);">ONES 文档</font>** | **<font style="color:rgb(255,255,255);">Jira+Confluence</font>** | **<font style="color:rgb(255,255,255);">Ydsz Plane</font>****<font style="color:rgb(255,255,255);"> </font>****<font style="color:rgb(255,255,255);">规划</font>** |
| --- | --- | --- | --- | --- | --- |
| <font style="color:rgb(51,51,51);">文档</font><font style="color:rgb(51,51,51);"> CRUD</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> |
| <font style="color:rgb(51,51,51);">Markdown 编辑</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> |
| <font style="color:rgb(51,51,51);">文档分类</font><font style="color:rgb(51,51,51);">/目录</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> |
| <font style="color:rgb(51,51,51);">关联需求/任务/缺陷</font> | <font style="color:rgb(51,51,51);">● 宏</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> |
| <font style="color:rgb(51,51,51);">关联版本</font> | <font style="color:rgb(51,51,51);">● 宏</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> |
| <font style="color:rgb(51,51,51);">版本管理</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">○ 规划中</font> |
| <font style="color:rgb(51,51,51);">文档评审</font> | <font style="color:rgb(51,51,51);">● Approval</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">○</font> |
| <font style="color:rgb(51,51,51);">文档模板库</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">○ 规划中</font> |
| <font style="color:rgb(51,51,51);">实时协作编辑</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">✕ 不做</font> |
| <font style="color:rgb(51,51,51);">附件上传</font> | <font style="color:rgb(51,51,51);">● 多类型</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> |
| <font style="color:rgb(51,51,51);">评论</font><font style="color:rgb(51,51,51);">/批注</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">○ 规划中</font> |
| <font style="color:rgb(51,51,51);">公开范围配置</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">○</font> |


## **<font style="color:rgb(46,117,182);">10.2 用户故事</font>**
• 作为团队成员，我希望能创建和编辑项目文档（PRD/设计文档/复盘报告）。

• 作为团队成员，我希望能将文档关联到对应的需求和版本。

• 作为团队负责人，我希望能对文档进行评审和审批。

## **<font style="color:rgb(46,117,182);">10.3 功能需求详述</font>**
### **<font style="color:rgb(51,51,51);">10.3.1 文档 CRUD</font>**
| **<font style="color:rgb(255,255,255);">功能</font>** | **<font style="color:rgb(255,255,255);">详细描述</font>** | **<font style="color:rgb(255,255,255);">验收标准</font>** | **<font style="color:rgb(255,255,255);">优先级</font>** |
| --- | --- | --- | --- |
| <font style="color:rgb(51,51,51);">创建文档</font> | <font style="color:rgb(51,51,51);">支持</font><font style="color:rgb(51,51,51);"> Markdown 编辑器，选择分类（PRD/设计/接口/测试报告/交付清单），填写标题和内容</font> | <font style="color:rgb(51,51,51);">标题必填，支持实时预览</font> | <font style="color:rgb(51,51,51);">P0</font> |
| <font style="color:rgb(51,51,51);">编辑文档</font> | <font style="color:rgb(51,51,51);">支持</font><font style="color:rgb(51,51,51);"> Markdown 编辑器和实时预览，保存时生成新版本</font> | <font style="color:rgb(51,51,51);">版本号自动递增，历史版本可查看</font> | <font style="color:rgb(51,51,51);">P0</font> |
| <font style="color:rgb(51,51,51);">删除文档</font> | <font style="color:rgb(51,51,51);">软删除，需二次确认</font> | <font style="color:rgb(51,51,51);">删除后进入回收站</font> | <font style="color:rgb(51,51,51);">P1</font> |
| <font style="color:rgb(51,51,51);">文档目录</font> | <font style="color:rgb(51,51,51);">按分类和项目组织文档树</font> | <font style="color:rgb(51,51,51);">支持拖拽排序和折叠展开</font> | <font style="color:rgb(51,51,51);">P1</font> |
| <font style="color:rgb(51,51,51);">文档搜索</font> | <font style="color:rgb(51,51,51);">按标题和内容全文搜索</font> | <font style="color:rgb(51,51,51);">搜索结果高亮展示</font> | <font style="color:rgb(51,51,51);">P1</font> |
| <font style="color:rgb(51,51,51);">文档分享</font> | <font style="color:rgb(51,51,51);">"分享链接"生成公开或内部链接</font> | <font style="color:rgb(51,51,51);">支持设置链接有效期和访问范围</font> | <font style="color:rgb(51,51,51);">P2</font> |


### **<font style="color:rgb(51,51,51);">10.3.2 文档关联</font>**
89. 关联需求/任务/缺陷：文档中 @提及需求/任务/缺陷，自动创建双向关联。

90. 关联迭代/版本：文档可归属到特定迭代/版本，在该迭代/版本详情中展示。

91. 文档模板：支持从模板创建（PRD 模板、技术方案模板），模板可自定义。

92. 版本管理：每次保存生成新版本，支持版本对比和回滚。

93. 评审流程：指定评审人，评审人可评论和打分，通过后文档状态变为「已评审」。

## **<font style="color:rgb(46,117,182);">10.4 交互流程</font>**
94. 创建文档：项目→「文档」→「新建文档」→ 选择分类/模板 → 编辑器中编写 → 保存

95. 关联需求/任务/缺陷：文档编辑器中输入 @ → 选择需求/任务/缺陷 → 自动插入关联链接

## **<font style="color:rgb(46,117,182);">10.5 数据模型设计</font>**
| **<font style="color:rgb(255,255,255);">📋</font>****<font style="color:rgb(255,255,255);"> document</font>** | | |
| --- | --- | --- |
| **<font style="color:rgb(255,255,255);">字段名</font>** | **<font style="color:rgb(255,255,255);">类型</font>****<font style="color:rgb(255,255,255);">/约束</font>** | **<font style="color:rgb(255,255,255);">说明</font>** |
| id | **PK (auto)** | 主键自增ID |
| project | **FK → project** | 外键关联 |
| title | VARCHAR(255) | 名称/标识 |
| description_html/stripped | TEXT |  |
| category | PRD/design/api/test/checklist | PRD/design/api/test/checklist |
| sprint | **FK → ****sprint** | 外键关联 |
| version | **FK → version** | 外键关联 |
| created_by | **FK → created_by** | 外键关联 |
| status | draft/reviewing/approved/archived | draft/reviewing/approved/archived |
| created_at | TIMESTAMP | 时间戳 |
| updated_at | TIMESTAMP | 时间戳 |
| **<font style="color:rgb(255,255,255);">📋</font>****<font style="color:rgb(255,255,255);"> document_version</font>** | | |
| **<font style="color:rgb(255,255,255);">字段名</font>** | **<font style="color:rgb(255,255,255);">类型</font>****<font style="color:rgb(255,255,255);">/约束</font>** | **<font style="color:rgb(255,255,255);">说明</font>** |
| document | **FK → document** | 外键关联 |
| version_number | TEXT |  |
| content_json | TEXT |  |
| description_html | TEXT |  |
| created_by | **FK → created_by** | 外键关联 |
| created_at | TIMESTAMP | 时间戳 |
| **<font style="color:rgb(255,255,255);">📋</font>****<font style="color:rgb(255,255,255);"> document_link</font>** | | |
| **<font style="color:rgb(255,255,255);">字段名</font>** | **<font style="color:rgb(255,255,255);">类型</font>****<font style="color:rgb(255,255,255);">/约束</font>** | **<font style="color:rgb(255,255,255);">说明</font>** |
| document | **FK → document** | 外键关联 |
| linkable_type | TEXT |  |
| linkable_id  // 多态关联到需求/任务/缺陷/迭代/版本 | TEXT |  |


  


**— ****Ydsz Plane**** ****产品需求文档核心业务功能细化****完****—**



# **<font style="color:rgb(31,78,121);">8.9 知识库 (Knowledge Base)</font>**
## **<font style="color:rgb(46,117,182);">10.1 功能概述</font>**
知识库是面向团队的知识沉淀与共享平台，支持结构化文档管理、版本控制、权限管理和全文检索。对标 Confluence、飞书文档、语雀、Notion。

核心定位：团队的技术文档中心、SOP 知识库、新人 onboarding 手册库；与需求/任务/缺陷双向关联，实现业务流程与知识沉淀的闭环。

注意：知识库是独立于「项目文档归档」的模块。项目文档归档是面向单个项目的轻量文档管理，知识库是跨项目的团队级知识管理平台。

## **<font style="color:rgb(46,117,182);">10.2 竞品对标</font>**
下表对标一线竞品。● 已支持 / △ 部分 / ○ 不支持 / ✕ 不适用

| **<font style="color:rgb(255,255,255);">功能点</font>** | **<font style="color:rgb(255,255,255);">Confluence</font>** | **<font style="color:rgb(255,255,255);">飞书文档</font>** | **<font style="color:rgb(255,255,255);">语雀</font>** | **<font style="color:rgb(255,255,255);">Notion</font>** | **<font style="color:rgb(255,255,255);">Ydsz Plane</font>****<font style="color:rgb(255,255,255);"> </font>****<font style="color:rgb(255,255,255);">规划</font>** |
| --- | --- | --- | --- | --- | --- |
| 文档 CRUD | ● | ● | ● | ● | ● |
| 富文本编辑器 | ● | ● | ● | ● | ● |
| Markdown 原生支持 | ● | ● | ● | ● | ● |
| 多级空间/目录 | ● | ● | ● | ● | ● |
| 文档模板 | ● | ● | ● | ● | ● |
| 文档版本管理 | ● | ● | ● | ○ | ○ 规划中 |
| 文档协作(多人实时) | ● | ● | ● | ● | ✕ 不做 |
| 文档评审/审批流 | ● | ● | ○ | ○ | ○ |
| 关联需求/任务/缺陷(双向) | ● Jira | ○ | ○ | ● | ● |
| 权限管理(空间级) | ● 细粒度 | ● | ● | ● | ● |
| 附件/嵌入内容 | ● 富媒体 | ● 富媒体 | ● | ● 富媒体 | ● |
| 全文检索 | ● | ● | ● | ● | ● |
| 开放 API | ● | ● | ● | ● | ● |
| 导入导出 | ● Markdown/HTML | ● | ● | ● | ● |


## **<font style="color:rgb(46,117,182);">10.3 用户故事</font>**
技术负责人：我需要一个可以沉淀团队技术方案、架构决策记录（ADR）的地方，新人能自助查阅。

测试负责人：我希望能把测试用例、测试报告模板放在知识库中，需求/任务/缺陷提 Defect 时能一键关联到知识库文章。

产品经理：我希望 PRD、竞品分析、用户调研报告能结构化管理，支持版本回溯与多人评审。

开发工程师：我需要一个地方写技术文档、接口文档，支持 Markdown 和代码高亮，能关联到 Git commit/PR。

## **<font style="color:rgb(46,117,182);">10.4 功能需求详述</font>**
| **<font style="color:rgb(255,255,255);">功能</font>** | **<font style="color:rgb(255,255,255);">详细描述与验收标准</font>** | **<font style="color:rgb(255,255,255);">优先级</font>** |
| --- | --- | --- |
| 创建空间(知识库) | 团队/项目级知识库容器，支持设置空间名称、描述、封面、公开/私有属性；Slug 唯一校验 | P0 |
| 文档树管理 | 支持无限层级目录结构，文档归类到目录节点，支持拖拽排序和层级调整；目录支持折叠/展开 | P0 |
| 创建文档 | 支持 Markdown 编辑器，实时预览；支持从模板创建；支持富文本编辑模式切换 | P0 |
| 编辑文档 | Markdown 编辑器，支持快捷键、代码高亮、表格、图片粘贴、数学公式、Mermaid 流程图 | P0 |
| 文档保存与历史 | 每次保存生成新版本，支持版本对比、回滚；保留 30 天历史版本 | P1 |
| 文档删除 | 软删除，进入回收站，30 天后自动清理；管理员可恢复 | P1 |
| 文档权限 | 空间级：Owner/Admin/Editor/Viewer 四级；文档级可继承或覆盖；支持公开链接分享 | P0 |
| 文档全文搜索 | 基于 Elasticsearch，支持标题+内容全文模糊匹配，结果高亮；支持按空间/标签过滤 | P0 |
| 文档导入 | 支持 Markdown/HTML/Word 批量导入，自动解析目录结构 | P1 |
| 文档导出 | 导出为 Markdown/HTML/PDF/Word 格式，支持批量导出整个目录 | P1 |
| 文档评论 | 支持全文评论和段落评论，支持 @提及、评论通知 | P1 |
| 文档关联需求/任务/缺陷 | @提及时关联需求/任务/缺陷，双向可追溯 | P0 |
| 文档关联版本 | 文档可归属到版本，在该版本详情中展示 | P1 |
| 文档评审 | 指定评审人，评审人可评论和打分；通过后文档状态变为「已评审」，支持审批流 | P1 |
| 文档标签/分类 | 支持多标签和多分类体系，可按标签聚合展示 | P1 |
| 文档订阅 | 关注某文档，当有新版本或评论时收到通知 | P2 |
| 文档使用统计 | 查看文档浏览量、编辑次数、最近编辑者、订阅人数 | P2 |
| 知识库模板 | 预置技术方案 / ADR / PRD / 测试报告 / 新人指南等模板，支持自定义模板 | P1 |
| 文档 API | 提供 Open API，支持机器人/自动化创建和更新文档 | P1 |


## **<font style="color:rgb(46,117,182);">10.5 交互流程</font>**
• 创建知识库空间→ 配置权限和模板 → 新建文档目录树

• 新建/编辑文档 → 选择模板 → 编写 Markdown → 实时预览 → 保存为草稿/发布

• 保存文档后，点击「关联需求/任务/缺陷」→ 搜索并关联需求/任务/缺陷，双向追溯

• 文档评审流程：提交评审→ 指定评审人 → 评审人打分/评论 → 通过/驳回

• 搜索：全文搜索/按空间/标签/作者过滤 → 结果高亮 → 点击跳转

• 分享：生成公开链接（可设置密码和有效期）→ 外部用户访问

## **<font style="color:rgb(46,117,182);">10.6 数据模型设计</font>**
| **<font style="color:rgb(255,255,255);">模型</font>** | **<font style="color:rgb(255,255,255);">字段</font>** | **<font style="color:rgb(255,255,255);">说明</font>** |
| --- | --- | --- |
| KnowledgeSpace | name, description, slug, cover_image, is_private | 知识库空间（跨项目或项目级） |
|  | owner(FK User), default_permission(enum) | 空间级权限默认配置 |
|  | workspace(FK) | 归属工作空间 |
|  | project(FK, nullable) | 可选归属项目（null 表示空间级） |
|  | created_at, updated_at | 时间戳 |
|  |  |  |
| KnowledgePage | title, content_md, content_html, version | 文档标题与内容 |
|  | parent(FK self), space(FK) | 目录层级与归属空间 |
|  | status(draft/published/archived) | 文档状态 |
|  | created_by, updated_by(FK User) | 作者信息 |
|  | reviewer(FK User, nullable) | 评审人（评审模式时） |
|  | labels(M2M) | 多标签 |
|  | is_pinned, is_featured | 置顶/推荐标记 |
|  | view_count | 浏览统计 |
|  |  |  |
| KnowledgePageVersion | page(FK), version, content_md | 文档版本快照 |
|  | created_by, created_at | 版本创建信息 |
|  |  |  |
| KnowledgePageRelation | page(FK), target_type(requirement/task/defect), target_id | 文档与需求/任务/缺陷双向关联 |
|  | relation_type(enum) | 关联类型（引用/被引用） |


# **<font style="color:rgb(31,78,121);">8.10 收件箱 (Intake/Inbox)</font>**
收件箱是用户将外部反馈（Defect 报告、功能建议、用户反馈）提交到项目的通道。Intake 提报经管理员审核后可转正为正式需求/任务/缺陷，或拒绝/归档。对标 Jira Service Management Queue、云效反馈。

## **<font style="color:rgb(46,117,182);">8.10 功能概述</font>**
收件箱是面向所有反馈来源的统一收集通道，支持外部Defect 报告、功能建议的筛选、转正和跟踪。对标 Jira Service Management, 云效需求池。

## **<font style="color:rgb(46,117,182);">8.10 竞品对标</font>**
| **<font style="color:rgb(255,255,255);">功能点</font>** | **<font style="color:rgb(255,255,255);">Jira Service Management</font>** | **<font style="color:rgb(255,255,255);">云效反馈</font>** | **<font style="color:rgb(255,255,255);">ONES 反馈</font>** | **<font style="color:rgb(255,255,255);">TAPD 反馈</font>** | **<font style="color:rgb(255,255,255);">Ydsz Plane</font>****<font style="color:rgb(255,255,255);"> </font>****<font style="color:rgb(255,255,255);">现有</font>** |
| --- | --- | --- | --- | --- | --- |
| <font style="color:rgb(51,51,51);">收件箱</font><font style="color:rgb(51,51,51);"> Channel 配置</font> | <font style="color:rgb(51,51,51);">● 多队列</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> |
| <font style="color:rgb(51,51,51);">Intake 提报提交</font> | <font style="color:rgb(51,51,51);">● 门户/邮件</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> |
| <font style="color:rgb(51,51,51);">自动分配规则</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> |
| <font style="color:rgb(51,51,51);">Intake 提报转正</font> | <font style="color:rgb(51,51,51);">→ 需求/任务/缺陷</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> |
| <font style="color:rgb(51,51,51);">审核拒绝</font><font style="color:rgb(51,51,51);">/归档</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> |
| <font style="color:rgb(51,51,51);">提交者跟踪门户</font> | <font style="color:rgb(51,51,51);">● 公开</font> | <font style="color:rgb(51,51,51);">○</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">○</font> | <font style="color:rgb(51,51,51);">●</font> |
| <font style="color:rgb(51,51,51);">邮件自动回复</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">○ 规划中</font> |
| <font style="color:rgb(51,51,51);">Webhook 事件</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> |
| <font style="color:rgb(51,51,51);">重复提报检测</font> | <font style="color:rgb(51,51,51);">● AI</font> | <font style="color:rgb(51,51,51);">○</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">○</font> | <font style="color:rgb(51,51,51);">○ 规划中</font> |
| <font style="color:rgb(51,51,51);">知识库</font><font style="color:rgb(51,51,51);">/FAQ 自助</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">○</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">○</font> | <font style="color:rgb(51,51,51);">○</font> |


## **<font style="color:rgb(46,117,182);">8.10 用户故事</font>**
• 作为外部用户，我希望能通过收件箱通道提交反馈，无需注册账号。

• 作为项目经理，我希望能审核 Intake 提报并决定转正为需求还是缺陷。

• 作为提交者，我希望能跟踪我的反馈处理进度。

## **<font style="color:rgb(46,117,182);">8.10 功能需求详述</font>**
### **<font style="color:rgb(51,51,51);">9.3.1 收件箱 Channel 配置</font>**
96. 创建 Channel：配置名称、描述、关联项目、默认需求/任务/缺陷类型、自动分配规则。

97. 访问控制：公开（无需登录）或私有（仅成员可访问）。

98. 门户页面：自动生成可嵌入的反馈门户页面，支持自定义 Logo 和说明。

### **<font style="color:rgb(51,51,51);">9.3.2 Intake 提报管理</font>**
| **<font style="color:rgb(255,255,255);">功能</font>** | **<font style="color:rgb(255,255,255);">详细描述</font>** | **<font style="color:rgb(255,255,255);">验收标准</font>** | **<font style="color:rgb(255,255,255);">优先级</font>** |
| --- | --- | --- | --- |
| <font style="color:rgb(51,51,51);">提交反馈</font> | <font style="color:rgb(51,51,51);">填写标题、描述、附件（截图）、优先级（可选）</font> | <font style="color:rgb(51,51,51);">提交成功返回跟踪</font><font style="color:rgb(51,51,51);"> ID</font> | <font style="color:rgb(51,51,51);">P0</font> |
| <font style="color:rgb(51,51,51);">转正为需求</font> | <font style="color:rgb(51,51,51);">管理员审核后转正为需求，系统创建关联，状态同步</font> | <font style="color:rgb(51,51,51);">Intake 状态变为关联的正式需求</font> | <font style="color:rgb(51,51,51);">P0</font> |
| <font style="color:rgb(51,51,51);">转正为缺陷</font> | <font style="color:rgb(51,51,51);">管理员审核后转正为缺陷，自动关联标题</font><font style="color:rgb(51,51,51);">/描述</font> | <font style="color:rgb(51,51,51);">Intake 状态变为关联的正式缺陷</font> | <font style="color:rgb(51,51,51);">P0</font> |
| <font style="color:rgb(51,51,51);">拒绝</font><font style="color:rgb(51,51,51);">/归档</font> | <font style="color:rgb(51,51,51);">标记为拒绝（无效反馈）或归档（暂不处理）</font> | <font style="color:rgb(51,51,51);">通知提交者（如公开门户）</font> | <font style="color:rgb(51,51,51);">P1</font> |
| <font style="color:rgb(51,51,51);">自动分配</font> | <font style="color:rgb(51,51,51);">基于模块</font><font style="color:rgb(51,51,51);">/类型自动分配给指定成员</font> | <font style="color:rgb(51,51,51);">分配后通知被指派人</font> | <font style="color:rgb(51,51,51);">P0</font> |
| <font style="color:rgb(51,51,51);">重复检测（</font><font style="color:rgb(51,51,51);">AI）</font> | <font style="color:rgb(51,51,51);">提交时</font><font style="color:rgb(51,51,51);"> AI 检测可能重复的提报，提示候选</font> | <font style="color:rgb(51,51,51);">检测结果高亮展示</font> | <font style="color:rgb(51,51,51);">P2</font> |


## **<font style="color:rgb(46,117,182);">8.10 交互流程</font>**
99. 提交反馈：收件箱门户/邮件 → 填写信息 → 提交 → 返回跟踪 ID

100. 审核转正：项目→ Intake 模块 → 点击提报 → 选择转正类型（需求/缺陷）→ 确认

101. 跟踪状态：提交者通过跟踪 ID + 邮件查看处理状态

## **<font style="color:rgb(46,117,182);">8.10 数据模型设计</font>**
| **<font style="color:rgb(255,255,255);">📋</font>****<font style="color:rgb(255,255,255);"> intake_channel</font>** | | |
| --- | --- | --- |
| **<font style="color:rgb(255,255,255);">字段名</font>** | **<font style="color:rgb(255,255,255);">类型</font>****<font style="color:rgb(255,255,255);">/约束</font>** | **<font style="color:rgb(255,255,255);">说明</font>** |
| id | **PK (auto)** | 主键自增ID |
| project | **FK → project** | 外键关联 |
| name | VARCHAR(255) | 名称/标识 |
| description | VARCHAR(255) | 名称/标识 |
| is_public | BOOLEAN | 布尔标志 |
| default_issue_type | TEXT |  |
| auto_assign_rules | JSON | JSON对象 |
| portal_url_slug | TEXT |  |
| **<font style="color:rgb(255,255,255);">📋</font>****<font style="color:rgb(255,255,255);"> intake_issue</font>** | | |
| **<font style="color:rgb(255,255,255);">字段名</font>** | **<font style="color:rgb(255,255,255);">类型</font>****<font style="color:rgb(255,255,255);">/约束</font>** | **<font style="color:rgb(255,255,255);">说明</font>** |
| id | **PK (auto)** | 主键自增ID |
| channel | **FK → channel** | 外键关联 |
| project | **FK → project** | 外键关联 |
| title | VARCHAR(255) | 名称/标识 |
| description_json | TEXT |  |
| status | open/accepted/rejected/archived | open/accepted/rejected/archived |
| converted_issue | **FK → converted_issue** | 外键关联 |
| submitter_name | TEXT |  |
| submitter_email | TEXT |  |
| submitted_at | TIMESTAMP | 时间戳 |
| reviewed_by | **FK → reviewed_by** | 外键关联 |
| reviewed_at | TIMESTAMP | 时间戳 |


  


# **<font style="color:rgb(31,78,121);">9. 效率增强</font>****<font style="color:rgb(31,78,121);">功能</font>**
本章详细描述项目仪表盘、个人工作台、通知中心、全局搜索、视图自定义、Webhook & API、自动化引擎、研发效能度量共 8 个效率增强模块的功能需求。

## **<font style="color:rgb(46,117,182);">9.1 项目仪表盘 (Project Dashboard)</font>**
面向项目经理、团队负责人和 PMO 的一站式项目聚合视图。通过可配置的可视化卡片，在一屏内呈现项目进度、质量、风险、资源等核心指标，支持逐层下钻。对标 TAPD 项目仪表盘、云效项目概览、ONES Performance。

DashboardConfig: name, layout(JSON), owner(FK User), is_default, scope(project/space) Widget: dashboard(FK), type(chart/counter/list), config(JSON), position, size WidgetData: widget(FK), data(JSON), refreshed_at

### **<font style="color:rgb(51,51,51);">9.1 数据模型设计</font>**
项目仪表盘的功能需求详述（详见下方详细功能描述）

### **<font style="color:rgb(51,51,51);">9.1 功能需求详述</font>**
选择项目/时间范围 → 查看仪表盘卡片 → 点击卡片元素下钻 → 进入单需求/任务/缺陷详情

### **<font style="color:rgb(51,51,51);">9.1 交互流程</font>**
项目经理: 我需要一张报表就能看到全项目的健康状态，发现风险点。 团队成员: 我想快速了解当前迭代的进度和我的贡献占比。 PMO: 我想同时查看多个项目的整体进展，做横向对比。

### **<font style="color:rgb(51,51,51);">9.1 用户故事</font>**
参见下方竞品对标表格

### **<font style="color:rgb(51,51,51);">9.1 竞品对标</font>**
项目仪表盘是面向项目经理、团队负责人、PMO 的一站式聚合视图，提供项目健康度、进度趋势、质量指标等关键数据。对标 TAPD 仪表盘, 云效项目概览, ONES Performance。

### **<font style="color:rgb(51,51,51);">9.1 功能概述</font>**
### **<font style="color:rgb(51,51,51);">用户故事</font>**
• 作为项目经理，我希望在一个页面看到所有项目的整体健康度。

• 作为 Scrum Master，我希望查看当前迭代的燃尽图和速率。

• 作为 PMO，我希望对比多个项目的交付速率和质量指标。

• 作为团队负责人，我希望按模块查看缺陷分布。

### **<font style="color:rgb(51,51,51);">卡片类型与指标</font>**
| **<font style="color:rgb(255,255,255);">卡片类型</font>** | **<font style="color:rgb(255,255,255);">数据维度</font>** | **<font style="color:rgb(255,255,255);">可视化形式</font>** | **<font style="color:rgb(255,255,255);">刷新频率</font>** | **<font style="color:rgb(255,255,255);">优先级</font>** |
| --- | --- | --- | --- | --- |
| <font style="color:rgb(51,51,51);">项目概览</font> | <font style="color:rgb(51,51,51);">需求</font><font style="color:rgb(51,51,51);">/任务/缺陷总数、完成率、遗留数</font> | <font style="color:rgb(51,51,51);">数字</font><font style="color:rgb(51,51,51);"> + 环形进度条</font> | <font style="color:rgb(51,51,51);">实时</font> | <font style="color:rgb(51,51,51);">P0</font> |
| <font style="color:rgb(51,51,51);">迭代进度</font> | <font style="color:rgb(51,51,51);">当前</font><font style="color:rgb(51,51,51);"> Sprint 完成度、故事点消耗</font> | <font style="color:rgb(51,51,51);">燃尽图</font><font style="color:rgb(51,51,51);"> / 燃起图</font> | <font style="color:rgb(51,51,51);">实时</font> | <font style="color:rgb(51,51,51);">P0</font> |
| <font style="color:rgb(51,51,51);">版本进度</font> | <font style="color:rgb(51,51,51);">聚合所有关联迭代完成度</font> | <font style="color:rgb(51,51,51);">甘特图</font><font style="color:rgb(51,51,51);"> + 进度条</font> | <font style="color:rgb(51,51,51);">实时</font> | <font style="color:rgb(51,51,51);">P0</font> |
| <font style="color:rgb(51,51,51);">模块分布</font> | <font style="color:rgb(51,51,51);">按模块聚合需求/任务/缺陷数、完成率</font> | <font style="color:rgb(51,51,51);">柱状图</font><font style="color:rgb(51,51,51);"> / 饼图</font> | <font style="color:rgb(51,51,51);">1小时</font> | <font style="color:rgb(51,51,51);">P0</font> |
| <font style="color:rgb(51,51,51);">质量指标</font> | <font style="color:rgb(51,51,51);">缺陷密度、准出率、逃逸率</font> | <font style="color:rgb(51,51,51);">数字</font><font style="color:rgb(51,51,51);"> + 趋势折线</font> | <font style="color:rgb(51,51,51);">1小时</font> | <font style="color:rgb(51,51,51);">P0</font> |
| <font style="color:rgb(51,51,51);">资源负载</font> | <font style="color:rgb(51,51,51);">成员任务饱和度、工时分布</font> | <font style="color:rgb(51,51,51);">热力图</font><font style="color:rgb(51,51,51);"> / 柱状图</font> | <font style="color:rgb(51,51,51);">4小时</font> | <font style="color:rgb(51,51,51);">P1</font> |
| <font style="color:rgb(51,51,51);">风险预警</font> | <font style="color:rgb(51,51,51);">逾期项、阻塞项、逾期缺陷</font> | <font style="color:rgb(51,51,51);">红色高亮列表</font> | <font style="color:rgb(51,51,51);">实时</font> | <font style="color:rgb(51,51,51);">P1</font> |
| <font style="color:rgb(51,51,51);">速率趋势</font> | <font style="color:rgb(51,51,51);">近</font><font style="color:rgb(51,51,51);"> 6 个 Sprint Velocity</font> | <font style="color:rgb(51,51,51);">折线图</font> | <font style="color:rgb(51,51,51);">Sprint 结束</font> | <font style="color:rgb(51,51,51);">P1</font> |
| <font style="color:rgb(51,51,51);">DORA 指标</font> | <font style="color:rgb(51,51,51);">部署频率</font><font style="color:rgb(51,51,51);">/变更周期/失败率/MTTR</font> | <font style="color:rgb(51,51,51);">四宫格数字</font> | <font style="color:rgb(51,51,51);">每日</font> | <font style="color:rgb(51,51,51);">P1</font> |
| <font style="color:rgb(51,51,51);">需求前置时间</font> | <font style="color:rgb(51,51,51);">从创建到完成的周期时间</font> | <font style="color:rgb(51,51,51);">散点图</font><font style="color:rgb(51,51,51);"> + 趋势线</font> | <font style="color:rgb(51,51,51);">每日</font> | <font style="color:rgb(51,51,51);">P2</font> |


### **<font style="color:rgb(51,51,51);">交互功能</font>**
102. 全局时间筛选：支持按日/周/月/季度/年/自定义时间范围切换所有卡片数据。

103. 项目下钻：从项目总览→ 进入单项目 → 进入单迭代 → 查看需求/任务/缺陷列表，逐层展开。

104. 卡片联动：点击某个图表元素，其他卡片自动过滤为该维度。

105. 全屏驾驶舱模式：支持大屏投影，自动轮播多个项目仪表盘。

106. 预警规则配置：可自定义阈值（如逾期率 > 10% 触发红色预警），支持邮件/IM 通知。

107. 导出报表：支持导出 PNG/Screenshot 到 PPT，或导出原始数据为 CSV。

### **<font style="color:rgb(51,51,51);">自定义能力</font>**
• PM 可拖拽调整卡片位置和大小，自由组合仪表盘布局。

• 支持保存多个仪表盘视图（如「PMO 视图」「开发 Leader 视图」「QA 视图」）。

• 系统预设模板：开箱即用的「敏捷项目模板」「版本交付模板」「PMO 多项目模板」。

• 卡片可见性：基于角色配置哪些卡片对哪些角色可见。

### **<font style="color:rgb(51,51,51);">竞品对标</font>**
| **<font style="color:rgb(255,255,255);">功能点</font>** | **<font style="color:rgb(255,255,255);">TAPD 仪表盘</font>** | **<font style="color:rgb(255,255,255);">云效概览</font>** | **<font style="color:rgb(255,255,255);">ONES Performance</font>** | **<font style="color:rgb(255,255,255);">PingCode</font>** | **<font style="color:rgb(255,255,255);">Jira</font>** |
| --- | --- | --- | --- | --- | --- |
| <font style="color:rgb(51,51,51);">自定义卡片布局</font> | <font style="color:rgb(51,51,51);">● 拖拽缩放</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">● 需插件</font> |
| <font style="color:rgb(51,51,51);">燃尽图</font><font style="color:rgb(51,51,51);">/燃起图</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> |
| <font style="color:rgb(51,51,51);">全屏驾驶舱</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">● 播放模式</font> | <font style="color:rgb(51,51,51);">○</font> | <font style="color:rgb(51,51,51);">○</font> |
| <font style="color:rgb(51,51,51);">预警规则配置</font> | <font style="color:rgb(51,51,51);">● 阈值+颜色</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">● 需插件</font> |
| <font style="color:rgb(51,51,51);">多项目聚合</font> | <font style="color:rgb(51,51,51);">● 项目集</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">● PMO模块</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> |
| <font style="color:rgb(51,51,51);">DORA 指标</font> | <font style="color:rgb(51,51,51);">○</font> | <font style="color:rgb(51,51,51);">○</font> | <font style="color:rgb(51,51,51);">○</font> | <font style="color:rgb(51,51,51);">○</font> | <font style="color:rgb(51,51,51);">○</font> |


  


## **<font style="color:rgb(46,117,182);">9.2 个人工作台 (My Workbench)</font>**
每个团队成员的首页，聚合与「我」相关的所有需求/任务/缺陷、迭代动态、通知提醒。对标 TAPD 工作台、云效工作台、Jira Your Work。

WorkbenchConfig: user(FK), layout(JSON), widgets(M2M WidgetDef) WidgetDef: name, type(todo/notification/calendar/iteration), default_config(JSON) TodoItem: user(FK), target_type(requirement/task/defect), target_id, is_pinned, order

### **<font style="color:rgb(51,51,51);">9.2 数据模型设计</font>**
个人工作台的功能需求详述（详见下方详细功能描述）

### **<font style="color:rgb(51,51,51);">9.2 功能需求详述</font>**
登录系统→ 打开工作台 → 查看待办/通知/日程 → 点击需求/任务/缺陷进入详情或批量操作

### **<font style="color:rgb(51,51,51);">9.2 交互流程</font>**
开发工程师: 我希望打开工作台就知道今天该做什么，没有切换成本。 产品经理: 我想一目了然看到与我相关的所有事项（我创建的、指派给我的、@我的）。 测试工程师: 我希望快速知道待验证的缺陷数量和优先级。

### **<font style="color:rgb(51,51,51);">9.2 用户故事</font>**
参见下方竞品对标表格

### **<font style="color:rgb(51,51,51);">9.2 竞品对标</font>**
个人工作台是面向每个团队成员的个人聚合首页，聚合待办、迭代、通知、搜索等个性化信息。对标 TAPD 工作台, 云效工作台。

### **<font style="color:rgb(51,51,51);">9.2 功能概述</font>**
### **<font style="color:rgb(51,51,51);">用户故事</font>**
• 作为开发工程师，我打开工作台就能看到今天应该做什么、有哪些逾期任务。

• 作为产品经理，我希望能一览本周待评审的需求和我负责的缺陷进展。

• 作为 Scrum Master，我想快速查看每个成员的负载是否均衡。

### **<font style="color:rgb(51,51,51);">区域划分</font>**
| **<font style="color:rgb(255,255,255);">区域</font>** | **<font style="color:rgb(255,255,255);">位置</font>** | **<font style="color:rgb(255,255,255);">内容</font>** | **<font style="color:rgb(255,255,255);">优先级</font>** |
| --- | --- | --- | --- |
| <font style="color:rgb(51,51,51);">快速创建</font> | <font style="color:rgb(51,51,51);">顶部固定栏</font> | <font style="color:rgb(51,51,51);">一键按钮：创建需求</font><font style="color:rgb(51,51,51);">/任务/缺陷，当前项目上下文自动带入</font> | <font style="color:rgb(51,51,51);">P0</font> |
| <font style="color:rgb(51,51,51);">我的待办</font> | <font style="color:rgb(51,51,51);">左侧主区域</font> | <font style="color:rgb(51,51,51);">按优先级排序：今日应开始、进行中、逾期未完成项</font> | <font style="color:rgb(51,51,51);">P0</font> |
| <font style="color:rgb(51,51,51);">今日日程</font> | <font style="color:rgb(51,51,51);">右侧顶栏</font> | <font style="color:rgb(51,51,51);">今日会议（关联周期性会议）</font><font style="color:rgb(51,51,51);">+ 个人日程</font> | <font style="color:rgb(51,51,51);">P0</font> |
| <font style="color:rgb(51,51,51);">迭代概览</font> | <font style="color:rgb(51,51,51);">右侧中部</font> | <font style="color:rgb(51,51,51);">我参与的迭代的当前进度、故事点消耗</font> | <font style="color:rgb(51,51,51);">P0</font> |
| <font style="color:rgb(51,51,51);">关注动态</font> | <font style="color:rgb(51,51,51);">右侧下部</font> | <font style="color:rgb(51,51,51);">我订阅</font><font style="color:rgb(51,51,51);">/参与的需求/任务/缺陷变更（状态更新、评论、@提及）</font> | <font style="color:rgb(51,51,51);">P1</font> |
| <font style="color:rgb(51,51,51);">最近访问</font> | <font style="color:rgb(51,51,51);">底部</font> | <font style="color:rgb(51,51,51);">最近查看的</font><font style="color:rgb(51,51,51);"> 10 个需求/任务/缺陷，支持一键跳转</font> | <font style="color:rgb(51,51,51);">P1</font> |
| <font style="color:rgb(51,51,51);">个人效率报告</font> | <font style="color:rgb(51,51,51);">底部</font><font style="color:rgb(51,51,51);">/独立页</font> | <font style="color:rgb(51,51,51);">本周完成故事点、工时分布、完成率趋势</font> | <font style="color:rgb(51,51,51);">P2</font> |


### **<font style="color:rgb(51,51,51);">交互能力</font>**
108. Focus Mode：沉浸模式，隐藏侧边栏，仅展示当前迭代待办，帮助深度工作。

109. 批量操作：支持勾选多个待办项，批量更新状态/指派/规划迭代。

110. 项目切换：顶部项目选择器，可跨项目查看所有待办。

111. 分组视图：支持按「项目/工作类型/截止日期/优先级」分组显示。

112. 拖拽排序：在待办列表中手动拖拽调整优先级顺序。

113. 快捷操作：在待办卡片上直接完成操作（流转状态、添加评论、记录工时），无需跳转详情页。

### **<font style="color:rgb(51,51,51);">自定义能力</font>**
• 布局配置：用户可拖拽调整区域位置，保存为个人布局。

• 卡片可见性：可开启/关闭「关注动态」「迭代概览」等区域。

• 默认视图：管理员可设置全局默认工作台布局。

## **<font style="color:rgb(46,117,182);">9.3 通知中心 (Notification Center)</font>**
聚合所有与用户相关的变更事件，支持站内、邮件、IM（企微/钉钉/飞书）多渠道推送。对标 Jira Notification、TAPD 消息中心。

Notification: recipient(FK User), target_type(requirement/task/defect), target_id, event_type, channel(in_app/email/im), is_read, created_at NotificationSubscription: user(FK), project(FK), event_types(JSON), channels(JSON), is_enabled NotificationDigest: user(FK), frequency(real_time/daily/weekly), last_sent_at

### **<font style="color:rgb(51,51,51);">9.3 数据模型设计</font>**
通知中心的功能需求详述（详见下方详细功能描述）

### **<font style="color:rgb(51,51,51);">9.3 功能需求详述</font>**
配置通知规则→ 接收事件通知 → 点击通知跳转/标记已读 → 查看通知历史

### **<font style="color:rgb(51,51,51);">9.3 交互流程</font>**
开发工程师: 我希望只收到与我相关的通知，不被噪音干扰。 项目经理: 我想实时知道关键状态变更，但不要每条变更都通知。 测试工程师: 我希望验证通过/失败时能收到 IM 提醒。

### **<font style="color:rgb(51,51,51);">9.3 用户故事</font>**
参见下方竞品对标表格

### **<font style="color:rgb(51,51,51);">9.3 竞品对标</font>**
通知中心聚合所有与用户相关的变更事件，支持站内、邮件、IM 等渠道推送和用户自定义。对标 Jira Notifications, 云效消息中心。

### **<font style="color:rgb(51,51,51);">9.3 功能概述</font>**
### **<font style="color:rgb(51,51,51);">通知类型</font>**
• 需求/任务/缺陷变更：状态流转、字段更新、指派变更、评论/@提及。

• 迭代事件：迭代开始/结束提醒、站会提醒。

• 版本发布：新版本发布、版本状态变更。

• 审批事件：需求评审、缺陷关闭请求。

• 自动化通知：自动化规则执行结果、阈值预警。

• @提及：评论或描述中被 @的用户。

### **<font style="color:rgb(51,51,51);">通知规则（默认）</font>**
| **<font style="color:rgb(255,255,255);">事件</font>** | **<font style="color:rgb(255,255,255);">站内通知</font>** | **<font style="color:rgb(255,255,255);">邮件</font>** | **<font style="color:rgb(255,255,255);">IM</font>** | **<font style="color:rgb(255,255,255);">默认接收人</font>** |
| --- | --- | --- | --- | --- |
| <font style="color:rgb(51,51,51);">我创建的需求/任务/缺陷发生变更</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">○</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">创建人</font> |
| <font style="color:rgb(51,51,51);">指派给我的需求/任务/缺陷</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">被指派人</font> |
| <font style="color:rgb(51,51,51);">我评论的需求/任务/缺陷有新回复</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">○</font> | <font style="color:rgb(51,51,51);">评论人</font> |
| <font style="color:rgb(51,51,51);">@我的</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">被</font><font style="color:rgb(51,51,51);">@人</font> |
| <font style="color:rgb(51,51,51);">迭代即将结束（剩余</font><font style="color:rgb(51,51,51);">1天）</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">迭代参与者</font> |
| <font style="color:rgb(51,51,51);">版本状态变更</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">●</font> | <font style="color:rgb(51,51,51);">版本参与者</font> |


### **<font style="color:rgb(51,51,51);">用户自定义</font>**
• 通知订阅：用户可在「设置→ 通知」中按项目/空间/需求/任务/缺陷类型开启或关闭特定通知。

• 免打扰时段：可设置免打扰时间段（如 22:00-08:00）。

• 汇总方式：支持即时/日报/周报三种模式，避免通知轰炸。

• 通知归档：已读通知自动归档，支持「全部标为已读」。

### **<font style="color:rgb(51,51,51);">IM 集成渠道</font>**
114. 企业微信：通过 Webhook 或自建应用推送通知，支持富文本卡片消息。

115. 钉钉：通过自定义机器人或企业内部应用推送，支持 Markdown 格式。

116. 飞书：通过 Bot 或自定义应用推送，支持消息卡片。

117. Email：SMTP 配置，支持模板定制，支持中英文。

118. Slack/Teams（国际版）：面向海外团队，支持 Webhook 和 Bot。

  


## **<font style="color:rgb(46,117,182);">9.4 全局搜索 (Global Search)</font>**
跨多对象的全文搜索和过滤器联动。对标 Jira Issue Search (JQL)、云效全局搜索。

SearchQuery: user(FK), query_text, filters(JSON), result_count, created_at SearchHistory: user(FK), query, last_used_at, use_count SearchIndex: content_type(FK ContentType), object_id, title_vector, content_vector, updated_at

### **<font style="color:rgb(51,51,51);">9.4 数据模型设计</font>**
全局搜索的功能需求详述（详见下方详细功能描述）

### **<font style="color:rgb(51,51,51);">9.4 功能需求详述</font>**
点击搜索框/快捷键 → 输入关键字/语法 → 查看结果 → 点击过滤/排序 → 进入需求/任务/缺陷详情

### **<font style="color:rgb(51,51,51);">9.4 交互流程</font>**
所有用户: 我想快速通过关键字找到相关需求/任务/缺陷，即使我记不清具体字段。 项目经理: 我想用高级语法（如 project:X status:Y）精确过滤需求/任务/缺陷。

### **<font style="color:rgb(51,51,51);">9.4 用户故事</font>**
参见下方竞品对标表格

### **<font style="color:rgb(51,51,51);">9.4 竞品对标</font>**
全局搜索提供跨多对象的全文检索和过滤器联动，支持类 JQL 搜索语法。对标 Jira Issue Search (JQL), 云效全局搜索。

### **<font style="color:rgb(51,51,51);">9.4 功能概述</font>**
### **<font style="color:rgb(51,51,51);">搜索能力</font>**
| **<font style="color:rgb(255,255,255);">能力</font>** | **<font style="color:rgb(255,255,255);">描述</font>** | **<font style="color:rgb(255,255,255);">优先级</font>** |
| --- | --- | --- |
| <font style="color:rgb(51,51,51);">全文搜索</font> | <font style="color:rgb(51,51,51);">支持标题、描述、评论全文模糊匹配</font> | <font style="color:rgb(51,51,51);">P0</font> |
| <font style="color:rgb(51,51,51);">关键词高亮</font> | <font style="color:rgb(51,51,51);">搜索结果中匹配关键词高亮显示</font> | <font style="color:rgb(51,51,51);">P0</font> |
| <font style="color:rgb(51,51,51);">多对象搜索</font> | <font style="color:rgb(51,51,51);">同时搜索需求</font><font style="color:rgb(51,51,51);">/任务/缺陷/文档/模块/版本</font> | <font style="color:rgb(51,51,51);">P0</font> |
| <font style="color:rgb(51,51,51);">过滤器联动</font> | <font style="color:rgb(51,51,51);">左右分栏：左侧过滤器，右侧结果</font> | <font style="color:rgb(51,51,51);">P0</font> |
| <font style="color:rgb(51,51,51);">高级过滤器</font> | <font style="color:rgb(51,51,51);">支持字段级筛选（如</font><font style="color:rgb(51,51,51);"> priority = high AND assignee = currentUser()）</font> | <font style="color:rgb(51,51,51);">P1</font> |
| <font style="color:rgb(51,51,51);">搜索语法</font> | <font style="color:rgb(51,51,51);">类</font><font style="color:rgb(51,51,51);"> JQL 语法（如 project:ABC status:todo assignee:me）</font> | <font style="color:rgb(51,51,51);">P1</font> |
| <font style="color:rgb(51,51,51);">搜索历史</font> | <font style="color:rgb(51,51,51);">保留最近</font><font style="color:rgb(51,51,51);"> 20 条搜索记录，支持收藏常用搜索</font> | <font style="color:rgb(51,51,51);">P1</font> |
| <font style="color:rgb(51,51,51);">快捷键</font> | <font style="color:rgb(51,51,51);">Ctrl/Cmd+K 聚焦搜索框，Esc 关闭搜索结果页</font> | <font style="color:rgb(51,51,51);">P1</font> |
| <font style="color:rgb(51,51,51);">导出结果</font> | <font style="color:rgb(51,51,51);">搜索结果可导出</font><font style="color:rgb(51,51,51);"> CSV/Excel</font> | <font style="color:rgb(51,51,51);">P2</font> |


### **<font style="color:rgb(51,51,51);">搜索范围与排序</font>**
• 搜索范围：默认当前空间内全项目，可限定为单项目。

• 排序方式：相关度（默认）、创建时间、更新时间、优先级。

• 结果分组：按需求/任务/缺陷类型分组展示，每类最多展示 10 条。

• 搜索建议：输入时实时展示搜索建议。

## **<font style="color:rgb(46,117,182);">9.5 视图与自定义 (Views & Customization)</font>**
视图与自定义能力让用户可以根据不同场景切换不同的需求/任务/缺陷展示形式。对标 Jira Views、TAPD 自定义视图。

数据模型

### **<font style="color:rgb(51,51,51);">9.5 数据模型设计</font>**
视图与自定义的功能需求详述（详见下方详细功能描述）

### **<font style="color:rgb(51,51,51);">9.5 功能需求详述</font>**
交互流程

### **<font style="color:rgb(51,51,51);">9.5 交互流程</font>**
用户故事描述

### **<font style="color:rgb(51,51,51);">9.5 用户故事</font>**
参见下方竞品对标表格

### **<font style="color:rgb(51,51,51);">9.5 竞品对标</font>**
视图与自定义功能描述

### **<font style="color:rgb(51,51,51);">9.5 功能概述</font>**
### **<font style="color:rgb(51,51,51);">视图类型</font>**
| **<font style="color:rgb(255,255,255);">视图类型</font>** | **<font style="color:rgb(255,255,255);">描述</font>** | **<font style="color:rgb(255,255,255);">适用场景</font>** | **<font style="color:rgb(255,255,255);">优先级</font>** |
| --- | --- | --- | --- |
| <font style="color:rgb(51,51,51);">看板</font><font style="color:rgb(51,51,51);"> (Kanban)</font> | <font style="color:rgb(51,51,51);">按状态列展示，支持列内卡片拖拽流转</font> | <font style="color:rgb(51,51,51);">敏捷开发、每日站会</font> | <font style="color:rgb(51,51,51);">P0</font> |
| <font style="color:rgb(51,51,51);">列表</font><font style="color:rgb(51,51,51);"> (List)</font> | <font style="color:rgb(51,51,51);">行式表格，支持列自定义、排序、批量操作</font> | <font style="color:rgb(51,51,51);">详细查看、批量更新</font> | <font style="color:rgb(51,51,51);">P0</font> |
| <font style="color:rgb(51,51,51);">甘特图</font><font style="color:rgb(51,51,51);"> (Gantt)</font> | <font style="color:rgb(51,51,51);">时间轴</font><font style="color:rgb(51,51,51);"> + 依赖关系，支持里程碑标记</font> | <font style="color:rgb(51,51,51);">项目计划、里程碑跟踪</font> | <font style="color:rgb(51,51,51);">P1</font> |
| <font style="color:rgb(51,51,51);">日历</font><font style="color:rgb(51,51,51);"> (Calendar)</font> | <font style="color:rgb(51,51,51);">按日期展示需求/任务/缺陷（基于开始</font><font style="color:rgb(51,51,51);">/截止日期）</font> | <font style="color:rgb(51,51,51);">排期规划、资源日历</font> | <font style="color:rgb(51,51,51);">P1</font> |
| <font style="color:rgb(51,51,51);">电子表格</font><font style="color:rgb(51,51,51);"> (Spreadsheet)</font> | <font style="color:rgb(51,51,51);">类</font><font style="color:rgb(51,51,51);"> Excel 的网格编辑，支持单元格编辑</font> | <font style="color:rgb(51,51,51);">数据批量编辑、导入导出</font> | <font style="color:rgb(51,51,51);">P2</font> |
| <font style="color:rgb(51,51,51);">时间线</font><font style="color:rgb(51,51,51);"> (Timeline)</font> | <font style="color:rgb(51,51,51);">按创建时间轴展示，适合回顾历史</font> | <font style="color:rgb(51,51,51);">项目复盘、审计</font> | <font style="color:rgb(51,51,51);">P2</font> |


### **<font style="color:rgb(51,51,51);">过滤</font>****<font style="color:rgb(51,51,51);">/分组/排序</font>**
119. 过滤器：支持状态/指派人/优先级/标签/日期/迭代/模块/类型等多维度筛选。

120. 分组：支持按状态/指派人/优先级/迭代/模块/类型等维度横向分组。

121. 排序：支持按任意字段升序/降序，支持多级排序。

122. 保存视图：支持保存为「个人视图」「团队视图」「系统默认视图」。

123. 分享视图：团队视图自动对项目成员可见，支持通过链接分享。

124. 默认视图：每个项目可配置默认打开视图（如看板或列表）。

### **<font style="color:rgb(51,51,51);">导入导出</font>**
• 支持 CSV/Excel 导入：从外部系统批量导入需求/任务/缺陷。

• 支持 CSV/Excel/JSON 导出：导出当前视图数据。

• 字段映射：导入时可配置外部字段与Ydsz Plane 字段的映射关系。

• 增量导入：支持识别已有需求/任务/缺陷（依据 external_id）进行更新而非重复创建。

  


## **<font style="color:rgb(46,117,182);">9.6 Webhook & API</font>**
Webhook 和 RESTful API 提供系统集成能力，支持与 CI/CD、IM、外部报表等第三方系统联动。对标 Jira REST API + Webhook、云效开放 API。

Webhook: url, secret, events(JSON), is_active, project(FK), created_by(FK User) WebhookLog: webhook(FK), event_type, payload(JSON), response_status, response_body, created_at, retry_count

### **<font style="color:rgb(51,51,51);">9.6 数据模型设计</font>**
Webhook的功能需求详述（详见下方详细功能描述）

### **<font style="color:rgb(51,51,51);">9.6 功能需求详述</font>**
创建 Webhook → 配置 URL、事件类型、重试规则 → 测试推送 → 查看推送日志

### **<font style="color:rgb(51,51,51);">9.6 交互流程</font>**
开发工程师: 我希望需求/任务/缺陷状态变更时能自动通知钉钉/企业微信。 DevOps: 我需要在需求/任务/缺陷创建时自动触发 Jenkins 构建。 集成方: 我需要 Webhook 支持签名验证和安全重试。

### **<font style="color:rgb(51,51,51);">9.6 用户故事</font>**
参见下方竞品对标表格

### **<font style="color:rgb(51,51,51);">9.6 竞品对标</font>**
Webhook & API 提供系统集成能力，支持事件订阅、请求签名、自动重试等。对标 Jira REST API + Webhook。

### **<font style="color:rgb(51,51,51);">9.6 功能概述</font>**
### **<font style="color:rgb(51,51,51);">Webhook 管理</font>**
125. Webhook CRUD：创建/编辑/删除/启用/禁用 Webhook。

126. 事件订阅：支持需求/任务/缺陷 CRUD/迭代/版本/成员变更等 30+ 事件类型。

127. 请求签名：支持 HMAC-SHA256 签名验证，确保请求来源可信。

128. 重试机制：HTTP 5xx 和 429 自动重试，最多 3 次，指数退避。

129. 日志管理：记录每次推送的请求头/请求体/响应状态/耗时，保留 30 天。

130. 测试模式：可手动触发测试事件，验证 Webhook 接收端是否正常。

### **<font style="color:rgb(51,51,51);">RESTful API</font>**
| **<font style="color:rgb(255,255,255);">能力</font>** | **<font style="color:rgb(255,255,255);">描述</font>** | **<font style="color:rgb(255,255,255);">优先级</font>** |
| --- | --- | --- |
| <font style="color:rgb(51,51,51);">OpenAPI Schema</font> | <font style="color:rgb(51,51,51);">提供</font><font style="color:rgb(51,51,51);"> OpenAPI 3.0 规范文档，支持 Swagger UI</font> | <font style="color:rgb(51,51,51);">P0</font> |
| <font style="color:rgb(51,51,51);">CRUD 端点</font> | <font style="color:rgb(51,51,51);">完整的需求/任务/缺陷</font><font style="color:rgb(51,51,51);">/项目/迭代/版本 CRUD API</font> | <font style="color:rgb(51,51,51);">P0</font> |
| <font style="color:rgb(51,51,51);">认证方式</font> | <font style="color:rgb(51,51,51);">支持</font><font style="color:rgb(51,51,51);"> API Token / OAuth2 / Session 认证</font> | <font style="color:rgb(51,51,51);">P0</font> |
| <font style="color:rgb(51,51,51);">分页查询</font> | <font style="color:rgb(51,51,51);">支持</font><font style="color:rgb(51,51,51);"> cursor 分页和 offset 分页</font> | <font style="color:rgb(51,51,51);">P0</font> |
| <font style="color:rgb(51,51,51);">字段筛选</font> | <font style="color:rgb(51,51,51);">支持</font><font style="color:rgb(51,51,51);"> fields 参数控制返回字段，减少传输</font> | <font style="color:rgb(51,51,51);">P1</font> |
| <font style="color:rgb(51,51,51);">批量操作</font> | <font style="color:rgb(51,51,51);">支持批量创建</font><font style="color:rgb(51,51,51);">/更新/删除（一次最多 100 条）</font> | <font style="color:rgb(51,51,51);">P1</font> |
| <font style="color:rgb(51,51,51);">Rate Limit</font> | <font style="color:rgb(51,51,51);">按用户</font><font style="color:rgb(51,51,51);">/IP 限流，默认 100 次/分钟</font> | <font style="color:rgb(51,51,51);">P0</font> |
| <font style="color:rgb(51,51,51);">SDK 支持</font> | <font style="color:rgb(51,51,51);">提供</font><font style="color:rgb(51,51,51);"> Python/JavaScript SDK 封装</font> | <font style="color:rgb(51,51,51);">P2</font> |


### **<font style="color:rgb(51,51,51);">Webhook 事件类型</font>**
• requirement.created / task.created / defect.created（以及 .updated / .deleted / .status_changed）

• sprint.created / sprint.started / sprint.completed / sprint.deleted

• version.released / version.created / version.updated

• member.added / member.removed / member.role_changed

• comment.created / comment.updated / comment.deleted

• automation.executed / automation.failed

  


## **<font style="color:rgb(46,117,182);">9.7 自动化引擎 (Automation Engine)</font>**
自动化引擎通过 Trigger → Condition → Action 规则模型，帮助用户自动化重复性操作。对标 Jira Automation、TAPD 自动化助手、云效 YAML Pipeline。

AutomationRule: name, project(FK), trigger(JSON), conditions(JSON), actions(JSON), is_active, execution_count, last_executed_at RuleExecution: rule(FK), trigger_event, input_data(JSON), results(JSON), status, duration_ms, created_at

### **<font style="color:rgb(51,51,51);">9.7 数据模型设计</font>**
自动化引擎的功能需求详述（详见下方详细功能描述）

### **<font style="color:rgb(51,51,51);">9.7 功能需求详述</font>**
创建自动化规则→ 配置 Trigger/Condition/Action → 测试规则 → 规则执行历史查看

### **<font style="color:rgb(51,51,51);">9.7 交互流程</font>**
项目经理: 我希望新建需求时自动设置默认优先级，减少手动操作。 Scrum Master: 我期望迭代结束时自动把未完成任务推回 Backlog。 QA: 我期望缺陷状态流转时自动通知相关人员。

### **<font style="color:rgb(51,51,51);">9.7 用户故事</font>**
参见下方竞品对标表格

### **<font style="color:rgb(51,51,51);">9.7 竞品对标</font>**
自动化引擎通过 Trigger → Condition → Action 规则模型，帮助用户自动化重复性操作。对标 Jira Automation, 云效自动化。

### **<font style="color:rgb(51,51,51);">9.7 功能概述</font>**
### **<font style="color:rgb(51,51,51);">规则结构</font>**
| **<font style="color:rgb(255,255,255);">组成</font>** | **<font style="color:rgb(255,255,255);">描述</font>** | **<font style="color:rgb(255,255,255);">示例</font>** |
| --- | --- | --- |
| <font style="color:rgb(51,51,51);">Trigger (触发器)</font> | <font style="color:rgb(51,51,51);">事件驱动或定时触发</font> | <font style="color:rgb(51,51,51);">需求/任务/缺陷 创建/流转/定时/代码提交</font> |
| <font style="color:rgb(51,51,51);">Condition (条件)</font> | <font style="color:rgb(51,51,51);">过滤条件，决定是否执行后续</font><font style="color:rgb(51,51,51);"> Action</font> | <font style="color:rgb(51,51,51);">Priority = High AND Assignee is Empty</font> |
| <font style="color:rgb(51,51,51);">Action (动作)</font> | <font style="color:rgb(51,51,51);">执行的操作</font> | <font style="color:rgb(51,51,51);">Transition / Assign / Notify / Create / Webhook</font> |


### **<font style="color:rgb(51,51,51);">内置规则模板</font>**
131. 子任务全完成时自动完成父需求/任务/缺陷（需求/任务/缺陷 Trigger + Condition + Transition Action）

132. 新缺陷自动指派关注者（缺陷创建时通知 Tech Lead）

133. 任务逾期自动提醒经办人（Scheduled Trigger + 3天未更新条件 + 邮件通知）

134. 版本发布时自动流转该版本中所有需求/任务/缺陷为已完成

135. 故事点变更时自动汇总 Epic 层故事点总数

136. 任务转入 In Progress 时自动设置开始日期为今天

137. 指派人为空时从团队列表中选择工作量最轻的人自动指派

### **<font style="color:rgb(51,51,51);">自定义规则能力</font>**
• IF / ELSE 分支：支持条件分支逻辑。

• 多动作链：一个规则可执行多个动作（最多 10 个）。

• 定时触发：支持 Cron 表达式配置（如每天 9:00 检查逾期任务）。

• 字段值复制：跨需求/任务/缺陷复制字段值（如子需求继承父需求模块）。

• Webhook 动作：执行时向外部系统发送自定义 HTTP 请求。

• 日志与审计：记录每次规则执行的时间、触发事件、执行结果、耗时。

### **<font style="color:rgb(51,51,51);">执行引擎</font>**
• 规则编译为 Celery Task，基于事件驱动异步执行。

• 支持规则启用/禁用、手动触发测试。

• 失败告警：连续 3 次执行失败时禁用规则并通知管理员。

• 并发控制：同一需求/任务/缺陷的自动化规则串行执行，避免状态竞争。

  


## **<font style="color:rgb(46,117,182);">9.8 研发效能度量 (Dev Efficiency Metrics)</font>**
研发效能度量体系基于 DORA 框架和国内信通院《软件研发效能度量规范》，从交付效率、交付质量、交付能力、资源效率四个维度量化团队效能。对标 ONES Performance、云效研发效能、Gitee Insight。

数据模型

### **<font style="color:rgb(51,51,51);">9.8 数据模型设计</font>**
研发效能度量的功能需求详述（详见下方详细功能描述）

### **<font style="color:rgb(51,51,51);">9.8 功能需求详述</font>**
交互流程

### **<font style="color:rgb(51,51,51);">9.8 交互流程</font>**
用户故事描述

### **<font style="color:rgb(51,51,51);">9.8 用户故事</font>**
参见下方竞品对标表格

### **<font style="color:rgb(51,51,51);">9.8 竞品对标</font>**
研发效能度量功能描述

### **<font style="color:rgb(51,51,51);">9.8 功能概述</font>**
### **<font style="color:rgb(51,51,51);">DORA 四大指标</font>**
| **<font style="color:rgb(255,255,255);">指标</font>** | **<font style="color:rgb(255,255,255);">定义</font>** | **<font style="color:rgb(255,255,255);">计算方法</font>** | **<font style="color:rgb(255,255,255);">优先级</font>** |
| --- | --- | --- | --- |
| <font style="color:rgb(51,51,51);">部署频率</font><font style="color:rgb(51,51,51);"> (DF)</font> | <font style="color:rgb(51,51,51);">单位时间部署到生产的次数</font> | <font style="color:rgb(51,51,51);">部署次数</font><font style="color:rgb(51,51,51);"> / 周期</font> | <font style="color:rgb(51,51,51);">P0</font> |
| <font style="color:rgb(51,51,51);">变更交付周期</font><font style="color:rgb(51,51,51);"> (LT)</font> | <font style="color:rgb(51,51,51);">代码提交到部署上线的平均时长</font> | <font style="color:rgb(51,51,51);">Mean(commit→deploy)</font> | <font style="color:rgb(51,51,51);">P0</font> |
| <font style="color:rgb(51,51,51);">变更失败率</font><font style="color:rgb(51,51,51);"> (CFR)</font> | <font style="color:rgb(51,51,51);">部署后导致故障</font><font style="color:rgb(51,51,51);">/降级的比例</font> | <font style="color:rgb(51,51,51);">(失败数/总数) × 100%</font> | <font style="color:rgb(51,51,51);">P0</font> |
| <font style="color:rgb(51,51,51);">服务恢复时间</font><font style="color:rgb(51,51,51);"> (MTTR)</font> | <font style="color:rgb(51,51,51);">故障发生到恢复的平均时长</font> | <font style="color:rgb(51,51,51);">Mean(故障→恢复)</font> | <font style="color:rgb(51,51,51);">P1</font> |


### **<font style="color:rgb(51,51,51);">流动效率指标</font>**
138. 端到端交付周期 (Lead Time)：需求创建 → 需求完成的平均天数。

139. 开发周期 (Sprint Time)：In Progress → Done 的平均天数。

140. 在制品数量 (WIP)：同时处于 In Progress 状态的需求数。

141. 吞吐量 (Throughput)：每周完成的需求/任务数量。

142. 流动效率 = 开发周期 / 端到端交付周期 × 100%。

### **<font style="color:rgb(51,51,51);">质量指标</font>**
• 缺陷密度：每需求/每故事点的缺陷数。

• 缺陷逃逸率：线上发现的缺陷 / 总缺陷数 × 100%。

• 返工率：被关联到原需求的缺陷数 / 该需求的总缺陷数 × 100%。

• 测试通过率：通过的测试用例 / 总测试用例 × 100%。

### **<font style="color:rgb(51,51,51);">效率视图与报表</font>**
143. 累计流图 (CFD)：展示各状态随时间的分布变化，反映流程瓶颈。

144. 控制图 (Control Chart)：展示周期时间的分布，识别异常值。

145. 散点图 (Scatterplot)：按完成日期展示每个需求的周期时间，叠加趋势线。

146. 速率图 (Velocity)：近 6-10 个 Sprint 完成故事点趋势。

147. 仪表盘模板：预置「工程效能仪表盘」「QA 质量仪表盘」「PMO 战略仪表盘」。

### **<font style="color:rgb(51,51,51);">数据采集与权限</font>**
• 自动采集：基于需求/任务/缺陷状态变更、迭代状态自动计算指标。

• 数据校准：允许管理员修正异常数据点。

• 权限隔离：效能数据按项目维度隔离，仅项目成员可见。

• 数据导出：效能数据支持 API 导出，可对接外部 BI 工具。

  


# **<font style="color:rgb(31,78,121);">10. 非功能需求</font>**
| **<font style="color:rgb(255,255,255);">维度</font>** | **<font style="color:rgb(255,255,255);">目标</font>** | **<font style="color:rgb(255,255,255);">措施</font>** | **<font style="color:rgb(255,255,255);">优先级</font>** |
| --- | --- | --- | --- |
| <font style="color:rgb(51,51,51);">性能</font> | <font style="color:rgb(51,51,51);">P95 页面加载 ≤ 2s</font> | <font style="color:rgb(51,51,51);">CDN/索引/Redis/按需查询</font> | <font style="color:rgb(51,51,51);">P0</font> |
| <font style="color:rgb(51,51,51);">可用性</font> | <font style="color:rgb(51,51,51);">SLA ≥ 99.9%</font> | <font style="color:rgb(51,51,51);">主从</font><font style="color:rgb(51,51,51);">/无状态/跨可用区</font> | <font style="color:rgb(51,51,51);">P0</font> |
| <font style="color:rgb(51,51,51);">安全</font> | <font style="color:rgb(51,51,51);">等保三级基线</font> | <font style="color:rgb(51,51,51);">HTTPS/OAuth/加盐/审计</font> | <font style="color:rgb(51,51,51);">P0</font> |
| <font style="color:rgb(51,51,51);">可扩展</font> | <font style="color:rgb(51,51,51);">100+ 并发项目</font> | <font style="color:rgb(51,51,51);">读写分离</font><font style="color:rgb(51,51,51);">/Celery/Redis集群</font> | <font style="color:rgb(51,51,51);">P1</font> |
| <font style="color:rgb(51,51,51);">可运维</font> | <font style="color:rgb(51,51,51);">一键部署</font> | <font style="color:rgb(51,51,51);">Docker Compose/Helm Chart</font> | <font style="color:rgb(51,51,51);">P0</font> |
| <font style="color:rgb(51,51,51);">国产化</font> | <font style="color:rgb(51,51,51);">信创兼容</font> | <font style="color:rgb(51,51,51);">麒麟</font><font style="color:rgb(51,51,51);">/统信/达梦/鲲鹏</font> | <font style="color:rgb(51,51,51);">P1</font> |


# **<font style="color:rgb(31,78,121);">11. 关键用户旅程</font>**
## **<font style="color:rgb(46,117,182);">11.1 产品经理的典型一天</font>**
• 打开个人工作台→ 本周待评审需求 10 条

• 需求池中 5 条 P0 加星标

• 创建 v2.6 版本 → 规划 3 个迭代 → 拆分大需求为子需求

• 进入 Sprint-14 规划 → 将 8 个需求拖拽入迭代

• 处理 @我通知并回复评论

• 查看项目仪表盘：当前进度略滞后，发出风险预警

## **<font style="color:rgb(46,117,182);">11.2 技术经理的一周</font>**
• 周一：创建迭代任务→ 5 个主任务拆为 18 个子任务 → 分配工时

• 周二：Daily Standup → 更新昨日完成 / 今日计划 / Blocker

• 周三：代码关联→ CR 通过 → 任务状态流转 '已提交'

• 周四：处理测试反馈缺陷→ 关联到任务

• 周五：迭代复盘→ 回顾 velocity → 归档完成需求/任务/缺陷

## **<font style="color:rgb(46,117,182);">11.3 测试经理的版本交付</font>**
• 版本前 2 周：编写测试用例、规划测试计划

• 版本前 1 周：执行回归测试

• 需求测试发现Defect → 一键提 Defect → 关联到需求

• Defect 修复后验证通过→ 流转 '已关闭' → 关联修复 PR

• 版本交付日：查看版本交付报告

• 查看项目仪表盘：版本进度、模块缺陷分布

# **<font style="color:rgb(31,78,121);">12. 成功指标</font>**
| **<font style="color:rgb(255,255,255);">指标</font>** | **<font style="color:rgb(255,255,255);">目标</font>** | **<font style="color:rgb(255,255,255);">周期</font>** | **<font style="color:rgb(255,255,255);">说明</font>** |
| --- | --- | --- | --- |
| <font style="color:rgb(51,51,51);">新用户</font><font style="color:rgb(51,51,51);"> 7 日激活率</font> | <font style="color:rgb(51,51,51);">≥ 60%</font> | <font style="color:rgb(51,51,51);">周</font> | <font style="color:rgb(51,51,51);">首次创建需求/任务/缺陷</font> |
| <font style="color:rgb(51,51,51);">需求/任务/缺陷创建数</font><font style="color:rgb(51,51,51);">/用户/周</font> | <font style="color:rgb(51,51,51);">≥ 5</font> | <font style="color:rgb(51,51,51);">周</font> | <font style="color:rgb(51,51,51);">核心活跃度</font> |
| <font style="color:rgb(51,51,51);">需求/任务/缺陷平均完成周期</font> | <font style="color:rgb(51,51,51);">≤ 5 天</font> | <font style="color:rgb(51,51,51);">月</font> | <font style="color:rgb(51,51,51);">从创建到完成</font> |
| <font style="color:rgb(51,51,51);">Sprint</font><font style="color:rgb(51,51,51);"> </font><font style="color:rgb(51,51,51);">按时完成率</font> | <font style="color:rgb(51,51,51);">≥ 70%</font> | <font style="color:rgb(51,51,51);">每</font><font style="color:rgb(51,51,51);">Sprint</font> | <font style="color:rgb(51,51,51);">交付准时性</font> |
| <font style="color:rgb(51,51,51);">版本按期发布率</font> | <font style="color:rgb(51,51,51);">≥ 80%</font> | <font style="color:rgb(51,51,51);">每</font><font style="color:rgb(51,51,51);"> Version</font> | <font style="color:rgb(51,51,51);">里程碑达成</font> |
| <font style="color:rgb(51,51,51);">30 日用户留存</font> | <font style="color:rgb(51,51,51);">≥ 40%</font> | <font style="color:rgb(51,51,51);">月</font> | <font style="color:rgb(51,51,51);">用户粘性</font> |
| <font style="color:rgb(51,51,51);">DAU/MAU</font> | <font style="color:rgb(51,51,51);">≥ 35%</font> | <font style="color:rgb(51,51,51);">月</font> | <font style="color:rgb(51,51,51);">月活占比</font> |
| <font style="color:rgb(51,51,51);">GitHub Stars / 月增长</font> | <font style="color:rgb(51,51,51);">8%+</font> | <font style="color:rgb(51,51,51);">月</font> | <font style="color:rgb(51,51,51);">开源热度</font> |


# **<font style="color:rgb(31,78,121);">13. 开放性问题</font>**
| **<font style="color:rgb(255,255,255);">问题</font>** | **<font style="color:rgb(255,255,255);">阻塞</font>** | **<font style="color:rgb(255,255,255);">负责人</font>** | **<font style="color:rgb(255,255,255);">状态</font>** |
| --- | --- | --- | --- |
| <font style="color:rgb(51,51,51);">多租户隔离：</font><font style="color:rgb(51,51,51);">Schema-per-tenant 还是 RLS？</font> | <font style="color:rgb(51,51,51);">是</font> | <font style="color:rgb(51,51,51);">架构师</font> | <font style="color:rgb(51,51,51);">方案讨论</font> |
| <font style="color:rgb(51,51,51);">IM 集成：仅 Webhook 还是原生机器人？</font> | <font style="color:rgb(51,51,51);">否</font> | <font style="color:rgb(51,51,51);">前端</font> | <font style="color:rgb(51,51,51);">评估中</font> |
| <font style="color:rgb(51,51,51);">测试用例：自建</font><font style="color:rgb(51,51,51);"> vs 对接 KiwiTCMS？</font> | <font style="color:rgb(51,51,51);">是</font> | <font style="color:rgb(51,51,51);">产品</font><font style="color:rgb(51,51,51);">+QA</font> | <font style="color:rgb(51,51,51);">待定</font> |
| <font style="color:rgb(51,51,51);">版本</font><font style="color:rgb(51,51,51);">-迭代映射规则与自动化策略</font> | <font style="color:rgb(51,51,51);">否</font> | <font style="color:rgb(51,51,51);">产品</font><font style="color:rgb(51,51,51);">+后端</font> | <font style="color:rgb(51,51,51);">待讨论</font> |
| <font style="color:rgb(51,51,51);">仪表盘小部件开放策略</font> | <font style="color:rgb(51,51,51);">否</font> | <font style="color:rgb(51,51,51);">产品</font><font style="color:rgb(51,51,51);">+架构</font> | <font style="color:rgb(51,51,51);">调研中</font> |
| <font style="color:rgb(51,51,51);">商业版</font><font style="color:rgb(51,51,51);"> vs 开源版功能边界</font> | <font style="color:rgb(51,51,51);">否</font> | <font style="color:rgb(51,51,51);">CEO</font> | <font style="color:rgb(51,51,51);">战略讨论</font> |


# **<font style="color:rgb(31,78,121);">14. 术语表</font>**
| **<font style="color:rgb(255,255,255);">术语</font>** | **<font style="color:rgb(255,255,255);">英文</font>** | **<font style="color:rgb(255,255,255);">定义</font>** |
| --- | --- | --- |
| <font style="color:rgb(51,51,51);">工作空间</font> | <font style="color:rgb(51,51,51);">Workspace</font> | <font style="color:rgb(51,51,51);">组织级容器，隔离企业</font><font style="color:rgb(51,51,51);">/团队数据</font> |
| <font style="color:rgb(51,51,51);">项目</font> | <font style="color:rgb(51,51,51);">Project</font> | <font style="color:rgb(51,51,51);">具体产品的研发项目，承载需求</font><font style="color:rgb(51,51,51);">/任务/缺陷</font> |
| <font style="color:rgb(51,51,51);">版本</font> | <font style="color:rgb(51,51,51);">Version</font> | <font style="color:rgb(51,51,51);">产品发版里程碑，含</font><font style="color:rgb(51,51,51);"> 1~N 个迭代</font> |
| <font style="color:rgb(51,51,51);">迭代</font> | <font style="color:rgb(51,51,51);">Sprint</font> | <font style="color:rgb(51,51,51);">固定周期开发单元</font> |
| <font style="color:rgb(51,51,51);">需求</font> | <font style="color:rgb(51,51,51);">Requirement</font> | <font style="color:rgb(51,51,51);">产品经理提出，可拆分子需求</font> |
| <font style="color:rgb(51,51,51);">子需求</font> | <font style="color:rgb(51,51,51);">Sub-requirement</font> | <font style="color:rgb(51,51,51);">需求拆分后的子需求/任务/缺陷</font> |
| <font style="color:rgb(51,51,51);">任务</font> | <font style="color:rgb(51,51,51);">Task</font> | <font style="color:rgb(51,51,51);">技术经理提出，可拆分子任务</font> |
| <font style="color:rgb(51,51,51);">子任务</font> | <font style="color:rgb(51,51,51);">Sub-task</font> | <font style="color:rgb(51,51,51);">任务拆分后的子需求/任务/缺陷</font> |
| <font style="color:rgb(51,51,51);">缺陷</font> | <font style="color:rgb(51,51,51);">Defect</font> | <font style="color:rgb(51,51,51);">由需求</font><font style="color:rgb(51,51,51);">/任务产生，可拆分子缺陷</font> |
| <font style="color:rgb(51,51,51);">子缺陷</font> | <font style="color:rgb(51,51,51);">Sub-defect</font> | <font style="color:rgb(51,51,51);">缺陷拆分后的子单元</font> |
| <font style="color:rgb(51,51,51);">模块</font> | <font style="color:rgb(51,51,51);">Module</font> | <font style="color:rgb(51,51,51);">需求/任务/缺陷归档属性，非独立对象</font> |
| <font style="color:rgb(51,51,51);">收件箱</font> | <font style="color:rgb(51,51,51);">Intake / Inbox</font> | <font style="color:rgb(51,51,51);">反馈收集通道</font> |
| <font style="color:rgb(51,51,51);">项目仪表盘</font> | <font style="color:rgb(51,51,51);">Project Dashboard</font> | <font style="color:rgb(51,51,51);">面向</font><font style="color:rgb(51,51,51);"> PM 的项目级聚合视图</font> |
| <font style="color:rgb(51,51,51);">个人工作台</font> | <font style="color:rgb(51,51,51);">My Work</font> | <font style="color:rgb(51,51,51);">面向成员的个人级聚合视图</font> |
| <font style="color:rgb(51,51,51);">WBS 拆分</font> | <font style="color:rgb(51,51,51);">WBS Breakdown</font> | <font style="color:rgb(51,51,51);">工作分解结构，支持三级层级</font> |


# **<font style="color:rgb(31,78,121);">15. 修订历史</font>**
| **<font style="color:rgb(255,255,255);">版本</font>** | **<font style="color:rgb(255,255,255);">日期</font>** | **<font style="color:rgb(255,255,255);">作者</font>** | **<font style="color:rgb(255,255,255);">变更说明</font>** |
| --- | --- | --- | --- |
| <font style="color:rgb(51,51,51);">v0.1</font> | <font style="color:rgb(51,51,51);">2026-08-06</font> | <font style="color:rgb(51,51,51);">Marvin Lee</font> | <font style="color:rgb(51,51,51);">初版</font><font style="color:rgb(51,51,51);"> PRD，基于 code base 结构分析</font> |
| <font style="color:rgb(51,51,51);">V</font><font style="color:rgb(51,51,51);">0</font><font style="color:rgb(51,51,51);">.</font><font style="color:rgb(51,51,51);">2</font> | <font style="color:rgb(51,51,51);">2026-08-06</font> | <font style="color:rgb(51,51,51);">Marvin Lee</font> | <font style="color:rgb(51,51,51);">竞品对标</font><font style="color:rgb(51,51,51);"> + 角色-需求-任务-缺陷 + 用户旅程</font> |
| <font style="color:rgb(51,51,51);">V</font><font style="color:rgb(51,51,51);">0</font><font style="color:rgb(51,51,51);">.</font><font style="color:rgb(51,51,51);">3</font> | <font style="color:rgb(51,51,51);">2026-08-06</font> | <font style="color:rgb(51,51,51);">Marvin Lee</font> | <font style="color:rgb(51,51,51);">模块</font><font style="color:rgb(51,51,51);">=归档属性 + 版本=1~N迭代</font> |
| <font style="color:rgb(51,51,51);">V</font><font style="color:rgb(51,51,51);">0</font><font style="color:rgb(51,51,51);">.</font><font style="color:rgb(51,51,51);">4</font> | <font style="color:rgb(51,51,51);">2026-08-06</font> | <font style="color:rgb(51,51,51);">Marvin Lee</font> | <font style="color:rgb(51,51,51);">核心修正：需求拆解不产生任务</font> |
| <font style="color:rgb(51,51,51);">V</font><font style="color:rgb(51,51,51);">0.5</font> | <font style="color:rgb(51,51,51);">2026-08-06</font> | <font style="color:rgb(51,51,51);">Marvin Lee</font> | <font style="color:rgb(51,51,51);">融合</font><font style="color:rgb(51,51,51);"> v</font><font style="color:rgb(51,51,51);">0.4</font><font style="color:rgb(51,51,51);"> </font><font style="color:rgb(51,51,51);">功能框架</font><font style="color:rgb(51,51,51);"> + v</font><font style="color:rgb(51,51,51);">0.5</font><font style="color:rgb(51,51,51);"> </font><font style="color:rgb(51,51,51);">属性竞品对比</font> |
| <font style="color:rgb(51,51,51);">V</font><font style="color:rgb(51,51,51);">0.6</font> | <font style="color:rgb(51,51,51);">2026-08-06</font> | <font style="color:rgb(51,51,51);">Marvin Lee</font> | <font style="color:rgb(51,51,51);">效率增强功能详细需求细化</font> |
| <font style="color:rgb(51,51,51);">V</font><font style="color:rgb(51,51,51);">1.0</font> | <font style="color:rgb(51,51,51);">2026-08-06</font> | <font style="color:rgb(51,51,51);">Marvin Lee</font> | <font style="color:rgb(51,51,51);">融合</font><font style="color:rgb(51,51,51);"> v</font><font style="color:rgb(51,51,51);">0.5</font><font style="color:rgb(51,51,51);"> </font><font style="color:rgb(51,51,51);">融合版</font><font style="color:rgb(51,51,51);"> + v</font><font style="color:rgb(51,51,51);">0.6</font><font style="color:rgb(51,51,51);"> </font><font style="color:rgb(51,51,51);">效率增强细化版，形成最终完整</font><font style="color:rgb(51,51,51);"> PRD</font> |




<font style="color:rgb(153,153,153);">— Ydsz Plane 产品需求文档 v1.0 最终完整版 完 —</font>

