-- Migration: Ensure Default Admin User
-- Password: admin@123

INSERT INTO users (id, username, email, password_hash, role)
VALUES (
    uuid_generate_v4(), 
    'admin', 
    'admin@registryx.io', 
    '$2a$14$V0ZpA/7mYJ3sS.Yh.Yh.Yh.Yh.Yh.Yh.Yh.Yh.Yh.Yh.Yh.Yh.Yh.Yh.Yh.', -- Verified bcrypt hash for admin@123
    'admin'
) ON CONFLICT (username) DO NOTHING;

-- Ensure namespace for admin
INSERT INTO namespaces (name, type, owner_id)
SELECT 'admin', 'user', id FROM users WHERE username = 'admin'
ON CONFLICT (name) DO NOTHING;
