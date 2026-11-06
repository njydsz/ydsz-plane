// Package customer — 客户管理应用服务。
//
// 职责：
//   - Customer CRUD + 列表筛选（stage / tier / industry / 关键字搜索）；
//   - CustomerRequest CRUD + 列表筛选（status / priority / customer）；
//   - LinkRequestToWorkItem：将客户请求关联到正式工作项（requirement/task/defect/epic）；
//   - GetCustomerStats：按状态聚合单个客户的请求数量；
//   - SearchCustomers：按名称 / 邮箱 / 行业模糊搜索。
package customer

import (
	"context"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/njydsz/ydsz-plane/pkg/errs"
)

// Service 客户管理应用服务。
type Service struct {
	db *pgxpool.Pool
}

// NewService 创建客户管理服务。
func NewService(db *pgxpool.Pool) *Service {
	return &Service{db: db}
}

// ============================================================
// 客户 CRUD
// ============================================================

// CreateCustomer 创建客户。
func (s *Service) CreateCustomer(ctx context.Context, in CreateCustomerInput) (*Customer, error) {
	if in.WorkspaceID <= 0 {
		return nil, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "workspace_id", Reason: "工作空间 ID 无效"})
	}
	name := strings.TrimSpace(in.Name)
	if name == "" {
		return nil, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "name", Reason: "客户名称不能为空"})
	}
	stage := CustomerStage(strings.TrimSpace(in.Stage))
	if stage == "" {
		stage = StageLead
	}
	if !stage.IsValid() {
		return nil, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "stage", Reason: "无效的客户阶段"})
	}
	tier := CustomerTier(strings.TrimSpace(in.Tier))
	if tier == "" {
		tier = TierFree
	}
	if !tier.IsValid() {
		return nil, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "tier", Reason: "无效的客户套餐档位"})
	}

	var c Customer
	err := s.db.QueryRow(ctx, `
		INSERT INTO customers
			(id, workspace_id, name, email, website, industry, stage, tier, region,
			 employees, revenue, account_manager_id, notes, created_by)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14)
		RETURNING id, workspace_id, name, COALESCE(email,''), COALESCE(website,''),
			COALESCE(industry,''), stage, tier, COALESCE(region,''), employees,
			COALESCE(revenue,''), account_manager_id, COALESCE(notes,''),
			created_by, created_at, updated_at`,
		genID(), in.WorkspaceID, name, nullIfEmpty(in.Email), nullIfEmpty(in.Website),
		nullIfEmpty(in.Industry), string(stage), string(tier), nullIfEmpty(in.Region),
		in.Employees, nullIfEmpty(in.Revenue), in.AccountManagerID,
		nullIfEmpty(in.Notes), in.CreatedBy,
	).Scan(&c.ID, &c.WorkspaceID, &c.Name, &c.Email, &c.Website, &c.Industry,
		&c.Stage, &c.Tier, &c.Region, &c.Employees, &c.Revenue, &c.AccountManagerID,
		&c.Notes, &c.CreatedBy, &c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}
	return &c, nil
}

// GetCustomer 获取单个客户。
func (s *Service) GetCustomer(ctx context.Context, wsID, customerID int64) (*Customer, error) {
	var c Customer
	err := s.db.QueryRow(ctx, `
		SELECT id, workspace_id, name, COALESCE(email,''), COALESCE(website,''),
			COALESCE(industry,''), stage, tier, COALESCE(region,''), employees,
			COALESCE(revenue,''), account_manager_id, COALESCE(notes,''),
			created_by, created_at, updated_at
		FROM customers
		WHERE id=$1 AND workspace_id=$2 AND deleted=false`,
		customerID, wsID,
	).Scan(&c.ID, &c.WorkspaceID, &c.Name, &c.Email, &c.Website, &c.Industry,
		&c.Stage, &c.Tier, &c.Region, &c.Employees, &c.Revenue, &c.AccountManagerID,
		&c.Notes, &c.CreatedBy, &c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, errs.ErrNotFound.WithCodeMessage("CUSTOMER.NOT_FOUND", "客户不存在")
		}
		return nil, errs.ErrInternal.Wrap(err)
	}
	return &c, nil
}

