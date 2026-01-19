-- Check current library namespace owner
SELECT id, name, owner_id FROM namespaces WHERE name = 'library';

-- If library exists but has no owner, we need to either:
-- 1. Make it public by setting a special flag, OR
-- 2. Assign it to a specific user

-- For now, let's check what users exist
SELECT id, username, role FROM users;
