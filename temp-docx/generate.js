const fs = require('fs');
const {
  Document, Packer, Paragraph, TextRun, Table, TableRow, TableCell,
  Header, Footer, AlignmentType, LevelFormat,
  TableOfContents, HeadingLevel, BorderStyle, WidthType, ShadingType,
  VerticalAlign, PageNumber, PageBreak, PageOrientation
} = require('docx');

// 样式常量
const border = { style: BorderStyle.SINGLE, size: 1, color: "CCCCCC" };
const borders = { top: border, bottom: border, left: border, right: border };
const headerShading = { fill: "1F4E79", type: ShadingType.CLEAR };
const altRowShading = { fill: "F2F7FB", type: ShadingType.CLEAR };

// 辅助函数 - 创建表头单元格
function headerCell(text, width) {
  return new TableCell({
    borders,
    width: { size: width, type: WidthType.DXA },
    shading: headerShading,
    margins: { top: 60, bottom: 60, left: 80, right: 80 },
    verticalAlign: VerticalAlign.CENTER,
    children: [new Paragraph({
      children: [new TextRun({ text, bold: true, color: "FFFFFF", font: "Microsoft YaHei", size: 20 })]
    })]
  });
}

// 辅助函数 - 创建普通单元格
function cell(text, width, shading) {
  const opts = {
    borders,
    width: { size: width, type: WidthType.DXA },
    margins: { top: 60, bottom: 60, left: 80, right: 80 },
    children: Array.isArray(text) ? text : [new Paragraph({
      children: [new TextRun({ text, font: "Microsoft YaHei", size: 20 })]
    })]
  };
  if (shading) opts.shading = shading;
  return new TableCell(opts);
}

// 辅助函数 - 创建带文本的单元格（支持多行）
function cellWithRuns(runs, width, shading) {
  const opts = {
    borders,
    width: { size: width, type: WidthType.DXA },
    margins: { top: 60, bottom: 60, left: 80, right: 80 },
    children: [new Paragraph({ children: runs })]
  };
  if (shading) opts.shading = shading;
  return new TableCell(opts);
}

// ============ 文档内容构建 ============

const children = [];

// 封面标题
children.push(new Paragraph({
  spacing: { before: 3600, after: 600 },
  alignment: AlignmentType.CENTER,
  children: [new TextRun({ text: "Ydsz Plane", font: "Microsoft YaHei", size: 56, bold: true, color: "1F4E79" })]
}));

children.push(new Paragraph({
  spacing: { after: 200 },
  alignment: AlignmentType.CENTER,
  children: [new TextRun({ text: "详细技术说明文档", font: "Microsoft YaHei", size: 36, color: "404040" })]
}));

children.push(new Paragraph({
  spacing: { after: 1200 },
  alignment: AlignmentType.CENTER,
  children: [new TextRun({ text: "版本 v1.0 | 2026-08-08", font: "Microsoft YaHei", size: 22, color: "808080" })]
}));

