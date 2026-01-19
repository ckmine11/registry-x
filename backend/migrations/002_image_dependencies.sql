-- Add health score columns if they don't exist (Idempotent)
DO $$ 
BEGIN 
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='manifests' AND column_name='health_score') THEN
        ALTER TABLE manifests ADD COLUMN health_score INT DEFAULT 0;
        ALTER TABLE manifests ADD COLUMN health_grade VARCHAR(5) DEFAULT 'F';
        ALTER TABLE manifests ADD COLUMN health_security INT DEFAULT 0;
        ALTER TABLE manifests ADD COLUMN health_freshness INT DEFAULT 0;
        ALTER TABLE manifests ADD COLUMN health_efficiency INT DEFAULT 0;
        ALTER TABLE manifests ADD COLUMN health_maintenance INT DEFAULT 0;
        ALTER TABLE manifests ADD COLUMN last_health_check TIMESTAMP;
    END IF;
END $$;

-- Image Dependencies Table
CREATE TABLE IF NOT EXISTS image_dependencies (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    manifest_id UUID NOT NULL REFERENCES manifests(id) ON DELETE CASCADE,
    parent_manifest_id UUID NOT NULL REFERENCES manifests(id) ON DELETE CASCADE,
    detected_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(manifest_id, parent_manifest_id)
);

CREATE INDEX IF NOT EXISTS idx_image_dependencies_manifest_id ON image_dependencies(manifest_id);
CREATE INDEX IF NOT EXISTS idx_image_dependencies_parent_id ON image_dependencies(parent_manifest_id);
