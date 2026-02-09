-- Add role column to users table (RBAC)
ALTER TABLE users ADD COLUMN IF NOT EXISTS role VARCHAR(50) DEFAULT 'developer';

-- Create Webhooks table (Notifications)
CREATE TABLE IF NOT EXISTS webhooks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    url TEXT NOT NULL,
    type VARCHAR(50) NOT NULL DEFAULT 'slack', -- slack, discord, generic
    events TEXT[] NOT NULL DEFAULT '{}', -- SCAN_COMPLETED, CRITICAL_VULN, POLICY_VIOLATION
    enabled BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Index for faster lookups
CREATE INDEX IF NOT EXISTS idx_webhooks_enabled ON webhooks(enabled);
