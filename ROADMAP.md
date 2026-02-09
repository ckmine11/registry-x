# RegistryX Roadmap 2026

## 🚀 Proposed New Features

### 1. 🧠 AI-Powered Remediation Assistant
**Concept:**
Instead of just listing vulnerabilities (e.g., "CVE-2023-1234 in alpine:3.14"), the system uses an AI model (like OpenAI or local Llama 3) to analyze the image and generate a specific **Fix Plan**.
- **Features:**
  - "Fix with AI" button on Vulnerability Report.
  - Generates patched `Dockerfile` lines.
  - Explains *why* the fix works.

### 2. 🧹 Automated Lifecycle Policies (Retention)
**Concept:**
Automate the cleanup of old and unused images to reduce storage costs (synergy with Cost Intelligence).
- **Features:**
  - Rule builder: "Keep last 5 tags", "Delete images older than 30 days".
  - "Dry Run" mode to see what would be deleted.
  - Scheduled background garbage collection.

### 3. 🛡️ Comprehensive Audit Dashboard
**Concept:**
A dedicated admin view to inspect the security audit trails collected by the backend.
- **Features:**
  - Timeline view of user actions (Push, Pull, Delete, Login).
  - Filter by User, Repository, or Severity.
  - Export logs for compliance (CSV/JSON).

### 4. 🔑 SSO & Enterprise Auth
**Concept:**
Integrate OpenID Connect (OIDC) for enterprise login.
- **Features:**
  - "Login with GitHub" / "Login with Google".
  - LDAP/AD integration.
  - Enforce MFA.

### 5. ⚡ Registry Mirror & Pull-Through Cache
**Concept:**
Cache images from upstream registries (Docker Hub, Quay, GCR) to speed up local builds and avoid rate limits.
- **Features:**
  - Transparent caching of `docker.io` images.
  - Offline mode support.

---

## ✅ Completed Core Features
- [x] Docker Registry V2 API
- [x] Trivy Vulnerability Scanning
- [x] Cost Intelligence (S3 Estimates)
- [x] Security Policies (OPA-like)
- [x] RBAC & Multi-tenancy
- [x] Image Signing (Cosign)
