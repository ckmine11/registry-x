-- 008_audit_logs.sql
CREATE TABLE audit_logs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    action VARCHAR(50) NOT NULL, -- 'LOGIN', 'PUSH', 'DELETE'
    repository_id UUID REFERENCES repositories(id) ON DELETE SET NULL,
    details JSONB, -- { "tag": "v1", "digest": "sha256:...", "ip": "1.2.3.4" }
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_audit_logs_user_id ON audit_logs(user_id);
CREATE INDEX idx_audit_logs_repo_id ON audit_logs(repository_id);
CREATE INDEX idx_audit_logs_created_at ON audit_logs(created_at DESC);
