CREATE TABLE IF NOT EXISTS `users` (
    `id` integer
  , `username` text NOT NULL
  , `password_hash` text NOT NULL
  , `role` text NOT NULL DEFAULT 'viewer'
  , `created_at` datetime
  , `updated_at` datetime
  , PRIMARY KEY(`id`)
  , UNIQUE(`username`)
);

-- username: admin, password: secret
INSERT INTO users(id, username, password_hash, role, created_at, updated_at)
VALUES(1, 'admin', '$2a$10$HNLoqiTO5oLwopczA/wcPOebfO79S.hnAA5HOkx5p6o3g5a2E30v2', 'admin', datetime('now'), datetime('now'));
