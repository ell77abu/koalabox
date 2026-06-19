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

CREATE TABLE IF NOT EXISTS activity_logs (
  id BIGSERIAL PRIMARY KEY,
  actor_id BIGINT NOT NULL,
  actor_name VARCHAR(20) NOT NULL,
  action_type VARCHAR(50) NOT NULL,
  target_type VARCHAR(20) NOT NULL,
  target_id BIGINT NOT NULL,
  scope VARCHAR(20) NOT NULL,
  team_id BIGINT NULL,
  content TEXT NOT NULL,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  CONSTRAINT chk_activity_action_type CHECK (
    action_type IN (
      'CREATE_TASK',
      'UPDATE_TASK_INFO',
      'DELETE_TASK',
      'TASK_ASSIGNEE_CHANGED_BY_MEMBER_LEAVE'
    )
  ),
  CONSTRAINT chk_activity_target_type CHECK (target_type IN ('TASK')),
  CONSTRAINT chk_activity_scope CHECK (scope IN ('PERSONAL', 'TEAM'))
);

CREATE INDEX IF NOT EXISTS idx_activity_logs_target_created
  ON activity_logs (target_type, target_id, created_at DESC);

-- Lab4-v2 assignment notification dismiss support.
-- This schema file is kept compatible with older base schemas where
-- task_assignment_requests may be created by lab2_merge.sql instead.
ALTER TABLE IF EXISTS task_assignment_requests
  ADD COLUMN IF NOT EXISTS dismiss_at TIMESTAMP NULL;

DO $$
BEGIN
  IF to_regclass('public.task_assignment_requests') IS NOT NULL THEN
    CREATE INDEX IF NOT EXISTS idx_task_assign_assignee_dismiss
      ON task_assignment_requests (assignee_id, dismiss_at, created_at DESC);
  END IF;
END $$;
