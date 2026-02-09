# RegistryX - Architecture Documentation

## 📐 System Architecture Overview

RegistryX follows a modern microservices-inspired architecture with clear separation of concerns, scalable components, and production-ready infrastructure.

---

## 🏗️ High-Level Architecture

```
┌─────────────────────────────────────────────────────────────────────┐
│                          Client Layer                                │
├─────────────────────────────────────────────────────────────────────┤
│                                                                       │
│  ┌──────────────┐    ┌──────────────┐    ┌──────────────┐          │
│  │   Web UI     │    │ Docker CLI   │    │  CI/CD       │          │
│  │  (Browser)   │    │              │    │  Pipeline    │          │
│  └──────┬───────┘    └──────┬───────┘    └──────┬───────┘          │
│         │                   │                    │                   │
│         │ HTTP/REST         │ OCI Registry API   │ OCI + REST       │
│         │                   │                    │                   │
└─────────┼───────────────────┼────────────────────┼───────────────────┘
          │                   │                    │
          ▼                   ▼                    ▼
┌─────────────────────────────────────────────────────────────────────┐
│                       Application Layer                              │
├─────────────────────────────────────────────────────────────────────┤
│                                                                       │
│  ┌─────────────────────────────────────────────────────────────┐   │
│  │                    Frontend (React)                          │   │
│  │  • React 18 + TypeScript                                     │   │
│  │  • Vite Build Tool                                           │   │
│  │  • TailwindCSS Styling                                       │   │
│  │  • TanStack Query (State Management)                         │   │
│  │  • React Router (Navigation)                                 │   │
│  └─────────────────────────────────────────────────────────────┘   │
│                              │                                        │
│                              │ REST API                               │
│                              ▼                                        │
│  ┌─────────────────────────────────────────────────────────────┐   │
│  │                    Backend (Go)                              │   │
│  │  ┌──────────────────────────────────────────────────────┐   │   │
│  │  │         API Layer (Gorilla Mux)                      │   │   │
│  │  │  • REST API Handlers                                 │   │   │
│  │  │  • OCI Registry Protocol                             │   │   │
│  │  │  • Authentication Middleware                         │   │   │
│  │  │  • CORS Middleware                                   │   │   │
│  │  └──────────────────────────────────────────────────────┘   │   │
│  │                              │                                │   │
│  │  ┌──────────────────────────────────────────────────────┐   │   │
│  │  │         Service Layer                                │   │   │
│  │  │  ┌────────────┐  ┌────────────┐  ┌────────────┐     │   │   │
│  │  │  │   Auth     │  │  Scanner   │  │   Costs    │     │   │   │
│  │  │  │  Service   │  │  Service   │  │  Service   │     │   │   │
│  │  │  └────────────┘  └────────────┘  └────────────┘     │   │   │
│  │  │  ┌────────────┐  ┌────────────┐  ┌────────────┐     │   │   │
│  │  │  │Intelligence│  │  Metadata  │  │   Policy   │     │   │   │
│  │  │  │  Service   │  │  Service   │  │  Service   │     │   │   │
│  │  │  └────────────┘  └────────────┘  └────────────┘     │   │   │
│  │  │  ┌────────────┐  ┌────────────┐  ┌────────────┐     │   │   │
│  │  │  │  Webhook   │  │   Email    │  │   Audit    │     │   │   │
│  │  │  │  Service   │  │  Service   │  │  Service   │     │   │   │
│  │  │  └────────────┘  └────────────┘  └────────────┘     │   │   │
│  │  └──────────────────────────────────────────────────────┘   │   │
│  │                              │                                │   │
│  │  ┌──────────────────────────────────────────────────────┐   │   │
│  │  │         Background Workers                           │   │   │
│  │  │  • Scan Queue Worker (Trivy)                         │   │   │
│  │  │  • EPSS Refresh Worker (Daily)                       │   │   │
│  │  │  • Cost Calculation Worker                           │   │   │
│  │  └──────────────────────────────────────────────────────┘   │   │
│  └─────────────────────────────────────────────────────────────┘   │
│                                                                       │
└───────────────────────────┬───────────────────────────────────────────┘
                            │
                            ▼
┌─────────────────────────────────────────────────────────────────────┐
│                         Data Layer                                   │
├─────────────────────────────────────────────────────────────────────┤
│                                                                       │
│  ┌──────────────┐    ┌──────────────┐    ┌──────────────┐          │
│  │ PostgreSQL   │    │    Redis     │    │    MinIO     │          │
│  │              │    │              │    │              │          │
│  │ • Users      │    │ • Queues     │    │ • Blobs      │          │
│  │ • Repos      │    │ • Sessions   │    │ • Layers     │          │
│  │ • Manifests  │    │ • Cache      │    │ • Manifests  │          │
│  │ • Scans      │    │              │    │              │          │
│  │ • Costs      │    │              │    │              │          │
│  │ • Policies   │    │              │    │              │          │
│  └──────────────┘    └──────────────┘    └──────────────┘          │
│                                                                       │
└─────────────────────────────────────────────────────────────────────┘
```

