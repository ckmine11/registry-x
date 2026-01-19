-- RegistryX Initial Schema

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Users
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    username VARCHAR(100) NOT NULL UNIQUE,
    email VARCHAR(255) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Namespaces (Organizations or Users)
CREATE TABLE namespaces (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL UNIQUE,
    type VARCHAR(50) NOT NULL DEFAULT 'organization', -- 'user' or 'organization'
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Repositories (e.g., 'nginx', 'my-app/backend')
CREATE TABLE repositories (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    namespace_id UUID NOT NULL REFERENCES namespaces(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (namespace_id, name)
);

-- Blobs (Content addressable storage tracker)
CREATE TABLE blobs (
    digest VARCHAR(255) PRIMARY KEY, -- sha256:...
    size BIGINT NOT NULL,
    media_type VARCHAR(255) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Manifests (The image definition)
CREATE TABLE manifests (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    repository_id UUID NOT NULL REFERENCES repositories(id) ON DELETE CASCADE,
    digest VARCHAR(255) NOT NULL,
    config_digest VARCHAR(255), -- Reference to the config blob
    media_type VARCHAR(255) NOT NULL,
    size BIGINT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (repository_id, digest)
);

-- Tags (Mutable references to manifests)
CREATE TABLE tags (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    repository_id UUID NOT NULL REFERENCES repositories(id) ON DELETE CASCADE,
    manifest_id UUID NOT NULL REFERENCES manifests(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (repository_id, name)
);

-- Manifest Layers (Join table for Manifest -> Blobs)
CREATE TABLE manifest_layers (
    manifest_id UUID NOT NULL REFERENCES manifests(id) ON DELETE CASCADE,
    blob_digest VARCHAR(255) NOT NULL REFERENCES blobs(digest) ON DELETE CASCADE,
    position INT NOT NULL, -- Layer order
    PRIMARY KEY (manifest_id, blob_digest, position)
);

-- Vulnerability Reports (Scan results)
CREATE TABLE vulnerability_reports (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    manifest_id UUID NOT NULL REFERENCES manifests(id) ON DELETE CASCADE,
    scanner VARCHAR(50) NOT NULL, -- 'trivy', 'grype'
    scanned_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    critical_count INT DEFAULT 0,
    high_count INT DEFAULT 0,
    medium_count INT DEFAULT 0,
    low_count INT DEFAULT 0,
    report_json JSONB, -- Full scan result for detailed display
    status VARCHAR(50) DEFAULT 'completed' -- 'pending', 'scanning', 'completed', 'failed'
);

-- Indexes for performance
CREATE INDEX idx_repositories_namespace_id ON repositories(namespace_id);
CREATE INDEX idx_tags_repository_id ON tags(repository_id);
CREATE INDEX idx_manifests_repository_id ON manifests(repository_id);
CREATE INDEX idx_vulnerability_reports_manifest_id ON vulnerability_reports(manifest_id);

-- Service Accounts (API Keys)
CREATE TABLE service_accounts (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    api_key_hash VARCHAR(255) NOT NULL, -- Storing hash of the key
    prefix VARCHAR(50) NOT NULL,        -- Storing prefix for display/lookup
    status VARCHAR(50) DEFAULT 'active', -- 'active', 'revoked'
    last_used_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

