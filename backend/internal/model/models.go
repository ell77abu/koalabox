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

// InvitationStatus 表示邀请/分配确认状态。
type InvitationStatus string

const (
	InvitationStatusPending  InvitationStatus = "PENDING"
	InvitationStatusAccepted InvitationStatus = "ACCEPTED"
	InvitationStatusDeclined InvitationStatus = "DECLINED"
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

// TaskDependency 映射 task_dependencies 表。
type TaskDependency struct {
	ID            uint64    `json:"id" db:"id"`
	PredecessorID uint64    `json:"predecessor_id" db:"predecessor_id"`
	SuccessorID   uint64    `json:"successor_id" db:"successor_id"`
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
}

// TaskDependencyNode 用于依赖关系返回。
type TaskDependencyNode struct {
	ID     uint64     `json:"id" db:"id"`
	Title  string     `json:"title" db:"title"`
	Status TaskStatus `json:"status" db:"status"`
}

// TaskDependencyLink 用于依赖图返回，source 指向 target。
type TaskDependencyLink struct {
	Source uint64 `json:"source" db:"source"`
	Target uint64 `json:"target" db:"target"`
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

// TeamInvitation 映射 team_invitations 表。
type TeamInvitation struct {
	ID        uint64           `json:"id" db:"id"`
	TeamID    uint64           `json:"team_id" db:"team_id"`
	InviterID uint64           `json:"inviter_id" db:"inviter_id"`
	InviteeID uint64           `json:"invitee_id" db:"invitee_id"`
	Status    InvitationStatus `json:"status" db:"status"`
	CreatedAt time.Time        `json:"created_at" db:"created_at"`
	UpdatedAt time.Time        `json:"updated_at" db:"updated_at"`
}

// TeamWithRole 用于“我的团队”列表返回。
type TeamWithRole struct {
	TeamID uint64   `json:"team_id" db:"team_id"`
	Name   string   `json:"name" db:"name"`
	Role   TeamRole `json:"role" db:"role"`
}

// TeamMemberDetail 用于团队成员列表返回。
type TeamMemberDetail struct {
	UserID   uint64   `json:"user_id" db:"user_id"`
	Username string   `json:"username" db:"username"`
	Role     TeamRole `json:"role" db:"role"`
}

// TeamInvitationDetail 用于待确认邀请列表。
type TeamInvitationDetail struct {
	ID              uint64           `json:"id" db:"id"`
	TeamID          uint64           `json:"team_id" db:"team_id"`
	TeamName        string           `json:"team_name" db:"team_name"`
	InviterID       uint64           `json:"inviter_id" db:"inviter_id"`
	InviterUsername string           `json:"inviter_username" db:"inviter_username"`
	Status          InvitationStatus `json:"status" db:"status"`
	CreatedAt       time.Time        `json:"created_at" db:"created_at"`
}

// TaskAssignmentRequest 映射 task_assignment_requests 表。
type TaskAssignmentRequest struct {
	ID          uint64           `json:"id" db:"id"`
	TaskID      uint64           `json:"task_id" db:"task_id"`
	TeamID      uint64           `json:"team_id" db:"team_id"`
	RequesterID uint64           `json:"requester_id" db:"requester_id"`
	AssigneeID  uint64           `json:"assignee_id" db:"assignee_id"`
	Status      InvitationStatus `json:"status" db:"status"`
	CreatedAt   time.Time        `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time        `json:"updated_at" db:"updated_at"`
}

// TaskAssignmentRequestDetail 用于待确认分配列表。
type TaskAssignmentRequestDetail struct {
	ID                uint64           `json:"id" db:"id"`
	TaskID            uint64           `json:"task_id" db:"task_id"`
	TaskTitle         string           `json:"task_title" db:"task_title"`
	TeamID            uint64           `json:"team_id" db:"team_id"`
	TeamName          string           `json:"team_name" db:"team_name"`
	RequesterID       uint64           `json:"requester_id" db:"requester_id"`
	RequesterUsername string           `json:"requester_username" db:"requester_username"`
	Status            InvitationStatus `json:"status" db:"status"`
	CreatedAt         time.Time        `json:"created_at" db:"created_at"`
}

// ActivityActionType 对应 Lab4 操作日志的任务操作类型。
type ActivityActionType string

const (
	ActivityActionCreateTask                   ActivityActionType = "CREATE_TASK"
	ActivityActionUpdateTaskInfo               ActivityActionType = "UPDATE_TASK_INFO"
	ActivityActionDeleteTask                   ActivityActionType = "DELETE_TASK"
	ActivityActionAssigneeChangedByMemberLeave ActivityActionType = "TASK_ASSIGNEE_CHANGED_BY_MEMBER_LEAVE"
)

// ActivityTargetType 表示操作对象类型，本期固定为 TASK。
type ActivityTargetType string

const (
	ActivityTargetTask ActivityTargetType = "TASK"
)

// ActivityScope 表示日志所属任务范围。
type ActivityScope string

const (
	ActivityScopePersonal ActivityScope = "PERSONAL"
	ActivityScopeTeam     ActivityScope = "TEAM"
)

// ActivityLog 映射 Lab4 新增的 activity_logs 表。
type ActivityLog struct {
	ID         uint64             `json:"id" db:"id"`
	ActorID    uint64             `json:"actor_id" db:"actor_id"`
	ActorName  string             `json:"actor_name" db:"actor_name"`
	ActionType ActivityActionType `json:"action_type" db:"action_type"`
	TargetType ActivityTargetType `json:"target_type" db:"target_type"`
	TargetID   uint64             `json:"target_id" db:"target_id"`
	Scope      ActivityScope      `json:"scope" db:"scope"`
	TeamID     *uint64            `json:"team_id" db:"team_id"`
	Content    string             `json:"content" db:"content"`
	CreatedAt  time.Time          `json:"created_at" db:"created_at"`
}

// ActivityLogListItem 是任务日志列表对前端暴露的最小展示字段。
type ActivityLogListItem struct {
	ID         uint64             `json:"id" db:"id"`
	ActorName  string             `json:"actor_name" db:"actor_name"`
	ActionType ActivityActionType `json:"action_type" db:"action_type"`
	Content    string             `json:"content" db:"content"`
	CreatedAt  time.Time          `json:"created_at" db:"created_at"`
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