---

## 🔄 Data Flow Diagrams

### 1. Image Push Flow

```
┌──────────┐
│  Docker  │
│   CLI    │
└────┬─────┘
     │
     │ 1. docker push localhost:5000/library/nginx:v1.0
     │
     ▼
┌─────────────────────────────────────────────────────┐
│              Backend (OCI Registry API)              │
├─────────────────────────────────────────────────────┤
│                                                       │
│  2. Authenticate User                                │
│     └─> Check JWT token or Basic Auth               │
│                                                       │
│  3. Start Blob Upload                                │
│     └─> POST /v2/{name}/blobs/uploads/              │
│     └─> Generate upload UUID                         │
│                                                       │
│  4. Upload Layers (Blobs)                            │
│     └─> PATCH /v2/{name}/blobs/uploads/{uuid}       │
│     └─> Stream data to MinIO                         │
│     └─> Calculate SHA256 digest                      │
│                                                       │
│  5. Complete Upload                                  │
│     └─> PUT /v2/{name}/blobs/uploads/{uuid}         │
│     └─> Store blob metadata in PostgreSQL           │
│                                                       │
│  6. Upload Manifest                                  │
│     └─> PUT /v2/{name}/manifests/{tag}              │
│     └─> Store manifest in PostgreSQL                │
│     └─> Link manifest to blobs                      │
│                                                       │
│  7. Queue Vulnerability Scan                         │
│     └─> Push scan job to Redis queue                │
│                                                       │
│  8. Trigger Webhooks                                 │
│     └─> Notify configured endpoints                 │
│                                                       │
└─────────────────────────────────────────────────────┘
     │
     │ 9. Background Processing
     │
     ▼
┌─────────────────────────────────────────────────────┐
│              Scan Worker (Background)                │
├─────────────────────────────────────────────────────┤
│                                                       │
│  10. Dequeue Scan Job                                │
│      └─> Pop from Redis scan_queue                   │
│                                                       │
│  11. Download Image Layers                           │
│      └─> Fetch blobs from MinIO                      │
│                                                       │
│  12. Run Trivy Scan                                  │
│      └─> Execute: trivy image --format json          │
│                                                       │
│  13. Parse Results                                   │
│      └─> Extract CVEs and severities                 │
│                                                       │
│  14. Store Scan Results                              │
│      └─> Insert into vulnerability_reports           │
│                                                       │
│  15. Fetch EPSS Data                                 │
│      └─> Query vulnerability_intelligence            │
│                                                       │
│  16. Calculate Priority Scores                       │
│      └─> Insert into manifest_vuln_priority          │
│                                                       │
│  17. Update Health Score                             │
│      └─> Calculate manifest health (0-100)           │
│                                                       │
│  18. Trigger Scan Complete Webhook                   │
│      └─> Notify external systems                     │
│                                                       │
└─────────────────────────────────────────────────────┘
```

---

### 2. Vulnerability Scan Flow

