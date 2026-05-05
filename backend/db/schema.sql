-- Database schema for Collaborative Task Management System API v1.0
-- Recommended engine: MySQL 8.0+

CREATE DATABASE IF NOT EXISTS task_system
  DEFAULT CHARACTER SET utf8mb4
  DEFAULT COLLATE utf8mb4_0900_ai_ci;

USE task_system;

-- Users
CREATE TABLE IF NOT EXISTS users (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  username VARCHAR(20) NOT NULL,
  password_hash VARCHAR(255) NOT NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  deleted_at DATETIME NULL,
  PRIMARY KEY (id),
  UNIQUE KEY uk_users_username (username),
  CONSTRAINT chk_users_username_format
    CHECK (username REGEXP '^[A-Za-z0-9_]{4,20}$')
) ENGINE=InnoDB;

-- Teams
CREATE TABLE IF NOT EXISTS teams (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  name VARCHAR(100) NOT NULL,
  owner_id BIGINT UNSIGNED NOT NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  deleted_at DATETIME NULL,
  PRIMARY KEY (id),
  CONSTRAINT fk_teams_owner_id FOREIGN KEY (owner_id) REFERENCES users(id),
  INDEX idx_teams_owner_id (owner_id)
) ENGINE=InnoDB;

-- Tasks (owned by user)
CREATE TABLE IF NOT EXISTS tasks (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  user_id BIGINT UNSIGNED NOT NULL,
  team_id BIGINT UNSIGNED NULL,
  assignee_id BIGINT UNSIGNED NULL,
  title VARCHAR(120) NOT NULL,
  description TEXT NULL,
  status ENUM('TODO', 'DOING', 'DONE', 'ERROR') NOT NULL DEFAULT 'TODO',
  priority TINYINT UNSIGNED NOT NULL DEFAULT 2,
  deadline DATETIME NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  deleted_at DATETIME NULL,
  PRIMARY KEY (id),
  CONSTRAINT fk_tasks_user_id FOREIGN KEY (user_id) REFERENCES users(id),
  CONSTRAINT fk_tasks_team_id FOREIGN KEY (team_id) REFERENCES teams(id),
  CONSTRAINT fk_tasks_assignee_id FOREIGN KEY (assignee_id) REFERENCES users(id),
  CONSTRAINT chk_tasks_priority CHECK (priority IN (1, 2, 3)),
  INDEX idx_tasks_user_deleted_created (user_id, deleted_at, created_at DESC),
  INDEX idx_tasks_user_status (user_id, status),
  INDEX idx_tasks_user_deadline (user_id, deadline),
  INDEX idx_tasks_user_priority (user_id, priority),
  INDEX idx_tasks_team_id (team_id),
  INDEX idx_tasks_assignee_id (assignee_id)
) ENGINE=InnoDB;

-- Team members
CREATE TABLE IF NOT EXISTS team_members (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  team_id BIGINT UNSIGNED NOT NULL,
  user_id BIGINT UNSIGNED NOT NULL,
  role ENUM('Owner', 'Admin', 'Member') NOT NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  deleted_at DATETIME NULL,
  PRIMARY KEY (id),
  CONSTRAINT fk_team_members_team_id FOREIGN KEY (team_id) REFERENCES teams(id),
  CONSTRAINT fk_team_members_user_id FOREIGN KEY (user_id) REFERENCES users(id),
  UNIQUE KEY uk_team_members_team_user (team_id, user_id),
  INDEX idx_team_members_team_id (team_id),
  INDEX idx_team_members_user_id (user_id)
) ENGINE=InnoDB;

-- Optional audit log for task operations (recommended for collaboration traceability)
CREATE TABLE IF NOT EXISTS task_audit_logs (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  task_id BIGINT UNSIGNED NOT NULL,
  user_id BIGINT UNSIGNED NOT NULL,
  action ENUM('CREATE', 'UPDATE', 'DELETE', 'STATUS_CHANGE') NOT NULL,
  before_data JSON NULL,
  after_data JSON NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  CONSTRAINT fk_task_audit_task_id FOREIGN KEY (task_id) REFERENCES tasks(id),
  CONSTRAINT fk_task_audit_user_id FOREIGN KEY (user_id) REFERENCES users(id),
  INDEX idx_task_audit_task_created (task_id, created_at DESC),
  INDEX idx_task_audit_user_created (user_id, created_at DESC)
) ENGINE=InnoDB;
