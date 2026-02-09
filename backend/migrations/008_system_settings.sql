-- Add system settings table for license and other persistent configs
CREATE TABLE IF NOT EXISTS system_settings (
    key VARCHAR(255) PRIMARY KEY,
    value TEXT NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Insert a placeholder for license key if not already present
INSERT INTO system_settings (key, value)
VALUES ('license_key', '')
ON CONFLICT (key) DO NOTHING;
