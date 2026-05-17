package repository

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"

	"task-backend/internal/model"
)

type TaskRepository interface {
	Create(ctx context.Context, task *model.Task) error
	ListPersonalAndAssignedTeam(ctx context.Context, userID uint64, page, size int) ([]model.Task, int64, error)
	ListByTeamID(ctx context.Context, teamID uint64, page, size int) ([]model.Task, int64, error)
	ListByUserID(ctx context.Context, userID uint64, page, size int) ([]model.Task, int64, error)
	ListAll(ctx context.Context, page, size int) ([]model.Task, int64, error)
	// 新增：支持排序的两个方法
	ListByUserIDWithOrder(ctx context.Context, userID uint64, page, size int, orderBy string) ([]model.Task, int64, error)
	ListAllWithOrder(ctx context.Context, page, size int, orderBy string) ([]model.Task, int64, error)
	//新增：支持搜索
	ListByUserIDWithSearch(ctx context.Context, userID uint64, page, size int, search, orderBy string) ([]model.Task, int64, error)
	ListAllWithSearch(ctx context.Context, page, size int, search, orderBy string) ([]model.Task, int64, error)
	GetByIDAndUserID(ctx context.Context, id, userID uint64) (*model.Task, error)
	GetByID(ctx context.Context, id uint64) (*model.Task, error)
	UpdateByIDAndUserID(ctx context.Context, task *model.Task) error
	UpdateByID(ctx context.Context, task *model.Task) error
	DeleteByIDAndUserID(ctx context.Context, id, userID uint64) error
	DeleteByID(ctx context.Context, id uint64) error
	ClearAssigneeForTeamTasks(ctx context.Context, teamID, assigneeID uint64) error
	AddDependency(ctx context.Context, predecessorID, successorID uint64) (*model.TaskDependency, error)
	RemoveDependency(ctx context.Context, predecessorID, successorID uint64) error
	ListPredecessors(ctx context.Context, taskID uint64) ([]model.TaskDependencyNode, error)
	ListSuccessors(ctx context.Context, taskID uint64) ([]model.TaskDependencyNode, error)
	CountIncompletePredecessors(ctx context.Context, taskID uint64) (int64, error)
	HasDependencyPath(ctx context.Context, fromID, toID uint64) (bool, error)
}

type taskRepository struct {
	db *gorm.DB
}

func NewTaskRepository(db *gorm.DB) TaskRepository {
	return &taskRepository{db: db}
}

func (r *taskRepository) Create(ctx context.Context, task *model.Task) error {
	return r.db.WithContext(ctx).Create(task).Error
}

func (r *taskRepository) ListPersonalAndAssignedTeam(ctx context.Context, userID uint64, page, size int) ([]model.Task, int64, error) {
	if page <= 0 {
		page = 1
	}
	if size <= 0 {
		size = 10
	}
	if size > 100 {
		size = 100
	}
	offset := (page - 1) * size

	query := r.db.WithContext(ctx).
		Model(&model.Task{}).
		Where("deleted_at IS NULL").
		Where("(team_id IS NULL AND user_id = ?) OR (team_id IS NOT NULL AND assignee_id = ?)", userID, userID)

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	tasks := make([]model.Task, 0)
	err := query.
		Order("created_at DESC").
		Limit(size).
		Offset(offset).
		Find(&tasks).Error
	if err != nil {
		return nil, 0, err
	}
	return tasks, total, nil
}

func (r *taskRepository) ListByTeamID(ctx context.Context, teamID uint64, page, size int) ([]model.Task, int64, error) {
	if page <= 0 {
		page = 1
	}
	if size <= 0 {
		size = 10
	}
	if size > 100 {
		size = 100
	}
	offset := (page - 1) * size

	query := r.db.WithContext(ctx).
		Model(&model.Task{}).
		Where("team_id = ?", teamID).
		Where("deleted_at IS NULL")

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	tasks := make([]model.Task, 0)
	err := query.
		Order("priority DESC").
		Order("created_at DESC").
		Limit(size).
		Offset(offset).
		Find(&tasks).Error
	if err != nil {
		return nil, 0, err
	}
	return tasks, total, nil
}

