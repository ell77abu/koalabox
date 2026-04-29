-- Lab2 Migration: Add team management, extend tasks, and add confirmation workflow tables
-- Safe to run multiple times (uses IF NOT EXISTS and guarded constraint additions)
-- Apply this to a PostgreSQL database that already has `users` and `tasks` (Lab1)

BEGIN;

-- 1) teams table (Lab2)
CREATE TABLE IF NOT EXISTS teams (
  id BIGSERIAL PRIMARY KEY,
  name VARCHAR(100) NOT NULL,
  owner_id BIGINT NOT NULL REFERENCES users(id),
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  deleted_at TIMESTAMP NULL
);

CREATE INDEX IF NOT EXISTS idx_teams_owner_id ON teams (owner_id);

-- 2) team_members table (Lab2)
CREATE TABLE IF NOT EXISTS team_members (
  id BIGSERIAL PRIMARY KEY,
  team_id BIGINT NOT NULL REFERENCES teams(id) ON DELETE CASCADE,
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

-- 3) extend tasks table with team_id and assignee_id (Lab2)
ALTER TABLE tasks
  ADD COLUMN IF NOT EXISTS team_id BIGINT NULL;

ALTER TABLE tasks
  ADD COLUMN IF NOT EXISTS assignee_id BIGINT NULL;

-- add foreign key constraints safely (ignore if already exists)
DO $$
BEGIN
  ALTER TABLE tasks
    ADD CONSTRAINT fk_tasks_team_id FOREIGN KEY (team_id) REFERENCES teams(id) ON DELETE SET NULL;
EXCEPTION WHEN duplicate_object THEN
  NULL;
END $$;

DO $$
BEGIN
  ALTER TABLE tasks
    ADD CONSTRAINT fk_tasks_assignee_id FOREIGN KEY (assignee_id) REFERENCES users(id) ON DELETE SET NULL;
EXCEPTION WHEN duplicate_object THEN
  NULL;
END $$;

CREATE INDEX IF NOT EXISTS idx_tasks_team_id ON tasks (team_id);
CREATE INDEX IF NOT EXISTS idx_tasks_assignee_id ON tasks (assignee_id);

-- 4) team_invitations table (Lab2 confirmation workflow)
CREATE TABLE IF NOT EXISTS team_invitations (
  id BIGSERIAL PRIMARY KEY,
  team_id BIGINT NOT NULL REFERENCES teams(id),
  inviter_id BIGINT NOT NULL REFERENCES users(id),
  invitee_id BIGINT NOT NULL REFERENCES users(id),
  status VARCHAR(10) NOT NULL DEFAULT 'PENDING',
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  CONSTRAINT chk_team_invite_status CHECK (status IN ('PENDING', 'ACCEPTED', 'DECLINED')),
  CONSTRAINT uk_team_invite_team_user_status UNIQUE (team_id, invitee_id, status)
);

CREATE INDEX IF NOT EXISTS idx_team_invite_invitee ON team_invitations (invitee_id, status);
CREATE INDEX IF NOT EXISTS idx_team_invite_team ON team_invitations (team_id);

-- 5) task_assignment_requests table (Lab2 confirmation workflow)
CREATE TABLE IF NOT EXISTS task_assignment_requests (
  id BIGSERIAL PRIMARY KEY,
  task_id BIGINT NOT NULL REFERENCES tasks(id),
  team_id BIGINT NOT NULL REFERENCES teams(id),
  requester_id BIGINT NOT NULL REFERENCES users(id),
  assignee_id BIGINT NOT NULL REFERENCES users(id),
  status VARCHAR(10) NOT NULL DEFAULT 'PENDING',
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  CONSTRAINT chk_task_assign_status CHECK (status IN ('PENDING', 'ACCEPTED', 'DECLINED')),
  CONSTRAINT uk_task_assign_task_user_status UNIQUE (task_id, assignee_id, status)
);

CREATE INDEX IF NOT EXISTS idx_task_assign_assignee ON task_assignment_requests (assignee_id, status);
CREATE INDEX IF NOT EXISTS idx_task_assign_task ON task_assignment_requests (task_id);

COMMIT;
