-- Lab2 Migration: Add team management and extend tasks table
-- This migration safely extends the Lab1 database with Lab2 requirements
-- Apply this to an existing database that has users and tasks tables (without team_id/assignee_id)

BEGIN;

CREATE TABLE IF NOT EXISTS teams (
  id BIGSERIAL PRIMARY KEY,
  name VARCHAR(100) NOT NULL,
  owner_id BIGINT NOT NULL REFERENCES users(id),
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  deleted_at TIMESTAMP NULL
);

CREATE INDEX IF NOT EXISTS idx_teams_owner_id ON teams (owner_id);

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


ALTER TABLE tasks
  ADD COLUMN IF NOT EXISTS team_id BIGINT NULL;

ALTER TABLE tasks
  ADD COLUMN IF NOT EXISTS assignee_id BIGINT NULL;


DO $$
BEGIN
  ALTER TABLE tasks
    ADD CONSTRAINT fk_tasks_team_id FOREIGN KEY (team_id) REFERENCES teams(id) ON DELETE SET NULL;
EXCEPTION WHEN duplicate_object THEN
  -- Constraint already exists, ignore
  NULL;
END $$;

DO $$
BEGIN
  ALTER TABLE tasks
    ADD CONSTRAINT fk_tasks_assignee_id FOREIGN KEY (assignee_id) REFERENCES users(id) ON DELETE SET NULL;
EXCEPTION WHEN duplicate_object THEN
  -- Constraint already exists, ignore
  NULL;
END $$;

CREATE INDEX IF NOT EXISTS idx_tasks_team_id ON tasks (team_id);
CREATE INDEX IF NOT EXISTS idx_tasks_assignee_id ON tasks (assignee_id);


COMMIT;

