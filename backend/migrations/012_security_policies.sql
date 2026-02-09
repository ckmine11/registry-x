-- Migration: Add Security Policies
-- Version: 010
-- Description: Tables for Global and Repository-specific Security Policies

CREATE TABLE IF NOT EXISTS security_policies (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    critical_threshold INTEGER DEFAULT 0,
    high_threshold INTEGER DEFAULT 0,
    medium_threshold INTEGER DEFAULT 0,
    low_threshold INTEGER DEFAULT 0,
    block_unscanned BOOLEAN DEFAULT FALSE,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS repository_policies (
    repository VARCHAR(255) PRIMARY KEY,
    critical_threshold INTEGER DEFAULT 0,
    high_threshold INTEGER DEFAULT 0,
    medium_threshold INTEGER DEFAULT 0,
    low_threshold INTEGER DEFAULT 0,
    block_unscanned BOOLEAN DEFAULT FALSE,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Insert default global policy if it doesn't exist
INSERT INTO security_policies (critical_threshold, high_threshold, low_threshold, medium_threshold, block_unscanned)
SELECT 0, 0, 0, 0, false
WHERE NOT EXISTS (SELECT 1 FROM security_policies);
