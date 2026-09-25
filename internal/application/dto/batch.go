// Package dto 定义跨领域共享的数据传输对象（DTO）。
//
// 包括标准化批量操作响应、分页错误格式等 API 契约结构体。
// 本包不含业务逻辑，仅承载 HTTP 请求/响应的 JSON 序列化契约。
package dto

// BatchResponse 标准批量操作响应。
//
// 规范约定：
//   - Jira Bulk Operation 风格 + Asana Batch API 风格的折中。
//   - 事务整体成功时 errors 为空切片（前端可区分 "全部成功" vs "部分失败"）。
//   - BatchError 绝不暴露 PII（internal stacktrace、credentials、SQL 片段等）。
//
// 示例：
//
//	{
//	  "success_count": 3,
//	  "failed_count": 1,
//	  "errors": [
//	    {"resource_id": 42, "code": "RESOURCE.NOT_FOUND", "message": "资源不存在"}
//	  ]
//	}
type BatchResponse struct {
	// SuccessCount 成功处理的条目数。
	SuccessCount int `json:"success_count"`
	// FailedCount 处理失败的条目数。
	FailedCount int `json:"failed_count"`
	// Errors 失败条目明细；全部成功时返回空切片（保证前端 null-safety）。
	Errors []BatchError `json:"errors"`
}

// BatchError 单个资源失败详情。
//
// 安全约束：
//   - Message 仅含面向终端用户的中文提示，不暴露内部堆栈/凭证/SQL。
//   - Code 复用 errs.AppError 的全局错误码体系。
type BatchError struct {
	// ResourceID 失败的资源 ID（工作项 ID 等）。
	ResourceID int64 `json:"resource_id"`
	// Code 错误码（eg. RESOURCE.NOT_FOUND / ISSUE.VERSION_CONFLICT）。
	Code string `json:"code"`
	// Message 面向终端用户的中文提示（无敏感信息）。
	Message string `json:"message"`
}

// NewBatchSuccess 构造全部成功的 BatchResponse。
func NewBatchSuccess(count int) BatchResponse {
	return BatchResponse{
		SuccessCount: count,
		FailedCount:  0,
		Errors:       []BatchError{},
	}
}
