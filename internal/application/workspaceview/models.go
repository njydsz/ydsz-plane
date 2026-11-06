// Package workspaceview — 工作空间级视图（Workspace Views）领域。
//
// 对标 Plane 的「Saved Views」：跨项目的工作空间级命名视图，保存布局 /
// 过滤器 / 展示选项 / 排序配置，支持公开与收藏。
//
// 数据表：workspace_views（id 由应用层 genID 生成）。
package workspaceview

import (
	"crypto/rand"
	"encoding/binary"
	"encoding/json"
	"time"
)

// Layout 视图布局类型。
type Layout string

const (
	LayoutBoard       Layout = "board"
	LayoutList        Layout = "list"
	LayoutSpreadsheet Layout = "spreadsheet"
	LayoutGantt       Layout = "gantt"
	LayoutCalendar    Layout = "calendar"
)

// IsValid 校验布局是否合法。
func (l Layout) IsValid() bool {
	switch l {
	case LayoutBoard, LayoutList, LayoutSpreadsheet, LayoutGantt, LayoutCalendar:
		return true
	default:
		return false
	}
}

// WorkspaceView 工作空间视图实体。
type WorkspaceView struct {
	ID             int64           `json:"id"`
	WorkspaceID    int64           `json:"workspace_id"`
	Name           string          `json:"name"`
	Description    string          `json:"description,omitempty"`
	Layout         Layout          `json:"layout"`
	Filters        json.RawMessage `json:"filters"`
	DisplayOptions json.RawMessage `json:"display_options"`
	Sort           json.RawMessage `json:"sort"`
	IsPublic       bool            `json:"is_public"`
	IsFavorite     bool            `json:"is_favorite"`
	CreatedBy      int64           `json:"created_by"`
	CreatedAt      time.Time       `json:"created_at"`
	UpdatedAt      time.Time       `json:"updated_at"`
}

// --- Input Types ---

// CreateInput 创建视图入参。
type CreateInput struct {
	WorkspaceID    int64
	Name           string
	Description    string
	Layout         Layout
	Filters        json.RawMessage
	DisplayOptions json.RawMessage
	Sort           json.RawMessage
	IsPublic       bool
	IsFavorite     bool
	CreatedBy      int64
}

// UpdateInput 更新视图入参（指针字段为 nil 时不更新）。
type UpdateInput struct {
	Name           *string
	Description    *string
	Layout         *Layout
	Filters        *json.RawMessage
	DisplayOptions *json.RawMessage
	Sort           *json.RawMessage
	IsPublic       *bool
	IsFavorite     *bool
}

// ListOptions 列表查询选项。
type ListOptions struct {
	WorkspaceID int64
	// CreatedBy > 0 时按创建人过滤（"我创建的视图"）。
	CreatedBy int64
	// PublicOnly=true 时仅返回 is_public=true 的视图。
	PublicOnly bool
	// FavoriteOnly=true 时仅返回当前用户收藏的视图。
	FavoriteOnly bool
	Limit        int
	Offset       int
}

// --- ID 生成 ---

// genID 生成基于时间戳的伪随机 ID（与其它应用层包策略一致）。
func genID() int64 {
	var b [8]byte
	_, _ = rand.Read(b[:])
	randBits := int64(binary.LittleEndian.Uint64(b[:]) & 0x3FFFFF)
	return (time.Now().UnixMilli() << 22) | randBits
}