func (r *taskRepository) ListByUserID(ctx context.Context, userID uint64, page, size int) ([]model.Task, int64, error) {
	if page <= 0 {
		page = 1
	}
	if size <= 0 {
		size = 10
	}
	if size > 100 {
		size = 100
	}
	offset := (page - 1) * size

	query := r.db.WithContext(ctx).
		Model(&model.Task{}).
		Where("user_id = ?", userID).
		Where("deleted_at IS NULL")

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var tasks []model.Task
	err := query.
		Order("created_at DESC").
		Limit(size).
		Offset(offset).
		Find(&tasks).Error
	if err != nil {
		return nil, 0, err
	}
	return tasks, total, nil
}

func (r *taskRepository) ListAll(ctx context.Context, page, size int) ([]model.Task, int64, error) {
	if page <= 0 {
		page = 1
	}
	if size <= 0 {
		size = 10
	}
	if size > 100 {
		size = 100
	}
	offset := (page - 1) * size

	query := r.db.WithContext(ctx).
		Model(&model.Task{}).
		Where("deleted_at IS NULL")

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var tasks []model.Task
	err := query.
		Order("created_at DESC").
		Limit(size).
		Offset(offset).
		Find(&tasks).Error
	if err != nil {
		return nil, 0, err
	}
	return tasks, total, nil
}

// ==================== 新增支持排序的方法（核心） ====================

func (r *taskRepository) ListByUserIDWithOrder(ctx context.Context, userID uint64, page, size int, orderBy string) ([]model.Task, int64, error) {
	if page <= 0 {
		page = 1
	}
	if size <= 0 {
		size = 10
	}
	if size > 100 {
		size = 100
	}

	offset := (page - 1) * size

	// 安全校验排序字段，防止 SQL 注入
	orderBy = sanitizeOrderBy(orderBy)

	query := r.db.WithContext(ctx).
		Model(&model.Task{}).
		Where("user_id = ?", userID).
		Where("deleted_at IS NULL")

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var tasks []model.Task
	err := query.
		Order(orderBy). // PostgreSQL 支持直接传入 "priority DESC"
		Limit(size).
		Offset(offset).
		Find(&tasks).Error

	if err != nil {
		return nil, 0, err
	}
	return tasks, total, nil
}

func (r *taskRepository) ListAllWithOrder(ctx context.Context, page, size int, orderBy string) ([]model.Task, int64, error) {
	if page <= 0 {
		page = 1
	}
	if size <= 0 {
		size = 10
	}
	if size > 100 {
		size = 100
	}

	offset := (page - 1) * size

	orderBy = sanitizeOrderBy(orderBy)

	query := r.db.WithContext(ctx).
		Model(&model.Task{}).
		Where("deleted_at IS NULL")

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var tasks []model.Task
	err := query.
		Order(orderBy).
		Limit(size).
		Offset(offset).
		Find(&tasks).Error

	if err != nil {
		return nil, 0, err
	}
	return tasks, total, nil
}

// ==================== 工具函数：排序安全校验 ====================

// 只允许指定的字段排序，防止 SQL 注入
func sanitizeOrderBy(orderBy string) string {
	// 默认排序
	if orderBy == "" {
		return "created_at DESC"
	}

	// 支持的排序字段 + 方向
	allowed := map[string]bool{
		"created_at DESC": true,
		"created_at ASC":  true,
		"deadline DESC":   true,
		"deadline ASC":    true,
		"priority DESC":   true,
		"priority ASC":    true,
	}

	// 如果传入的就是完整 "字段 方向"，直接校验
	if allowed[orderBy] {
		return orderBy
	}

	// 支持前端只传字段名的情况（默认 DESC）
	switch strings.ToLower(strings.TrimSpace(orderBy)) {
	case "created_at":
		return "created_at DESC"
	case "deadline":
		return "deadline DESC"
	case "priority":
		return "priority DESC"
	default:
		return "created_at DESC" // 默认
	}
}

// ========== 新增：带搜索的查询方法 ==========