// UpdateCustomer 更新客户。
func (s *Service) UpdateCustomer(ctx context.Context, wsID, customerID int64, in UpdateCustomerInput) (*Customer, error) {
	cur, err := s.GetCustomer(ctx, wsID, customerID)
	if err != nil {
		return nil, err
	}

	if in.Name != nil {
		name := strings.TrimSpace(*in.Name)
		if name == "" {
			return nil, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "name", Reason: "客户名称不能为空"})
		}
		cur.Name = name
	}
	if in.Email != nil {
		cur.Email = strings.TrimSpace(*in.Email)
	}
	if in.Website != nil {
		cur.Website = strings.TrimSpace(*in.Website)
	}
	if in.Industry != nil {
		cur.Industry = strings.TrimSpace(*in.Industry)
	}
	if in.Stage != nil {
		stage := CustomerStage(strings.TrimSpace(*in.Stage))
		if !stage.IsValid() {
			return nil, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "stage", Reason: "无效的客户阶段"})
		}
		cur.Stage = stage
	}
	if in.Tier != nil {
		tier := CustomerTier(strings.TrimSpace(*in.Tier))
		if !tier.IsValid() {
			return nil, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "tier", Reason: "无效的客户套餐档位"})
		}
		cur.Tier = tier
	}
	if in.Region != nil {
		cur.Region = strings.TrimSpace(*in.Region)
	}
	if in.Employees != nil {
		cur.Employees = *in.Employees
	}
	if in.Revenue != nil {
		cur.Revenue = strings.TrimSpace(*in.Revenue)
	}
	if in.AccountManagerID != nil {
		cur.AccountManagerID = *in.AccountManagerID
	}
	if in.Notes != nil {
		cur.Notes = strings.TrimSpace(*in.Notes)
	}

	if _, err := s.db.Exec(ctx, `
		UPDATE customers
		SET name=$1, email=$2, website=$3, industry=$4, stage=$5, tier=$6, region=$7,
		    employees=$8, revenue=$9, account_manager_id=$10, notes=$11, updated_at=now()
		WHERE id=$12 AND workspace_id=$13 AND deleted=false`,
		cur.Name, nullIfEmpty(cur.Email), nullIfEmpty(cur.Website), nullIfEmpty(cur.Industry),
		string(cur.Stage), string(cur.Tier), nullIfEmpty(cur.Region), cur.Employees,
		nullIfEmpty(cur.Revenue), cur.AccountManagerID, nullIfEmpty(cur.Notes),
		customerID, wsID); err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}
	return s.GetCustomer(ctx, wsID, customerID)
}

// DeleteCustomer 删除客户（软删除；同时软删除其请求）。
func (s *Service) DeleteCustomer(ctx context.Context, wsID, customerID int64) error {
	tag, err := s.db.Exec(ctx,
		`UPDATE customers SET deleted=true, updated_at=now() WHERE id=$1 AND workspace_id=$2 AND deleted=false`,
		customerID, wsID)
	if err != nil {
		return errs.ErrInternal.Wrap(err)
	}
	if tag.RowsAffected() == 0 {
		return errs.ErrNotFound.WithCodeMessage("CUSTOMER.NOT_FOUND", "客户不存在")
	}
	_, _ = s.db.Exec(ctx,
		`UPDATE customer_requests SET deleted=true, updated_at=now() WHERE customer_id=$1 AND deleted=false`,
		customerID)
	return nil
}

