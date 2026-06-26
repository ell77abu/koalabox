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
	ListVisible(ctx context.Context, query TaskListQuery) ([]model.Task, int64, error)
	ListPersonalAndAssignedTeam(ctx context.Context, userID uint64, page, size int) ([]model.Task, int64, error)
	ListByTeamID(ctx context.Context, teamID uint64, page, size int) ([]model.Task, int64, error)
	ListByUserID(ctx context.Context, userID uint64, page, size int) ([]model.Task, int64, error)
	ListAll(ctx context.Context, page, size int) ([]model.Task, int64, error)
	ListPersonalTrashByUserID(ctx context.Context, userID uint64, page, size int) ([]model.Task, int64, error)
	ListTeamTrashByTeamID(ctx context.Context, teamID uint64, page, size int) ([]model.Task, int64, error)
	GetByIDAndUserID(ctx context.Context, id, userID uint64) (*model.Task, error)
	GetByIDForUserScope(ctx context.Context, id, userID uint64) (*model.Task, error)
	GetByID(ctx context.Context, id uint64) (*model.Task, error)
	GetByIDIncludingDeleted(ctx context.Context, id uint64) (*model.Task, error)
	UpdateByIDAndUserID(ctx context.Context, task *model.Task) error
	UpdateByID(ctx context.Context, task *model.Task) error
	UpdateByIDForOwnerOrTeamManagerScope(ctx context.Context, task *model.Task, userID uint64) error
	UpdateStatusByIDForAssignedMemberScope(ctx context.Context, taskID, userID uint64, status model.TaskStatus) error
	DeleteByIDAndUserID(ctx context.Context, id, userID uint64) error
	DeleteByIDForOwnerOrTeamManagerScope(ctx context.Context, id, userID uint64) error
	DeleteByID(ctx context.Context, id uint64) error
	ClearAssigneeForTeamTasks(ctx context.Context, teamID, assigneeID uint64) error
	AddDependency(ctx context.Context, predecessorID, successorID uint64) (*model.TaskDependency, error)
	AddDependencyWithCycleCheck(ctx context.Context, predecessorID, successorID uint64) (*model.TaskDependency, error)
	RemoveDependency(ctx context.Context, predecessorID, successorID uint64) error
	ListPredecessors(ctx context.Context, taskID uint64) ([]model.TaskDependencyNode, error)
	ListSuccessors(ctx context.Context, taskID uint64) ([]model.TaskDependencyNode, error)
	CountIncompletePredecessors(ctx context.Context, taskID uint64) (int64, error)
	ListIncompletePredecessorIDs(ctx context.Context, taskID uint64) ([]uint64, error)
	ListIncompletePredecessors(ctx context.Context, taskID uint64) ([]model.TaskDependencyNode, error)
	ListDoneSuccessors(ctx context.Context, taskID uint64) ([]model.TaskDependencyNode, error)
	HasDependencyPath(ctx context.Context, fromID, toID uint64) (bool, error)
	GetDependencyGraph(ctx context.Context, taskID uint64) ([]model.TaskDependencyNode, []model.TaskDependencyLink, error)
}

type TaskListQuery struct {
	UserID    uint64
	UserLevel uint8
	TeamID    *uint64
	Scope     string
	Status    string
	Search    string
	Sort      string
	Page      int
	Size      int
}

type taskRepository struct {
	db *gorm.DB
}

func NewTaskRepository(db *gorm.DB) TaskRepository {
	return &taskRepository{db: db}
}

func (r *taskRepository) GetDB() *gorm.DB {
	return r.db
}

func (r *taskRepository) Create(ctx context.Context, task *model.Task) error {
	return r.db.WithContext(ctx).Create(task).Error
}

func (r *taskRepository) ListVisible(ctx context.Context, query TaskListQuery) ([]model.Task, int64, error) {
	if query.Page <= 0 {
		query.Page = 1
	}
	if query.Size <= 0 {
		query.Size = 10
	}
	if query.Size > 100 {
		query.Size = 100
	}
	offset := (query.Page - 1) * query.Size

	baseQuery := r.db.WithContext(ctx).
		Table("tasks").
		Where("tasks.deleted_at IS NULL")

	baseQuery = applyTaskVisibility(baseQuery, query)
	baseQuery = applyTaskStatusFilter(baseQuery, query.Status)

	if keyword := strings.TrimSpace(query.Search); keyword != "" {
		baseQuery = baseQuery.Where("tasks.title ILIKE ?", "%"+keyword+"%")
	}

	var total int64
	if err := baseQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	tasks := make([]model.Task, 0)
	err := baseQuery.
		Select("tasks.*").
		Order(taskListOrderBy(query.Sort)).
		Limit(query.Size).
		Offset(offset).
		Find(&tasks).Error
	if err != nil {
		return nil, 0, err
	}
	return tasks, total, nil
}

