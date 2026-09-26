// Package issue — 工作项域 HTTP handlers + DDD 聚合根。
//
// S18 P0-9 架构解耦：issue 模块按子域拆分，本文件为主聚合入口，
// 为上层（router / cmd/api）提供统一的服务接口。
//
// 子域结构：
//   - shared	: 共享类型（枚举、常量、DTO、错误）
//   - requirement	: 需求子域（requirement_service.go）
//   - task		: 任务子域（task_service.go + timelog_service.go）
//   - defect	: 缺陷子域（defect_service.go + defect_analytics.go）
//
// 依赖方向：router → Handler（本包）→ 子域 service → 子域 repository
package issue

// 注：实际文件保留在 issue/ 根目录以兼容现有导入路径（渐进式重构）。
// 子域包独立的类型定义在各自目录下。