// ListCustomers 列出客户（支持 stage / tier / industry / 关键字搜索）。
func (s *Service) ListCustomers(ctx context.Context, f ListCustomersFilter) ([]Customer, int, error) {
	where := []string{"c.workspace_id=$1", "c.deleted=false"}
	args := []any{f.WorkspaceID}
	idx := 2
	if f.Stage != "" {
		where = append(where, "c.stage=$"+strconv.Itoa(idx))
		args = append(args, f.Stage)
		idx++
	}
	if f.Tier != "" {
		where = append(where, "c.tier=$"+strconv.Itoa(idx))
		args = append(args, f.Tier)
		idx++
	}
	if f.Industry != "" {
		where = append(where, "c.industry=$"+strconv.Itoa(idx))
		args = append(args, f.Industry)
		idx++
	}
	if q := strings.TrimSpace(f.Search); q != "" {
		where = append(where, "(c.name ILIKE $"+strconv.Itoa(idx)+" OR c.email ILIKE $"+strconv.Itoa(idx)+" OR c.industry ILIKE $"+strconv.Itoa(idx)+")")
		args = append(args, "%"+q+"%")
		idx++
	}
	cond := strings.Join(where, " AND ")

	var total int
	if err := s.db.QueryRow(ctx, `SELECT COUNT(*) FROM customers c WHERE `+cond, args...).Scan(&total); err != nil {
		return nil, 0, errs.ErrInternal.Wrap(err)
	}

	limit, offset := f.Limit, f.Offset
	if limit <= 0 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}
	args = append(args, limit, offset)

	rows, err := s.db.Query(ctx, `
		SELECT c.id, c.workspace_id, c.name, COALESCE(c.email,''), COALESCE(c.website,''),
			COALESCE(c.industry,''), c.stage, c.tier, COALESCE(c.region,''), c.employees,
			COALESCE(c.revenue,''), c.account_manager_id, COALESCE(c.notes,''),
			c.created_by, c.created_at, c.updated_at
		FROM customers c
		WHERE `+cond+`
		ORDER BY c.created_at DESC
		LIMIT $`+strconv.Itoa(idx)+` OFFSET $`+strconv.Itoa(idx+1),
		args...)
	if err != nil {
		return nil, 0, errs.ErrInternal.Wrap(err)
	}
	defer rows.Close()

	out := make([]Customer, 0)
	for rows.Next() {
		var c Customer
		if err := rows.Scan(&c.ID, &c.WorkspaceID, &c.Name, &c.Email, &c.Website, &c.Industry,
			&c.Stage, &c.Tier, &c.Region, &c.Employees, &c.Revenue, &c.AccountManagerID,
			&c.Notes, &c.CreatedBy, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, 0, errs.ErrInternal.Wrap(err)
		}
		out = append(out, c)
	}
	return out, total, rows.Err()
}

// SearchCustomers 按名称 / 邮箱 / 行业模糊搜索客户。
func (s *Service) SearchCustomers(ctx context.Context, wsID int64, q string, limit int) ([]Customer, error) {
	q = strings.TrimSpace(q)
	if q == "" {
		return []Customer{}, nil
	}
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	rows, err := s.db.Query(ctx, `
		SELECT id, workspace_id, name, COALESCE(email,''), COALESCE(website,''),
			COALESCE(industry,''), stage, tier, COALESCE(region,''), employees,
			COALESCE(revenue,''), account_manager_id, COALESCE(notes,''),
			created_by, created_at, updated_at
		FROM customers
		WHERE workspace_id=$1 AND deleted=false
		  AND (name ILIKE $2 OR email ILIKE $2 OR industry ILIKE $2)
		ORDER BY name ASC
		LIMIT $3`,
		wsID, "%"+q+"%", limit)
	if err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}
	defer rows.Close()

	out := make([]Customer, 0)
	for rows.Next() {
		var c Customer
		if err := rows.Scan(&c.ID, &c.WorkspaceID, &c.Name, &c.Email, &c.Website, &c.Industry,
			&c.Stage, &c.Tier, &c.Region, &c.Employees, &c.Revenue, &c.AccountManagerID,
			&c.Notes, &c.CreatedBy, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, errs.ErrInternal.Wrap(err)
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

// ============================================================
// 客户请求 CRUD
// ============================================================

// CreateCustomerRequest 创建客户请求。
func (s *Service) CreateCustomerRequest(ctx context.Context, in CreateCustomerRequestInput) (*CustomerRequest, error) {
	if in.WorkspaceID <= 0 {
		return nil, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "workspace_id", Reason: "工作空间 ID 无效"})
	}
	if in.CustomerID <= 0 {
		return nil, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "customer_id", Reason: "必须关联客户"})
	}
	title := strings.TrimSpace(in.Title)
	if title == "" {
		return nil, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "title", Reason: "请求标题不能为空"})
	}
	// 校验客户存在且属于该工作空间。
	var exists int
	if err := s.db.QueryRow(ctx,
		`SELECT COUNT(*) FROM customers WHERE id=$1 AND workspace_id=$2 AND deleted=false`,
		in.CustomerID, in.WorkspaceID).Scan(&exists); err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}
	if exists == 0 {
		return nil, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "customer_id", Reason: "客户不存在"})
	}

	status := RequestStatus(strings.TrimSpace(in.Status))
	if status == "" {
		status = RequestNew
	}
	if !status.IsValid() {
		return nil, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "status", Reason: "无效的请求状态"})
	}
	priority := RequestPriority(strings.TrimSpace(in.Priority))
	if priority == "" {
		priority = PriorityMedium
	}
	if !priority.IsValid() {
		return nil, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "priority", Reason: "无效的优先级"})
	}
	source := RequestSource(strings.TrimSpace(in.Source))
	if source == "" {
		source = SourceEmail
	}
	if !source.IsValid() {
		return nil, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "source", Reason: "无效的请求来源"})
	}
	if in.WorkItemID != nil && *in.WorkItemID > 0 {
		if err := s.validateWorkItem(ctx, in.WorkspaceID, *in.WorkItemID, in.WorkItemType); err != nil {
			return nil, err
		}
	}

	var r CustomerRequest
	err := s.db.QueryRow(ctx, `
		INSERT INTO customer_requests
			(id, workspace_id, customer_id, title, description, status, priority, source,
			 work_item_id, work_item_type, assigned_to)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
		RETURNING id, workspace_id, customer_id, title, COALESCE(description,''),
			status, priority, source, work_item_id, COALESCE(work_item_type,''),
			assigned_to, created_at, updated_at`,
		genID(), in.WorkspaceID, in.CustomerID, title, nullIfEmpty(in.Description),
		string(status), string(priority), string(source), in.WorkItemID,
		nullIfEmpty(in.WorkItemType), in.AssignedTo,
	).Scan(&r.ID, &r.WorkspaceID, &r.CustomerID, &r.Title, &r.Description,
		&r.Status, &r.Priority, &r.Source, &r.WorkItemID, &r.WorkItemType,
		&r.AssignedTo, &r.CreatedAt, &r.UpdatedAt)
	if err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}
	return &r, nil
}

