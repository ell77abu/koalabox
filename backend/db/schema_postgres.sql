-- Database schema for Collaborative Task Management System API v1.0
-- Recommended engine: PostgreSQL 14+

-- Create database manually if needed:
-- CREATE DATABASE task_system;

-- Connect to task_system before running below statements.

CREATE TABLE IF NOT EXISTS users (
  id BIGSERIAL PRIMARY KEY,
  username VARCHAR(20) NOT NULL UNIQUE,
  level SMALLINT NOT NULL DEFAULT 0,
  password_hash VARCHAR(255) NOT NULL,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  deleted_at TIMESTAMP NULL,
  CONSTRAINT chk_users_level CHECK (level IN (0, 1)),
  CONSTRAINT chk_users_username_format
    CHECK (username ~ '^[A-Za-z0-9_]{4,20}$')
);

CREATE TABLE IF NOT EXISTS teams (
  id BIGSERIAL PRIMARY KEY,
  name VARCHAR(100) NOT NULL,
  owner_id BIGINT NOT NULL REFERENCES users(id),
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  deleted_at TIMESTAMP NULL
);

CREATE INDEX IF NOT EXISTS idx_teams_owner_id ON teams (owner_id);

CREATE TABLE IF NOT EXISTS tasks (
  id BIGSERIAL PRIMARY KEY,
  user_id BIGINT NOT NULL REFERENCES users(id),
  team_id BIGINT NULL REFERENCES teams(id),
  assignee_id BIGINT NULL REFERENCES users(id),
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
CREATE INDEX IF NOT EXISTS idx_tasks_team_id ON tasks (team_id);
CREATE INDEX IF NOT EXISTS idx_tasks_assignee_id ON tasks (assignee_id);

CREATE TABLE IF NOT EXISTS team_members (
  id BIGSERIAL PRIMARY KEY,
  team_id BIGINT NOT NULL REFERENCES teams(id),
  user_id BIGINT NOT NULL REFERENCES users(id),
  role VARCHAR(10) NOT NULL,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  deleted_at TIMESTAMP NULL,
  CONSTRAINT chk_team_members_role CHECK (role IN ('Owner', 'Admin', 'Member')),
  CONSTRAINT uk_team_members_team_user UNIQUE (team_id, user_id)
);

CREATE INDEX IF NOT EXISTS idx_team_members_team_id ON team_members (team_id);
CREATE INDEX IF NOT EXISTS idx_team_members_user_id ON team_members (user_id);

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
