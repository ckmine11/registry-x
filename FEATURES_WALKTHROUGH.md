# RegistryX - Complete Features Walkthrough

## 📋 Table of Contents
1. [Project Overview](#project-overview)
2. [Architecture & Technology Stack](#architecture--technology-stack)
3. [Core Features](#core-features)
4. [Advanced Security Features](#advanced-security-features)
5. [Cost Intelligence & Optimization](#cost-intelligence--optimization)
6. [User Management & Authentication](#user-management--authentication)
7. [API Reference](#api-reference)
8. [Database Schema](#database-schema)
9. [Deployment Options](#deployment-options)
10. [Testing & Verification](#testing--verification)

---

## 🎯 Project Overview

**RegistryX** is a production-ready, OCI-compliant container registry that goes far beyond simple image storage. It's designed for enterprises that need:

- **Security-First Approach**: Real-time vulnerability scanning with EPSS-based prioritization
- **Cost Optimization**: Intelligent cost tracking and zombie image detection
- **Enterprise Features**: Multi-tenant isolation, RBAC, audit logging, and webhooks
- **Developer Experience**: Modern React UI with real-time updates and comprehensive API

### Key Differentiators
- ✅ **Smart Vulnerability Prioritization** using EPSS (Exploit Prediction Scoring System)
- ✅ **Cost Intelligence** with zombie image detection and cleanup
- ✅ **Repository-Specific Security Policies** with global defaults
- ✅ **Image Signing Support** with Cosign integration
- ✅ **Complete Multi-Tenant Isolation** at the database level
- ✅ **Real-time Webhooks** for CI/CD integration
- ✅ **Comprehensive Audit Logging** for compliance

---

## 🏗️ Architecture & Technology Stack

### Backend Stack
```
Language: Go 1.22
Framework: Gorilla Mux (REST API)
Database: PostgreSQL 16
Cache/Queue: Redis 7
Storage: MinIO (S3-compatible)
Scanner: Trivy (latest)
```

### Frontend Stack
```
Framework: React 18 + TypeScript
Build Tool: Vite
Styling: TailwindCSS
State Management: TanStack Query
Routing: React Router v6
UI Components: Lucide Icons
Charts: ReactFlow (dependency graphs)
PDF Generation: jsPDF
```

### Infrastructure Components
```
├── PostgreSQL 16      → Metadata, users, scan results
├── Redis 7            → Queue management, session storage
├── MinIO              → Blob storage (layers, manifests)
├── Trivy              → Vulnerability scanning engine
└── Nginx (optional)   → Reverse proxy for production
```

### Microservices Architecture

```
backend/pkg/
├── api/              → REST API handlers
├── auth/             → Authentication & authorization
├── audit/            → Audit logging service
├── config/           → Configuration management
├── costs/            → Cost intelligence service
├── database/         → Database connection & migrations
├── email/            → Email notification service
├── epss/             → EPSS data fetching client
├── health/           → Health score calculation
├── intelligence/     → Vulnerability prioritization
├── metadata/         → Image metadata service
├── middleware/       → HTTP middleware (auth, CORS)
├── policy/           → OPA policy engine integration
├── queue/            → Redis queue management
├── registry/         → OCI registry protocol handlers
├── reports/          → PDF report generation
├── scanner/          → Trivy integration
├── storage/          → S3/MinIO storage driver
└── webhook/          → Webhook notification service
```

---

## 🎨 Core Features

### 1. OCI-Compliant Registry

**Full Docker Registry V2 API Support**

RegistryX implements the complete OCI Distribution Specification:

#### Supported Operations
- ✅ **Image Push/Pull**: Standard `docker push` and `docker pull`
- ✅ **Manifest Management**: V2 Schema 2 and OCI manifests
- ✅ **Blob Storage**: Content-addressable layer storage
- ✅ **Tag Management**: Mutable tag references
- ✅ **Catalog API**: Repository listing
- ✅ **Blob Uploads**: Chunked and monolithic uploads

#### API Endpoints
```
GET  /v2/                              → Base check
GET  /v2/_catalog                      → List repositories
GET  /v2/{name}/tags/list              → List tags
GET  /v2/{name}/manifests/{reference}  → Get manifest
PUT  /v2/{name}/manifests/{reference}  → Push manifest
GET  /v2/{name}/blobs/{digest}         → Get blob
POST /v2/{name}/blobs/uploads/         → Start upload
PUT  /v2/{name}/blobs/uploads/{uuid}   → Complete upload
```

#### Example Usage
```bash
# Login
docker login localhost:5000
Username: admin
Password: password123

# Tag and push
docker tag nginx:latest localhost:5000/library/nginx:v1.0
docker push localhost:5000/library/nginx:v1.0

# Pull
docker pull localhost:5000/library/nginx:v1.0

# List repositories
curl http://localhost:5000/v2/_catalog

# List tags
curl http://localhost:5000/v2/library/nginx/tags/list
```

---

### 2. Modern Web Dashboard

**React-Based UI with Real-Time Updates**

#### Pages & Features

##### 📊 Dashboard (`/dashboard`)
- **Real-time Statistics**
  - Total repositories count
  - Total images count
  - Storage usage (GB)
  - Vulnerability summary (Critical/High/Medium/Low)
- **Recent Activity Feed**
  - Image pushes
  - Scan completions
  - Policy violations
- **Quick Actions**
  - Create repository
  - View scan reports
  - Access settings

##### 📦 Repositories (`/repositories`)
- **Repository Grid View**
  - Repository name and namespace
  - Image count per repository
  - Last updated timestamp
  - Vulnerability badge
- **Search & Filter**
  - Search by name
  - Filter by namespace
  - Sort by date/name
- **Actions**
  - View details
  - Delete repository
  - Create new repository

##### 🔍 Repository Details (`/repositories/:name`)
- **Tag Management**
  - List all tags
  - View manifest details
  - Delete tags
  - Copy pull command
- **Vulnerability Scanning**
  - Scan status indicator
  - Vulnerability breakdown
  - Download Trivy JSON report
  - Manual scan trigger
  - Scan history timeline
- **Image Information**
  - Digest (SHA256)
  - Size (MB/GB)
  - Created date
  - Pull count
  - Last pulled timestamp
- **Layer Details**
  - Layer digests
  - Layer sizes
  - Layer commands
- **Signature Status**
  - Cosign signature detection
  - Verification status

##### 🛡️ Security Policies (`/policies`)
- **Global Security Policy**
  - Minimum severity threshold
  - Maximum allowed vulnerabilities
  - Block unsigned images
  - Require image signing
- **Repository-Specific Overrides**
  - Custom policies per repository
  - Override global settings
  - Policy inheritance
- **Policy Preview**
  - Test policy against images
  - View affected images

##### 💰 Cost Intelligence (`/costs`)
- **Cost Dashboard**
  - Total monthly cost
  - Storage costs breakdown
  - Bandwidth costs
  - Cost trends
- **Zombie Image Detection**
  - Images not pulled in 90+ days
  - Potential savings calculation
  - One-click cleanup
- **Cost Optimization Suggestions**
  - Large image identification
  - Deduplication opportunities
  - Compression recommendations

##### 🔗 Dependency Graph (`/lineage`)
- **Visual Dependency Mapping**
  - Base image relationships
  - Layer sharing visualization
  - Interactive graph navigation
- **Impact Analysis**
  - See which images are affected by base image vulnerabilities

##### ⚙️ Settings (`/settings`)
- **User Management**
  - List users
  - Invite users (email)
  - Update roles (admin/user)
  - Delete users
- **Service Accounts**
  - Create API keys
  - Revoke tokens
  - View last used
- **Webhooks**
  - Configure webhook endpoints
  - Event types (push, scan, delete)
  - Test webhooks
- **Notifications**
  - Email notifications
  - Slack integration
  - Webhook configuration
- **System Configuration**
  - Storage quotas
  - Retention policies
  - Garbage collection

##### 👤 User Profile (`/profile`)
- **Account Information**
  - Username and email
  - Account creation date
  - Last login
- **Security**
  - Change password
  - View active sessions
  - Revoke sessions
- **Audit Log**
  - View personal activity
  - Download audit reports

##### 🔐 Sessions (`/sessions`)
- **Active Session Management**
  - View all active sessions
  - Session details (IP, user agent, location)
  - Last activity timestamp
  - Revoke individual sessions
  - Revoke all sessions

---

## 🛡️ Advanced Security Features

### 1. Real-Time Vulnerability Scanning

**Powered by Trivy**

#### How It Works
1. **Automatic Trigger**: When an image is pushed, a scan job is queued
2. **Background Processing**: Redis queue processes scan jobs asynchronously
3. **Trivy Execution**: Trivy scans the image layers for vulnerabilities
4. **Result Storage**: Scan results stored in PostgreSQL with full JSON report
5. **Notification**: Webhooks triggered on scan completion

#### Scan Status Flow
```
pending → scanning → completed/failed
```

#### API Endpoints

**Get Scan Status**
```bash
GET /api/v1/repositories/{name}/manifests/{reference}/scan/status

Response:
{
  "status": "completed",
  "scanned_at": "2026-02-07T16:30:00Z",
  "summary": {
    "critical": 5,
    "high": 12,
    "medium": 23,
    "low": 8
  }
}
```

**Download Trivy Report**
```bash
GET /api/v1/repositories/{name}/manifests/{reference}/scan/report

# Downloads: trivy-report-{repository}-{reference}.json
```

**View Scan History**
```bash
GET /api/v1/repositories/{name}/manifests/{reference}/scan/history

Response:
{
  "scans": [
    {
      "id": "uuid-1",
      "status": "completed",
      "scanned_at": "2026-02-07T16:30:00Z",
      "summary": { "critical": 5, "high": 12, ... }
    }
  ]
}
```

**Trigger Manual Scan**
```bash
POST /api/v1/repositories/{name}/manifests/{reference}/scan/trigger

Response:
{
  "message": "Scan triggered successfully",
  "status": "scanning"
}
```

#### Database Schema
```sql
CREATE TABLE vulnerability_reports (
    id UUID PRIMARY KEY,
    manifest_id UUID REFERENCES manifests(id),
    scanner VARCHAR(50),              -- 'trivy'
    status VARCHAR(50),                -- 'pending', 'scanning', 'completed', 'failed'
    scanned_at TIMESTAMP,
    critical_count INT,
    high_count INT,
    medium_count INT,
    low_count INT,
    report_json JSONB                  -- Full Trivy output
);
```

---

### 2. EPSS-Based Vulnerability Prioritization

**Smart Threat Intelligence**

#### What is EPSS?
EPSS (Exploit Prediction Scoring System) predicts the likelihood of a CVE being exploited in the wild within the next 30 days. Unlike CVSS (which measures severity), EPSS measures **exploitability**.

#### How RegistryX Uses EPSS

1. **Daily EPSS Data Refresh**
   - Background worker fetches latest EPSS scores from FIRST.org
   - Updates `vulnerability_intelligence` table
   - Runs every 24 hours

2. **Priority Score Calculation**
   ```
   Priority Score (0-100) = 
     Base Severity Weight (40%) +
     EPSS Score Weight (40%) +
     Exploit Maturity Weight (20%)
   ```

3. **Recommended Actions**
   - **Urgent** (90-100): Active exploit, high EPSS, critical severity
   - **High** (70-89): High EPSS or critical severity
   - **Medium** (40-69): Moderate EPSS, high severity
   - **Low** (20-39): Low EPSS, medium severity
   - **Monitor** (0-19): Low risk, track for changes

#### API Endpoints

**Get Prioritized Vulnerabilities**
```bash
GET /api/v1/vulnerabilities/prioritized?manifest_id={uuid}

Response:
{
  "vulnerabilities": [
    {
      "cve_id": "CVE-2024-1234",
      "base_severity": "CRITICAL",
      "epss_score": 0.9523,
      "epss_percentile": 0.9987,
      "priority_score": 95,
      "recommended_action": "urgent",
      "has_active_exploit": true
    }
  ]
}
```

**Get CVE Intelligence**
```bash
GET /api/v1/vulnerabilities/intelligence/CVE-2024-1234

Response:
{
  "cve_id": "CVE-2024-1234",
  "epss_score": 0.9523,
  "epss_percentile": 0.9987,
  "has_active_exploit": true,
  "exploit_maturity": "functional",
  "trending_score": 85,
  "last_updated": "2026-02-07T00:00:00Z"
}
```

**Refresh EPSS Data**
```bash
POST /api/v1/vulnerabilities/refresh-epss

Response:
{
  "message": "EPSS refresh started",
  "status": "processing"
}
```

#### Database Schema
```sql
CREATE TABLE vulnerability_intelligence (
    id UUID PRIMARY KEY,
    cve_id VARCHAR(50) UNIQUE,
    epss_score DECIMAL(5,4),           -- 0.0000 to 1.0000
    epss_percentile DECIMAL(5,4),      -- 0.0000 to 1.0000
    has_active_exploit BOOLEAN,
    exploit_maturity VARCHAR(50),      -- 'unproven', 'poc', 'functional', 'high'
    trending_score INT,                -- 0-100
    last_updated TIMESTAMP
);

CREATE TABLE manifest_vuln_priority (
    id UUID PRIMARY KEY,
    manifest_id UUID REFERENCES manifests(id),
    cve_id VARCHAR(50),
    base_severity VARCHAR(20),         -- 'CRITICAL', 'HIGH', 'MEDIUM', 'LOW'
    epss_score DECIMAL(5,4),
    runtime_exposed BOOLEAN,
    priority_score INT,                -- 0-100
    recommended_action VARCHAR(50)     -- 'urgent', 'high', 'medium', 'low', 'monitor'
);
```

---

### 3. Image Signing with Cosign

**Supply Chain Security**

#### Supported Signing Methods
- ✅ **Cosign Key-Based Signing**
- ✅ **Cosign Keyless Signing** (Fulcio/Rekor)
- ✅ **Automatic Signature Detection**

#### How It Works

1. **Sign Image with Cosign**
   ```bash
   # Generate keys (one-time)
   cosign generate-key-pair
   
   # Sign image
   cosign sign --key cosign.key localhost:5000/library/nginx:latest
   ```

2. **Signature Storage**
   - Cosign creates a signature tag: `sha256-{digest}.sig`
   - RegistryX detects this pattern automatically
   - UI displays "Signed" badge

3. **Verification**
   ```bash
   cosign verify --key cosign.pub localhost:5000/library/nginx:latest
   ```

#### Automated Signing Script
```powershell
# scripts/sign_images.ps1
# Signs all images in the registry automatically
.\scripts\sign_images.ps1
```

#### Policy Enforcement
```rego
# Require signed images in production
package registry

deny[msg] {
    input.operation == "pull"
    input.environment == "production"
    not input.signed
    msg := "Only signed images allowed in production"
}
```

---

### 4. Repository-Specific Security Policies

**Granular Policy Control**

#### Policy Hierarchy
```
Global Policy (Default)
    ↓
Repository Override (Optional)
    ↓
Final Policy Applied
```

#### Global Security Policy

**Configuration**
```json
{
  "min_severity": "HIGH",
  "max_critical": 0,
  "max_high": 5,
  "block_unsigned": false,
  "require_scan": true
}
```

**API Endpoints**
```bash
# Get global policy
GET /api/v1/system/security/policy

# Update global policy
PUT /api/v1/system/security/policy
Content-Type: application/json

{
  "min_severity": "CRITICAL",
  "max_critical": 0,
  "max_high": 0,
  "block_unsigned": true
}
```

#### Repository-Specific Overrides

**Use Cases**
- **Development Repos**: More lenient policies
- **Production Repos**: Strict policies
- **Third-Party Images**: Custom rules

**API Endpoints**
```bash
# List all overrides
GET /api/v1/system/security/policy/overrides

# Create/Update override
POST /api/v1/system/security/policy/overrides
Content-Type: application/json

{
  "repository": "library/nginx",
  "policy": {
    "min_severity": "MEDIUM",
    "max_critical": 2,
    "max_high": 10,
    "block_unsigned": false
  }
}

# Delete override
DELETE /api/v1/system/security/policy/overrides/{repository}
```

#### Database Schema
```sql
CREATE TABLE security_policies (
    id UUID PRIMARY KEY,
    repository_id UUID REFERENCES repositories(id),  -- NULL for global
    min_severity VARCHAR(20),
    max_critical INT,
    max_high INT,
    max_medium INT,
    block_unsigned BOOLEAN,
    require_scan BOOLEAN,
    created_at TIMESTAMP,
    updated_at TIMESTAMP
);
```

---

## 💰 Cost Intelligence & Optimization

### 1. Real-Time Cost Tracking

**Comprehensive Cost Analytics**

#### Cost Calculation Model

**Storage Costs**
```
Storage Cost = (Image Size in GB) × (Storage Rate per GB/month) × (Days Stored / 30)
```

**Bandwidth Costs**
```
Bandwidth Cost = (Pull Count) × (Image Size in GB) × (Bandwidth Rate per GB)
```

**Total Cost**
```
Total Cost = Storage Cost + Bandwidth Cost
```

#### Default Pricing (Configurable)
```
Storage: $0.023 per GB/month
Bandwidth: $0.09 per GB
```

#### API Endpoints

**Get Cost Dashboard**
```bash
GET /api/v1/costs/dashboard

Response:
{
  "total_cost_usd": 45.67,
  "storage_cost_usd": 23.45,
  "bandwidth_cost_usd": 22.22,
  "total_images": 150,
  "total_storage_gb": 1024.5,
  "zombie_images_count": 23,
  "potential_savings_usd": 12.34,
  "top_expensive_images": [
    {
      "repository": "library/nginx",
      "tag": "latest",
      "cost_usd": 5.67,
      "size_gb": 2.3
    }
  ]
}
```

**Refresh Costs**
```bash
POST /api/v1/costs/refresh

Response:
{
  "message": "Cost recalculation started",
  "status": "processing"
}
```

---

### 2. Zombie Image Detection

**Identify Unused Images**

#### Detection Criteria
- Image not pulled in **90+ days**
- Storage cost > $0
- No active tags (optional)

#### API Endpoints

**Get Zombie Images**
```bash
GET /api/v1/costs/zombie-images

Response:
{
  "zombies": [
    {
      "manifest_id": "uuid-1",
      "repository": "library/old-app",
      "tag": "v1.0",
      "days_since_last_pull": 120,
      "storage_cost_usd": 2.34,
      "size_gb": 1.5,
      "recommended_action": "delete"
    }
  ],
  "total_zombies": 23,
  "total_savings_usd": 45.67
}
```

**Cleanup Zombies**
```bash
POST /api/v1/costs/cleanup-zombies

Response:
{
  "message": "Cleanup started",
  "deleted_count": 23,
  "reclaimed_storage_gb": 34.5,
  "savings_usd": 45.67
}
```

#### Database Schema
```sql
CREATE TABLE zombie_images (
    id UUID PRIMARY KEY,
    manifest_id UUID REFERENCES manifests(id),
    days_since_last_pull INT,
    storage_cost_usd DECIMAL(10,4),
    recommended_action VARCHAR(50),
    detected_at TIMESTAMP
);

CREATE TABLE storage_costs (
    id UUID PRIMARY KEY,
    manifest_id UUID REFERENCES manifests(id),
    blob_size_bytes BIGINT,
    storage_cost_usd DECIMAL(10,4),
    bandwidth_cost_usd DECIMAL(10,4),
    total_cost_usd DECIMAL(10,4),
    pull_count_30d INT,
    last_pulled_at TIMESTAMP,
    cost_per_pull DECIMAL(10,6),
    calculated_at TIMESTAMP
);
```

---

### 3. Optimization Suggestions

**AI-Powered Recommendations**

#### Suggestion Types

1. **Base Image Optimization**
   - Recommend Alpine/Distroless variants
   - Identify outdated base images

2. **Multi-Stage Build Detection**
   - Detect single-stage builds
   - Suggest multi-stage optimization

3. **Package Removal**
   - Identify unnecessary packages
   - Suggest cleanup commands

4. **Layer Merging**
   - Detect excessive layers
   - Recommend layer consolidation

#### Database Schema
```sql
CREATE TABLE optimization_suggestions (
    id UUID PRIMARY KEY,
    manifest_id UUID REFERENCES manifests(id),
    suggestion_type VARCHAR(50),
    current_state TEXT,
    recommended_change TEXT,
    estimated_size_reduction_mb INT,
    confidence_score INT,              -- 0-100
    implementation_difficulty VARCHAR(20),
    created_at TIMESTAMP
);
```

---

## 👥 User Management & Authentication

### 1. Multi-Tenant Isolation

**Complete Data Segregation**

#### Isolation Levels

1. **Namespace Isolation**
   - Each user has a personal namespace
   - Organizations can have shared namespaces

2. **Repository Isolation**
   - Users can only see their own repositories
   - Admin users can see all repositories

3. **Database-Level Isolation**
   ```sql
   -- All queries filtered by user ownership
   SELECT * FROM repositories 
   WHERE namespace_id IN (
     SELECT id FROM namespaces WHERE owner_id = current_user_id
   );
   ```

#### User Roles

**Admin**
- Full system access
- User management
- Global policy configuration
- All repositories visible

**User**
- Personal namespace access
- Own repositories only
- Limited settings access

---

### 2. Authentication System

**JWT-Based Authentication**

#### Login Flow
```
1. User submits credentials
2. Backend validates against PostgreSQL
3. JWT token generated (24h expiry)
4. Token stored in Redis for session management
5. Frontend stores token in localStorage
6. Token included in Authorization header for API calls
```

#### API Endpoints

**Register**
```bash
POST /api/v1/auth/register
Content-Type: application/json

{
  "username": "john",
  "email": "john@example.com",
  "password": "SecurePass123!"
}

Response:
{
  "message": "User created successfully",
  "user_id": "uuid-1"
}
```

**Login**
```bash
POST /api/v1/auth/login
Content-Type: application/json

{
  "username": "john",
  "password": "SecurePass123!"
}

Response:
{
  "token": "eyJhbGciOiJIUzI1NiIs...",
  "user": {
    "id": "uuid-1",
    "username": "john",
    "email": "john@example.com",
    "role": "user"
  }
}
```

**Logout**
```bash
POST /api/v1/auth/logout
Authorization: Bearer {token}

Response:
{
  "message": "Logged out successfully"
}
```

---

### 3. Password Reset Flow

**Email-Based Recovery**

#### Flow
```
1. User requests password reset
2. Backend generates reset token
3. Email sent with reset link
4. User clicks link with token
5. User sets new password
6. Token invalidated
```

#### API Endpoints

**Forgot Password**
```bash
POST /api/v1/auth/forgot-password
Content-Type: application/json

{
  "email": "john@example.com"
}

Response:
{
  "message": "Password reset email sent"
}
```

**Reset Password**
```bash
POST /api/v1/auth/reset-with-key
Content-Type: application/json

{
  "reset_key": "abc123...",
  "new_password": "NewSecurePass123!"
}

Response:
{
  "message": "Password reset successful"
}
```

#### Database Schema
```sql
CREATE TABLE password_reset_tokens (
    id UUID PRIMARY KEY,
    user_id UUID REFERENCES users(id),
    token VARCHAR(255) UNIQUE,
    expires_at TIMESTAMP,
    used BOOLEAN DEFAULT false,
    created_at TIMESTAMP
);
```

---

### 4. Session Management

**Redis-Backed Sessions**

#### Features
- Track active sessions per user
- View session details (IP, user agent, location)
- Revoke individual sessions
- Revoke all sessions (force logout)

#### API Endpoints

**Get Active Sessions**
```bash
GET /api/v1/system/sessions
Authorization: Bearer {token}

Response:
{
  "sessions": [
    {
      "id": "session-1",
      "user_id": "uuid-1",
      "ip_address": "192.168.1.100",
      "user_agent": "Mozilla/5.0...",
      "created_at": "2026-02-07T10:00:00Z",
      "last_activity": "2026-02-07T16:30:00Z"
    }
  ]
}
```

**Revoke Session**
```bash
DELETE /api/v1/system/sessions/{session_id}
Authorization: Bearer {token}

Response:
{
  "message": "Session revoked successfully"
}
```

---

### 5. Service Accounts (API Keys)

**Machine-to-Machine Authentication**

#### Features
- Generate API keys for CI/CD
- Prefix-based key format: `rgx_live_...`
- Key rotation support
- Last used tracking

#### API Endpoints

**Create Service Account**
```bash
POST /api/v1/service-accounts
Authorization: Bearer {token}
Content-Type: application/json

{
  "name": "CI/CD Pipeline",
  "description": "GitHub Actions deployment"
}

Response:
{
  "id": "uuid-1",
  "name": "CI/CD Pipeline",
  "api_key": "rgx_live_abc123def456...",
  "prefix": "rgx_live_abc123",
  "created_at": "2026-02-07T16:30:00Z"
}
```

**List Service Accounts**
```bash
GET /api/v1/service-accounts
Authorization: Bearer {token}

Response:
{
  "accounts": [
    {
      "id": "uuid-1",
      "name": "CI/CD Pipeline",
      "prefix": "rgx_live_abc123",
      "status": "active",
      "last_used_at": "2026-02-07T15:00:00Z",
      "created_at": "2026-02-07T10:00:00Z"
    }
  ]
}
```

**Revoke Service Account**
```bash
DELETE /api/v1/service-accounts/{id}
Authorization: Bearer {token}

Response:
{
  "message": "Service account revoked"
}
```

---

### 6. Audit Logging

**Comprehensive Activity Tracking**

#### Logged Events
- User login/logout
- Image push/pull
- Repository creation/deletion
- Policy changes
- User management actions
- API key usage

#### API Endpoints

**Get Audit Logs**
```bash
GET /api/v1/user/audit-logs?limit=50&offset=0
Authorization: Bearer {token}

Response:
{
  "logs": [
    {
      "id": "uuid-1",
      "user_id": "uuid-2",
      "action": "image.push",
      "resource": "library/nginx:latest",
      "ip_address": "192.168.1.100",
      "user_agent": "docker/20.10.7",
      "timestamp": "2026-02-07T16:30:00Z",
      "details": {
        "digest": "sha256:abc123...",
        "size_bytes": 142857600
      }
    }
  ],
  "total": 1234
}
```

#### Database Schema
```sql
CREATE TABLE audit_logs (
    id UUID PRIMARY KEY,
    user_id UUID REFERENCES users(id),
    action VARCHAR(100),               -- 'image.push', 'user.login', etc.
    resource VARCHAR(255),             -- Affected resource
    ip_address VARCHAR(50),
    user_agent TEXT,
    timestamp TIMESTAMP,
    details JSONB                      -- Additional context
);
```

---

## 📡 API Reference

### Complete API Endpoint List

#### Authentication & Users
```
POST   /api/v1/auth/register           → Register new user
POST   /api/v1/auth/login              → Login
POST   /api/v1/auth/logout             → Logout
POST   /api/v1/auth/forgot-password    → Request password reset
POST   /api/v1/auth/reset-with-key     → Reset password with token
POST   /api/v1/auth/change-password    → Change password (authenticated)
GET    /api/v1/auth/token              → Get Docker registry token
```

#### Dashboard & Stats
```
GET    /api/v1/stats                   → Dashboard statistics
GET    /api/v1/health-check            → System health check
GET    /api/v1/system/config           → System configuration
```

#### Repositories & Images
```
POST   /api/v1/repositories            → Create repository
GET    /v2/_catalog                    → List repositories (OCI)
GET    /v2/{name}/tags/list            → List tags (OCI)
GET    /api/v1/repositories/{name}/manifests/{ref}        → Manifest details
DELETE /api/v1/repositories/{name}                        → Delete repository
DELETE /api/v1/repositories/{name}/tags/{tag}             → Delete tag
DELETE /api/v1/repositories/{name}/manifests/{reference}  → Delete manifest
```

#### Vulnerability Scanning
```
GET    /api/v1/repositories/{name}/manifests/{ref}/scan/status   → Scan status
GET    /api/v1/repositories/{name}/manifests/{ref}/scan/report   → Download report
GET    /api/v1/repositories/{name}/manifests/{ref}/scan/history  → Scan history
POST   /api/v1/repositories/{name}/manifests/{ref}/scan/trigger  → Trigger scan
```

#### Vulnerability Intelligence
```
GET    /api/v1/vulnerabilities/prioritized                → Prioritized CVEs
GET    /api/v1/vulnerabilities/intelligence/{cve}         → CVE intelligence
POST   /api/v1/vulnerabilities/refresh-epss               → Refresh EPSS data
```

#### Cost Intelligence
```
GET    /api/v1/costs/dashboard         → Cost dashboard
GET    /api/v1/costs/zombie-images     → Zombie images
POST   /api/v1/costs/refresh            → Recalculate costs
POST   /api/v1/costs/cleanup-zombies    → Cleanup zombies
```

#### Security Policies
```
GET    /api/v1/system/security/policy                     → Get global policy
PUT    /api/v1/system/security/policy                     → Update global policy
GET    /api/v1/system/security/policy/overrides           → List overrides
POST   /api/v1/system/security/policy/overrides           → Create override
DELETE /api/v1/system/security/policy/overrides/{repo}    → Delete override
```

#### User Management
```
GET    /api/v1/users                   → List users (admin)
POST   /api/v1/users                   → Invite user (admin)
DELETE /api/v1/users/{id}              → Delete user (admin)
PUT    /api/v1/users/{id}/role         → Update role (admin)
```

#### Session Management
```
GET    /api/v1/system/sessions         → List active sessions
DELETE /api/v1/system/sessions/{id}    → Revoke session
```

#### Service Accounts
```
GET    /api/v1/service-accounts        → List service accounts
POST   /api/v1/service-accounts        → Create service account
DELETE /api/v1/service-accounts/{id}   → Revoke service account
```

#### Webhooks
```
GET    /api/v1/system/webhooks         → List webhooks
POST   /api/v1/system/webhooks         → Create webhook
DELETE /api/v1/system/webhooks/{id}    → Delete webhook
POST   /api/v1/system/webhooks/{id}/test → Test webhook
```

#### Audit Logs
```
GET    /api/v1/user/audit-logs         → Get audit logs
```

#### Dependency Graph
```
GET    /api/v1/dependencies            → Get dependency graph
```

#### System Operations
```
POST   /api/v1/system/gc               → Garbage collection
```

---

## 🗄️ Database Schema

### Complete Schema Overview

```sql
-- Users & Authentication
users
password_reset_tokens
service_accounts

-- Multi-Tenancy
namespaces
repositories

-- OCI Storage
blobs
manifests
tags
manifest_layers

-- Security
vulnerability_reports
vulnerability_intelligence
manifest_vuln_priority
security_policies

-- Cost Intelligence
storage_costs
zombie_images
cost_savings_opportunities
optimization_suggestions

-- System
audit_logs
webhooks
sessions (Redis)
scan_queue (Redis)

-- Metadata
manifest_health_scores
image_dependencies
```

### Key Relationships

```
users (1) ──→ (N) namespaces
namespaces (1) ──→ (N) repositories
repositories (1) ──→ (N) manifests
repositories (1) ──→ (N) tags
manifests (1) ──→ (N) tags
manifests (1) ──→ (N) manifest_layers
manifest_layers (N) ──→ (1) blobs
manifests (1) ──→ (N) vulnerability_reports
manifests (1) ──→ (N) manifest_vuln_priority
manifests (1) ──→ (N) storage_costs
```

---

## 🚀 Deployment Options

### 1. Local Development (Docker Compose)

**Quick Start**
```powershell
# Windows
.\scripts\start_production.ps1

# Linux/Mac
docker-compose -f deploy/docker-compose.yml up -d
```

**Access Points**
- Frontend: http://localhost:5173
- Backend API: http://localhost:5000
- MinIO Console: http://localhost:9001
- PostgreSQL: localhost:5432
- Redis: localhost:6379

---

### 2. Production VPS Deployment

**Requirements**
- Ubuntu 22.04+ or similar
- 4GB+ RAM
- 50GB+ storage
- Docker & Docker Compose

**Steps**
```bash
# 1. Install Docker
sudo apt update
sudo apt install -y docker.io docker-compose
sudo systemctl enable --now docker

# 2. Clone/Upload project
git clone https://github.com/ckmine11/registry-x.git
cd registry-x

# 3. Configure environment
cp .env.example .env
nano .env  # Update passwords, secrets

# 4. Start services
docker-compose -f deploy/docker-compose.yml up -d

# 5. Verify
curl http://localhost:5000/api/v1/health-check
```

---

### 3. Kubernetes Deployment

**Manifests Included**
```
deploy/k8s/
├── namespace.yaml
├── postgres.yaml
├── redis.yaml
├── minio.yaml
├── backend.yaml
├── frontend.yaml
└── ingress.yaml
```

**Deploy**
```bash
# 1. Build and push images
docker build -t your-registry/registryx-backend:latest ./backend
docker push your-registry/registryx-backend:latest

docker build -t your-registry/registryx-frontend:latest ./frontend
docker push your-registry/registryx-frontend:latest

# 2. Update image references in k8s manifests
# Edit deploy/k8s/*.yaml files

# 3. Apply manifests
kubectl apply -f deploy/k8s/

# 4. Verify
kubectl get pods -n registryx
kubectl get svc -n registryx

# 5. Access (port-forward or ingress)
kubectl port-forward svc/frontend-svc 5173:80 -n registryx
```

---

### 4. SSL/HTTPS Setup (Production)

**Using Nginx + Certbot**

```bash
# 1. Install Nginx and Certbot
sudo apt install nginx certbot python3-certbot-nginx

# 2. Configure Nginx
sudo nano /etc/nginx/sites-available/registryx

# Add configuration:
server {
    server_name registry.yourdomain.com;
    client_max_body_size 20G;

    location / {
        proxy_pass http://localhost:5173;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }

    location /api/ {
        proxy_pass http://localhost:5000;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }
}

# 3. Enable site
sudo ln -s /etc/nginx/sites-available/registryx /etc/nginx/sites-enabled/
sudo nginx -t
sudo systemctl reload nginx

# 4. Get SSL certificate
sudo certbot --nginx -d registry.yourdomain.com

# 5. Auto-renewal
sudo certbot renew --dry-run
```

---

## 🧪 Testing & Verification

### Included Test Scripts

```powershell
# Push demo images
.\scripts\push_demo_images.ps1

# Push vulnerable images for testing
.\scripts\push_vulnerable_images.ps1

# Sign images with Cosign
.\scripts\sign_images.ps1

# Verify vulnerability scanning
.\scripts\verify_scan.ps1

# Verify cost intelligence
.\scripts\verify_costs.ps1

# Test user isolation
.\scripts\test_complete_isolation.ps1

# Verify catalog access
.\scripts\verify_catalog.ps1
```

---

## 📊 Monitoring & Observability

### Health Check Endpoint

```bash
GET /api/v1/health-check

Response:
{
  "status": "ok",
  "version": "2.4",
  "database": "connected",
  "redis": "connected",
  "storage": "connected",
  "timestamp": "2026-02-07T16:30:00Z"
}
```

### Logs

```bash
# View all logs
docker-compose -f deploy/docker-compose.yml logs -f

# Backend logs only
docker logs registryx-backend -f

# Database logs
docker logs registryx-db -f

# Scan worker logs
docker logs registryx-backend -f | grep "Worker:"
```

---

## 🔐 Security Best Practices

### Production Checklist

- [ ] Change default passwords in `.env`
- [ ] Generate strong `JWT_SECRET`
- [ ] Enable HTTPS with valid SSL certificate
- [ ] Configure CORS for specific origins
- [ ] Enable `MINIO_SECURE=true` for production S3
- [ ] Set `POLICY_ENVIRONMENT=prod`
- [ ] Configure SMTP for email notifications
- [ ] Enable audit logging
- [ ] Set up regular database backups
- [ ] Configure firewall rules
- [ ] Implement rate limiting (Nginx)
- [ ] Enable image signing enforcement
- [ ] Set up monitoring and alerting

---

## 📈 Performance Optimization

### Recommended Settings

**PostgreSQL**
```sql
-- Increase connection pool
max_connections = 200

-- Optimize for SSD
random_page_cost = 1.1

-- Increase shared buffers
shared_buffers = 2GB
```

**Redis**
```
maxmemory 1gb
maxmemory-policy allkeys-lru
```

**MinIO**
```
MINIO_STORAGE_CLASS_STANDARD=EC:2
```

---

## 🎓 Advanced Use Cases

### 1. CI/CD Integration

**GitHub Actions Example**
```yaml
name: Build and Push to RegistryX

on:
  push:
    branches: [main]

jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      
      - name: Login to RegistryX
        run: |
          echo "${{ secrets.REGISTRY_PASSWORD }}" | docker login registry.yourdomain.com -u ${{ secrets.REGISTRY_USERNAME }} --password-stdin
      
      - name: Build and Push
        run: |
          docker build -t registry.yourdomain.com/myapp:${{ github.sha }} .
          docker push registry.yourdomain.com/myapp:${{ github.sha }}
      
      - name: Wait for Scan
        run: |
          sleep 30  # Wait for scan to complete
          curl -f https://registry.yourdomain.com/api/v1/repositories/myapp/manifests/${{ github.sha }}/scan/status
```

---

### 2. Webhook Integration

**Slack Notifications**
```bash
# Create webhook
POST /api/v1/system/webhooks
{
  "url": "https://hooks.slack.com/services/YOUR/WEBHOOK/URL",
  "events": ["scan.completed", "image.pushed"],
  "active": true
}
```

**Custom Webhook Handler**
```python
from flask import Flask, request

app = Flask(__name__)

@app.route('/webhook', methods=['POST'])
def handle_webhook():
    data = request.json
    event_type = data['event']
    
    if event_type == 'scan.completed':
        if data['summary']['critical'] > 0:
            # Send alert
            send_alert(f"Critical vulnerabilities found in {data['repository']}")
    
    return {'status': 'ok'}
```

---

## 🆘 Troubleshooting

### Common Issues

**1. Scan stuck in "pending"**
```bash
# Check Redis queue
docker exec registryx-redis redis-cli LLEN scan_queue

# Check backend logs
docker logs registryx-backend | grep "Worker:"

# Manually trigger scan
curl -X POST http://localhost:5000/api/v1/repositories/{name}/manifests/{ref}/scan/trigger
```

**2. Images not showing in UI**
```bash
# Check database
docker exec registryx-db psql -U registryx -d registryx -c "SELECT COUNT(*) FROM manifests;"

# Check namespace ownership
docker exec registryx-db psql -U registryx -d registryx -c "SELECT * FROM namespaces;"
```

**3. Cost showing as $0**
```bash
# Trigger cost recalculation
curl -X POST http://localhost:5000/api/v1/costs/refresh

# Check storage_costs table
docker exec registryx-db psql -U registryx -d registryx -c "SELECT * FROM storage_costs LIMIT 5;"
```

---

## 📚 Additional Resources

- [Trivy Documentation](https://aquasecurity.github.io/trivy/)
- [EPSS Specification](https://www.first.org/epss/)
- [OCI Distribution Spec](https://github.com/opencontainers/distribution-spec)
- [Cosign Documentation](https://docs.sigstore.dev/cosign/overview/)
- [Docker Registry API](https://docs.docker.com/registry/spec/api/)

---

## 🎯 Summary

RegistryX provides:

✅ **Full OCI compliance** with Docker Registry V2 API  
✅ **Real-time vulnerability scanning** with Trivy  
✅ **Smart prioritization** using EPSS scores  
✅ **Cost intelligence** with zombie detection  
✅ **Image signing** support with Cosign  
✅ **Repository-specific policies** with global defaults  
✅ **Multi-tenant isolation** at database level  
✅ **Comprehensive audit logging** for compliance  
✅ **Modern React UI** with real-time updates  
✅ **Production-ready** with Docker Compose and Kubernetes support  

**RegistryX is ready for production deployment today!** 🚀