// GetCustomerRequest 获取单个客户请求。
func (s *Service) GetCustomerRequest(ctx context.Context, wsID, requestID int64) (*CustomerRequest, error) {
	var r CustomerRequest
	err := s.db.QueryRow(ctx, `
		SELECT id, workspace_id, customer_id, title, COALESCE(description,''),
			status, priority, source, work_item_id, COALESCE(work_item_type,''),
			assigned_to, created_at, updated_at
		FROM customer_requests
		WHERE id=$1 AND workspace_id=$2 AND deleted=false`,
		requestID, wsID,
	).Scan(&r.ID, &r.WorkspaceID, &r.CustomerID, &r.Title, &r.Description,
		&r.Status, &r.Priority, &r.Source, &r.WorkItemID, &r.WorkItemType,
		&r.AssignedTo, &r.CreatedAt, &r.UpdatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, errs.ErrNotFound.WithCodeMessage("CUSTOMER_REQUEST.NOT_FOUND", "客户请求不存在")
		}
		return nil, errs.ErrInternal.Wrap(err)
	}
	return &r, nil
}

// UpdateCustomerRequest 更新客户请求。
func (s *Service) UpdateCustomerRequest(ctx context.Context, wsID, requestID int64, in UpdateCustomerRequestInput) (*CustomerRequest, error) {
	cur, err := s.GetCustomerRequest(ctx, wsID, requestID)
	if err != nil {
		return nil, err
	}

	if in.Title != nil {
		title := strings.TrimSpace(*in.Title)
		if title == "" {
			return nil, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "title", Reason: "请求标题不能为空"})
		}
		cur.Title = title
	}
	if in.Description != nil {
		cur.Description = strings.TrimSpace(*in.Description)
	}
	if in.Status != nil {
		status := RequestStatus(strings.TrimSpace(*in.Status))
		if !status.IsValid() {
			return nil, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "status", Reason: "无效的请求状态"})
		}
		cur.Status = status
	}
	if in.Priority != nil {
		priority := RequestPriority(strings.TrimSpace(*in.Priority))
		if !priority.IsValid() {
			return nil, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "priority", Reason: "无效的优先级"})
		}
		cur.Priority = priority
	}
	if in.Source != nil {
		source := RequestSource(strings.TrimSpace(*in.Source))
		if !source.IsValid() {
			return nil, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "source", Reason: "无效的请求来源"})
		}
		cur.Source = source
	}
	if in.CustomerID != nil {
		if *in.CustomerID <= 0 {
			return nil, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "customer_id", Reason: "客户 ID 无效"})
		}
		var exists int
		if err := s.db.QueryRow(ctx,
			`SELECT COUNT(*) FROM customers WHERE id=$1 AND workspace_id=$2 AND deleted=false`,
			*in.CustomerID, wsID).Scan(&exists); err != nil {
			return nil, errs.ErrInternal.Wrap(err)
		}
		if exists == 0 {
			return nil, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "customer_id", Reason: "客户不存在"})
		}
		cur.CustomerID = *in.CustomerID
	}
	if in.WorkItemID != nil {
		if *in.WorkItemID <= 0 {
			cur.WorkItemID = nil
			cur.WorkItemType = ""
		} else {
			// 优先使用请求中携带的类型，否则回退到当前已关联的类型（自动探测）。
			wit := cur.WorkItemType
			if in.WorkItemType != nil {
				wit = strings.TrimSpace(*in.WorkItemType)
			}
			if err := s.validateWorkItem(ctx, wsID, *in.WorkItemID, wit); err != nil {
				return nil, err
			}
			cur.WorkItemID = in.WorkItemID
			cur.WorkItemType = wit
		}
	} else if in.WorkItemType != nil {
		cur.WorkItemType = strings.TrimSpace(*in.WorkItemType)
	}
	if in.AssignedTo != nil {
		cur.AssignedTo = *in.AssignedTo
	}

	if _, err := s.db.Exec(ctx, `
		UPDATE customer_requests
		SET title=$1, description=$2, status=$3, priority=$4, source=$5, customer_id=$6,
		    work_item_id=$7, work_item_type=$8, assigned_to=$9, updated_at=now()
		WHERE id=$10 AND workspace_id=$11 AND deleted=false`,
		cur.Title, nullIfEmpty(cur.Description), string(cur.Status), string(cur.Priority),
		string(cur.Source), cur.CustomerID, cur.WorkItemID, nullIfEmpty(cur.WorkItemType),
		cur.AssignedTo, requestID, wsID); err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}
	return s.GetCustomerRequest(ctx, wsID, requestID)
}

