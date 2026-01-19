-- Migration: Add advanced features tables
-- Version: 003
-- Description: Add tables for EPSS vulnerability prioritization, cost intelligence, and optimization

-- Vulnerability Intelligence Table
CREATE TABLE IF NOT EXISTS vulnerability_intelligence (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    cve_id VARCHAR(50) UNIQUE NOT NULL,
    epss_score DECIMAL(5,4),
    epss_percentile DECIMAL(5,4),
    has_active_exploit BOOLEAN DEFAULT false,
    exploit_maturity VARCHAR(50),
    trending_score INT DEFAULT 0,
    last_updated TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_vuln_intel_cve ON vulnerability_intelligence(cve_id);
CREATE INDEX idx_vuln_intel_epss ON vulnerability_intelligence(epss_score DESC);

-- Manifest Vulnerability Priority Table
CREATE TABLE IF NOT EXISTS manifest_vuln_priority (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    manifest_id UUID REFERENCES manifests(id) ON DELETE CASCADE,
    cve_id VARCHAR(50) NOT NULL,
    base_severity VARCHAR(20),
    epss_score DECIMAL(5,4),
    runtime_exposed BOOLEAN DEFAULT false,
    priority_score INT, -- 0-100, AI-calculated
    recommended_action VARCHAR(50), -- 'urgent', 'high', 'medium', 'low', 'monitor'
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_manifest_vuln_priority_manifest ON manifest_vuln_priority(manifest_id);
CREATE INDEX idx_manifest_vuln_priority_score ON manifest_vuln_priority(priority_score DESC);

-- Storage Costs Table
CREATE TABLE IF NOT EXISTS storage_costs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    manifest_id UUID REFERENCES manifests(id) ON DELETE CASCADE,
    blob_size_bytes BIGINT NOT NULL,
    storage_cost_usd DECIMAL(10,4),
    bandwidth_cost_usd DECIMAL(10,4),
    total_cost_usd DECIMAL(10,4),
    pull_count_30d INT DEFAULT 0,
    last_pulled_at TIMESTAMP,
    cost_per_pull DECIMAL(10,6),
    calculated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_storage_costs_manifest ON storage_costs(manifest_id);
CREATE INDEX idx_storage_costs_total ON storage_costs(total_cost_usd DESC);

-- Zombie Images Table
CREATE TABLE IF NOT EXISTS zombie_images (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    manifest_id UUID REFERENCES manifests(id) ON DELETE CASCADE,
    days_since_last_pull INT NOT NULL,
    storage_cost_usd DECIMAL(10,4),
    recommended_action VARCHAR(50), -- 'delete', 'archive', 'keep'
    detected_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_zombie_images_manifest ON zombie_images(manifest_id);
CREATE INDEX idx_zombie_images_days ON zombie_images(days_since_last_pull DESC);

-- Cost Savings Opportunities Table
CREATE TABLE IF NOT EXISTS cost_savings_opportunities (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    opportunity_type VARCHAR(50) NOT NULL, -- 'zombie_cleanup', 'deduplication', 'compression'
    estimated_savings_usd DECIMAL(10,2),
    affected_images INT,
    implementation_effort VARCHAR(20), -- 'low', 'medium', 'high'
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_cost_savings_type ON cost_savings_opportunities(opportunity_type);
CREATE INDEX idx_cost_savings_amount ON cost_savings_opportunities(estimated_savings_usd DESC);

-- Optimization Suggestions Table
CREATE TABLE IF NOT EXISTS optimization_suggestions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    manifest_id UUID REFERENCES manifests(id) ON DELETE CASCADE,
    suggestion_type VARCHAR(50) NOT NULL, -- 'base_image', 'multi_stage', 'package_removal', 'layer_merge'
    current_state TEXT,
    recommended_change TEXT,
    estimated_size_reduction_mb INT,
    confidence_score INT, -- 0-100
    implementation_difficulty VARCHAR(20), -- 'easy', 'medium', 'hard'
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_optimization_manifest ON optimization_suggestions(manifest_id);
CREATE INDEX idx_optimization_type ON optimization_suggestions(suggestion_type);

-- Add last_pulled_at column to manifests table if it doesn't exist
DO $$ 
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'manifests' AND column_name = 'last_pulled_at'
    ) THEN
        ALTER TABLE manifests ADD COLUMN last_pulled_at TIMESTAMP;
    END IF;
END $$;

-- Add pull_count column to manifests table if it doesn't exist
DO $$ 
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'manifests' AND column_name = 'pull_count'
    ) THEN
        ALTER TABLE manifests ADD COLUMN pull_count INT DEFAULT 0;
    END IF;
END $$;

-- Create index on last_pulled_at for zombie detection
CREATE INDEX IF NOT EXISTS idx_manifests_last_pulled ON manifests(last_pulled_at);
