-- Database schema for Collaborative Task Management System API v1.0
-- Recommended engine: PostgreSQL 14+

-- Create database manually if needed:
-- CREATE DATABASE task_system;

-- Connect to task_system before running below statements.

CREATE TABLE IF NOT EXISTS users (
  id BIGSERIAL PRIMARY KEY,
  username VARCHAR(20) NOT NULL UNIQUE,
  password_hash VARCHAR(255) NOT NULL,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  deleted_at TIMESTAMP NULL,
  CONSTRAINT chk_users_username_format
    CHECK (username ~ '^[A-Za-z0-9_]{4,20}$')
);

CREATE TABLE IF NOT EXISTS tasks (
  id BIGSERIAL PRIMARY KEY,
  user_id BIGINT NOT NULL REFERENCES users(id),
  title VARCHAR(120) NOT NULL,
  description TEXT NULL,
  status VARCHAR(10) NOT NULL DEFAULT 'TODO',
  priority SMALLINT NOT NULL DEFAULT 2,
  deadline TIMESTAMP NULL,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  deleted_at TIMESTAMP NULL,
  CONSTRAINT chk_tasks_status CHECK (status IN ('TODO', 'DOING', 'DONE', 'ERROR')),
  CONSTRAINT chk_tasks_priority CHECK (priority IN (1, 2, 3))
);

CREATE INDEX IF NOT EXISTS idx_tasks_user_deleted_created
  ON tasks (user_id, deleted_at, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_tasks_user_status ON tasks (user_id, status);
CREATE INDEX IF NOT EXISTS idx_tasks_user_deadline ON tasks (user_id, deadline);
CREATE INDEX IF NOT EXISTS idx_tasks_user_priority ON tasks (user_id, priority);

CREATE TABLE IF NOT EXISTS task_audit_logs (
  id BIGSERIAL PRIMARY KEY,
  task_id BIGINT NOT NULL REFERENCES tasks(id),
  user_id BIGINT NOT NULL REFERENCES users(id),
  action VARCHAR(20) NOT NULL,
  before_data JSONB NULL,
  after_data JSONB NULL,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  CONSTRAINT chk_task_audit_action CHECK (action IN ('CREATE', 'UPDATE', 'DELETE', 'STATUS_CHANGE'))
);

CREATE INDEX IF NOT EXISTS idx_task_audit_task_created
  ON task_audit_logs (task_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_task_audit_user_created
  ON task_audit_logs (user_id, created_at DESC);
