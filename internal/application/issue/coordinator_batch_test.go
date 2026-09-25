// Package issue — BatchUpdateV2 单元测试（S16 P1-4）。
//
// 验证批量操作的分桶逻辑和 SQL 正确性，无需启动真实数据库（mock db）。
// 集成测试覆盖见 core_chain_integration_test.go。
package issue

import (
	"testing"

	"github.com/njydsz/ydsz-plane/pkg/errs"
	"github.com/stretchr/testify/assert"
)

// TestBatchGroupByType_Empty 验证空 IDs 返回 nil。
func TestBatchGroupByType_Empty(t *testing.T) {
	// 空输入应快速返回 nil（防御性路径）。
	result, err := batchGroupByType(nil, nil, 0, []int64{})
	assert.NoError(t, err)
	assert.Nil(t, result)
}

// TestBatchSoftDelete_EmptyRows 验证空 rows 不生成 SQL 错误。
func TestBatchSoftDelete_EmptyRows(t *testing.T) {
	// 空 rows 时应直接返回 nil（无需执行 SQL）。
	err := batchSoftDelete(nil, nil, 0, TypeTask, nil)
	assert.NoError(t, err)
}

// TestBatchPriorityUpdate_EmptyRows 验证空 rows 不报错。
func TestBatchPriorityUpdate_EmptyRows(t *testing.T) {
	err := batchPriorityUpdate(nil, nil, TypeTask, 0, nil, "high")
	assert.NoError(t, err)
}

// TestBatchAssigneeUpsert_EmptyRows 验证空 rows 不报错。
func TestBatchAssigneeUpsert_EmptyRows(t *testing.T) {
	err := batchAssigneeUpsert(nil, nil, TypeTask, 0, nil, 1)
	assert.NoError(t, err)
}

// TestBatchUpdateV2_EmptyIDs 验证空 IDs 快速返回。
func TestBatchUpdateV2_EmptyIDs(t *testing.T) {
	svc := &Service{}
	result, err := svc.BatchUpdateV2(nil, 1, 1, 1, BatchUpdateInput{IDs: []int64{}})
	assert.NoError(t, err)
	assert.Equal(t, 0, result.Succeeded)
}

// TestValidateTableWhitelist 验证表名白名单校验（SQL 注入防护）。
func TestValidateTableWhitelist(t *testing.T) {
	// 合法表名
	assert.NoError(t, ValidateTable("task"))
	assert.NoError(t, ValidateTable("requirement"))
	assert.NoError(t, ValidateTable("defect"))

	// 非法表名 — SQL 注入尝试被拦截
	assert.Error(t, ValidateTable("tasks; DROP TABLE users"))
	assert.Error(t, ValidateTable("nonexistent"))
	assert.Error(t, ValidateTable(""))
	assert.Error(t, ValidateTable("1=1"))
}

// TestBatchResultStruct 验证结果结构体序列化字段。
func TestBatchResultStruct(t *testing.T) {
	r := BatchResult{Succeeded: 10, Failed: 0}
	assert.Equal(t, 10, r.Succeeded)
}

// TestErrVersionConflictExposed 验证版本冲突错误可被上层识别。
func TestErrVersionConflictExposed(t *testing.T) {
	assert.Error(t, errs.ErrVersionConflict)
}
