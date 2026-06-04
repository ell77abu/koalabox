-- Lab4 Activity Logs
-- Adds the dedicated operation log table used by Lab4.

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
