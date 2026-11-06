// Package workitemtype — 自定义工作项类型域。
//
// 对标 Plane Issue Type：项目可定义自定义工作项类型（Story / Bug / ...），
// 每种类型可配置一组自定义属性（字段），工作项按类型携带属性值。
//
// 数据表：
//   - work_item_types（id 由应用层 genID 生成）
//   - work_item_type_properties（id 由应用层 genID 生成）
//   - work_item_custom_values（id 由应用层 genID 生成）
//
// 设计要点：
//   - 类型（WorkItemType）为项目级，code 在同一项目内唯一；
//   - 类型可归档（archived_at）/软删除（deleted_at），归档后不再出现在列表；
//   - 属性（WorkItemTypeProperty）的 code 在同一类型内唯一；
//   - 属性值（WorkItemCustomFieldValue）采用 EAV 模型，value 统一存 TEXT，
//     由 field_type 解释（number/date/checkbox 等按前端约定序列化）；
//   - work_item_type 字段记录值归属的工作项类型（内置 requirement/task/defect
//     或自定义类型 code），便于跨类型查询与索引。
package workitemtype

import (
	"crypto/rand"
	"encoding/binary"
	"time"
)

// FieldType 自定义属性字段类型枚举。
type FieldType string

const (
	// FieldTypeText 单行文本。
	FieldTypeText FieldType = "text"
	// FieldTypeNumber 数字。
	FieldTypeNumber FieldType = "number"
	// FieldTypeSelect 单选下拉（options 为候选项数组）。
	FieldTypeSelect FieldType = "select"
	// FieldTypeMultiselect 多选下拉（options 为候选项数组）。
	FieldTypeMultiselect FieldType = "multiselect"
	// FieldTypeDate 日期（ISO 8601 字符串）。
	FieldTypeDate FieldType = "date"
	// FieldTypeCheckbox 布尔勾选。
	FieldTypeCheckbox FieldType = "checkbox"
	// FieldTypeURL 链接。
	FieldTypeURL FieldType = "url"
)

// IsValid 校验字段类型是否合法。
func (f FieldType) IsValid() bool {
	switch f {
	case FieldTypeText, FieldTypeNumber, FieldTypeSelect, FieldTypeMultiselect,
		FieldTypeDate, FieldTypeCheckbox, FieldTypeURL:
		return true
	default:
		return false
	}
}

// WorkItemType 自定义工作项类型定义（项目级）。
type WorkItemType struct {
	ID          int64      `json:"id"`
	ProjectID   int64      `json:"project_id"`
	Name        string     `json:"name"`        // e.g. "Story", "Bug"
	Code        string     `json:"code"`        // unique slug e.g. "story"
	Color       string     `json:"color"`
	Icon        string     `json:"icon"`
	Description string     `json:"description"`
	Level       int        `json:"level"` // 0=task level, 1=epic level
	IsActive    bool       `json:"is_active"`
	IsDefault   bool       `json:"is_default"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	ArchivedAt  *time.Time `json:"archived_at,omitempty"`
}

// WorkItemTypeProperty 类型的自定义属性定义。
type WorkItemTypeProperty struct {
	ID        int64     `json:"id"`
	TypeID    int64     `json:"type_id"`
	Name      string    `json:"name"`
	Code      string    `json:"code"`
	FieldType string    `json:"field_type"` // text/number/select/multiselect/date/checkbox/url
	Options   string    `json:"options"`    // JSON array for select types
	Required  bool      `json:"required"`
	SortOrder int       `json:"sort_order"`
	CreatedAt time.Time `json:"created_at"`
}

// WorkItemCustomFieldValue 工作项的自定义属性值。
type WorkItemCustomFieldValue struct {
	ID           int64     `json:"id"`
	PropertyID   int64     `json:"property_id"`
	WorkItemID   int64     `json:"work_item_id"`
	WorkItemType string    `json:"work_item_type"` // requirement/task/defect/custom type code
	Value        string    `json:"value"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// --- Input Types ---

// CreateTypeInput 创建工作项类型入参。
type CreateTypeInput struct {
	ProjectID   int64
	Name        string
	Code        string
	Color       string
	Icon        string
	Description string
	Level       int
	IsDefault   bool
}

// UpdateTypeInput 更新工作项类型入参（字段为指针，nil 表示不更新）。
type UpdateTypeInput struct {
	Name        *string
	Code        *string
	Color       *string
	Icon        *string
	Description *string
	Level       *int
	IsActive    *bool
	IsDefault   *bool
}

// ListTypesOptions 类型列表查询选项。
type ListTypesOptions struct {
	ProjectID int64
	Limit     int
	Offset    int
}

// CreatePropertyInput 新增属性入参。
type CreatePropertyInput struct {
	TypeID    int64
	Name      string
	Code      string
	FieldType string
	Options   string
	Required  bool
	SortOrder int
}

// SetCustomFieldValueInput 写入自定义属性值入参。
type SetCustomFieldValueInput struct {
	PropertyID   int64
	WorkItemID   int64
	WorkItemType string
	Value        string
}

// --- ID 生成 ---

// genID 生成基于时间戳的伪随机 ID（与其它应用层包策略一致）。
//
// 结构：高 42 位为毫秒时间戳，低 22 位为密码学随机数。
func genID() int64 {
	var b [8]byte
	_, _ = rand.Read(b[:])
	randBits := int64(binary.LittleEndian.Uint64(b[:]) & 0x3FFFFF)
	return (time.Now().UnixMilli() << 22) | randBits
}
