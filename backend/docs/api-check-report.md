# API Check Report (Auth + Tasks)

## Scope
Checked and aligned with `api.md` for:
- Auth: register/login
- Tasks: list/create/update/delete
- JWT bearer auth and ownership constraints

## Implemented checks

### Auth
- `POST /api/v1/auth/register`
  - username regex: `^[A-Za-z0-9_]{4,20}$`
  - password: length >= 6 and includes letter + digit
  - duplicate username -> `409`
- `POST /api/v1/auth/login`
  - invalid credentials -> `401`
  - success returns token + username

### Tasks
- JWT required for all `/api/v1/tasks` endpoints
- Token parsed from `Authorization: Bearer <token>`
- user id injected into request context
- Ownership enforced by repository query conditions:
  - list by `user_id`
  - update/delete by `id + user_id`

### Task request validation
- title: required, trimmed non-empty, max 120
- status: one of `TODO|DOING|DONE|ERROR`
- priority: one of `1|2|3`
- deadline: format `YYYY-MM-DD HH:mm:ss`

### Pagination
- query params: `page`, `size`
- default: `page=1`, `size=10`
- max size: `100`
- response includes `list/page/size/total`

## Routes
- `POST /api/v1/auth/register`
- `POST /api/v1/auth/login`
- `GET /api/v1/tasks`
- `POST /api/v1/tasks`
- `PUT /api/v1/tasks/:id`
- `PATCH /api/v1/tasks/:id`
- `DELETE /api/v1/tasks/:id`

## Unified response
All handlers use unified response helpers in `pkg/response/response.go`.

## Notes for frontend
- Always send JWT in `Authorization` header for tasks APIs.
- For deadline field, send string format: `2026-04-01 12:00:00`.