func applyTaskVisibility(db *gorm.DB, query TaskListQuery) *gorm.DB {
	if query.TeamID != nil {
		return db.Where("tasks.team_id = ?", *query.TeamID)
	}

	scope := strings.ToLower(strings.TrimSpace(query.Scope))
	if query.UserLevel == 1 {
		switch scope {
		case "personal":
			return db.Where("tasks.team_id IS NULL")
		case "team":
			return db.Where("tasks.team_id IS NOT NULL")
		default:
			return db
		}
	}

	switch scope {
	case "personal":
		return db.Where("tasks.team_id IS NULL AND tasks.user_id = ?", query.UserID)
	case "team":
		return db.Where(
			`tasks.team_id IS NOT NULL AND tasks.assignee_id = ? AND EXISTS (
				SELECT 1
				FROM team_members tm
				JOIN teams t ON t.id = tm.team_id AND t.deleted_at IS NULL
				WHERE tm.team_id = tasks.team_id
					AND tm.user_id = ?
					AND tm.deleted_at IS NULL
			)`,
			query.UserID, query.UserID,
		)
	default:
		return db.Where(
			`(tasks.team_id IS NULL AND tasks.user_id = ?) OR
			(tasks.team_id IS NOT NULL AND tasks.assignee_id = ? AND EXISTS (
				SELECT 1
				FROM team_members tm
				JOIN teams t ON t.id = tm.team_id AND t.deleted_at IS NULL
				WHERE tm.team_id = tasks.team_id
					AND tm.user_id = ?
					AND tm.deleted_at IS NULL
			))`,
			query.UserID, query.UserID, query.UserID,
		)
	}
}

func applyTaskStatusFilter(db *gorm.DB, status string) *gorm.DB {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case "active":
		return db.Where("tasks.status IN ?", []model.TaskStatus{
			model.TaskStatusTODO,
			model.TaskStatusDOING,
			model.TaskStatusERROR,
		})
	case "done":
		return db.Where("tasks.status = ?", model.TaskStatusDONE)
	default:
		return db
	}
}