```
┌─────────────────────────────────────────────────────────────┐
│                    Scan Trigger                              │
└─────────────────────────────────────────────────────────────┘
                            │
                ┌───────────┴───────────┐
                │                       │
                ▼                       ▼
        ┌──────────────┐        ┌──────────────┐
        │ Automatic    │        │   Manual     │
        │ (on push)    │        │ (API call)   │
        └──────┬───────┘        └──────┬───────┘
               │                       │
               └───────────┬───────────┘
                           │
                           ▼
        ┌──────────────────────────────────────┐
        │   Queue Scan Job in Redis            │
        │   {                                  │
        │     manifest_id: "uuid",             │
        │     repository: "library/nginx",     │
        │     reference: "v1.0"                │
        │   }                                  │
        └──────────────┬───────────────────────┘
                       │
                       ▼
        ┌──────────────────────────────────────┐
        │   Background Worker Picks Up Job     │
        └──────────────┬───────────────────────┘
                       │
                       ▼
        ┌──────────────────────────────────────┐
        │   Update Status: "scanning"          │
        └──────────────┬───────────────────────┘
                       │
                       ▼
        ┌──────────────────────────────────────┐
        │   Download Image from MinIO          │
        │   • Fetch all layer blobs            │
        │   • Reconstruct image                │
        └──────────────┬───────────────────────┘
                       │
                       ▼
        ┌──────────────────────────────────────┐
        │   Execute Trivy Scan                 │
        │   $ trivy image --format json        │
        └──────────────┬───────────────────────┘
                       │
                       ▼
        ┌──────────────────────────────────────┐
        │   Parse Trivy Output                 │
        │   • Extract CVEs                     │
        │   • Count by severity                │
        │   • Store full JSON report           │
        └──────────────┬───────────────────────┘
                       │
                       ▼
        ┌──────────────────────────────────────┐
        │   Enrich with EPSS Intelligence      │
        │   • Query vulnerability_intelligence │
        │   • Calculate priority scores        │
        │   • Determine recommended actions    │
        └──────────────┬───────────────────────┘
                       │
                       ▼
        ┌──────────────────────────────────────┐
        │   Store Results in PostgreSQL        │
        │   • vulnerability_reports            │
        │   • manifest_vuln_priority           │
        └──────────────┬───────────────────────┘
                       │
                       ▼
        ┌──────────────────────────────────────┐
        │   Calculate Health Score             │
        │   Score = f(critical, high, epss)    │
        └──────────────┬───────────────────────┘
                       │
                       ▼
        ┌──────────────────────────────────────┐
        │   Update Status: "completed"         │
        └──────────────┬───────────────────────┘
                       │
                       ▼
        ┌──────────────────────────────────────┐
        │   Trigger Webhooks                   │
        │   • scan.completed event             │
        │   • Include summary                  │
        └──────────────────────────────────────┘
```

---

### 3. Cost Intelligence Flow

```
┌─────────────────────────────────────────────────────────────┐
│                  Cost Calculation Trigger                    │
└─────────────────────────────────────────────────────────────┘
                            │
                ┌───────────┴───────────┐
                │                       │
                ▼                       ▼
        ┌──────────────┐        ┌──────────────┐
        │  Periodic    │        │   Manual     │
        │  (Daily)     │        │  (API call)  │
        └──────┬───────┘        └──────┬───────┘
               │                       │
               └───────────┬───────────┘
                           │
                           ▼
        ┌──────────────────────────────────────┐
        │   Fetch All Manifests                │
        │   SELECT * FROM manifests            │
        └──────────────┬───────────────────────┘
                       │
                       ▼
        ┌──────────────────────────────────────┐
        │   For Each Manifest:                 │
        │                                      │
        │   1. Calculate Storage Cost          │
        │      cost = size_gb × $0.023/mo      │
        │                                      │
        │   2. Calculate Bandwidth Cost        │
        │      cost = pulls × size_gb × $0.09  │
        │                                      │
        │   3. Calculate Total Cost            │
        │      total = storage + bandwidth     │
        │                                      │
        │   4. Check Last Pull Date            │
        │      days_since = now - last_pulled  │
        │                                      │
        │   5. Identify Zombies                │
        │      if days_since > 90:             │
        │        mark as zombie                │
        │                                      │
        └──────────────┬───────────────────────┘
                       │
                       ▼
        ┌──────────────────────────────────────┐
        │   Store in storage_costs Table       │
        │   • manifest_id                      │
        │   • storage_cost_usd                 │
        │   • bandwidth_cost_usd               │
        │   • total_cost_usd                   │
        │   • pull_count_30d                   │
        │   • last_pulled_at                   │
        └──────────────┬───────────────────────┘
                       │
                       ▼
        ┌──────────────────────────────────────┐
        │   Store Zombies in zombie_images     │
        │   • manifest_id                      │
        │   • days_since_last_pull             │
        │   • storage_cost_usd                 │
        │   • recommended_action               │
        └──────────────┬───────────────────────┘
                       │
                       ▼
        ┌──────────────────────────────────────┐
        │   Calculate Savings Opportunities    │
        │   • Total zombie cost                │
        │   • Deduplication potential          │
        │   • Compression opportunities        │
        └──────────────┬───────────────────────┘
                       │
                       ▼
        ┌──────────────────────────────────────┐
        │   Update Dashboard Metrics           │
        │   • Total cost                       │
        │   • Zombie count                     │
        │   • Potential savings                │
        └──────────────────────────────────────┘
```

