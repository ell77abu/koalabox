# 错误映射说明（Error Mapping）

本文档用于明确后端当前实现中的错误语义与 HTTP 状态码映射，便于前后端联调和测试。

## 1. 统一错误语义（Repository 层）

定义位置：`internal/repository/errors.go`

- `ErrNotFound`：资源不存在（例如任务 ID 不存在，或用户不存在）。
- `ErrConflict`：资源冲突（当前主要用于注册时用户名唯一键冲突）。
- `ErrForbidden`：资源存在，但当前用户无权操作（越权访问/修改/删除）。

## 2. Auth 接口错误映射

对应文件：`internal/handler/auth_handler.go`

### 2.1 注册 `POST /api/v1/auth/register`

- `400 Bad Request`
  - 请求体非法（JSON 结构错误）
  - 用户名不合法（`ErrInvalidUsername`）
  - 密码不合法（`ErrInvalidPassword`）
- `409 Conflict`
  - 用户名已存在（`repository.ErrConflict`）
  - 响应 message：`name already exists`
- `500 Internal Server Error`
  - 其他未预期错误

### 2.2 登录 `POST /api/v1/auth/login`

- `400 Bad Request`
  - 请求体非法（JSON 结构错误）
- `401 Unauthorized`
  - 用户名或密码错误（`ErrInvalidCredentials`）
  - 响应 message：`invalid username or password`
- `500 Internal Server Error`
  - 其他未预期错误

### 2.3 修改密码 `PUT /api/v1/auth/password`

- `401 Unauthorized`
  - 未携带或无效登录态
  - 登录用户不存在（业务映射为 `ErrInvalidCredentials`）
- `400 Bad Request`
  - 请求体非法（JSON 结构错误）
  - 新密码不合法（`ErrInvalidPassword`）
  - 旧密码错误（`ErrInvalidOldPassword`）
- `500 Internal Server Error`
  - 其他未预期错误

## 3. Task 接口错误映射

对应文件：`internal/handler/task_handler.go`

### 3.1 查询任务 `GET /api/v1/tasks`

- `401 Unauthorized`
  - 未携带或无效登录态
- `500 Internal Server Error`
  - 查询失败

### 3.2 创建任务 `POST /api/v1/tasks`

- `401 Unauthorized`
  - 未携带或无效登录态
- `400 Bad Request`
  - 请求体非法
  - 标题、状态、优先级、截止时间校验失败
- `500 Internal Server Error`
  - 创建失败

### 3.3 更新任务 `PUT /api/v1/tasks/:id`

- `401 Unauthorized`
  - 未携带或无效登录态
- `400 Bad Request`
  - 任务 ID 非法
  - 请求体非法
  - 字段校验失败
- `403 Forbidden`
  - 任务存在，但不属于当前普通用户（`repository.ErrForbidden`）
- `404 Not Found`
  - 任务不存在（`repository.ErrNotFound`）
- `500 Internal Server Error`
  - 更新失败

### 3.4 删除任务 `DELETE /api/v1/tasks/:id`

- `401 Unauthorized`
  - 未携带或无效登录态
- `400 Bad Request`
  - 任务 ID 非法
- `403 Forbidden`
  - 任务存在，但不属于当前普通用户（`repository.ErrForbidden`）
- `404 Not Found`
  - 任务不存在（`repository.ErrNotFound`）
- `500 Internal Server Error`
  - 删除失败

## 4. 权限相关说明

- 普通用户（`level=0`）
  - 只能查看和操作自己的任务。
  - 对他人任务执行更新/删除时返回 `403 Forbidden`。
- 管理员（`level=1`）
  - 可查看并管理所有任务。

## 5. 前端联调建议

- 对 `401`：统一做登录态失效处理（清理 token 并跳转登录页）。
- 对 `403`：提示“无权限执行该操作”。
- 对 `404`：提示“资源不存在或已被删除”。
- 对 `409`：在注册场景提示“用户名已存在”。
- 对 `400`：优先展示后端返回的 message（用于表单字段错误提示）。

## 6. 备注

- 文档以当前代码行为为准。若后续新增业务错误，请同步更新：
  - `internal/repository/errors.go`
  - 对应 service 透传/转换逻辑
  - 对应 handler HTTP 映射
  - 本文档
