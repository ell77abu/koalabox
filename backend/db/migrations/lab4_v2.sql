-- Lab4-v2: task trash and assignment notification dismiss support.
-- Apply this after lab2_merge.sql.

BEGIN;

ALTER TABLE task_assignment_requests
  ADD COLUMN IF NOT EXISTS dismiss_at TIMESTAMP NULL;

CREATE INDEX IF NOT EXISTS idx_task_assign_assignee_dismiss
  ON task_assignment_requests (assignee_id, dismiss_at, created_at DESC);

COMMIT;