---

## 🗄️ Database Schema Architecture

### Entity Relationship Diagram

```
┌─────────────┐
│    users    │
└──────┬──────┘
       │ 1
       │
       │ N
┌──────▼──────┐
│ namespaces  │
└──────┬──────┘
       │ 1
       │
       │ N
┌──────▼──────────┐
│  repositories   │
└──────┬──────────┘
       │ 1
       │
       ├─────────────────┐
       │ N               │ N
┌──────▼──────┐   ┌──────▼──────┐
│  manifests  │   │    tags     │
└──────┬──────┘   └─────────────┘
       │ 1               │
       │                 │ N
       │                 │
       │            ┌────▼────┐
       │            │manifests│
       │            └─────────┘
       │
       ├─────────────────┬─────────────────┬─────────────────┐
       │ 1               │ 1               │ 1               │ 1
       │ N               │ N               │ N               │ N
┌──────▼──────────┐ ┌───▼──────────┐ ┌───▼──────────┐ ┌───▼──────────┐
│manifest_layers  │ │vulnerability │ │manifest_vuln │ │storage_costs │
│                 │ │   _reports   │ │  _priority   │ │              │
└──────┬──────────┘ └──────────────┘ └──────────────┘ └──────────────┘
       │ N
       │ 1
┌──────▼──────┐
│    blobs    │
└─────────────┘

Additional Tables:
┌──────────────────────┐
│vulnerability_        │
│  intelligence        │
│ (EPSS data)          │
└──────────────────────┘

┌──────────────────────┐
│ zombie_images        │
└──────────────────────┘

┌──────────────────────┐
│ security_policies    │
└──────────────────────┘

┌──────────────────────┐
│ audit_logs           │
└──────────────────────┘

┌──────────────────────┐
│ webhooks             │
└──────────────────────┘

┌──────────────────────┐
│ service_accounts     │
└──────────────────────┘
```

---

## 🔐 Security Architecture

### Authentication Flow

```
┌──────────────┐
│   Client     │
└──────┬───────┘
       │
       │ 1. POST /api/v1/auth/login
       │    { username, password }
       │
       ▼
┌─────────────────────────────────────┐
│         Backend API                 │
├─────────────────────────────────────┤
│                                     │
│  2. Validate Credentials            │
│     └─> Query users table           │
│     └─> Compare bcrypt hash         │
│                                     │
│  3. Generate JWT Token              │
│     └─> Sign with JWT_SECRET        │
│     └─> Expiry: 24 hours            │
│                                     │
│  4. Store Session in Redis          │
│     └─> Key: session:{user_id}     │
│     └─> Value: {token, metadata}   │
│                                     │
│  5. Return Token                    │
│     └─> { token, user }             │
│                                     │
└─────────────┬───────────────────────┘
              │
              ▼
┌──────────────────────────────────────┐
│   Client Stores Token                │
│   • localStorage.setItem('token')    │
└──────────────────────────────────────┘

Subsequent Requests:
┌──────────────┐
│   Client     │
└──────┬───────┘
       │
       │ GET /api/v1/repositories
       │ Authorization: Bearer {token}
       │
       ▼
┌─────────────────────────────────────┐
│    Auth Middleware                  │
├─────────────────────────────────────┤
│                                     │
│  1. Extract Token from Header       │
│                                     │
│  2. Verify JWT Signature            │
│     └─> Use JWT_SECRET              │
│                                     │
│  3. Check Expiry                    │
│                                     │
│  4. Validate Session in Redis       │
│     └─> Check session:{user_id}    │
│                                     │
│  5. Attach User to Request Context  │
│                                     │
│  6. Pass to Handler                 │
│                                     │
└─────────────────────────────────────┘
```

