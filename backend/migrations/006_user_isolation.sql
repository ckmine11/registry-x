-- 006_user_isolation.sql
-- Add role to users for access control
ALTER TABLE users ADD COLUMN IF NOT EXISTS role VARCHAR(50) DEFAULT 'user';

-- Add owner_id to namespaces to link them to users
ALTER TABLE namespaces ADD COLUMN IF NOT EXISTS owner_id UUID REFERENCES users(id);
CREATE INDEX IF NOT EXISTS idx_namespaces_owner ON namespaces(owner_id);

-- Migration logic:
-- 1. If username = 'admin', set role = 'admin'
UPDATE users SET role = 'admin' WHERE username = 'admin';

-- 2. If a namespace exists with the same name as a user, link them
UPDATE namespaces 
SET owner_id = users.id 
FROM users 
WHERE namespaces.name = users.username;
