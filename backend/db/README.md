# 数据库初始化与迁移说明

本文档用于说明 `backend/db` 目录下 SQL 脚本的用途、执行顺序与推荐实践，覆盖：

- 全新环境初始化（从 0 到当前版本）
- 已有旧版本数据库的增量迁移（Lab1/Lab2/Lab3 -> Lab4）

## 1. 目前脚本

- `schema_postgres.sql`
  - 当前项目的基础建库脚本（仅保留通用基础结构：`users`、`tasks`、`activity_logs` 及其索引）
  - 适合数据库初始化的第一步
  - 不包含 Lab2 团队协作结构和 Lab3 任务依赖结构

- `migrations/lab2_merge.sql`
  - 将 Lab2 能力增量加到 Lab1（团队体系、任务扩展字段、邀请/指派工作流）
  - 设计为可重复执行（大量 `IF NOT EXISTS` / 受保护约束）

- `migrations/lab3_merge.sql`
  - Lab3 增量：新增 `task_dependencies`、索引、以及跨任务域校验触发器
  - 设计为可重复执行（`CREATE TABLE IF NOT EXISTS` + `DROP TRIGGER IF EXISTS` + `CREATE OR REPLACE FUNCTION`）

- `migrations/lab4_activity_logs.sql`
  - Lab4 增量：新增 `activity_logs` 操作日志表与查询索引
  - 设计为可重复执行（`CREATE TABLE IF NOT EXISTS` + `CREATE INDEX IF NOT EXISTS`）

## 2. 脚本执行策略

### 全新数据库初始化

适用：本地首次搭建、测试环境新建库、部署新实例。

执行顺序：

1. 创建数据库（示例）
   - `CREATE DATABASE task_system;`
2. 执行基础架构
   - `backend/db/schema_postgres.sql`
3. 执行 Lab2 增量
   - `backend/db/migrations/lab2_merge.sql`
4. 执行 Lab3 增量
   - `backend/db/migrations/lab3_merge.sql`
5. 执行 Lab4 增量
   - `backend/db/migrations/lab4_activity_logs.sql`

按上述顺序执行后，数据库的基线架构和后续增量开发分离，同时每个脚本职责更清晰。

### 已有 Lab1 数据库，迁移升级到当前 Lab4 版本

执行顺序：

1. 执行 `backend/db/migrations/lab2_merge.sql`
2. 执行 `backend/db/migrations/lab3_merge.sql`
3. 执行 `backend/db/migrations/lab4_activity_logs.sql`

### 已有 Lab2 数据库，迁移升级到 Lab4

执行顺序：

1. 执行 `backend/db/migrations/lab3_merge.sql`
2. 执行 `backend/db/migrations/lab4_activity_logs.sql`

### 已有 Lab3 数据库，迁移升级到 Lab4

执行顺序：

1. 仅执行 `backend/db/migrations/lab4_activity_logs.sql`

## 3. 执行示例（psql）

以下命令在仓库根目录执行：

```bash
psql "postgresql://<user>:<password>@<host>:<port>/<dbname>" -f backend/db/schema_postgres.sql
psql "postgresql://<user>:<password>@<host>:<port>/<dbname>" -f backend/db/migrations/lab2_merge.sql
psql "postgresql://<user>:<password>@<host>:<port>/<dbname>" -f backend/db/migrations/lab3_merge.sql
psql "postgresql://<user>:<password>@<host>:<port>/<dbname>" -f backend/db/migrations/lab4_activity_logs.sql
```

增量迁移（Lab1 -> Lab4）：

```bash
psql "postgresql://<user>:<password>@<host>:<port>/<dbname>" -f backend/db/migrations/lab2_merge.sql
psql "postgresql://<user>:<password>@<host>:<port>/<dbname>" -f backend/db/migrations/lab3_merge.sql
psql "postgresql://<user>:<password>@<host>:<port>/<dbname>" -f backend/db/migrations/lab4_activity_logs.sql
```
或先连接 PostgreSQL，再在目标数据库执行对应 SQL 脚本。

## 4. 初始化后最小校验

建议执行以下 SQL 做结构校验：

```sql
-- Lab3 核心表是否存在
SELECT to_regclass('public.task_dependencies') AS task_dependencies_table;

-- Lab2/Lab3/Lab4 核心表数量检查
SELECT table_name
FROM information_schema.tables
WHERE table_schema = 'public'
  AND table_name IN (
    'users','tasks','teams','team_members',
    'team_invitations','task_assignment_requests',
    'task_dependencies','activity_logs'
  )
ORDER BY table_name;

-- 触发器是否存在
SELECT tgname
FROM pg_trigger
WHERE tgname = 'trg_task_dependencies_validate';
```

## 5. 迁移与回滚建议

- 迁移前先备份：
  - `pg_dump -Fc -d <dbname> -f backup_before_migration.dump`
- 在低峰期执行迁移，避免与业务写入高并发冲突。
- 当前脚本未提供完整“自动回滚脚本”，推荐通过备份恢复回退。
- 所有脚本均为 PostgreSQL 方言，建议 PostgreSQL 14+。

## 6. 与后端配置的关系

- 后端通过 `configs/config.yaml`（或环境变量）中的 `database.dsn` 连接数据库。
- 只要迁移完成且 DSN 指向正确实例，服务可直接启动：
  - `go run ./cmd/server`