---

### Multi-Tenant Isolation

```
┌─────────────────────────────────────────────────────────┐
│                    User Request                          │
│   GET /api/v1/repositories                              │
│   Authorization: Bearer {token}                          │
└───────────────────────┬─────────────────────────────────┘
                        │
                        ▼
┌─────────────────────────────────────────────────────────┐
│              Auth Middleware                             │
│  • Extract user_id from JWT                             │
│  • Attach to request context                            │
└───────────────────────┬─────────────────────────────────┘
                        │
                        ▼
┌─────────────────────────────────────────────────────────┐
│              Repository Handler                          │
│                                                          │
│  1. Get user_id from context                            │
│                                                          │
│  2. Query with isolation:                               │
│     SELECT r.*                                           │
│     FROM repositories r                                  │
│     JOIN namespaces n ON r.namespace_id = n.id          │
│     WHERE n.owner_id = {user_id}                        │
│                                                          │
│  3. Return only user's repositories                     │
│                                                          │
└─────────────────────────────────────────────────────────┘

Database-Level Isolation:
┌─────────────────────────────────────────────────────────┐
│                PostgreSQL Database                       │
├─────────────────────────────────────────────────────────┤
│                                                          │
│  User A's Data:                                         │
│  ┌──────────────────────────────────────┐              │
│  │ namespace: user-a                    │              │
│  │   └─> repository: user-a/app1        │              │
│  │   └─> repository: user-a/app2        │              │
│  └──────────────────────────────────────┘              │
│                                                          │
│  User B's Data:                                         │
│  ┌──────────────────────────────────────┐              │
│  │ namespace: user-b                    │              │
│  │   └─> repository: user-b/service1    │              │
│  │   └─> repository: user-b/service2    │              │
│  └──────────────────────────────────────┘              │
│                                                          │
│  ❌ User A CANNOT access User B's data                  │
│  ✅ Admin users can access all data                     │
│                                                          │
└─────────────────────────────────────────────────────────┘
```

---

## 🔄 Background Workers Architecture

```
┌─────────────────────────────────────────────────────────┐
│                   Backend Process                        │
├─────────────────────────────────────────────────────────┤
│                                                          │
│  Main Goroutine:                                        │
│  ┌────────────────────────────────────────┐            │
│  │  HTTP Server (Gorilla Mux)             │            │
│  │  • Handle API requests                 │            │
│  │  • Serve OCI registry protocol         │            │
│  └────────────────────────────────────────┘            │
│                                                          │
│  Worker Goroutine 1:                                    │
│  ┌────────────────────────────────────────┐            │
│  │  Scan Queue Worker                     │            │
│  │  • Poll Redis scan_queue               │            │
│  │  • Process scan jobs                   │            │
│  │  • Run Trivy scans                     │            │
│  │  • Store results                       │            │
│  └────────────────────────────────────────┘            │
│                                                          │
│  Worker Goroutine 2:                                    │
│  ┌────────────────────────────────────────┐            │
│  │  EPSS Refresh Worker                   │            │
│  │  • Run every 24 hours                  │            │
│  │  • Fetch EPSS data from FIRST.org      │            │
│  │  • Update vulnerability_intelligence   │            │
│  │  • Recalculate priorities              │            │
│  └────────────────────────────────────────┘            │
│                                                          │
│  Worker Goroutine 3 (Optional):                         │
│  ┌────────────────────────────────────────┐            │
│  │  Cost Calculation Worker               │            │
│  │  • Run periodically                    │            │
│  │  • Calculate storage costs             │            │
│  │  • Identify zombie images              │            │
│  │  • Update cost tables                  │            │
│  └────────────────────────────────────────┘            │
│                                                          │
└─────────────────────────────────────────────────────────┘
```

---

## 📦 Storage Architecture

### Blob Storage (MinIO)