func taskListOrderBy(sort string) string {
	switch strings.ToLower(strings.TrimSpace(sort)) {
	case "priority":
		return "tasks.priority DESC, tasks.created_at DESC"
	case "deadline":
		return "tasks.deadline ASC NULLS LAST, tasks.created_at DESC"
	case "created_at", "default":
		return "tasks.created_at DESC"
	default:
		return "tasks.created_at DESC"
	}
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

	baseQuery := r.db.WithContext(ctx).
		Table("tasks").
		Where("tasks.deleted_at IS NULL").
		Where(
			`(tasks.team_id IS NULL AND tasks.user_id = ?) OR
			(tasks.team_id IS NOT NULL AND tasks.assignee_id = ? AND EXISTS (
				SELECT 1
				FROM team_members tm
				JOIN teams t ON t.id = tm.team_id AND t.deleted_at IS NULL
				WHERE tm.team_id = tasks.team_id
					AND tm.user_id = ?
					AND tm.deleted_at IS NULL
			))`,
			userID, userID, userID,
		)

	var total int64
	if err := baseQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	tasks := make([]model.Task, 0)
	err := baseQuery.
		Select("tasks.*").
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

func (r *taskRepository) ListPersonalTrashByUserID(ctx context.Context, userID uint64, page, size int) ([]model.Task, int64, error) {
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
		Where("team_id IS NULL").
		Where("deleted_at IS NOT NULL")

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	tasks := make([]model.Task, 0)
	err := query.
		Order("deleted_at DESC").
		Limit(size).
		Offset(offset).
		Find(&tasks).Error
	if err != nil {
		return nil, 0, err
	}
	return tasks, total, nil
}

func (r *taskRepository) ListTeamTrashByTeamID(ctx context.Context, teamID uint64, page, size int) ([]model.Task, int64, error) {
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
		Where("deleted_at IS NOT NULL")

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	tasks := make([]model.Task, 0)
	err := query.
		Order("deleted_at DESC").
		Limit(size).
		Offset(offset).
		Find(&tasks).Error
	if err != nil {
		return nil, 0, err
	}
	return tasks, total, nil
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

func (r *taskRepository) GetByIDForUserScope(ctx context.Context, id, userID uint64) (*model.Task, error) {
	var task model.Task
	err := r.db.WithContext(ctx).
		Table("tasks").
		Select("tasks.*").
		Where("tasks.id = ?", id).
		Where("tasks.deleted_at IS NULL").
		Where(
			`(tasks.team_id IS NULL AND tasks.user_id = ?) OR
			(tasks.team_id IS NOT NULL AND EXISTS (
				SELECT 1
				FROM team_members tm
				JOIN teams t ON t.id = tm.team_id AND t.deleted_at IS NULL
				WHERE tm.team_id = tasks.team_id
					AND tm.user_id = ?
					AND tm.deleted_at IS NULL
			))`,
			userID, userID,
		).
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

func (r *taskRepository) GetByIDIncludingDeleted(ctx context.Context, id uint64) (*model.Task, error) {
	var task model.Task
	err := r.db.WithContext(ctx).
		Where("id = ?", id).
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
	res := r.db.WithContext(ctx).
		Model(&model.Task{}).
		Where("id = ?", task.ID).
		Where("deleted_at IS NULL").
		Updates(taskUpdateMap(task))
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func taskUpdateMap(task *model.Task) map[string]interface{} {
	return map[string]interface{}{
		"title":       task.Title,
		"description": task.Description,
		"status":      task.Status,
		"priority":    task.Priority,
		"deadline":    task.Deadline,
		"team_id":     task.TeamID,
		"assignee_id": task.AssigneeID,
		"updated_at":  time.Now(),
	}
}

func activeTeamMemberScopeSQL() string {
	return `EXISTS (
		SELECT 1
		FROM team_members tm
		JOIN teams teams_scope ON teams_scope.id = tm.team_id AND teams_scope.deleted_at IS NULL
		WHERE tm.team_id = tasks.team_id
			AND tm.user_id = ?
			AND tm.deleted_at IS NULL
	)`
}

func ownerOrTeamManagerScopeSQL() string {
	return `(team_id IS NULL AND user_id = ?) OR
		(team_id IS NOT NULL AND EXISTS (
			SELECT 1
			FROM team_members tm
			JOIN teams teams_scope ON teams_scope.id = tm.team_id AND teams_scope.deleted_at IS NULL
			WHERE tm.team_id = tasks.team_id
				AND tm.user_id = ?
				AND tm.role IN (?, ?)
				AND tm.deleted_at IS NULL
		))`
}

func (r *taskRepository) UpdateByIDForOwnerOrTeamManagerScope(ctx context.Context, task *model.Task, userID uint64) error {
	res := r.db.WithContext(ctx).
		Model(&model.Task{}).
		Where("id = ?", task.ID).
		Where("deleted_at IS NULL").
		Where(ownerOrTeamManagerScopeSQL(), userID, userID, model.TeamRoleOwner, model.TeamRoleAdmin).
		Updates(taskUpdateMap(task))
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *taskRepository) UpdateStatusByIDForAssignedMemberScope(ctx context.Context, taskID, userID uint64, status model.TaskStatus) error {
	res := r.db.WithContext(ctx).
		Model(&model.Task{}).
		Where("id = ?", taskID).
		Where("deleted_at IS NULL").
		Where("team_id IS NOT NULL").
		Where("assignee_id = ?", userID).
		Where(activeTeamMemberScopeSQL(), userID).
		Updates(map[string]interface{}{
			"status":     status,
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

func (r *taskRepository) DeleteByIDAndUserID(ctx context.Context, id, userID uint64) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.
			Where("predecessor_id = ? OR successor_id = ?", id, id).
			Delete(&model.TaskDependency{}).Error; err != nil {
			return err
		}

		res := tx.
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
	})
}

func (r *taskRepository) DeleteByIDForOwnerOrTeamManagerScope(ctx context.Context, id, userID uint64) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		res := tx.
			Model(&model.Task{}).
			Where("id = ?", id).
			Where("deleted_at IS NULL").
			Where(ownerOrTeamManagerScopeSQL(), userID, userID, model.TeamRoleOwner, model.TeamRoleAdmin).
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

		return tx.
			Where("predecessor_id = ? OR successor_id = ?", id, id).
			Delete(&model.TaskDependency{}).Error
	})
}

func (r *taskRepository) DeleteByID(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.
			Where("predecessor_id = ? OR successor_id = ?", id, id).
			Delete(&model.TaskDependency{}).Error; err != nil {
			return err
		}

		res := tx.
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
	})
}