// 版本信息表
children.push(new Table({
  width: { size: 9000, type: WidthType.DXA },
  columnWidths: [3000, 6000],
  rows: [
    new TableRow({
      children: [
        cell("项目名称", 3000, altRowShading),
        cell("Ydsz Plane", 6000, altRowShading)
      ]
    },
    new TableRow({
      children: [
        cell("项目定位", 3000),
        cell("面向中国软件团队的开源项目管理平台", 6000)
      ]
    },
    new TableRow({
      children: [
        cell("技术栈", 3000, altRowShading),
        cell("Go 1.26 + Gin 1.12 | Vue 3.5 + Vite 6 | PostgreSQL 18 | Redis 8 | RabbitMQ 4", 6000, altRowShading)
      ]
    },
    new TableRow({
      children: [
        cell("架构模式", 3000),
        cell("模块化单体 (Modular Monolith) + 异步 Worker 进程", 6000)
      ]
    },
    new TableRow({
      children: [
        cell("里程碑状态", 3000, altRowShading),
        cell("M0-M6 全部完成，后端 API + 前端视图全部交付", 6000, altRowShading)
      ]
    },
    new TableRow({
      children: [
        cell("许可协议", 3000),
        cell("MIT License", 6000)
      ]
    })
  ]
}));

children.push(new Paragraph({ spacing: { before: 1200 }, children: [new PageBreak()] }));

// ========== 目录 ==========
children.push(new Paragraph({
  heading: HeadingLevel.HEADING_1,
  children: [new TextRun({ text: "目录", font: "Microsoft YaHei", size: 32, bold: true })],
  pageBreakBefore: true
}));

children.push(new TableOfContents("目录", { hyperlink: true, headingStyleRange: "1-3" }));

children.push(new Paragraph({ children: [new PageBreak()] }));

// ========== 第一章：项目概述 ==========
children.push(new Paragraph({
  heading: HeadingLevel.HEADING_1,
  pageBreakBefore: true,
  children: [new TextRun({ text: "1. 项目概述", font: "Microsoft YaHei", size: 32, bold: true })]
}));

children.push(new Paragraph({
  heading: HeadingLevel.HEADING_2,
  children: [new TextRun({ text: "1.1 项目定位", font: "Microsoft YaHei", size: 26, bold: true })]
}));

children.push(new Paragraph({
  spacing: { after: 200 },
  children: [new TextRun({
    text: "Ydsz Plane 是一款面向中国软件团队的开源、自托管项目管理平台。核心域为「工作空间 → 项目 → 工作项（需求/任务/缺陷）」，辅以迭代管理、版本日、仪表盘、个人工作台、全文搜索、通知系统、Webhook 集成、自动化规则引擎、效能度量等增强能力。",
    font: "Microsoft YaHei", size: 22
  })]
}));

children.push(new Paragraph({
  spacing: { after: 200 },
  children: [new TextRun({
    text: "项目以模块化单体架构起步，基于领域驱动设计（DDD）划分限界上下文，事件驱动保证模块间解耦，为将来微服务拆分预留接缝。同时全面适配信创生态（麒麟/统信/openEuler + ARM64 + 国密算法），满足等保三级基线要求。",
    font: "Microsoft YaHei", size: 22
  })]
}));

children.push(new Paragraph({
  heading: HeadingLevel.HEADING_2,
  children: [new TextRun({ text: "1.2 技术选型总览", font: "Microsoft YaHei", size: 26, bold: true })]
}));

children.push(new Table({
  width: { size: 9000, type: WidthType.DXA },
  columnWidths: [1800, 2800, 4400],
  rows: [
    new TableRow({
      children: [
        headerCell("层次", 1800),
        headerCell("技术选型", 2800),
        headerCell("说明", 4400)
      ]
    }),
    new TableRow({
      children: [
        cell("后端", 1800, altRowShading),
        cell("Go 1.26.5 + Gin 1.12", 2800, altRowShading),
        cell("模块化单体，DDD 轻量分层", 4400, altRowShading)
      ]
    }),
    new TableRow({
      children: [
        cell("前端", 1800),
        cell("Vue 3.5 + TypeScript", 2800),
        cell("组合式 API + Pinia + Vite 6", 4800)
      ]
    }),
    new TableRow({
      children: [
        cell("数据库", 1800, altRowShading),
        cell("PostgreSQL 18", 2800, altRowShading),
        cell("ACID + JSONB + RLS + 全文检索", 4400, altRowShading)
      ]
    }),
    new TableRow({
      children: [
        cell("缓存", 1800),
        cell("Redis 8", 2800),
        cell("分布式锁、限流、会话、WS 扇出", 4400)
      ]
    }),
    new TableRow({
      children: [
        cell("消息队列", 1800, altRowShading),
        cell("RabbitMQ 4", 2800, altRowShading),
        cell("Outbox 投递 + 任务队列 + DLX/DLQ", 4400, altRowShading)
      ]
    }),
    new TableRow({
      children: [
        cell("对象存储", 1800),
        cell("MinIO (可选)", 2800),
        cell("附件、Logo 预签名访问", 4400)
      ]
    }),
    new TableRow({
      children: [
        cell("部署", 1800, altRowShading),
        cell("Docker Compose", 2800, altRowShading),
        cell("一键启动核心栈 + full profile", 4400, altRowShading)
      ]
    })
  ]
}));

children.push(new Paragraph({ children: [new PageBreak()] }));

// ========== 第二章：系统架构 ==========
children.push(new Paragraph({
  heading: HeadingLevel.HEADING_1,
  pageBreakBefore: true,
  children: [new TextRun({ text: "2. 系统架构", font: "Microsoft YaHei", size: 32, bold: true })]
}));

children.push(new Paragraph({
  heading: HeadingLevel.HEADING_2,
  children: [new TextRun({ text: "2.1 架构形态", font: "Microsoft YaHei", size: 26, bold: true })]
}));

children.push(new Paragraph({
  spacing: { after: 200 },
  children: [new TextRun({
    text: "采用「模块化单体 + 异步 Worker 进程」的部署形态。API 进程无状态可水平扩展，Worker 进程承载后台任务（事件投递、通知分发、搜索索引、Webhook 重试、定时任务）。",
    font: "Microsoft YaHei", size: 22
  })]
}));

children.push(new Paragraph({
  spacing: { after: 200 },
  children: [new TextRun({
    text: "分层架构：Interfaces (HTTP Handler) → Application (用例编排) → Domain (限界上下文) → Infrastructure (存储、缓存、消息队列)。",
    font: "Microsoft YaHei", size: 22
  })]
}));

children.push(new Paragraph({
  heading: HeadingLevel.HEADING_2,
  children: [new TextRun({ text: "2.2 进程模型", font: "Microsoft YaHei", size: 26, bold: true })]
}));

children.push(new Table({
  width: { size: 9000, type: WidthType.DXA },
  columnWidths: [2000, 2000, 5000],
  rows: [
    new TableRow({
      children: [
        headerCell("进程", 2000),
        headerCell("入口", 2000),
        headerCell("职责", 5000)
      ]
    }),
    new TableRow({
      children: [
        cell("API Server", 2000, altRowShading),
        cell("cmd/api/main.go", 2000, altRowShading),
        cell("HTTP 请求处理、认证鉴权、中间件链、路由分发", 5000, altRowShading)
      ]
    }),
    new TableRow({
      children: [
        cell("Worker", 2000),
        cell("cmd/worker/main.go", 2000),
        cell("Outbox Relay、任务消费、定时事件、通知投递、指标聚合", 5000)
      ]
    }),
    new TableRow({
      children: [
        cell("Migrate", 2000, altRowShading),
        cell("cmd/migrate/main.go", 2000, altRowShading),
        cell("数据库迁移执行 (golang-migrate)", 5000, altRowShading)
      ]
    })
  ]
}));

children.push(new Paragraph({ children: [new PageBreak()] }));

children.push(new Paragraph({
  heading: HeadingLevel.HEADING_2,
  children: [new TextRun({ text: "2.3 关键架构决策 (ADR)", font: "Microsoft YaHei", size: 26, bold: true })]
}));

children.push(new Table({
  width: { size: 9000, type: WidthType.DXA },
  columnWidths: [800, 2800, 2200, 3200],
  rows: [
    new TableRow({
      children: [
        headerCell("编号", 800),
        headerCell("决策", 2800),
        headerCell("结论", 2200),
        headerCell("理由", 3200)
      ]
    }),
    new TableRow({
      children: [
        cell("ADR-1", 800, altRowShading),
        cell("单体 vs 微服务", 2800, altRowShading),
        cell("模块化单体", 2200, altRowShading),
        cell("交付优先，事件驱动保可拆分性", 3200, altRowShading)
      ]
    }),
    new TableRow({
      children: [
        cell("ADR-2", 800),
        cell("多租户隔离", 2800),
        cell("共享 Schema + RLS", 2200),
        cell("运维成本低 + DB 级兜底", 3200)
      ]
    }),
    new TableRow({
      children: [
        cell("ADR-3", 800, altRowShading),
        cell("工作项建模", 2800, altRowShading),
        cell("单表 + type 区分", 2200, altRowShading),
        cell("简化查询 + 类型化 JSONB", 3200, altRowShading)
      ]
    }),
    new TableRow({
      children: [
        cell("ADR-4", 800),
        cell("事件机制", 2800),
        cell("Outbox → RabbitMQ", 2200),
        cell("可靠投递 + 死信队列", 3200)
      ]
    }),
    new TableRow({
      children: [
        cell("ADR-5", 800, altRowShading),
        cell("自动化引擎", 2800, altRowShading),
        cell("JSON DSL + TCA", 2200, altRowShading),
        cell("Go 栈对齐 + 复用任务队列", 3200, altRowShading)
      ]
    }),
    new TableRow({
      children: [
        cell("ADR-6", 800),
        cell("实时推送", 2800),
        cell("WSS + Redis", 2200),
        cell("低延迟 + 多节点扇出", 3200)
      ]
    }),
    new TableRow({
      children: [
        cell("ADR-7", 800, altRowShading),
        cell("全文检索", 2800, altRowShading),
        cell("PG FTS (ES 可选)", 2200, altRowShading),
        cell("DB 为唯一事实源", 3200, altRowShading)
      ]
    })
  ]
}));

children.push(new Paragraph({ children: [new PageBreak()] }));

// ========== 第三章：目录结构 ==========
children.push(new Paragraph({
  heading: HeadingLevel.HEADING_1,
  pageBreakBefore: true,
  children: [new TextRun({ text: "3. 项目目录结构与模块说明", font: "Microsoft YaHei", size: 32, bold: true })]
}));

children.push(new Paragraph({
  heading: HeadingLevel.HEADING_2,
  children: [new TextRun({ text: "3.1 顶层目录布局", font: "Microsoft YaHei", size: 26, bold: true })]
}));

children.push(new Table({
  width: { size: 9000, type: WidthType.DXA },
  columnWidths: [2400, 6600],
  rows: [
    new TableRow({
      children: [
        headerCell("路径", 2400),
        headerCell("说明", 6600)
      ]
    }),
    new TableRow({
      children: [
        new TableCell({
          borders, width: { size: 2400, type: WidthType.DXA }, margins: { top: 60, bottom: 60, left: 80, right: 80 },
          shading: altRowShading,
          children: [new Paragraph({ children: [new TextRun({ text: "cmd/", font: "Consolas", size: 20, color: "1F4E79" })] })]
        }),
        new TableCell({
          borders, width: { size: 6600, type: WidthType.DXA }, margins: { top: 60, bottom: 60, left: 80, right: 80 },
          shading: altRowShading,
          children: [new Paragraph({ children: [new TextRun({ text: "进程入口：api / worker / migrate", font: "Microsoft YaHei", size: 20 })] })]
        })
      ]
    }),
    new TableRow({
      children: [
        new TableCell({
          borders, width: { size: 2400, type: WidthType.DXA }, margins: { top: 60, bottom: 60, left: 80, right: 80 },
          children: [new Paragraph({ children: [new TextRun({ text: "internal/", font: "Consolas", size: 20, color: "1F4E79" })] })]
        }),
        new TableCell({
          borders, width: { size: 6600, type: WidthType.DXA }, margins: { top: 60, bottom: 60, left: 80, right: 80 },
          children: [new Paragraph({ children: [new TextRun({ text: "私有代码：application + infrastructure + interfaces + config + rbac", font: "Microsoft YaHei", size: 20 })] })]
        })
      ]
    }),
    new TableRow({
      children: [
        new TableCell({
          borders, width: { size: 2400, type: WidthType.DXA }, margins: { top: 60, bottom: 60, left: 80, right: 80 },
          shading: altRowShading,
          children: [new Paragraph({ children: [new TextRun({ text: "pkg/", font: "Consolas", size: 20, color: "1F4E79" })] })]
        }),
        new TableCell({
          borders, width: { size: 6600, type: WidthType.DXA }, margins: { top: 60, bottom: 60, left: 80, right: 80 },
          shading: altRowShading,
          children: [new Paragraph({ children: [new TextRun({ text: "可复用库：errs / crypto / searchql", font: "Microsoft YaHei", size: 20 })] })]
        })
      ]
    }),
    new TableRow({
      children: [
        new TableCell({
          borders, width: { size: 2400, type: WidthType.DXA }, margins: { top: 60, bottom: 60, left: 80, right: 80 },
          children: [new Paragraph({ children: [new TextRun({ text: "web/", font: "Consolas", size: 20, color: "1F4E79" })] })]
        }),
        new TableCell({
          borders, width: { size: 6600, type: WidthType.DXA }, margins: { top: 60, bottom: 60, left: 80, right: 80 },
          children: [new Paragraph({ children: [new TextRun({ text: "前端 pnpm workspace (Vue 3 SPA + Playwright E2E + Storybook)", font: "Microsoft YaHei", size: 20 })] })]
        })
      ]
    }),
    new TableRow({
      children: [
        new TableCell({
          borders, width: { size: 2400, type: WidthType.DXA }, margins: { top: 60, bottom: 60, left: 80, right: 80 },
          shading: altRowShading,
          children: [new Paragraph({ children: [new TextRun({ text: "sql/", font: "Consolas", size: 20, color: "1F4E79" })] })]
        }),
        new TableCell({
          borders, width: { size: 6600, type: WidthType.DXA }, margins: { top: 60, bottom: 60, left: 80, right: 80 },
          shading: altRowShading,
          children: [new Paragraph({ children: [new TextRun({ text: "递增编号迁移脚本 (0001~0017) + 完整导出 SQL", font: "Microsoft YaHei", size: 20 })] })]
        })
      ]
    }),
    new TableRow({
      children: [
        new TableCell({
          borders, width: { size: 2400, type: WidthType.DXA }, margins: { top: 60, bottom: 60, left: 80, right: 80 },
          children: [new Paragraph({ children: [new TextRun({ text: "docs/", font: "Consolas", size: 20, color: "1F4E79" })] })]
        }),
        new TableCell({
          borders, width: { size: 6600, type: WidthType.DXA }, margins: { top: 60, bottom: 60, left: 80, right: 80 },
          children: [new Paragraph({ children: [new TextRun({ text: "架构文档、部署配置、性能报告", font: "Microsoft YaHei", size: 20 })] })]
        })
      ]
    }),
    new TableRow({
      children: [
        new TableCell({
          borders, width: { size: 2400, type: WidthType.DXA }, margins: { top: 60, bottom: 60, left: 80, right: 80 },
          shading: altRowShading,
          children: [new Paragraph({ children: [new TextRun({ text: "scripts/", font: "Consolas", size: 20, color: "1F4E79" })] })]
        }),
        new TableCell({
          borders, width: { size: 6600, type: WidthType.DXA }, margins: { top: 60, bottom: 60, left: 80, right: 80 },
          shading: altRowShading,
          children: [new Paragraph({ children: [new TextRun({ text: "种子数据 + 大规模造数 + k6 压测 + reindex", font: "Microsoft YaHei", size: 20 })] })]
        })
      ]
    })
  ]
}));

children.push(new Paragraph({ children: [new PageBreak()] }));

children.push(new Paragraph({
  heading: HeadingLevel.HEADING_2,
  children: [new TextRun({ text: "3.2 后端分层详解 (internal/)", font: "Microsoft YaHei", size: 26, bold: true })]
}));

children.push(new Paragraph({
  heading: HeadingLevel.HEADING_3,
  children: [new TextRun({ text: "interfaces/http - 接口层", font: "Microsoft YaHei", size: 24, bold: true })]
}));

children.push(new Paragraph({
  spacing: { after: 200 },
  children: [new TextRun({
    text: "组装 Gin 引擎、中间件链与路由表。中间件顺序：SecurityHeaders → RequestID → Recovery → CORS → AccessLog → Metrics。Handlers 只负责参数绑定/响应编排，业务逻辑全部委托给 Application 服务。",
    font: "Microsoft YaHei", size: 22
  })]
}));

children.push(new Paragraph({
  heading: HeadingLevel.HEADING_3,
  children: [new TextRun({ text: "application/* - 应用服务层", font: "Microsoft YaHei", size: 24, bold: true })]
}));

children.push(new Table({
  width: { size: 9000, type: WidthType.DXA },
  columnWidths: [2000, 7000],
  rows: [
    new TableRow({
      children: [
        headerCell("模块", 2000),
        headerCell("核心职责", 7000)
      ]
    }),
    new TableRow({
      children: [
        cell("auth", 2000, altRowShading),
        cell("JWT 签发/校验、RBAC 主体、密码重置、OIDC 预留", 7000, altRowShading)
      ]
    }),
    new TableRow({
      children: [
        cell("workspace", 2000),
        cell("工作空间 CRUD、成员邀请、审计日志、项目模板", 7000)
      ]
    }),
    new TableRow({
      children: [
        cell("issue", 2000, altRowShading),
        cell("工作项 CRUD、状态流转、WBS、关联、依赖、缺陷、评论、工时", 7000, altRowShading)
      ]
    }),
    new TableRow({
      children: [
        cell("sprint", 2000),
        cell("迭代生命周期、燃尽快照、容量规划、速率建议、复盘", 7000)
      ]
    }),
    new TableRow({
      children: [
        cell("version", 2000, altRowShading),
        cell("版本日 CRUD、状态机、Release Notes、交付报告", 7000, altRowShading)
      ]
    }),
    new TableRow({
      children: [
        cell("notification", 2000),
        cell("通知管道、订阅、摘要、多渠道分发", 7000)
      ]
    }),
    new TableRow({
      children: [
        cell("automation", 2000, altRowShading),
        cell("规则引擎 (TCA DSL)、熔断、锁、Cron、执行审计", 7000, altRowShading)
      ]
    }),
    new TableRow({
      children: [
        cell("search", 2000),
        cell("全文检索、搜索历史、书签、索引同步", 7000)
      ]
    }),
    new TableRow({
      children: [
        cell("dashboard", 2000, altRowShading),
        cell("Widget 框架、数据模板、风险预警、解决方案", 7000, altRowShading)
      ]
    }),
    new TableRow({
      children: [
        cell("webhook", 2000),
        cell("订阅管理、HMAC 签名、投递日志、手动重试", 7000)
      ]
    }),
    new TableRow({
      children: [
        cell("intake", 2000, altRowShading),
        cell("收件箱通道、公开提交、工单审核转正", 7000, altRowShading)
      ]
    }),
    new TableRow({
      children: [
        cell("metrics", 2000),
        cell("DORA 四指标、迭代速率、前置时间、质量指标、资源负载", 7000)
      ]
    }),
    new TableRow({
      children: [
        cell("ai", 2000, altRowShading),
        cell("智能指派、重复检测、自动分类、摘要生成 (LLM / Fallback)", 7000, altRowShading)
      ]
    })
  ]
}));

children.push(new Paragraph({ children: [new PageBreak()] }));

children.push(new Paragraph({
  heading: HeadingLevel.HEADING_3,
  children: [new TextRun({ text: "infrastructure/* - 基础设施层", font: "Microsoft YaHei", size: 24, bold: true })]
}));

children.push(new Table({
  width: { size: 9000, type: WidthType.DXA },
  columnWidths: [1800, 7200],
  rows: [
    new TableRow({
      children: [
        headerCell("模块", 1800),
        headerCell("职责", 7200)
      ]
    }),
    new TableRow({
      children: [
        cell("persistence", 1800, altRowShading),
        cell("pgx 连接池 + 租户上下文 + 事务管理", 7200, altRowShading)
      ]
    }),
    new TableRow({
      children: [
        cell("cache", 1800),
        cell("Redis 客户端封装", 7200)
      ]
    }),
    new TableRow({
      children: [
        cell("mq", 1800, altRowShading),
        cell("RabbitMQ 连接 + Exchange/Queue 声明 + 任务客户端", 7200, altRowShading)
      ]
    }),
    new TableRow({
      children: [
        cell("events", 1800),
        cell("Outbox Relay (DB → RabbitMQ 事件总线)", 7200)
      ]
    }),
    new TableRow({
      children: [
        cell("mail", 1800, altRowShading),
        cell("SMTP 抽象 + Noop/SMTP 自动切换 + 双 MIME 模板", 7200, altRowShading)
      ]
    }),
    new TableRow({
      children: [
        cell("storage", 1800),
        cell("MinIO/S3 客户端 + 预签名 URL", 7200)
      ]
    }),
    new TableRow({
      children: [
        cell("telemetry", 1800, altRowShading),
        cell("zap 结构化日志 + Prometheus 指标 + OpenTelemetry 埋点", 7200, altRowShading)
      ]
    }),
    new TableRow({
      children: [
        cell("ws", 1800),
        cell("WebSocket Hub (多节点扇出 + 断线补偿)", 7200)
      ]
    }),
    new TableRow({
      children: [
        cell("worker", 1800, altRowShading),
        cell("Jitter 工具 (幂等定时偏移)", 7200, altRowShading)
      ]
    })
  ]
}));

children.push(new Paragraph({ children: [new PageBreak()] }));

// ========== 第四章：领域模型与限界上下文 ==========
children.push(new Paragraph({
  heading: HeadingLevel.HEADING_1,
  pageBreakBefore: true,
  children: [new TextRun({ text: "4. 领域模型与限界上下文", font: "Microsoft YaHei", size: 32, bold: true })]
}));

children.push(new Paragraph({
  heading: HeadingLevel.HEADING_2,
  children: [new TextRun({ text: "4.1 限界上下文映射", font: "Microsoft YaHei", size: 26, bold: true })]
}));

children.push(new Table({
  width: { size: 9000, type: WidthType.DXA },
  columnWidths: [1800, 3000, 4200],
  rows: [
    new TableRow({
      children: [
        headerCell("上下文", 1800),
        headerCell("核心聚合", 3000),
        headerCell("职责", 4200)
      ]
    }),
    new TableRow({
      children: [
        cell("iam", 1800, altRowShading),
        cell("User, ApiToken, Session", 3000, altRowShading),
        cell("认证、用户、令牌签发", 4200, altRowShading)
      ]
    }),
    new TableRow({
      children: [
        cell("workspace", 1800),
        cell("Workspace, Member, Invitation", 3000),
        cell("租户、成员、角色、邀请", 4200)
      ]
    }),
    new TableRow({
      children: [
        cell("project", 1800, altRowShading),
        cell("Project, Module, Label, State", 3000, altRowShading),
        cell("项目配置、模块、标签、状态机", 4200, altRowShading)
      ]
    }),
    new TableRow({
      children: [
        cell("issue", 1800),
        cell("Issue, Comment, Activity", 3000),
        cell("需求/任务/缺陷统一聚合", 4200)
      ]
    }),
    new TableRow({
      children: [
        cell("sprint", 1800, altRowShading),
        cell("Sprint, SprintIssue, Snapshot", 3000, altRowShading),
        cell("迭代生命周期、燃尽快照", 4200, altRowShading)
      ]
    }),
    new TableRow({
      children: [
        cell("version", 1800),
        cell("Version, Checklist", 3000),
        cell("版本日聚合、发布", 4200)
      ]
    }),
    new TableRow({
      children: [
        cell("notification", 1800, altRowShading),
        cell("Notification, Preference", 3000, altRowShading),
        cell("通知编排、订阅、投递", 4200, altRowShading)
      ]
    }),
    new TableRow({
      children: [
        cell("automation", 1800),
        cell("Rule, Execution", 3000),
        cell("TCA 规则引擎", 4200)
      ]
    }),
    new TableRow({
      children: [
        cell("search", 1800, altRowShading),
        cell("Query, History, Bookmark", 3000, altRowShading),
        cell("查询语法解析、检索", 4200, altRowShading)
      ]
    })
  ]
}));

children.push(new Paragraph({ children: [new PageBreak()] }));

children.push(new Paragraph({
  heading: HeadingLevel.HEADING_2,
  children: [new TextRun({ text: "4.2 Issue 聚合核心模型", font: "Microsoft YaHei", size: 26, bold: true })]
}));

children.push(new Paragraph({
  spacing: { after: 200 },
  children: [new TextRun({
    text: "Issue 是系统的核心聚合根，需求/任务/缺陷统一使用 issues 单表，通过 type 字段区分。类型差异字段（如 severity、found_phase 等缺陷专有字段）独立列存以支持高频过滤，其余扩展属性使用 type_attrs JSONB。",
    font: "Microsoft YaHei", size: 22
  })]
}));

children.push(new Table({
  width: { size: 9000, type: WidthType.DXA },
  columnWidths: [2400, 1800, 4800],
  rows: [
    new TableRow({
      children: [
        headerCell("字段", 2400),
        headerCell("类型", 1800),
        headerCell("说明", 4800)
      ]
    }),
    new TableRow({
      children: [
        cell("id", 2400, altRowShading),
        cell("bigint", 1800, altRowShading),
        cell("全局唯一 ID", 4800, altRowShading)
      ]
    }),
    new TableRow({
      children: [
        cell("project_id", 2400),
        cell("bigint", 1800),
        cell("所属项目 (复合唯一索引前缀)", 4800)
      ]
    }),
    new TableRow({
      children: [
        cell("sequence_id", 2400, altRowShading),
        cell("bigint", 1800, altRowShading),
        cell("项目内自增 → IDENTIFIER-123 展示", 4800, altRowShading)
      ]
    }),
    new TableRow({
      children: [
        cell("type", 2400),
        cell("enum", 1800),
        cell("requirement | task | defect", 4800)
      ]
    }),
    new TableRow({
      children: [
        cell("parent_id", 2400, altRowShading),
        cell("bigint", 1800, altRowShading),
        cell("WBS 父级引用 (限 3 层)", 4800, altRowShading)
      ]
    }),
    new TableRow({
      children: [
        cell("state_id", 2400),
        cell("bigint", 1800),
        cell("状态外键 → states 表 (project 维度配置)", 4800)
      ]
    }),
    new TableRow({
      children: [
        cell("severity", 2400, altRowShading),
        cell("int", 1800, altRowShading),
        cell("缺陷专有 (P0=1...P4=5)", 4800, altRowShading)
      ]
    }),
    new TableRow({
      children: [
        cell("assignee_ids", 2400),
        cell("bigint[]", 1800),
        cell("指派用户集合", 4800)
      ]
    }),
    new TableRow({
      children: [
        cell("sprint_id", 2400, altRowShading),
        cell("bigint", 1800, altRowShading),
        cell("所属迭代 (单 active sprint 约束)", 4800, altRowShading)
      ]
    }),
    new TableRow({
      children: [
        cell("version_id", 2400),
        cell("bigint", 1800),
        cell("首次发布版本", 4800)
      ]
    }),
    new TableRow({
      children: [
        cell("description_json", 2400, altRowShading),
        cell("jsonb", 1800, altRowShading),
        cell("富文本结构化内容 (TipTap 协议)", 4800, altRowShading)
      ]
    })
  ]
}));

children.push(new Paragraph({ children: [new PageBreak()] }));

children.push(new Paragraph({
  heading: HeadingLevel.HEADING_2,
  children: [new TextRun({ text: "4.3 状态机设计", font: "Microsoft YaHei", size: 26, bold: true })]
}));

children.push(new Paragraph({
  spacing: { after: 200 },
  children: [new TextRun({
    text: "State 按 project 维度配置，包含 name/color/group/sequence 四要素。group 决定统计口径（backlog/unstarted/started/completed/cancelled/triage），sequence 决定看板列序。流转规则由 state_transition 表控制：",
    font: "Microsoft YaHei", size: 22
  })]
}));

children.push(new Paragraph({
  spacing: { after: 100 },
  children: [new TextRun({ text: "• from_state：源状态", font: "Microsoft YaHei", size: 22 })]
}));
children.push(new Paragraph({
  spacing: { after: 100 },
  children: [new TextRun({ text: "• to_state：目标状态", font: "Microsoft YaHei", size: 22 })]
}));
children.push(new Paragraph({
  spacing: { after: 100 },
  children: [new TextRun({ text: "required_fields[]：流转时必须填写的字段", font: "Microsoft YaHei", size: 22 })]
}));
children.push(new Paragraph({
  spacing: { after: 200 },
  children: [new TextRun({ text: "allowed_roles[]：允许触发流转的角色集合", font: "Microsoft YaHei", size: 22 })]
}));

children.push(new Paragraph({
  spacing: { after: 200 },
  children: [new TextRun({
    text: "内置三组状态机模板：开发流、缺陷流（新建→确认→处理中→修复→待验→关闭/拒绝/重开）、需求评审流。自定义流转规则放 Phase 3 与自动化引擎打通。",
    font: "Microsoft YaHei", size: 22
  })]
}));

children.push(new Paragraph({ children: [new PageBreak()] }));

children.push(new Paragraph({
  heading: HeadingLevel.HEADING_2,
  children: [new TextRun({ text: "4.4 迭代与版本模型", font: "Microsoft YaHei", size: 26, bold: true })]
}));

children.push(new Paragraph({
  heading: HeadingLevel.HEADING_3,
  children: [new TextRun({ text: "Sprint 生命周期", font: "Microsoft YaHei", size: 24, bold: true })]
}));

children.push(new Paragraph({
  spacing: { after: 200 },
  children: [new TextRun({
    text: "planned → active → completed。项目维度唯一 active 约束 (unique partial index)。每个 sprint 包含 goal/start_date/end_date/capacity 字段。通过 /start 与 /complete 端点触发状态转换。",
    font: "Microsoft YaHei", size: 22
  })]
}));

children.push(new Paragraph({
  heading: HeadingLevel.HEADING_3,
  children: [new TextRun({ text: "Version 状态机", font: "Microsoft YaHei", size: 24, bold: true })]
}));

children.push(new Paragraph({
  spacing: { after: 200 },
  children: [new TextRun({
    text: "planning → active → released → archived。发布时执行准出校验（checklist），生成 Release Notes（模板渲染）+ 交付报告。版本日通过 version_sprints 关联表聚合多个迭代，进度跨迭代汇总。",
    font: "Microsoft YaHei", size: 22
  })]
}));

children.push(new Paragraph({ children: [new PageBreak()] }));

// ========== 第五章：API 设计规范 ==========
children.push(new Paragraph({
  heading: HeadingLevel.HEADING_1,
  pageBreakBefore: true,
  children: [new TextRun({ text: "5. API 设计规范", font: "Microsoft YaHei", size: 32, bold: true })]
}));

children.push(new Paragraph({
  heading: HeadingLevel.HEADING_2,
  children: [new TextRun({ text: "5.1 RESTful 约定", font: "Microsoft YaHei", size: 26, bold: true })]
}));

children.push(new Table({
  width: { size: 9000, type: WidthType.DXA },
  columnWidths: [1800, 7200],
  rows: [
    new TableRow({
      children: [
        headerCell("规则", 1800),
        headerCell("说明", 7200)
      ]
    }),
    new TableRow({
      children: [
        cell("基础路径", 1800, altRowShading),
        cell("/api/v1/", 7200, altRowShading)
      ]
    }),
    new TableRow({
      children: [
        cell("认证", 1800),
        cell("Cookie (Web) / Bearer Token / API Token (第三方集成)", 7200)
      ]
    }),
    new TableRow({
      children: [
        cell("分页", 1800, altRowShading),
        cell("无限滚动：?cursor=  |  报表类：?page=&per_page=", 7200, altRowShading)
      ]
    }),
    new TableRow({
      children: [
        cell("过滤", 1800),
        cell("?state__group=started&priority__in=urgent,high&assignee=me", 7200)
      ]
    }),
    new TableRow({
      children: [
        cell("字段裁剪", 1800, altRowShading),
        cell("?fields=id,name,state", 7200, altRowShading)
      ]
    }),
    new TableRow({
      children: [
        cell("版本隔离", 1800),
        cell("URL 前缀 /api/v1 稳定，非兼容变更升级 /api/v2", 7200)
      ]
    })
  ]
}));

children.push(new Paragraph({ spacing: { before: 200 } }));

children.push(new Paragraph({
  heading: HeadingLevel.HEADING_2,
  children: [new TextRun({ text: "5.2 统一 Envelope 与错误格式", font: "Microsoft YaHei", size: 26, bold: true })]
}));

children.push(new Paragraph({
  spacing: { after: 100 },
  children: [new TextRun({ text: "成功响应:", font: "Microsoft YaHei", size: 22, bold: true })]
}));
children.push(new Paragraph({
  spacing: { after: 200 },
  children: [new TextRun({ text: '{"id":1,"name":"示例","workspace":{"slug":"demo"}}', font: "Consolas", size: 20, color: "404040" })]
}));

children.push(new Paragraph({
  spacing: { after: 100 },
  children: [new TextRun({ text: "错误响应:", font: "Microsoft YaHei", size: 22, bold: true })]
}));
children.push(new Paragraph({
  spacing: { after: 200 },
  children: [new TextRun({ text: '{"error":{"code":"ISSUE.INVALID_STATE_TRANSITION","message":"状态流转失败","details":{"from":"started","backlog":"backlog"}}}', font: "Consolas", size: 20, color: "404040" })]
}));

children.push(new Paragraph({
  spacing: { after: 200 },
  children: [new TextRun({
    text: "错误码采用 {DOMAIN}.{ACTION} 命名，HTTP 状态码语义化（400/401/403/404/422/429/500）。请求级限流 100 req/min/用户，登录接口 10 req/min/IP。",
    font: "Microsoft YaHei", size: 22
  })]
}));

children.push(new Paragraph({ children: [new PageBreak()] }));

children.push(new Paragraph({
  heading: HeadingLevel.HEADING_2,
  children: [new TextRun({ text: "5.3 已交付 API 清单（按限界上下文）", font: "Microsoft YaHei", size: 26, bold: true })]
}));

children.push(new Paragraph({
  heading: HeadingLevel.HEADING_3,
  children: [new TextRun({ text: "鉴权 (iam)", font: "Microsoft YaHei", size: 24, bold: true })]
}));

children.push(new Paragraph({
  spacing: { after: 100 },
  children: [new TextRun({
    text: "POST /auth/login | /refresh | /register | /forgot-password | /reset-password",
    font: "Consolas", size: 20
  })]
}));
children.push(new Paragraph({
  spacing: { after: 200 },
  children: [new TextRun({
    text: "GET  /me  ·  GET/POST/DELETE /me/api-tokens",
    font: "Consolas", size: 20
  })]
}));

children.push(new Paragraph({
  heading: HeadingLevel.HEADING_3,
  children: [new TextRun({ text: "工作空间 (workspace)", font: "Microsoft YaHei", size: 24, bold: true })]
}));

children.push(new Paragraph({
  spacing: { after: 100 },
  children: [new TextRun({
    text: "POST/GET/PATCH/DELETE /workspaces/:id",
    font: "Consolas", size: 20
  })]
}));
children.push(new Paragraph({
  spacing: { after: 100 },
  children: [new TextRun({
    text: "GET/PATCH/DELETE /workspaces/:id/members",
    font: "Consolas", size: 20
  })]
}));
children.push(new Paragraph({
  spacing: { after: 100 },
  children: [new TextRun({
    text: "POST/GET/DELETE  /workspaces/:id/invitations",
    font: "Consolas", size: 20
  })]
}));
children.push(new Paragraph({
  spacing: { after: 200 },
  children: [new TextRun({
    text: "GET /workspaces/:id/audit-logs  ·  GET /workspaces/:id/projects",
    font: "Consolas", size: 20
  })]
}));

children.push(new Paragraph({
  heading: HeadingLevel.HEADING_3,
  children: [new TextRun({ text: "工作项 (issue)", font: "Microsoft YaHei", size: 24, bold: true })]
}));

children.push(new Paragraph({
  spacing: { after: 100 },
  children: [new TextRun({
    text: "GET/POST /issues  ·  POST /issues/batch  ·  GET /issues/export",
    font: "Consolas", size: 20
  })]
}));
children.push(new Paragraph({
  spacing: { after: 100 },
  children: [new TextRun({
    text: "GET/PATCH/DELETE /issues/:iid  ·  POST /issues/:iid/transition",
    font: "Consolas", size: 20
  })]
}));
children.push(new Paragraph({
  spacing: { after: 100 },
  children: [new TextRun({
    text: "GET/POST/PATCH/DELETE /issues/:iid/comments",
    font: "Consolas", size: 20
  })]
}));
children.push(new Paragraph({
  spacing: { after: 100 },
  children: [new TextRun({
    text: "GET/POST/DELETE /issues/:iid/relations  ·  /dependencies",
    font: "Consolas", size: 20
  })]
}));
children.push(new Paragraph({
  spacing: { after: 200 },
  children: [new TextRun({
    text: "GET /issues/:iid/activities  ·  GET/POST/PATCH/DELETE /issues/:iid/time-logs",
    font: "Consolas", size: 20
  })]
}));

children.push(new Paragraph({
  heading: HeadingLevel.HEADING_3,
  children: [new TextRun({ text: "迭代 (sprint)", font: "Microsoft YaHei", size: 24, bold: true })]
}));

children.push(new Paragraph({
  spacing: { after: 200 },
  children: [new TextRun({
    text: "CRUD /sprints  ·  POST /sprints/:sid:start|complete  ·  GET /sprints/:sid/burndown|progress|review",
    font: "Consolas", size: 20
  })]
}));

children.push(new Paragraph({
  heading: HeadingLevel.HEADING_3,
  children: [new TextRun({ text: "版本 (version) + 其他域", font: "Microsoft YaHei", size: 24, bold: true })]
}));

children.push(new Paragraph({
  spacing: { after: 100 },
  children: [new TextRun({
    text: "CRUD /versions  ·  POST /versions/:vid/activate|release|archive",
    font: "Consolas", size: 20
  })]
}));
children.push(new Paragraph({
  spacing: { after: 100 },
  children: [new TextRun({
    text: "GET/POST /workspaces/:id/search[/history|bookmarks]",
    font: "Consolas", size: 20
  })]
}));
children.push(new Paragraph({
  spacing: { after: 100 },
  children: [new TextRun({
    text: "GET/POST /workspaces/:id/notifications[/preferences]",
    font: "Consolas", size: 20
  })]
}));
children.push(new Paragraph({
  spacing: { after: 100 },
  children: [new TextRun({
    text: "CRUD /automation[/templates|/executions]  ·  POST /automation/dry-run",
    font: "Consolas", size: 20
  })]
}));
children.push(new Paragraph({
  spacing: { after: 100 },
  children: [new TextRun({
    text: "GET /metrics/{velocity|lead-time|quality|dora|resource-load}",
    font: "Consolas", size: 20
  })]
}));
children.push(new Paragraph({
  spacing: { after: 100 },
  children: [new TextRun({
    text: "WS /ws/:workspace_id  ·  Webhook/Intake 公开 API /api/v1/public/intake",
    font: "Consolas", size: 20
  })]
}));

children.push(new Paragraph({ children: [new PageBreak()] }));

// ========== 第六章：数据模型与持久化 ==========
children.push(new Paragraph({
  heading: HeadingLevel.HEADING_1,
  pageBreakBefore: true,
  children: [new TextRun({ text: "6. 数据模型与持久化", font: "Microsoft YaHei", size: 32, bold: true })]
}));

children.push(new Paragraph({
  heading: HeadingLevel.HEADING_2,
  children: [new TextRun({ text: "6.1 租户隔离 (RLS)", font: "Microsoft YaHei", size: 26, bold: true })]
}));

children.push(new Paragraph({
  spacing: { after: 200 },
  children: [new TextRun({
    text: "所有业务表带 workspace_id 冗余列，启用 PostgreSQL Row Level Security (RLS)。API 中间件解析 workspace slug 后在连接上执行 SET LOCAL app.workspace_id = ?，事务内租户过滤自动生效。",
    font: "Microsoft YaHei", size: 22
  })]
}));

children.push(new Paragraph({
  heading: HeadingLevel.HEADING_2,
  children: [new TextRun({ text: "6.2 Outbox 表 (事务性事件)", font: "Microsoft YaHei", size: 26, bold: true })]
}));

children.push(new Paragraph({
  spacing: { after: 200 },
  children: [new TextRun({
    text: "domain_events 表记录领域事件，API 进程写入后立即返回 (at-least-once)。Worker 进程 (Relay) 轮询 unpublished 事件并发布到 RabbitMQ EventExchange。幂等键 = event_id，确保消费端去重。事件类型覆盖 30+ 场景。",
    font: "Microsoft YaHei", size: 22
  })]
}));

children.push(new Paragraph({
  heading: HeadingLevel.HEADING_2,
  children: [new TextRun({ text: "6.3 迁移脚本清单", font: "Microsoft YaHei", size: 26, bold: true })]
}));

children.push(new Table({
  width: { size: 9000, type: WidthType.DXA },
  columnWidths: [1600, 7400],
  rows: [
    new TableRow({
      children: [
        headerCell("版本", 1600),
        headerCell("内容", 7400)
      ]
    }),
    new TableRow({
      children: [
        cell("0001", 1600, altRowShading),
        cell("基础表：users/workspace/workspace_members/projects/issues/etc.", 7400, altRowShading)
      ]
    }),
    new TableRow({
      children: [
        cell("0002", 1600),
        cell("密码重置 token + 邮件事件表", 7400)
      ]
    }),
    new TableRow({
      children: [
        cell("0003", 1600, altRowShading),
        cell("领域事件 (domain_events/outbox)", 7400, altRowShading)
      ]
    }),
    new TableRow({
      children: [
        cell("0004", 1600),
        cell("RLS 租户策略 + 员工日志", 7400)
      ]
    }),
    new TableRow({
      children: [
        cell("0005-0017", 1600, altRowShading),
        cell("迭代/版本/评论/工时/默认流转/搜索/通知/附件/Webhook/自动化/效能度量等", 7400, altRowShading)
      ]
    })
  ]
}));

children.push(new Paragraph({ children: [new PageBreak()] }));

children.push(new Paragraph({
  heading: HeadingLevel.HEADING_2,
  children: [new TextRun({ text: "6.4 缓存策略 (Redis)", font: "Microsoft YaHei", size: 26, bold: true })]
}));

children.push(new Table({
  width: { size: 9000, type: WidthType.DXA },
  columnWidths: [2800, 3000, 3200],
  rows: [
    new TableRow({
      children: [
        headerCell("Key 模式", 2800),
        headerCell("内容", 3000),
        headerCell("TTL", 3200)
      ]
    }),
    new TableRow({
      children: [
        cell("ws:{slug}:meta", 2800, altRowShading),
        cell("工作空间元信息", 3000, altRowShading),
        cell("1h，写时失效", 3200, altRowShading)
      ]
    }),
    new TableRow({
      children: [
        cell("proj:{id}:config", 2800),
        cell("项目配置", 3000),
        cell("30m，写时失效", 3200)
      ]
    }),
    new TableRow({
      children: [
        cell("issue:{id}:detail", 2800, altRowShading),
        cell("工作项详情", 3000, altRowShading),
        cell("10m，事件失效", 3200, altRowShading)
      ]
    }),
    new TableRow({
      children: [
        cell("dash:{id}:widget:{wid}", 2800),
        cell("仪表盘 Widget", 3000),
        cell("按刷新频率", 3200)
      ]
    }),
    new TableRow({
      children: [
        cell("ratelimit:{uid}", 2800, altRowShading),
        cell("限流计数", 3000, altRowShading),
        cell("—", 3200, altRowShading)
      ]
    }),
    new TableRow({
      children: [
        cell("lock:{resource}", 2800),
        cell("分布式锁", 3000),
        cell("—", 3200)
      ]
    })
  ]
}));

children.push(new Paragraph({ children: [new PageBreak()] }));

// ========== 第七章：前端架构 ==========
children.push(new Paragraph({
  heading: HeadingLevel.HEADING_1,
  pageBreakBefore: true,
  children: [new TextRun({ text: "7. 前端架构", font: "Microsoft YaHei", size: 32, bold: true })]
}));

children.push(new Paragraph({
  heading: HeadingLevel.HEADING_2,
  children: [new TextRun({ text: "7.1 工程结构", font: "Microsoft YaHei", size: 26, bold: true })]
}));

children.push(new Table({
  width: { size: 9000, type: WidthType.DXA },
  columnWidths: [2200, 6800],
  rows: [
    new TableRow({
      children: [
        headerCell("路径", 2200),
        headerCell("说明", 6800)
      ]
    }),
    new TableRow({
      children: [
        cell("api/", 2200, altRowShading),
        cell("Axios 客户端 + 401 单飞刷新 + 限流回调 + Service 层", 6800, altRowShading)
      ]
    }),
    new TableRow({
      children: [
        cell("views/", 2200),
        cell("45+ 页面组件 (auth/project/workspace/settings)", 6800)
      ]
    }),
    new TableRow({
      children: [
        cell("stores/", 2200, altRowShading),
        cell("Pinia 状态管理 (auth/workspace/sprint/...)", 6800, altRowShading)
      ]
    }),
    new TableRow({
      children: [
        cell("router/", 2200),
        cell("路由表 + 权限守卫 + 视图偏好持久化", 6800)
      ]
    }),
    new TableRow({
      children: [
        cell("components/", 2200, altRowShading),
        cell("通用组件 (Command Palette/Toast/沉团引导等)", 6800, altRowShading)
      ]
    }),
    new TableRow({
      children: [
        cell("composables/", 2200),
        cell("组合式 API 封装", 6800)
      ]
    }),
    new TableRow({
      children: [
        cell("design/", 2200, altRowShading),
        cell("设计令牌 (CSS 变量 + light/dark 主题)", 6800, altRowShading)
      ]
    }),
    new TableRow({
      children: [
        cell("lib/", 2200),
        cell("工具库 (ws-client/shortcut/...)", 6800)
      ]
    }),
    new TableRow({
      children: [
        cell("locales/", 2200, altRowShading),
        cell("国际化资源", 6800, altRowShading)
      ]
    })
  ]
}));

children.push(new Paragraph({ children: [new PageBreak()] }));

children.push(new Paragraph({
  heading: HeadingLevel.HEADING_2,
  children: [new TextRun({ text: "7.2 路由与视图", font: "Microsoft YaHei", size: 26, bold: true })]
}));

children.push(new Table({
  width: { size: 9000, type: WidthType.DXA },
  columnWidths: [3000, 6000],
  rows: [
    new TableRow({
      children: [
        headerCell("路由", 3000),
        headerCell("视图", 6000)
      ]
    }),
    new TableRow({
      children: [
        cell("/", 3000, altRowShading),
        cell("WorkspaceListView (工作空间列表)", 6000, altRowShading)
      ]
    }),
    new TableRow({
      children: [
        cell(":wsId/workbench", 3000),
        cell("WorkbenchView (个人工作台)", 6000)
      ]
    }),
    new TableRow({
      children: [
        cell(":wsId/search", 3000, altRowShading),
        cell("SearchView (全局搜索)", 6000, altRowShading)
      ]
    }),
    new TableRow({
      children: [
        cell(":wsId/projects", 3000),
        cell("ProjectListView", 6000)
      ]
    }),
    new TableRow({
      children: [
        cell(":wsId/projects/:pid/board", 3000, altRowShading),
        cell("KanbanBoardView", 6000, altRowShading)
      ]
    }),
    new TableRow({
      children: [
        cell(":wsId/projects/:pid/list", 3000),
        cell("IssueListView", 6000)
      ]
    }),
    new TableRow({
      children: [
        cell(":wsId/projects/:pid/issues/:iid", 3000, altRowShading),
        cell("IssueDetailView (含 FocusMode)", 6000, altRowShading)
      ]
    }),
    new TableRow({
      children: [
        cell(":wsId/projects/:pid/sprints/...", 3000),
        cell("SprintList/Detail/Planning/Standup/Burndown", 6000)
      ]
    }),
    new TableRow({
      children: [
        cell(":wsId/projects/:pid/versions/...", 3000, altRowShading),
        cell("VersionList/Detail/Release/DeliveryReport", 6000, altRowShading)
      ]
    }),
    new TableRow({
      children: [
        cell(":wsId/projects/:pid/dashboard", 3000),
        cell("DashboardView (可配置 Widget)", 6000)
      ]
    }),
    new TableRow({
      children: [
        cell(":wsId/projects/:pid/automation", 3000, altRowShading),
        cell("AutomationView", 6000, altRowShading)
      ]
    }),
    new TableRow({
      children: [
        cell(":wsId/projects/:pid/metrics", 3000),
        cell("MetricsView (效能度量)", 6000)
      ]
    }),
    new TableRow({
      children: [
        cell(":wsId/projects/:pid/{gantt,calendar,spreadsheet}", 3000, altRowShading),
        cell("三视图已交付", 6000, altRowShading)
      ]
    }),
    new TableRow({
      children: [
        cell(":wsId/settings/...", 3000),
        cell("Workspace/Intake/Webhook/Notification/RBAC/DLQ", 6000)
      ]
    }),
    new TableRow({
      children: [
        cell("/settings/api-tokens", 3000, altRowShading),
        cell("ApiTokensView", 6000, altRowShading)
      ]
    }),
    new TableRow({
      children: [
        cell("/intake/:wsId/:slug", 3000),
        cell("IntakePublicView (公开表单)", 6000)
      ]
    })
  ]
}));

children.push(new Paragraph({ children: [new PageBreak()] }));

children.push(new Paragraph({
  heading: HeadingLevel.HEADING_2,
  children: [new TextRun({ text: "7.3 状态与数据流", font: "Microsoft YaHei", size: 26, bold: true })]
}));

children.push(new Paragraph({
  spacing: { after: 200 },
  children: [new TextRun({
    text: "Pinia 按域分 store (auth / workspace / sprint / issue / notification 等)，服务端数据使用 TanStack Query (Vue Query) 管理缓存/失效，避免手写 loading 态。实时层由单例 ws-client 订阅 workspace:{id} 频道，收到事件后精确失效对应 Query Key。用户路由级鉴权守卫集成 RBAC 校验：workspace:update、webhook:manage、automation:manage、audit:read 等 10 项权限。",
    font: "Microsoft YaHei", size: 22
  })]
}));

children.push(new Paragraph({ children: [new PageBreak()] }));

// ========== 第八章：异步与 Worker ==========
children.push(new Paragraph({
  heading: HeadingLevel.HEADING_1,
  pageBreakBefore: true,
  children: [new TextRun({ text: "8. 异步任务与 Worker 机制", font: "Microsoft YaHei", size: 32, bold: true })]
}));

children.push(new Paragraph({
  heading: HeadingLevel.HEADING_2,
  children: [new TextRun({ text: "8.1 双管线架构", font: "Microsoft YaHei", size: 26, bold: true })]
}));

children.push(new Paragraph({
  spacing: { after: 200 },
  children: [new TextRun({
    text: "Worker 进程承载两条独立管线：(1) Outbox Relay —— 轮询 PostgreSQL domain_events 并发布到 RabbitMQ（at-least-once，idempotent consumer）；(2) Task Worker —— 消费任务队列（通知、索引、Webhook、自动化）+ 定时任务。Redis 不参与 Worker (专注缓存、限流、WS)。",
    font: "Microsoft YaHei", size: 22
  })]
}));

children.push(new Paragraph({
  heading: HeadingLevel.HEADING_2,
  children: [new TextRun({ text: "8.2 RabbitMQ 双 Exchange 分工", font: "Microsoft YaHei", size: 26, bold: true })]
}));

children.push(new Table({
  width: { size: 9000, type: WidthType.DXA },
  columnWidths: [2000, 3500, 3500],
  rows: [
    new TableRow({
      children: [
        headerCell("Exchange", 2000),
        headerCell("类型", 3500),
        headerCell("用途", 3500)
      ]
    }),
    new TableRow({
      children: [
        cell("EventExchange", 2000, altRowShading),
        cell("topic", 3500, altRowShading),
        cell("领域事件广播 (30+ 类型)", 3500, altRowShading)
      ]
    }),
    new TableRow({
      children: [
        cell("TaskExchange", 2000),
        cell("direct", 3500),
        cell("异步任务分发", 3500)
      ]
    }),
    new TableRow({
      children: [
        cell("DLX/DLQ", 2000, altRowShading),
        cell("—", 3500, altRowShading),
        cell("死信/重放", 3500, altRowShading)
      ]
    })
  ]
}));

children.push(new Paragraph({ children: [new PageBreak()] }));

children.push(new Paragraph({
  heading: HeadingLevel.HEADING_2,
  children: [new TextRun({ text: "8.3 领域事件与扇出消费者", font: "Microsoft YaHei", size: 26, bold: true })]
}));

children.push(new Table({
  width: { size: 9000, type: WidthType.DXA },
  columnWidths: [2400, 6600],
  rows: [
    new TableRow({
      children: [
        headerCell("消费者", 2400),
        headerCell("动作", 6600)
      ]
    }),
    new TableRow({
      children: [
        cell("notif-consumer", 2400, altRowShading),
        cell("创建站内通知 (issue.* / comment.*)", 6600, altRowShading)
      ]
    }),
    new TableRow({
      children: [
        cell("webhook-consumer", 2400),
        cell("匹配订阅 + 签名投递 + 失败入队", 6600)
      ]
    }),
    new TableRow({
      children: [
        cell("automation-consumer", 2400, altRowShading),
        cell("执行匹配规则 (Trigger-Condition-Action)", 6600, altRowShading)
      ]
    }),
    new TableRow({
      children: [
        cell("search-indexer", 2400),
        cell("upsert / delete / sync{issue,sprint,version}", 6600)
      ]
    }),
    new TableRow({
      children: [
        cell("ws-fanout", 2400, altRowShading),
        cell("广播到 WebSocket Hub (Redis Pub/Sub)", 6600, altRowShading)
      ]
    })
  ]
}));

children.push(new Paragraph({ children: [new PageBreak()] }));

children.push(new Paragraph({
  heading: HeadingLevel.HEADING_2,
  children: [new TextRun({ text: "8.4 Worker 定时任务", font: "Microsoft YaHei", size: 26, bold: true })]
}));

children.push(new Table({
  width: { size: 9000, type: WidthType.DXA },
  columnWidths: [2600, 2000, 4400],
  rows: [
    new TableRow({
      children: [
        headerCell("任务", 2600),
        headerCell("频率", 2000),
        headerCell("说明", 4400)
      ]
    }),
    new TableRow({
      children: [
        cell("sprint-snapshot", 2600, altRowShading),
        cell("00:05", 2000, altRowShading),
        cell("燃尽快照 (幂等, ON CONFLICT)", 4400, altRowShading)
      ]
    }),
    new TableRow({
      children: [
        cell("metrics-snapshot", 2600),
        cell("01:30", 2000),
        cell("DORA/velocity/lead-time 日聚合", 4400)
      ]
    }),
    new TableRow({
      children: [
        cell("webhook-cleanup", 2600, altRowShading),
        cell("01:00", 2000, altRowShading),
        cell("清理 30 天前投递日志", 4400, altRowShading)
      ]
    }),
    new TableRow({
      children: [
        cell("digest-runner", 2600),
        cell("60s", 2000),
        cell("Daily/Weekly 邮件摘要聚合", 4400)
      ]
    }),
    new TableRow({
      children: [
        cell("dispatch-worker", 2600, altRowShading),
        cell("30s", 2000, altRowShading),
        cell("通知多渠道投递 (邮件/IM)", 4400, altRowShading)
      ]
    }),
    new TableRow({
      children: [
        cell("scheduled-cron", 2600),
        cell("60s", 2000),
        cell("自动化定时触发器", 4400)
      ]
    })
  ]
}));

children.push(new Paragraph({
  spacing: { after: 200 },
  children: [new TextRun({
    text: "定时任务采用幂等设计 (DB upsert / ON CONFLICT)，配合 Day 抖动避免多实例并发命中同一时间点，容忍单点失败 (SnapshotAllActive 容错，单项目失败不阻塞全局)。",
    font: "Microsoft YaHei", size: 22
  })]
}));

children.push(new Paragraph({ children: [new PageBreak()] }));

// ========== 第九章：安全设计 ==========
children.push(new Paragraph({
  heading: HeadingLevel.HEADING_1,
  pageBreakBefore: true,
  children: [new TextRun({ text: "9. 安全设计", font: "Microsoft YaHei", size: 32, bold: true })]
}));

children.push(new Paragraph({
  heading: HeadingLevel.HEADING_2,
  children: [new TextRun({ text: "9.1 认证体系", font: "Microsoft YaHei", size: 26, bold: true })]
}));

children.push(new Table({
  width: { size: 9000, type: WidthType.DXA },
  columnWidths: [2200, 6800],
  rows: [
    new TableRow({
      children: [
        headerCell("机制", 2200),
        headerCell("说明", 6800)
      ]
    }),
    new TableRow({
      children: [
        cell("密码登录", 2200, altRowShading),
        cell("bcrypt (cost=12) + 邮箱验证", 6800, altRowShading)
      ]
    }),
    new TableRow({
      children: [
        cell("JWT", 2200),
        cell("Access Token (15m) + Refresh Token (30d)，HttpOnly Cookie", 6800)
      ]
    }),
    new TableRow({
      children: [
        cell("API Token", 2200, altRowShading),
        cell("ydz_ 前缀，Personal Access Token 鉴权", 6800, altRowShading)
      ]
    }),
    new TableRow({
      children: [
        cell("OIDC", 2200),
        cell("预留接口 (SSO / SAML P3)", 6800)
      ]
    })
  ]
}));

children.push(new Paragraph({ children: [new PageBreak()] }));

children.push(new Paragraph({
  heading: HeadingLevel.HEADING_2,
  children: [new TextRun({ text: "9.2 RBAC 权限模型", font: "Microsoft YaHei", size: 26, bold: true })]
}));

children.push(new Paragraph({
  spacing: { after: 200 },
  children: [new TextRun({
    text: "双层 RBAC：workspace 角色与 project 角色。四角色 (owner/admin/member/guest) 映射到 10 项权限码。权限码由 DB-backed 的 RBACStore 加载，中间件按路由校验：",
    font: "Microsoft YaHei", size: 22
  })]
}));

children.push(new Paragraph({
  spacing: { after: 200 },
  children: [new TextRun({
    text: 'workspace:read|update|delete · project:create|update|delete · member:invite|change_role|remove · automation:manage · audit:read · analytics:read',
    font: "Consolas", size: 20, color: "1F4E79"
  })]
}));

children.push(new Paragraph({
  heading: HeadingLevel.HEADING_2,
  children: [new TextRun({ text: "9.3 安全纵深", font: "Microsoft YaHei", size: 26, bold: true })]
}));

children.push(new Paragraph({
  spacing: { after: 200 },
  children: [new TextRun({
    text: "HTTP 安全头覆盖 8 项：CSP/HSTS/COOP/CORP/Permissions-Policy/X-Frame-Options/X-Content-Type-Options/Referrer-Policy。参数化查询防 SQL 注入，限流防暴力破解 + CC 攻击，CORS 白名单 + 凭证许可，Webhook HMAC-SHA256 签名校验。附件私有桶 + 预签名 URL，Token 仅存 hash。审计日志记录管理类操作。",
    font: "Microsoft YaHei", size: 22
  })]
}));

children.push(new Paragraph({
  heading: HeadingLevel.HEADING_2,
  children: [new TextRun({ text: "9.4 国密算法预留", font: "Microsoft YaHei", size: 26, bold: true })]
}));

children.push(new Paragraph({
  spacing: { after: 200 },
  children: [new TextRun({
    text: "通过 build tag 隔离实现 SM2/SM3/SM4 国密算法 (pkg/crypto/)，TLS 与存储加密抽象为接口，便于切换信创合规加密模块。等保三级基线要求的密码策略、审计、防篡改均已规划。",
    font: "Microsoft YaHei", size: 22
  })]
}));

children.push(new Paragraph({ children: [new PageBreak()] }));

// ========== 第十章：部署与运维 ==========
children.push(new Paragraph({
  heading: HeadingLevel.HEADING_1,
  pageBreakBefore: true,
  children: [new TextRun({ text: "10. 部署与运维", font: "Microsoft YaHei", size: 32, bold: true })]
}));

children.push(new Paragraph({
  heading: HeadingLevel.HEADING_2,
  children: [new TextRun({ text: "10.1 部署架构", font: "Microsoft YaHei", size: 26, bold: true })]
}));

children.push(new Table({
  width: { size: 9000, type: WidthType.DXA },
  columnWidths: [1800, 1600, 5600],
  rows: [
    new TableRow({
      children: [
        headerCell("组件", 1800),
        headerCell("端口", 1600),
        headerCell("用途", 5600)
      ]
    }),
    new TableRow({
      children: [
        cell("nginx", 1800, altRowShading),
        cell("80/443", 1600, altRowShading),
        cell("反代/WAF/静态资源/限流", 5600, altRowShading)
      ]
    }),
    new TableRow({
      children: [
        cell("web", 1800),
        cell("静态产物", 1600),
        cell("Vite 构建 SPA，由 nginx serve", 5600)
      ]
    }),
    new TableRow({
      children: [
        cell("api", 1800, altRowShading),
        cell("8080", 1600, altRowShading),
        cell("后端 HTTP，可水平扩展", 5600, altRowShading)
      ]
    }),
    new TableRow({
      children: [
        cell("worker", 1800),
        cell("—", 1600),
        cell("后台异步任务，同镜像不同 CMD", 5600)
      ]
    }),
    new TableRow({
      children: [
        cell("postgres", 1800, altRowShading),
        cell("5432", 1600, altRowShading),
        cell("主业务 DB", 5600, altRowShading)
      ]
    }),
    new TableRow({
      children: [
        cell("redis", 1800),
        cell("6379", 1600),
        cell("缓存/限流/WS 扇出", 5600)
      ]
    }),
    new TableRow({
      children: [
        cell("rabbitmq", 1800, altRowShading),
        cell("5672", 1600, altRowShading),
        cell("事件总线/任务队列", 5600, altRowShading)
      ]
    }),
    new TableRow({
      children: [
        cell("mailpit", 1800),
        cell("1025/8025", 1600),
        cell("本地邮件测试 (dev)", 5600)
      ]
    }),
    new TableRow({
      children: [
        cell("es", 1800, altRowShading),
        cell("9200", 1600, altRowShading),
        cell("可选 full profile", 5600, altRowShading)
      ]
    }),
    new TableRow({
      children: [
        cell("minio", 1800),
        cell("9000", 1600),
        cell("可选附件存储", 5600)
      ]
    })
  ]
}));

children.push(new Paragraph({ children: [new PageBreak()] }));

children.push(new Paragraph({
  heading: HeadingLevel.HEADING_2,
  children: [new TextRun({ text: "10.2 一键启动命令", font: "Microsoft YaHei", size: 26, bold: true })]
}));

children.push(new Paragraph({
  spacing: { after: 100 },
  children: [new TextRun({ text: "docker compose -f docs/deployments/docker-compose.yml up -d", font: "Consolas", size: 20, color: "1F4E79" })]
}));
children.push(new Paragraph({
  spacing: { after: 200 },
  children: [new TextRun({ text: "docker compose --profile full up -d  # 含 ES + MinIO", font: "Consolas", size: 20, color: "808080" })]
}));

children.push(new Paragraph({
  heading: HeadingLevel.HEADING_2,
  children: [new TextRun({ text: "10.3 Makefile 命令速查", font: "Microsoft YaHei", size: 26, bold: true })]
}));

children.push(new Table({
  width: { size: 9000, type: WidthType.DXA },
  columnWidths: [3000, 6000],
  rows: [
    new TableRow({
      children: [
        headerCell("命令", 3000),
        headerCell("说明", 6000)
      ]
    }),
    new TableRow({
      children: [
        cell("make dev", 3000, altRowShading),
        cell("启动基础设施容器", 6000, altRowShading)
      ]
    }),
    new TableRow({
      children: [
        cell("make up|down", 3000),
        cell("起停服务", 6000)
      ]
    }),
    new TableRow({
      children: [
        cell("make migrate", 3000, altRowShading),
        cell("迁移到最新", 6000, altRowShading)
      ]
    }),
    new TableRow({
      children: [
        cell("make seed", 3000),
        cell("导入幂等种子数据", 6000)
      ]
    }),
    new TableRow({
      children: [
        cell("make seed-scale", 3000, altRowShading),
        cell("百万级工作项造数", 6000, altRowShading)
      ]
    }),
    new TableRow({
      children: [
        cell("make dev-api", 3000),
        cell("air 热重载 API", 6000)
      ]
    }),
    new TableRow({
      children: [
        cell("make dev-worker", 3000, altRowShading),
        cell("启动 worker", 6000, altRowShading)
      ]
    }),
    new TableRow({
      children: [
        cell("make dev-web", 3000),
        cell("前端 dev server", 6000)
      ]
    }),
    new TableRow({
      children: [
        cell("make lint|test|build", 3000, altRowShading),
        cell("全链路 lint/测试/构建", 6000, altRowShading)
      ]
    }),
    new TableRow({
      children: [
        cell("make openapi", 3000),
        cell("生成 swaggo 文档", 6000)
      ]
    })
  ]
}));

children.push(new Paragraph({ children: [new PageBreak()] }));

children.push(new Paragraph({
  heading: HeadingLevel.HEADING_2,
  children: [new TextRun({ text: "10.4 健康检查与监控", font: "Microsoft YaHei", size: 26, bold: true })]
}));

children.push(new Table({
  width: { size: 9000, type: WidthType.DXA },
  columnWidths: [2400, 6600],
  rows: [
    new TableRow({
      children: [
        headerCell("端点", 2400),
        headerCell("用途", 6600)
      ]
    }),
    new TableRow({
      children: [
        cell("/healthz", 2400, altRowShading),
        cell("健康检查 (200 ok)", 6600, altRowShading)
      ]
    }),
    new TableRow({
      children: [
        cell("/readyz", 2400),
        cell("就绪检查 (+PG/Redis 探针)", 6600)
      ]
    }),
    new TableRow({
      children: [
        cell("/metrics", 2400, altRowShading),
        cell("Prometheus RED 指标", 6600, altRowShading)
      ]
    }),
    new TableRow({
      children: [
        cell("/swagger", 2400),
        cell("Swagger UI (仅 dev)", 6600)
      ]
    })
  ]
}));

children.push(new Paragraph({ spacing: { before: 200 } }));

children.push(new Paragraph({
  spacing: { after: 200 },
  children: [new TextRun({
    text: "日志由 ELK / Loki 收集，Prometheus + Grafana 看板监控 QPS/延迟/队列积压/WS 连接数。OpenTelemetry 埋点覆盖 Gin/HTTP/DB，Jaeger 可选。",
    font: "Microsoft YaHei", size: 22
  })]
}));

children.push(new Paragraph({ children: [new PageBreak()] }));

// ========== 第十一章：里程碑与路线图 ==========
children.push(new Paragraph({
  heading: HeadingLevel.HEADING_1,
  pageBreakBefore: true,
  children: [new TextRun({ text: "11. 里程碑与路线图", font: "Microsoft YaHei", size: 32, bold: true })]
}));

children.push(new Table({
  width: { size: 9000, type: WidthType.DXA },
  columnWidths: [1200, 1600, 6200],
  rows: [
    new TableRow({
      children: [
        headerCell("阶段", 1200),
        headerCell("状态", 1600),
        headerCell("核心交付物", 6200)
      ]
    }),
    new TableRow({
      children: [
        cell("M0 基座", 1200, altRowShading),
        cell("✅", 1600, altRowShading),
        cell("Monorepo / CI / Docker / Gin 骨架 / 鉴权链路 / 种子数据", 6200, altRowShading)
      ]
    }),
    new TableRow({
      children: [
        cell("M0.5 增强", 1200),
        cell("✅", 1600),
        cell("Prometheus / RBAC / 安全头 / swaggo / 邮件抽象 / Axios 拦截器", 6200)
      ]
    }),
    new TableRow({
      children: [
        cell("M1 租户", 1200, altRowShading),
        cell("✅", 1600, altRowShading),
        cell("Workspace / Project CRUD / 成员邀请 / 审计日志 / 前端导航", 6200, altRowShading)
      ]
    }),
    new TableRow({
      children: [
        cell("M2 工作项", 1200),
        cell("✅", 1600),
        cell("Issue 聚合 / 状态机 / WBS / 看板 / 批量 / 评论 / 工时 / 关联 / 导出", 6200)
      ]
    }),
    new TableRow({
      children: [
        cell("M3 迭代", 1200, altRowShading),
        cell("✅", 1600, altRowShading),
        cell("Sprint 生命周期 / 燃尽图 / Backlog / 复盘 / 前端视图", 6200, altRowShading)
      ]
    }),
    new TableRow({
      children: [
        cell("M4 版本日", 1200),
        cell("✅", 1600),
        cell("Version CRUD / 状态机 / Release Notes / 交付报告 / 缺陷过滤", 6200)
      ]
    }),
    new TableRow({
      children: [
        cell("M5 MVP", 1200, altRowShading),
        cell("✅", 1600, altRowShading),
        cell("附件 / 评论 / 通知 / WS / 性能基线 / 导出基础设施", 6200, altRowShading)
      ]
    }),
    new TableRow({
      children: [
        cell("M6 开放", 1200),
        cell("✅", 1600),
        cell("搜索 / 工作台 / 仪表盘 / Webhook / Intake / 自动化 / 效能 / 前端完整视图", 6200)
      ]
    })
  ]
}));

children.push(new Paragraph({ spacing: { before: 400, after: 200 } }));

children.push(new Paragraph({
  children: [new TextRun({ text: "Phase 3+ (S12+): 通知 IM 多渠道 / ES 升级 / 甘特图 / 日历 / 电子表格 / SSO / 国际化 / PWA / 数据迁移 / AI 功能 / 信创实测", font: "Microsoft YaHei", size: 20, color: "808080", italics: true })]
}));

children.push(new Paragraph({ children: [new PageBreak()] }));

// ========== 第十二章：本地开发 ==========
children.push(new Paragraph({
  heading: HeadingLevel.HEADING_1,
  pageBreakBefore: true,
  children: [new TextRun({ text: "12. 本地开发指南", font: "Microsoft YaHei", size: 32, bold: true })]
}));

children.push(new Paragraph({
  heading: HeadingLevel.HEADING_2,
  children: [new TextRun({ text: "12.1 服务依赖", font: "Microsoft YaHei", size: 26, bold: true })]
}));

children.push(new Table({
  width: { size: 9000, type: WidthType.DXA },
  columnWidths: [2000, 2200, 2400, 2400],
  rows: [
    new TableRow({
      children: [
        headerCell("服务", 2000),
        headerCell("地址", 2200),
        headerCell("账密", 2400),
        headerCell("说明", 2400)
      ]
    }),
    new TableRow({
      children: [
        cell("PostgreSQL", 2000, altRowShading),
        cell("127.0.0.1:5432", 2200, altRowShading),
        cell("postgres/Limw1020", 2400, altRowShading),
        cell("主 DB", 2400, altRowShading)
      ]
    }),
    new TableRow({
      children: [
        cell("Redis", 2000),
        cell("127.0.0.1:6379", 2200),
        cell("Limw1020", 2400),
        cell("缓存/限流", 2400)
      ]
    }),
    new TableRow({
      children: [
        cell("RabbitMQ", 2000, altRowShading),
        cell("127.0.0.1:5672", 2200, altRowShading),
        cell("guest/guest", 2400, altRowShading),
        cell("事件/任务", 2400, altRowShading)
      ]
    }),
    new TableRow({
      children: [
        cell("Mailpit", 2000),
        cell(":1025 / :8025", 2200),
        cell("—", 2400),
        cell("邮件测试", 2400)
      ]
    })
  ]
}));

children.push(new Paragraph({ children: [new PageBreak()] }));

children.push(new Paragraph({
  heading: HeadingLevel.HEADING_2,
  children: [new TextRun({ text: "12.2 种子数据账密", font: "Microsoft YaHei", size: 26, bold: true })]
}));

children.push(new Paragraph({
  spacing: { after: 200 },
  children: [new TextRun({ text: "admin@ydsz.dev / Admin@123", font: "Consolas", size: 22, color: "1F4E79" })]
}));

children.push(new Paragraph({
  heading: HeadingLevel.HEADING_2,
  children: [new TextRun({ text: "12.3 启动流程", font: "Microsoft YaHei", size: 26, bold: true })]
}));

const steps = [
  "make up                    # 启动基础设施",
  "cp .env.example .env       # 环境变量",
  "make migrate               # 数据库迁移",
  "make seed                  # 导入种子数据",
  "make dev-api               # API (air 热重载)",
  "make dev-worker            # 异步 worker",
  "cd web && pnpm install && pnpm dev   # 前端 dev server"
];

steps.forEach(step => {
  children.push(new Paragraph({
    spacing: { after: 100 },
    children: [new TextRun({ text: step, font: "Consolas", size: 20, color: "404040" })]
  }));
});

children.push(new Paragraph({ spacing: { before: 400, after: 200 } }));

// ========== 附录 ==========
children.push(new Paragraph({
  heading: HeadingLevel.HEADING_1,
  pageBreakBefore: true,
  children: [new TextRun({ text: "附录：关键数值与指标", font: "Microsoft YaHei", size: 32, bold: true })]
}));

children.push(new Table({
  width: { size: 9000, type: WidthType.DXA },
  columnWidths: [3000, 3000, 3000],
  rows: [
    new TableRow({
      children: [
        headerCell("指标", 3000),
        headerCell("目标", 3000),
        headerCell("说明", 3000)
      ]
    }),
    new TableRow({
      children: [
        cell("API P95", 3000, altRowShading),
        cell("≤ 200ms", 3000, altRowShading),
        cell("单实例基线", 3000, altRowShading)
      ]
    }),
    new TableRow({
      children: [
        cell("页面加载 P95", 3000),
        cell("≤ 2s", 3000),
        cell("首屏", 3000)
      ]
    }),
    new TableRow({
      children: [
        cell("并发用户", 3000, altRowShading),
        cell("≥ 1000", 3000, altRowShading),
        cell("单实例", 3000, altRowShading)
      ]
    }),
    new TableRow({
      children: [
        cell("单项目工作项", 3000),
        cell("≥ 100 万", 3000),
        cell("PostgreSQL 容量", 3000)
      ]
    }),
    new TableRow({
      children: [
        cell("可用性 SLA", 3000, altRowShading),
        cell("≥ 99.9%", 3000, altRowShading),
        cell("年度停机 <8.76h", 3000, altRowShading)
      ]
    }),
    new TableRow({
      children: [
        cell("已交付 API", 3000),
        cell("80+", 3000),
        cell("REST + WS + Public", 3000)
      ]
    }),
    new TableRow({
      children: [
        cell("前端视图", 3000, altRowShading),
        cell("45+", 3000, altRowShading),
        cell("多视图 + 设置页", 3000, altRowShading)
      ]
    }),
    new TableRow({
      children: [
        cell("自动化模板", 3000),
        cell("7", 3000),
        cell("内置业务模板", 3000)
      ]
    }),
    new TableRow({
      children: [
        cell("Widget 类型", 3000, altRowShading),
        cell("10", 3000, altRowShading),
        cell("仪表盘可配置", 3000, altRowShading)
      ]
    })
  ]
}));

children.push(new Paragraph({ spacing: { before: 400 } }));

children.push(new Paragraph({
  alignment: AlignmentType.CENTER,
  children: [new TextRun({ text: "--- Ydsz Plane 详细技术说明文档 | 2026-08-08 ---", font: "Microsoft YaHei", size: 20, color: "808080", italics: true })]
}));

// 创建文档
const doc = new Document({
  creator: "Ydsz Plane Team",
  title: "Ydsz Plane 详细技术说明文档",
  description: "基于最新代码的详细技术文档",
  styles: {
    default: {
      document: {
        run: { font: "Microsoft YaHei", size: 22 }
      }
    },
    paragraphStyles: [
      {
        id: "Heading1", name: "Heading 1", basedOn: "Normal", next: "Normal", quickFormat: true,
        run: { size: 32, bold: true, font: "Microsoft YaHei", color: "1F4E79" },
        paragraph: { spacing: { before: 360, after: 200 }, outlineLevel: 0 }
      },
      {
        id: "Heading2", name: "Heading 2", basedOn: "Normal", next: "Normal", quickFormat: true,
        run: { size: 26, bold: true, font: "Microsoft YaHei", color: "2E75B6" },
        paragraph: { spacing: { before: 280, after: 160 }, outlineLevel: 1 }
      },
      {
        id: "Heading3", name: "Heading 3", basedOn: "Normal", next: "Normal", quickFormat: true,
        run: { size: 24, bold: true, font: "Microsoft YaHei", color: "404040" },
        paragraph: { spacing: { before: 200, after: 120 }, outlineLevel: 2 }
      },
      {
        id: "Heading4", name: "Heading 4", basedOn: "Normal", next: "Normal", quickFormat: true,
        run: { size: 22, bold: true, font: "Microsoft YaHei", color: "404040" },
        paragraph: { spacing: { before: 160, after: 100 }, outlineLevel: 3 }
      }
    ]
  },
  numbering: {
    config: []
  },
  sections: [{
    properties: {
      page: {
        size: { width: 11906, height: 16838 }, // A4
        margin: { top: 1440, right: 1440, bottom: 1440, left: 1440 }
      }
    },
    headers: {
      default: new Header({
        children: [new Paragraph({
          alignment: AlignmentType.RIGHT,
          children: [new TextRun({ text: "Ydsz Plane 技术文档", font: "Microsoft YaHei", size: 18, color: "808080" })]
        })]
      })
    },
    footers: {
      default: new Footer({
        children: [new Paragraph({
          alignment: AlignmentType.CENTER,
          children: [
            new TextRun({ font: "Microsoft YaHei", size: 18, color: "808080" }),
            new TextRun({ children: [PageNumber.CURRENT], font: "Microsoft YaHei", size: 18, color: "808080" })
          ]
        })]
      })
    },
    children
  }]
});

// 生成文件
Packer.toBuffer(doc).then(buffer => {
  fs.writeFileSync("D:/Code/open/ydsz-plane/docs/Ydsz Plane 详细技术说明文档.docx", buffer);
  console.log("文档生成成功: D:/Code/open/ydsz-plane/docs/Ydsz Plane 详细技术说明文档.docx");
  console.log(`文件大小: ${(buffer.length / 1024).toFixed(0)} KB`);
});