```
MinIO Bucket: registryx-data
│
├── blobs/
│   ├── sha256/
│   │   ├── ab/
│   │   │   └── abc123...def (layer blob)
│   │   ├── cd/
│   │   │   └── cde456...ghi (layer blob)
│   │   └── ...
│   │
│   └── uploads/
│       └── {uuid}/
│           └── data (temporary upload)
│
└── manifests/
    └── {repository}/
        ├── sha256:abc123... (manifest)
        └── sha256:def456... (manifest)

Storage Strategy:
• Content-addressable (SHA256 digest)
• Deduplication by design (same blob = same digest)
• Chunked uploads for large layers
• Garbage collection for orphaned blobs
```

---

## 🌐 Network Architecture

### Port Mapping

```
┌─────────────────────────────────────────────────────────┐
│                    Docker Network                        │
│                   (registryx-net)                        │
├─────────────────────────────────────────────────────────┤
│                                                          │
│  ┌──────────────────┐                                   │
│  │  Frontend        │                                   │
│  │  Port: 80        │◄──────┐                           │
│  │  Exposed: 5173   │       │                           │
│  └──────────────────┘       │                           │
│                              │                           │
│  ┌──────────────────┐       │                           │
│  │  Backend         │       │                           │
│  │  Port: 5000      │◄──────┤                           │
│  │  Exposed: 5000   │       │                           │
│  └────────┬─────────┘       │                           │
│           │                 │                           │
│           ├─────────────────┘                           │
│           │                                              │
│           ├──────────────────┐                          │
│           │                  │                          │
│           ▼                  ▼                          │
│  ┌──────────────┐   ┌──────────────┐                   │
│  │ PostgreSQL   │   │    Redis     │                   │
│  │ Port: 5432   │   │  Port: 6379  │                   │
│  │ Exposed:5432 │   │  Exposed:6379│                   │
│  └──────────────┘   └──────────────┘                   │
│                                                          │
│           │                                              │
│           ▼                                              │
│  ┌──────────────┐                                       │
│  │    MinIO     │                                       │
│  │ Port: 9000   │                                       │
│  │ Port: 9001   │                                       │
│  │ Exposed:9000 │                                       │
│  │ Exposed:9001 │                                       │
│  └──────────────┘                                       │
│                                                          │
└─────────────────────────────────────────────────────────┘

External Access:
• Frontend:      http://localhost:5173
• Backend API:   http://localhost:5000
• MinIO Console: http://localhost:9001
• PostgreSQL:    localhost:5432
• Redis:         localhost:6379
```

---

## 🚀 Deployment Architectures

### 1. Single Server (Docker Compose)

```
┌─────────────────────────────────────────────┐
│           VPS / Cloud Server                 │
│         (Ubuntu 22.04, 4GB RAM)             │
├─────────────────────────────────────────────┤
│                                              │
│  ┌────────────────────────────────────┐    │
│  │      Docker Compose                │    │
│  │                                    │    │
│  │  ┌──────┐ ┌──────┐ ┌──────┐      │    │
│  │  │Front │ │Back  │ │ DB   │      │    │
│  │  │ end  │ │ end  │ │      │      │    │
│  │  └──────┘ └──────┘ └──────┘      │    │
│  │                                    │    │
│  │  ┌──────┐ ┌──────┐               │    │
│  │  │Redis │ │MinIO │               │    │
│  │  └──────┘ └──────┘               │    │
│  └────────────────────────────────────┘    │
│                                              │
│  ┌────────────────────────────────────┐    │
│  │         Nginx (Optional)           │    │
│  │  • Reverse Proxy                   │    │
│  │  • SSL Termination                 │    │
│  │  • Rate Limiting                   │    │
│  └────────────────────────────────────┘    │
│                                              │
└─────────────────────────────────────────────┘
```

---

### 2. Kubernetes Cluster