// DeleteCustomerRequest 删除客户请求（软删除）。
func (s *Service) DeleteCustomerRequest(ctx context.Context, wsID, requestID int64) error {
	tag, err := s.db.Exec(ctx,
		`UPDATE customer_requests SET deleted=true, updated_at=now() WHERE id=$1 AND workspace_id=$2 AND deleted=false`,
		requestID, wsID)
	if err != nil {
		return errs.ErrInternal.Wrap(err)
	}
	if tag.RowsAffected() == 0 {
		return errs.ErrNotFound.WithCodeMessage("CUSTOMER_REQUEST.NOT_FOUND", "客户请求不存在")
	}
	return nil
}

// ListCustomerRequests 列出客户请求（支持 status / priority / customer 筛选）。
func (s *Service) ListCustomerRequests(ctx context.Context, f ListCustomerRequestsFilter) ([]CustomerRequest, int, error) {
	where := []string{"r.workspace_id=$1", "r.deleted=false"}
	args := []any{f.WorkspaceID}
	idx := 2
	if f.CustomerID != nil {
		where = append(where, "r.customer_id=$"+strconv.Itoa(idx))
		args = append(args, *f.CustomerID)
		idx++
	}
	if f.Status != "" {
		where = append(where, "r.status=$"+strconv.Itoa(idx))
		args = append(args, f.Status)
		idx++
	}
	if f.Priority != "" {
		where = append(where, "r.priority=$"+strconv.Itoa(idx))
		args = append(args, f.Priority)
		idx++
	}
	cond := strings.Join(where, " AND ")

	var total int
	if err := s.db.QueryRow(ctx, `SELECT COUNT(*) FROM customer_requests r WHERE `+cond, args...).Scan(&total); err != nil {
		return nil, 0, errs.ErrInternal.Wrap(err)
	}

	limit, offset := f.Limit, f.Offset
	if limit <= 0 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}
	args = append(args, limit, offset)

	rows, err := s.db.Query(ctx, `
		SELECT r.id, r.workspace_id, r.customer_id, r.title, COALESCE(r.description,''),
			r.status, r.priority, r.source, r.work_item_id, COALESCE(r.work_item_type,''),
			r.assigned_to, r.created_at, r.updated_at
		FROM customer_requests r
		WHERE `+cond+`
		ORDER BY r.created_at DESC
		LIMIT $`+strconv.Itoa(idx)+` OFFSET $`+strconv.Itoa(idx+1),
		args...)
	if err != nil {
		return nil, 0, errs.ErrInternal.Wrap(err)
	}
	defer rows.Close()

	out := make([]CustomerRequest, 0)
	for rows.Next() {
		var r CustomerRequest
		if err := rows.Scan(&r.ID, &r.WorkspaceID, &r.CustomerID, &r.Title, &r.Description,
			&r.Status, &r.Priority, &r.Source, &r.WorkItemID, &r.WorkItemType,
			&r.AssignedTo, &r.CreatedAt, &r.UpdatedAt); err != nil {
			return nil, 0, errs.ErrInternal.Wrap(err)
		}
		out = append(out, r)
	}
	return out, total, rows.Err()
}

