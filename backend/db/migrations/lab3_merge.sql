-- Lab3 Merge Migration: Add task dependencies with isolation rules
-- This merge script applies Lab3 changes on top of Lab1+Lab2 DB (safe to run multiple times)

BEGIN;

-- adjacency-list edge table for task dependencies
CREATE TABLE IF NOT EXISTS task_dependencies (
  id BIGSERIAL PRIMARY KEY,
  predecessor_id BIGINT NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
  successor_id BIGINT NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

  -- same dependency cannot be added twice
  CONSTRAINT uk_task_dependencies UNIQUE (predecessor_id, successor_id),
  -- a task cannot depend on itself
  CONSTRAINT chk_no_self_dependency CHECK (predecessor_id != successor_id)
);

CREATE INDEX IF NOT EXISTS idx_task_deps_pred ON task_dependencies(predecessor_id);
CREATE INDEX IF NOT EXISTS idx_task_deps_succ ON task_dependencies(successor_id);

-- Enforce domain rules:
-- 1) personal tasks (team_id IS NULL) may only depend on tasks owned by the same user
-- 2) team tasks (team_id IS NOT NULL) may only depend on tasks within the same team
-- 3) personal and team tasks cannot be mixed in a dependency
-- Implemented with a trigger because CHECK constraints cannot reference other tables.

CREATE OR REPLACE FUNCTION fn_task_dependencies_validate() RETURNS trigger AS $$
DECLARE
  pred_user BIGINT;
  pred_team BIGINT;
  succ_user BIGINT;
  succ_team BIGINT;
BEGIN
  SELECT user_id, team_id INTO pred_user, pred_team FROM tasks WHERE id = NEW.predecessor_id;
  IF NOT FOUND THEN
    RAISE EXCEPTION 'Predecessor task % does not exist', NEW.predecessor_id;
  END IF;

  SELECT user_id, team_id INTO succ_user, succ_team FROM tasks WHERE id = NEW.successor_id;
  IF NOT FOUND THEN
    RAISE EXCEPTION 'Successor task % does not exist', NEW.successor_id;
  END IF;

  -- both personal tasks: owners must match
  IF pred_team IS NULL AND succ_team IS NULL THEN
    IF pred_user IS DISTINCT FROM succ_user THEN
      RAISE EXCEPTION 'Personal tasks must belong to the same user: predecessor=% owner=%, successor=% owner=%', NEW.predecessor_id, pred_user, NEW.successor_id, succ_user;
    END IF;

  ELSIF pred_team IS NOT NULL AND succ_team IS NOT NULL THEN
    -- both team tasks: team must match
    IF pred_team IS DISTINCT FROM succ_team THEN
      RAISE EXCEPTION 'Team tasks must belong to the same team: predecessor=% team=%, successor=% team=%', NEW.predecessor_id, pred_team, NEW.successor_id, succ_team;
    END IF;

  ELSE
    -- one personal and one team task: not allowed
    RAISE EXCEPTION 'Cannot create dependency between personal and team tasks: predecessor=% team=% , successor=% team=%', NEW.predecessor_id, pred_team, NEW.successor_id, succ_team;
  END IF;

  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- recreate trigger in a safe way
DROP TRIGGER IF EXISTS trg_task_dependencies_validate ON task_dependencies;

CREATE TRIGGER trg_task_dependencies_validate
  BEFORE INSERT OR UPDATE ON task_dependencies
  FOR EACH ROW EXECUTE FUNCTION fn_task_dependencies_validate();

COMMIT;