```
┌─────────────────────────────────────────────────────────┐
│                  Kubernetes Cluster                      │
├─────────────────────────────────────────────────────────┤
│                                                          │
│  Namespace: registryx                                   │
│                                                          │
│  ┌────────────────────────────────────────────────┐    │
│  │              Ingress Controller                 │    │
│  │  • HTTPS Termination                           │    │
│  │  • Load Balancing                              │    │
│  └────────────────┬───────────────────────────────┘    │
│                   │                                      │
│         ┌─────────┴─────────┐                           │
│         │                   │                           │
│         ▼                   ▼                           │
│  ┌─────────────┐     ┌─────────────┐                   │
│  │  Frontend   │     │   Backend   │                   │
│  │  Service    │     │   Service   │                   │
│  │             │     │             │                   │
│  │ ┌─────────┐ │     │ ┌─────────┐ │                   │
│  │ │  Pod 1  │ │     │ │  Pod 1  │ │                   │
│  │ └─────────┘ │     │ └─────────┘ │                   │
│  │ ┌─────────┐ │     │ ┌─────────┐ │                   │
│  │ │  Pod 2  │ │     │ │  Pod 2  │ │                   │
│  │ └─────────┘ │     │ └─────────┘ │                   │
│  └─────────────┘     └──────┬──────┘                   │
│                              │                           │
│              ┌───────────────┼───────────────┐          │
│              │               │               │          │
│              ▼               ▼               ▼          │
│       ┌───────────┐   ┌───────────┐   ┌───────────┐   │
│       │PostgreSQL │   │   Redis   │   │   MinIO   │   │
│       │StatefulSet│   │StatefulSet│   │StatefulSet│   │
│       │           │   │           │   │           │   │
│       │ PVC: 50GB │   │ PVC: 10GB │   │ PVC:500GB │   │
│       └───────────┘   └───────────┘   └───────────┘   │
│                                                          │
└─────────────────────────────────────────────────────────┘
```

---

## 🔄 Scalability Considerations

### Horizontal Scaling

```
Component          | Scalable | Strategy
-------------------|----------|----------------------------------
Frontend           | ✅ Yes   | Multiple replicas behind LB
Backend API        | ✅ Yes   | Multiple replicas, stateless
Scan Workers       | ✅ Yes   | Multiple workers, Redis queue
PostgreSQL         | ⚠️ Limited| Read replicas, connection pooling
Redis              | ✅ Yes   | Redis Cluster or Sentinel
MinIO              | ✅ Yes   | Distributed mode (4+ nodes)
```

### Performance Optimization

```
Layer              | Optimization
-------------------|----------------------------------
Frontend           | • CDN for static assets
                   | • Code splitting
                   | • Lazy loading
                   | • Caching
-------------------|----------------------------------
Backend            | • Connection pooling
                   | • Response caching (Redis)
                   | • Async processing (queues)
                   | • Database indexing
-------------------|----------------------------------
Database           | • Proper indexes
                   | • Query optimization
                   | • Partitioning (large tables)
                   | • Vacuum and analyze
-------------------|----------------------------------
Storage            | • Blob deduplication
                   | • Compression
                   | • Garbage collection
                   | • CDN for blob delivery
```

---

## 📊 Monitoring Architecture

### Observability Stack (Recommended)

```
┌─────────────────────────────────────────────────────────┐
│                    RegistryX                             │
└───────────────────────┬─────────────────────────────────┘
                        │
                        │ Metrics, Logs, Traces
                        │
        ┌───────────────┼───────────────┐
        │               │               │
        ▼               ▼               ▼
┌──────────────┐ ┌──────────────┐ ┌──────────────┐
│ Prometheus   │ │     Loki     │ │    Jaeger    │
│  (Metrics)   │ │    (Logs)    │ │   (Traces)   │
└──────┬───────┘ └──────┬───────┘ └──────┬───────┘
       │                │                │
       └────────────────┼────────────────┘
                        │
                        ▼
                ┌──────────────┐
                │   Grafana    │
                │ (Dashboard)  │
                └──────────────┘
```

---

## 🎯 Summary

RegistryX architecture is designed for:

✅ **Scalability** - Horizontal scaling of all components  
✅ **Reliability** - Background workers, queue-based processing  
✅ **Security** - Multi-tenant isolation, JWT auth, RBAC  
✅ **Performance** - Caching, async processing, optimized queries  
✅ **Maintainability** - Clear separation of concerns, modular design  
✅ **Observability** - Comprehensive logging, metrics, health checks  

**Production-ready architecture for enterprise container registry needs!** 🚀
