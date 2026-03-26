# Repository + Pagination + Unified Response Guide

## 1. New files

- `internal/repository/errors.go`
- `internal/repository/user_repository.go`
- `internal/repository/task_repository.go`
- `pkg/response/response.go` (enhanced)

## 2. User repository

```go
repo := repository.NewUserRepository(db)
```

Provided methods:
- `Create(ctx, user)`
- `GetByUsername(ctx, username)`
- `GetByID(ctx, id)`

`GetByUsername/GetByID` return `repository.ErrNotFound` when not found.

## 3. Task repository

```go
repo := repository.NewTaskRepository(db)
```

Provided methods:
- `Create(ctx, task)`
- `ListByUserID(ctx, userID, page, size)`
- `GetByIDAndUserID(ctx, id, userID)`
- `UpdateByIDAndUserID(ctx, task)`
- `DeleteByIDAndUserID(ctx, id, userID)`

### Pagination behavior
- default: `page=1,size=10`
- max size: `100`
- return: `([]model.Task, total, error)`

## 4. Unified response helpers

Use helpers in `pkg/response/response.go`:
- `response.Success(...)`
- `response.Created(...)`
- `response.SuccessPage(...)`
- `response.BadRequest(...)`
- `response.Unauthorized(...)`
- `response.Forbidden(...)`
- `response.NotFound(...)`
- `response.Conflict(...)`
- `response.InternalError(...)`

## 5. Handler usage example (task list)

```go
func (h *TaskHandler) List(c *gin.Context) {
    page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
    size, _ := strconv.Atoi(c.DefaultQuery("size", "10"))

    userID := uint64(1) // replace with jwt middleware parsed value
    list, total, err := h.taskRepo.ListByUserID(c.Request.Context(), userID, page, size)
    if err != nil {
        response.InternalError(c, "query tasks failed")
        return
    }

    response.SuccessPage(c, "query success", list, page, size, total)
}
```

## 6. Permission guarantee

All task write/read methods include `user_id` in SQL conditions to guarantee "only operate own tasks".
