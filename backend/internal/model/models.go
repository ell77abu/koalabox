package model

import "time"

// TaskStatus 对应 API 中的任务状态枚举：TODO、DOING、DONE、ERROR。
type TaskStatus string

const (
	TaskStatusTODO  TaskStatus = "TODO"
	TaskStatusDOING TaskStatus = "DOING"
	TaskStatusDONE  TaskStatus = "DONE"
	TaskStatusERROR TaskStatus = "ERROR"
)

// Priority 任务优先级：1=低，2=中，3=高。
type Priority uint8

const (
	PriorityLow    Priority = 1
	PriorityMedium Priority = 2
	PriorityHigh   Priority = 3
)

// TeamRole 对应 team_members 里的成员角色。
type TeamRole string

const (
	TeamRoleOwner  TeamRole = "Owner"
	TeamRoleAdmin  TeamRole = "Admin"
	TeamRoleMember TeamRole = "Member"
)

// User 映射 users 表。
type User struct {
	ID    uint64 `json:"id" db:"id"`
	Level uint8  `json:"level" db:"level"`
	// 用户名正则 ^[A-Za-z0-9_]{4,20}$ 在请求 DTO/Service 层校验。
	Username     string     `json:"username" db:"username"`
	PasswordHash string     `json:"-" db:"password_hash"`
	CreatedAt    time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at" db:"updated_at"`
	DeletedAt    *time.Time `json:"-" db:"deleted_at"`
}

// Task 映射 tasks 表。
type Task struct {
	ID          uint64     `json:"id" db:"id"`
	UserID      uint64     `json:"user_id" db:"user_id"`
	TeamID      *uint64    `json:"team_id" db:"team_id"`
	AssigneeID  *uint64    `json:"assignee_id" db:"assignee_id"`
	Title       string     `json:"title" db:"title" validate:"required,min=1,max=120"`
	Description string     `json:"description" db:"description"`
	Status      TaskStatus `json:"status" db:"status" validate:"required,oneof=TODO DOING DONE ERROR"`
	Priority    Priority   `json:"priority" db:"priority" validate:"required,oneof=1 2 3"`
	Deadline    *time.Time `json:"deadline" db:"deadline"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at" db:"updated_at"`
	DeletedAt   *time.Time `json:"deleted_at,omitempty" db:"deleted_at"`
}

// Team 映射 teams 表。
type Team struct {
	ID        uint64     `json:"id" db:"id"`
	Name      string     `json:"name" db:"name"`
	OwnerID   uint64     `json:"owner_id" db:"owner_id"`
	CreatedAt time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt time.Time  `json:"updated_at" db:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty" db:"deleted_at"`
}

// TeamMember 映射 team_members 表。
type TeamMember struct {
	ID        uint64     `json:"id" db:"id"`
	TeamID    uint64     `json:"team_id" db:"team_id"`
	UserID    uint64     `json:"user_id" db:"user_id"`
	Role      TeamRole   `json:"role" db:"role"`
	CreatedAt time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt time.Time  `json:"updated_at" db:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty" db:"deleted_at"`
}

// TaskAuditLog 映射 task_audit_logs 表。
type TaskAuditLog struct {
	ID        uint64    `json:"id" db:"id"`
	TaskID    uint64    `json:"task_id" db:"task_id"`
	UserID    uint64    `json:"user_id" db:"user_id"`
	Action    string    `json:"action" db:"action"`
	Before    *string   `json:"before_data" db:"before_data"`
	After     *string   `json:"after_data" db:"after_data"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// CommonResponse 为 API 统一响应结构。
type CommonResponse[T any] struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    T      `json:"data"`
}

// LoginResponseData 对应 /auth/login 成功返回的数据结构。
type LoginResponseData struct {
	Token    string `json:"token"`
	Username string `json:"username"`
	Level    uint8  `json:"level"`
}

// Paging 定义列表查询通用分页参数。
type Paging struct {
	Page int `json:"page" form:"page" validate:"min=1"`
	Size int `json:"size" form:"size" validate:"min=1,max=100"`
}