func (r *taskRepository) ListByUserIDWithSearch(ctx context.Context, userID uint64, page, size int, search, orderBy string) ([]model.Task, int64, error) {
	if page <= 0 {
		page = 1
	}
	if size <= 0 {
		size = 10
	}
	if size > 100 {
		size = 100
	}

	offset := (page - 1) * size
	orderBy = sanitizeOrderBy(orderBy)

	query := r.db.WithContext(ctx).
		Model(&model.Task{}).
		Where("user_id = ?", userID).
		Where("deleted_at IS NULL")

	// 如果有搜索关键词，则对 title 和 description 进行模糊搜索
	if search != "" {
		likePattern := "%" + search + "%"
		query = query.Where("title ILIKE ? OR description ILIKE ?", likePattern, likePattern)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var tasks []model.Task
	err := query.
		Order(orderBy).
		Limit(size).
		Offset(offset).
		Find(&tasks).Error

	return tasks, total, err
}

func (r *taskRepository) ListAllWithSearch(ctx context.Context, page, size int, search, orderBy string) ([]model.Task, int64, error) {
	if page <= 0 {
		page = 1
	}
	if size <= 0 {
		size = 10
	}
	if size > 100 {
		size = 100
	}

	offset := (page - 1) * size
	orderBy = sanitizeOrderBy(orderBy)

	query := r.db.WithContext(ctx).
		Model(&model.Task{}).
		Where("deleted_at IS NULL")

	if search != "" {
		likePattern := "%" + search + "%"
		query = query.Where("title ILIKE ? OR description ILIKE ?", likePattern, likePattern)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var tasks []model.Task
	err := query.
		Order(orderBy).
		Limit(size).
		Offset(offset).
		Find(&tasks).Error

	return tasks, total, err
}

func (r *taskRepository) GetByIDAndUserID(ctx context.Context, id, userID uint64) (*model.Task, error) {
	var task model.Task
	err := r.db.WithContext(ctx).
		Where("id = ?", id).
		Where("user_id = ?", userID).
		Where("deleted_at IS NULL").
		First(&task).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &task, nil
}

func (r *taskRepository) GetByID(ctx context.Context, id uint64) (*model.Task, error) {
	var task model.Task
	err := r.db.WithContext(ctx).
		Where("id = ?", id).
		Where("deleted_at IS NULL").
		First(&task).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &task, nil
}

func (r *taskRepository) UpdateByIDAndUserID(ctx context.Context, task *model.Task) error {
	updates := map[string]interface{}{
		"title":       task.Title,
		"description": task.Description,
		"status":      task.Status,
		"priority":    task.Priority,
		"deadline":    task.Deadline,
		"team_id":     task.TeamID,
		"assignee_id": task.AssigneeID,
		"updated_at":  time.Now(),
	}

	res := r.db.WithContext(ctx).
		Model(&model.Task{}).
		Where("id = ?", task.ID).
		Where("user_id = ?", task.UserID).
		Where("deleted_at IS NULL").
		Updates(updates)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *taskRepository) UpdateByID(ctx context.Context, task *model.Task) error {
	updates := map[string]interface{}{
		"title":       task.Title,
		"description": task.Description,
		"status":      task.Status,
		"priority":    task.Priority,
		"deadline":    task.Deadline,
		"team_id":     task.TeamID,
		"assignee_id": task.AssigneeID,
		"updated_at":  time.Now(),
	}

	res := r.db.WithContext(ctx).
		Model(&model.Task{}).
		Where("id = ?", task.ID).
		Where("deleted_at IS NULL").
		Updates(updates)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *taskRepository) DeleteByIDAndUserID(ctx context.Context, id, userID uint64) error {
	res := r.db.WithContext(ctx).
		Model(&model.Task{}).
		Where("id = ?", id).
		Where("user_id = ?", userID).
		Where("deleted_at IS NULL").
		Updates(map[string]interface{}{
			"deleted_at": time.Now(),
			"updated_at": time.Now(),
		})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *taskRepository) DeleteByID(ctx context.Context, id uint64) error {
	res := r.db.WithContext(ctx).
		Model(&model.Task{}).
		Where("id = ?", id).
		Where("deleted_at IS NULL").
		Updates(map[string]interface{}{
			"deleted_at": time.Now(),
			"updated_at": time.Now(),
		})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *taskRepository) ClearAssigneeForTeamTasks(ctx context.Context, teamID, assigneeID uint64) error {
	res := r.db.WithContext(ctx).
		Model(&model.Task{}).
		Where("team_id = ?", teamID).
		Where("assignee_id = ?", assigneeID).
		Where("deleted_at IS NULL").
		Where("status IN ?", []model.TaskStatus{model.TaskStatusTODO, model.TaskStatusDOING}).
		Updates(map[string]interface{}{
			"assignee_id": nil,
			"updated_at":  time.Now(),
		})
	if res.Error != nil {
		return res.Error
	}
	return nil
}

func (r *taskRepository) AddDependency(ctx context.Context, predecessorID, successorID uint64) (*model.TaskDependency, error) {
	dep := &model.TaskDependency{
		PredecessorID: predecessorID,
		SuccessorID:   successorID,
	}
	err := r.db.WithContext(ctx).Create(dep).Error
	if err == nil {
		return dep, nil
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return nil, ErrConflict
	}
	return nil, err
}

func (r *taskRepository) RemoveDependency(ctx context.Context, predecessorID, successorID uint64) error {
	res := r.db.WithContext(ctx).
		Where("predecessor_id = ?", predecessorID).
		Where("successor_id = ?", successorID).
		Delete(&model.TaskDependency{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *taskRepository) ListPredecessors(ctx context.Context, taskID uint64) ([]model.TaskDependencyNode, error) {
	var list []model.TaskDependencyNode
	err := r.db.WithContext(ctx).
		Table("task_dependencies td").
		Select("t.id, t.title, t.status").
		Joins("JOIN tasks t ON t.id = td.predecessor_id AND t.deleted_at IS NULL").
		Where("td.successor_id = ?", taskID).
		Order("t.created_at ASC").
		Find(&list).Error
	if err != nil {
		return nil, err
	}
	if list == nil {
		list = make([]model.TaskDependencyNode, 0)
	}
	return list, nil
}

func (r *taskRepository) ListSuccessors(ctx context.Context, taskID uint64) ([]model.TaskDependencyNode, error) {
	var list []model.TaskDependencyNode
	err := r.db.WithContext(ctx).
		Table("task_dependencies td").
		Select("t.id, t.title, t.status").
		Joins("JOIN tasks t ON t.id = td.successor_id AND t.deleted_at IS NULL").
		Where("td.predecessor_id = ?", taskID).
		Order("t.created_at ASC").
		Find(&list).Error
	if err != nil {
		return nil, err
	}
	if list == nil {
		list = make([]model.TaskDependencyNode, 0)
	}
	return list, nil
}

func (r *taskRepository) CountIncompletePredecessors(ctx context.Context, taskID uint64) (int64, error) {
	var total int64
	err := r.db.WithContext(ctx).
		Table("task_dependencies td").
		Joins("JOIN tasks t ON t.id = td.predecessor_id AND t.deleted_at IS NULL").
		Where("td.successor_id = ?", taskID).
		Where("t.status <> ?", model.TaskStatusDONE).
		Count(&total).Error
	if err != nil {
		return 0, err
	}
	return total, nil
}

func (r *taskRepository) HasDependencyPath(ctx context.Context, fromID, toID uint64) (bool, error) {
	if fromID == 0 || toID == 0 {
		return false, nil
	}
	const query = `
WITH RECURSIVE dep(successor_id) AS (
  SELECT successor_id FROM task_dependencies WHERE predecessor_id = ?
  UNION
  SELECT td.successor_id FROM task_dependencies td
  JOIN dep ON td.predecessor_id = dep.successor_id
)
SELECT 1 AS found FROM dep WHERE successor_id = ? LIMIT 1`
	var found int
	if err := r.db.WithContext(ctx).Raw(query, fromID, toID).Scan(&found).Error; err != nil {
		return false, err
	}
	return found == 1, nil
}

func (r *taskRepository) SoftDeleteByTeamID(ctx context.Context, teamID uint64) error {
	res := r.db.WithContext(ctx).
		Model(&model.Task{}).
		Where("team_id = ?", teamID).
		Where("deleted_at IS NULL").
		Updates(map[string]interface{}{
			"deleted_at": time.Now(),
			"updated_at": time.Now(),
		})
	if res.Error != nil {
		return res.Error
	}
	return nil
}