// LinkRequestToWorkItem 将客户请求关联到正式工作项。
func (s *Service) LinkRequestToWorkItem(ctx context.Context, in LinkRequestInput) (*CustomerRequest, error) {
	if in.WorkspaceID <= 0 || in.RequestID <= 0 || in.WorkItemID <= 0 {
		return nil, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "work_item_id", Reason: "参数不完整"})
	}
	// 校验请求存在。
	if _, err := s.GetCustomerRequest(ctx, in.WorkspaceID, in.RequestID); err != nil {
		return nil, err
	}
	// 校验工作项存在。
	if err := s.validateWorkItem(ctx, in.WorkspaceID, in.WorkItemID, in.WorkItemType); err != nil {
		return nil, err
	}

	if _, err := s.db.Exec(ctx, `
		UPDATE customer_requests
		SET work_item_id=$1, work_item_type=$2, updated_at=now()
		WHERE id=$3 AND workspace_id=$4 AND deleted=false`,
		in.WorkItemID, nullIfEmpty(in.WorkItemType), in.RequestID, in.WorkspaceID); err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}
	return s.GetCustomerRequest(ctx, in.WorkspaceID, in.RequestID)
}

// GetCustomerStats 统计单个客户的请求数量（按状态分组）。
func (s *Service) GetCustomerStats(ctx context.Context, wsID, customerID int64) (*CustomerStats, error) {
	// 校验客户存在。
	if _, err := s.GetCustomer(ctx, wsID, customerID); err != nil {
		return nil, err
	}
	rows, err := s.db.Query(ctx, `
		SELECT status, COUNT(*) FROM customer_requests
		WHERE customer_id=$1 AND workspace_id=$2 AND deleted=false
		GROUP BY status`, customerID, wsID)
	if err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}
	defer rows.Close()

	stats := &CustomerStats{
		CustomerID: customerID,
		ByStatus:   map[string]int{},
	}
	for rows.Next() {
		var status string
		var count int
		if err := rows.Scan(&status, &count); err != nil {
			return nil, errs.ErrInternal.Wrap(err)
		}
		stats.ByStatus[status] = count
		stats.Total += count
	}
	return stats, rows.Err()
}

// ============================================================
// Helpers
// ============================================================

// validateWorkItem 校验工作项是否存在且属于该工作空间。
// workItemType 为空时自动探测类型。
func (s *Service) validateWorkItem(ctx context.Context, wsID, workItemID int64, workItemType string) error {
	typeCode := strings.TrimSpace(workItemType)
	if typeCode == "" {
		// 自动探测类型。
		for _, t := range []string{"requirement", "task", "defect"} {
			if s.workItemExists(ctx, wsID, workItemID, t) {
				return nil
			}
		}
		return errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "work_item_id", Reason: "工作项不存在"})
	}
	if !validWorkItemType(typeCode) {
		return errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "work_item_type", Reason: "无效的工作项类型"})
	}
	if !s.workItemExists(ctx, wsID, workItemID, typeCode) {
		return errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "work_item_id", Reason: "工作项不存在"})
	}
	return nil
}

// workItemExists 检查指定类型的工作项是否存在于该工作空间。
func (s *Service) workItemExists(ctx context.Context, wsID, workItemID int64, typeCode string) bool {
	var exists bool
	var err error
	switch typeCode {
	case "requirement":
		err = s.db.QueryRow(ctx,
			`SELECT EXISTS(SELECT 1 FROM requirement WHERE id=$1 AND workspace_id=$2 AND deleted=false)`,
			workItemID, wsID).Scan(&exists)
	case "task":
		err = s.db.QueryRow(ctx,
			`SELECT EXISTS(SELECT 1 FROM task WHERE id=$1 AND workspace_id=$2 AND deleted=false)`,
			workItemID, wsID).Scan(&exists)
	case "defect":
		err = s.db.QueryRow(ctx,
			`SELECT EXISTS(SELECT 1 FROM defect WHERE id=$1 AND workspace_id=$2 AND deleted=false)`,
			workItemID, wsID).Scan(&exists)
	default:
		return false
	}
	if err != nil {
		return false
	}
	return exists
}

// validWorkItemType 校验工作项类型。
func validWorkItemType(t string) bool {
	switch t {
	case "requirement", "task", "defect":
		return true
	}
	return false
}

func nullIfEmpty(s string) any {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	return s
}
