-- Create manifest_health table for storing health scores
CREATE TABLE IF NOT EXISTS manifest_health (
    manifest_id UUID PRIMARY KEY REFERENCES manifests(id) ON DELETE CASCADE,
    overall_score INTEGER NOT NULL,
    grade VARCHAR(5) NOT NULL,
    calculated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_manifest_health_score ON manifest_health(overall_score);
CREATE INDEX idx_manifest_health_grade ON manifest_health(grade);
