// Package teamspace — 团队空间域。
//
// 对标 Plane Teamspaces：团队 = 一组人 + 一组关联项目。
// 团队空间聚合展示：
//   - 团队成员被分配到的工作项；
//   - 关联项目的迭代（Cycle）rollup；
//   - 团队进度（逾期/进度/汇总）；
//   - 团队页面与活动。
//
// 数据表：
//   - teamspaces
//   - teamspace_members
//   - teamspace_projects
package teamspace

import (
	"crypto/rand"
	"encoding/binary"
	"time"
)

// Teamspace 团队空间。
type Teamspace struct {
	ID          int64      `json:"id"`
	WorkspaceID int64      `json:"workspace_id"`
	Name        string     `json:"name"`
	Description string     `json:"description,omitempty"`
	Color       string     `json:"color,omitempty"`
	LeadID      int64      `json:"lead_id"`
	CreatedBy   int64      `json:"created_by"`
	ArchivedAt  *time.Time `json:"archived_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// TeamspaceMember 团队成员。
type TeamspaceMember struct {
	TeamspaceID int64     `json:"teamspace_id"`
	UserID      int64     `json:"user_id"`
	Role        string    `json:"role"` // lead / member
	CreatedAt   time.Time `json:"created_at"`
}

// TeamspaceProject 团队关联项目。
type TeamspaceProject struct {
	TeamspaceID int64     `json:"teamspace_id"`
	ProjectID   int64     `json:"project_id"`
	CreatedAt   time.Time `json:"created_at"`
}

// TeamspaceProgress 团队进度汇总。
type TeamspaceProgress struct {
	TeamspaceID int64 `json:"teamspace_id"`
	// 成员维度
	MemberCount int `json:"member_count"`
	// 工作项维度
	TotalWorkItems      int `json:"total_work_items"`
	CompletedWorkItems  int `json:"completed_work_items"`
	OverdueWorkItems    int `json:"overdue_work_items"`
	InProgressWorkItems int `json:"in_progress_work_items"`
	// 项目维度
	ProjectCount int `json:"project_count"`
	// 迭代维度
	ActiveCycles int `json:"active_cycles"`
	// 完成率（0-100）
	CompletionRate float64 `json:"completion_rate"`
}

// TeamspaceMemberWorkload 成员工作负载。
type TeamspaceMemberWorkload struct {
	UserID          int64   `json:"user_id"`
	TotalWorkItems  int     `json:"total_work_items"`
	InProgressItems int     `json:"in_progress_items"`
	OverdueItems    int     `json:"overdue_items"`
	CompletionRate  float64 `json:"completion_rate"`
}

// --- Input Types ---

// CreateTeamspaceInput 创建团队空间入参。
type CreateTeamspaceInput struct {
	WorkspaceID int64
	Name        string
	Description string
	Color       string
	LeadID      int64
	CreatedBy   int64
	MemberIDs   []int64
	ProjectIDs  []int64
}

// UpdateTeamspaceInput 更新团队空间入参。
type UpdateTeamspaceInput struct {
	Name        *string
	Description *string
	Color       *string
	LeadID      *int64
}

// ListTeamspacesOptions 列表查询选项。
type ListTeamspacesOptions struct {
	WorkspaceID int64
	Limit       int
	Offset      int
}

// --- 枚举 ---

// MemberRole 团队成员角色。
type MemberRole string

const (
	// MemberRoleLead 团队负责人。
	MemberRoleLead MemberRole = "lead"
	// MemberRoleMember 普通成员。
	MemberRoleMember MemberRole = "member"
)

// IsValid 校验成员角色。
func (r MemberRole) IsValid() bool {
	return r == MemberRoleLead || r == MemberRoleMember
}

// --- ID 生成 ---

// genID 生成基于时间戳的伪随机 ID（与其它应用层包策略一致）。
func genID() int64 {
	var b [8]byte
	_, _ = rand.Read(b[:])
	randBits := int64(binary.LittleEndian.Uint64(b[:]) & 0x3FFFFF)
	return (time.Now().UnixMilli() << 22) | randBits
}
