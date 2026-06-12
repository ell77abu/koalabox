# Collaborative Task Management - Database Design (Go backend)

## 1. Database Choice
- RDBMS: PostgreSQL 14+
- Reason:
  - Strong consistency for task CRUD and ownership control.
  - Native constraints for enum and priority checks.
  - Good indexing support for pagination and filtering.

## 2. Core Tables
- users
- tasks
- activity_logs

See SQL file: `db/schema_postgres.sql`.

## 3. Field Design Mapping to API
### users
- id: primary key
- username: unique, matches `[A-Za-z0-9_]{4,20}`
- password_hash: store bcrypt/argon2 hash only `[a-zA-Z0-9_]{6,30}`
- created_at / updated_at / deleted_at

### tasks
- id: primary key
- user_id: owner user id
- title / description
- message
- status: `TODO|DOING|DONE|ERROR`
- priority: `1|2|3`
- deadline: nullable datetime
- created_at / updated_at / deleted_at

### activity_logs
- actor_id / actor_name / action_type
- target_type / target_id
- scope / team_id
- content
- created_at

## 4. Ownership Enforcement
For all operations under `/tasks`, use user id from JWT and scope SQL by `user_id`:

```sql
SELECT * FROM tasks
WHERE user_id = ? AND deleted_at IS NULL
ORDER BY created_at DESC
LIMIT ? OFFSET ?;
```

Update/Delete must also include ownership filter:

```sql
UPDATE tasks
SET title = ?, description = ?, status = ?, priority = ?, deadline = ?
WHERE id = ? AND user_id = ? AND deleted_at IS NULL;

UPDATE tasks
SET deleted_at = NOW()
WHERE id = ? AND user_id = ? AND deleted_at IS NULL;
```

If affected rows = 0, return `403` (or `404`, depending on your policy).

## 5. Pagination Recommendation
- Request: `page`, `size`
- Convert:
  - `limit = size`
  - `offset = (page - 1) * size`
- Add max size cap (recommended <= 100).

## 6. Input Validation Notes
- Register:
  - username regex: `^[A-Za-z0-9_]{4,20}$`
  - password: length >= 6 and includes letters + numbers
- Task status/priority: whitelist only
- deadline parse format: `2006-01-02 15:04:05`

## 7. Suggested Future Extensions
- tags + task_tag_rel table
- comments table for collaboration
- shared_task_members table for multi-user collaboration
- notification table for due-date reminders
