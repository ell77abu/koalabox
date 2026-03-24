package model

import "time"

// TaskStatus matches API enum: TODO, DOING, DONE, ERROR.
type TaskStatus string

const (
	TaskStatusTODO  TaskStatus = "TODO"
	TaskStatusDOING TaskStatus = "DOING"
	TaskStatusDONE  TaskStatus = "DONE"
	TaskStatusERROR TaskStatus = "ERROR"
)

// Priority: 1=low, 2=medium, 3=high.
type Priority uint8

const (
	PriorityLow    Priority = 1
	PriorityMedium Priority = 2
	PriorityHigh   Priority = 3
)

// User maps to users table.
type User struct {
	ID uint64 `json:"id" db:"id"`
	// Enforce username regex ^[A-Za-z0-9_]{4,20}$ in request DTO/service layer.
	Username     string     `json:"username" db:"username"`
	PasswordHash string     `json:"-" db:"password_hash"`
	CreatedAt    time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at" db:"updated_at"`
	DeletedAt    *time.Time `json:"-" db:"deleted_at"`
}

// Task maps to tasks table.
type Task struct {
	ID          uint64     `json:"id" db:"id"`
	UserID      uint64     `json:"user_id" db:"user_id"`
	Title       string     `json:"title" db:"title" validate:"required,min=1,max=120"`
	Description string     `json:"description" db:"description"`
	Status      TaskStatus `json:"status" db:"status" validate:"required,oneof=TODO DOING DONE ERROR"`
	Priority    Priority   `json:"priority" db:"priority" validate:"required,oneof=1 2 3"`
	Deadline    *time.Time `json:"deadline" db:"deadline"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at" db:"updated_at"`
	DeletedAt   *time.Time `json:"deleted_at,omitempty" db:"deleted_at"`
}

// TaskAuditLog maps to task_audit_logs table.
type TaskAuditLog struct {
	ID        uint64    `json:"id" db:"id"`
	TaskID    uint64    `json:"task_id" db:"task_id"`
	UserID    uint64    `json:"user_id" db:"user_id"`
	Action    string    `json:"action" db:"action"`
	Before    *string   `json:"before_data" db:"before_data"`
	After     *string   `json:"after_data" db:"after_data"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// CommonResponse follows the API unified response style.
type CommonResponse[T any] struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    T      `json:"data"`
}

// LoginResponseData follows /auth/login success payload.
type LoginResponseData struct {
	Token    string `json:"token"`
	Username string `json:"username"`
}

// Paging defines common list query params.
type Paging struct {
	Page int `json:"page" form:"page" validate:"min=1"`
	Size int `json:"size" form:"size" validate:"min=1,max=100"`
}
