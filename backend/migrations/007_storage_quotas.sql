-- 007_storage_quotas.sql
-- Add quota to namespaces (Default 5GB)
ALTER TABLE namespaces ADD COLUMN IF NOT EXISTS quota_bytes BIGINT DEFAULT 5368709120;