func (r *taskRepository) ClearAssigneeForTeamTasks(ctx context.Context, teamID, assigneeID uint64) error {
	return r.ClearAssigneeForOpenTeamTasks(ctx, teamID, assigneeID)
}

func (r *taskRepository) ListOpenAssignedTeamTasks(ctx context.Context, teamID, assigneeID uint64) ([]model.Task, error) {
	tasks := make([]model.Task, 0)
	err := r.db.WithContext(ctx).
		Model(&model.Task{}).
		Where("team_id = ?", teamID).
		Where("assignee_id = ?", assigneeID).
		Where("deleted_at IS NULL").
		Where("status IN ?", []model.TaskStatus{model.TaskStatusTODO, model.TaskStatusDOING}).
		Order("id ASC").
		Find(&tasks).Error
	if err != nil {
		return nil, err
	}
	return tasks, nil
}

func (r *taskRepository) ClearAssigneeForOpenTeamTasks(ctx context.Context, teamID, assigneeID uint64) error {
	res := r.db.WithContext(ctx).
		Model(&model.Task{}).
		Where("team_id = ?", teamID).
		Where("assignee_id = ?", assigneeID).
		Where("deleted_at IS NULL").
		Where("status = ?", model.TaskStatusTODO).
		Updates(map[string]interface{}{
			"assignee_id": nil,
			"updated_at":  time.Now(),
		})
	if res.Error != nil {
		return res.Error
	}

	res = r.db.WithContext(ctx).
		Model(&model.Task{}).
		Where("team_id = ?", teamID).
		Where("assignee_id = ?", assigneeID).
		Where("deleted_at IS NULL").
		Where("status = ?", model.TaskStatusDOING).
		Updates(map[string]interface{}{
			"assignee_id": nil,
			"status":      model.TaskStatusTODO,
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

func (r *taskRepository) AddDependencyWithCycleCheck(ctx context.Context, predecessorID, successorID uint64) (*model.TaskDependency, error) {
	var dep *model.TaskDependency
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := lockTaskIDs(tx, predecessorID, successorID); err != nil {
			return err
		}

		hasCycle, err := hasDependencyPathTx(tx, successorID, predecessorID)
		if err != nil {
			return err
		}
		if hasCycle {
			return ErrDependencyCycle
		}

		dep = &model.TaskDependency{
			PredecessorID: predecessorID,
			SuccessorID:   successorID,
		}
		if err := tx.Create(dep).Error; err != nil {
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) && pgErr.Code == "23505" {
				return ErrConflict
			}
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return dep, nil
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

func (r *taskRepository) ListIncompletePredecessorIDs(ctx context.Context, taskID uint64) ([]uint64, error) {
	var ids []uint64
	err := r.db.WithContext(ctx).
		Table("task_dependencies td").
		Select("t.id").
		Joins("JOIN tasks t ON t.id = td.predecessor_id AND t.deleted_at IS NULL").
		Where("td.successor_id = ?", taskID).
		Where("t.status <> ?", model.TaskStatusDONE).
		Order("t.id ASC").
		Find(&ids).Error
	if err != nil {
		return nil, err
	}
	if ids == nil {
		ids = make([]uint64, 0)
	}
	return ids, nil
}

func (r *taskRepository) ListIncompletePredecessors(ctx context.Context, taskID uint64) ([]model.TaskDependencyNode, error) {
	var list []model.TaskDependencyNode
	err := r.db.WithContext(ctx).
		Table("task_dependencies td").
		Select("t.id, t.title, t.status").
		Joins("JOIN tasks t ON t.id = td.predecessor_id AND t.deleted_at IS NULL").
		Where("td.successor_id = ?", taskID).
		Where("t.status <> ?", model.TaskStatusDONE).
		Order("t.id ASC").
		Find(&list).Error
	if err != nil {
		return nil, err
	}
	if list == nil {
		list = make([]model.TaskDependencyNode, 0)
	}
	return list, nil
}

func (r *taskRepository) ListDoneSuccessors(ctx context.Context, taskID uint64) ([]model.TaskDependencyNode, error) {
	var list []model.TaskDependencyNode
	err := r.db.WithContext(ctx).
		Table("task_dependencies td").
		Select("t.id, t.title, t.status").
		Joins("JOIN tasks t ON t.id = td.successor_id AND t.deleted_at IS NULL").
		Where("td.predecessor_id = ?", taskID).
		Where("t.status = ?", model.TaskStatusDONE).
		Order("t.id ASC").
		Find(&list).Error
	if err != nil {
		return nil, err
	}
	if list == nil {
		list = make([]model.TaskDependencyNode, 0)
	}
	return list, nil
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

func hasDependencyPathTx(tx *gorm.DB, fromID, toID uint64) (bool, error) {
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
	if err := tx.Raw(query, fromID, toID).Scan(&found).Error; err != nil {
		return false, err
	}
	return found == 1, nil
}

func lockTaskIDs(tx *gorm.DB, firstID, secondID uint64) error {
	if firstID == 0 && secondID == 0 {
		return nil
	}
	first := firstID
	second := secondID
	if second < first {
		first, second = second, first
	}
	if first != 0 {
		if err := tx.Raw("SELECT pg_advisory_xact_lock(?)", int64(first)).Error; err != nil {
			return err
		}
	}
	if second != 0 && second != first {
		if err := tx.Raw("SELECT pg_advisory_xact_lock(?)", int64(second)).Error; err != nil {
			return err
		}
	}
	return nil
}

func (r *taskRepository) GetDependencyGraph(ctx context.Context, taskID uint64) ([]model.TaskDependencyNode, []model.TaskDependencyLink, error) {
	const nodeQuery = `
WITH RECURSIVE
ancestors(id) AS (
  SELECT predecessor_id FROM task_dependencies WHERE successor_id = ?
  UNION
  SELECT td.predecessor_id FROM task_dependencies td JOIN ancestors a ON td.successor_id = a.id
),
descendants(id) AS (
  SELECT successor_id FROM task_dependencies WHERE predecessor_id = ?
  UNION
  SELECT td.successor_id FROM task_dependencies td JOIN descendants d ON td.predecessor_id = d.id
),
related(id) AS (
  SELECT ?
  UNION SELECT id FROM ancestors
  UNION SELECT id FROM descendants
)
SELECT DISTINCT t.id, t.title, t.status
FROM related r
JOIN tasks t ON t.id = r.id AND t.deleted_at IS NULL
ORDER BY t.id ASC`

	nodes := make([]model.TaskDependencyNode, 0)
	if err := r.db.WithContext(ctx).Raw(nodeQuery, taskID, taskID, taskID).Scan(&nodes).Error; err != nil {
		return nil, nil, err
	}

	const linkQuery = `
WITH RECURSIVE
ancestors(id) AS (
  SELECT predecessor_id FROM task_dependencies WHERE successor_id = ?
  UNION
  SELECT td.predecessor_id FROM task_dependencies td JOIN ancestors a ON td.successor_id = a.id
),
descendants(id) AS (
  SELECT successor_id FROM task_dependencies WHERE predecessor_id = ?
  UNION
  SELECT td.successor_id FROM task_dependencies td JOIN descendants d ON td.predecessor_id = d.id
),
related(id) AS (
  SELECT ?
  UNION SELECT id FROM ancestors
  UNION SELECT id FROM descendants
)
SELECT DISTINCT td.predecessor_id AS source, td.successor_id AS target
FROM task_dependencies td
JOIN related rp ON rp.id = td.predecessor_id
JOIN related rs ON rs.id = td.successor_id
JOIN tasks pred ON pred.id = td.predecessor_id AND pred.deleted_at IS NULL
JOIN tasks succ ON succ.id = td.successor_id AND succ.deleted_at IS NULL
ORDER BY td.predecessor_id ASC, td.successor_id ASC`

	links := make([]model.TaskDependencyLink, 0)
	if err := r.db.WithContext(ctx).Raw(linkQuery, taskID, taskID, taskID).Scan(&links).Error; err != nil {
		return nil, nil, err
	}
	return nodes, links, nil
}

func (r *taskRepository) SoftDeleteByTeamID(ctx context.Context, teamID uint64) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.
			Where("predecessor_id IN (SELECT id FROM tasks WHERE team_id = ?) OR successor_id IN (SELECT id FROM tasks WHERE team_id = ?)", teamID, teamID).
			Delete(&model.TaskDependency{}).Error; err != nil {
			return err
		}

		res := tx.
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
	})
}
