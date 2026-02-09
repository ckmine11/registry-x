# RegistryX - Enterprise Container Registry

<div align="center">

![RegistryX Logo](https://via.placeholder.com/800x200/1a1a2e/00d4ff?text=RegistryX+-+Secure+%26+Intelligent+Container+Registry)

**Production-Ready OCI Registry with Advanced Security & Cost Intelligence**

[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Go Version](https://img.shields.io/badge/Go-1.22-00ADD8?logo=go)](https://go.dev/)
[![React](https://img.shields.io/badge/React-18-61DAFB?logo=react)](https://reactjs.org/)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-16-336791?logo=postgresql)](https://www.postgresql.org/)

[Features](#-key-features) • [Quick Start](#-quick-start) • [Documentation](#-documentation) • [Architecture](#-architecture) • [API Reference](#-api-reference)

</div>

---

## 🎯 What is RegistryX?

RegistryX is a **next-generation container registry** that combines OCI compliance with enterprise-grade security and cost optimization. Unlike traditional registries that simply store images, RegistryX actively helps you:

- 🛡️ **Secure your supply chain** with real-time vulnerability scanning and EPSS-based prioritization
- 💰 **Optimize costs** by identifying zombie images and providing actionable insights
- 🔐 **Enforce policies** with repository-specific security rules and image signing
- 📊 **Gain visibility** through comprehensive dashboards and dependency graphs
- 🚀 **Scale confidently** with multi-tenant isolation and production-ready architecture

---

## ✨ Key Features

### 🛡️ Advanced Security Pipeline

| Feature | Description |
|---------|-------------|
| **Real-Time Scanning** | Automatic Trivy scans on every image push with background processing |
| **EPSS Intelligence** | Smart vulnerability prioritization using Exploit Prediction Scoring System |
| **Image Signing** | Cosign integration with automatic signature detection |
| **Policy Enforcement** | Repository-specific security policies with global defaults |
| **Audit Logging** | Comprehensive activity tracking for compliance |

### 💰 Cost Intelligence

| Feature | Description |
|---------|-------------|
| **Cost Tracking** | Real-time storage and bandwidth cost calculation per image |
| **Zombie Detection** | Identify unused images (90+ days) with one-click cleanup |
| **Optimization Tips** | AI-powered suggestions for reducing image sizes |
| **Cost Dashboard** | Visual breakdown of expenses and savings opportunities |

### 🎨 Modern User Experience

| Feature | Description |
|---------|-------------|
| **React Dashboard** | Sleek, responsive UI built with React 18 + TailwindCSS |
| **Real-Time Updates** | Live scan status and cost calculations |
| **Dependency Graph** | Visual representation of image relationships |
| **Multi-Tenant** | Complete data isolation between users and organizations |

### 🔧 Enterprise Ready

| Feature | Description |
|---------|-------------|
| **OCI Compliant** | Full Docker Registry V2 API implementation |
| **S3 Storage** | MinIO-based scalable object storage |
| **Webhooks** | Event notifications for CI/CD integration |
| **API Keys** | Service accounts for machine-to-machine auth |
| **Session Management** | Track and revoke active sessions |

---

## 🏗️ Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                     RegistryX Platform                       │
├─────────────────────────────────────────────────────────────┤
│                                                               │
│  ┌──────────────┐      ┌──────────────┐                     │
│  │   Frontend   │      │   Backend    │                     │
│  │  React + TS  │◄────►│   Go 1.22    │                     │
│  │   Vite       │      │  Gorilla Mux │                     │
│  └──────────────┘      └───────┬──────┘                     │
│                                 │                             │
│         ┌───────────────────────┼───────────────────────┐    │
│         │                       │                       │    │
│         ▼                       ▼                       ▼    │
│  ┌─────────────┐       ┌──────────────┐       ┌──────────┐ │
│  │ PostgreSQL  │       │    Redis     │       │  MinIO   │ │
│  │  Metadata   │       │ Queue/Cache  │       │ Storage  │ │
│  │  Users      │       │  Sessions    │       │  Blobs   │ │
│  │  Scans      │       │  Scan Jobs   │       │ Layers   │ │
│  └─────────────┘       └──────────────┘       └──────────┘ │
│                                                               │
│  ┌──────────────────────────────────────────────────────┐   │
│  │              Background Workers                       │   │
│  │  • Trivy Scanner  • EPSS Refresh  • Cost Calculator  │   │
│  └──────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────┘
```

### Technology Stack

**Backend**
- Go 1.22 with Gorilla Mux
- PostgreSQL 16 (metadata, users, scans)
- Redis 7 (queues, sessions)
- MinIO (S3-compatible blob storage)
- Trivy (vulnerability scanner)

**Frontend**
- React 18 + TypeScript
- Vite (build tool)
- TailwindCSS (styling)
- TanStack Query (data fetching)
- React Router v6 (routing)
- ReactFlow (dependency graphs)

**Infrastructure**
- Docker & Docker Compose
- Kubernetes (optional)
- Nginx (reverse proxy)

---

## 🚀 Quick Start

### Prerequisites

- Docker & Docker Compose
- Git
- 4GB+ RAM
- 20GB+ free disk space

### Installation (5 Minutes)

#### Windows (PowerShell)

```powershell
# 1. Clone the repository
git clone https://github.com/ckmine11/registry-x.git
cd registry-x

# 2. Start all services
.\scripts\start_production.ps1

# 3. Wait for services to start (30-60 seconds)
# Then access: http://localhost:5173
```

#### Linux/macOS

```bash
# 1. Clone the repository
git clone https://github.com/ckmine11/registry-x.git
cd registry-x

# 2. Start all services
docker-compose -f deploy/docker-compose.yml up -d

# 3. Wait for services to start (30-60 seconds)
# Then access: http://localhost:5173
```

### First Login

```
URL: http://localhost:5173
Username: admin
Password: admin@123
```

⚠️ **Important**: Change the default password immediately in production!

---

## 📖 Documentation

| Document | Description |
|----------|-------------|
| [FEATURES_WALKTHROUGH.md](FEATURES_WALKTHROUGH.md) | Complete feature documentation with examples |
| [DEPLOYMENT_GUIDE.md](DEPLOYMENT_GUIDE.md) | Production deployment instructions |
| [docs/TRIVY_SCAN_FEATURES.md](docs/TRIVY_SCAN_FEATURES.md) | Vulnerability scanning guide |
| [docs/IMAGE_SIGNING.md](docs/IMAGE_SIGNING.md) | Image signing with Cosign |
| [MANUAL_TESTING.md](MANUAL_TESTING.md) | Testing procedures |

---

## 💻 Usage Examples

### Pushing Your First Image

```bash
# 1. Login to RegistryX
docker login localhost:5000
Username: admin
Password: password123

# 2. Tag your image
docker tag nginx:latest localhost:5000/library/nginx:v1.0

# 3. Push the image
docker push localhost:5000/library/nginx:v1.0

# 4. Automatic scan starts in background
# View results in UI: http://localhost:5173/repositories/library/nginx
```

### Checking Vulnerability Status

```bash
# Get scan status
curl http://localhost:5000/api/v1/repositories/library/nginx/manifests/v1.0/scan/status

# Download full Trivy report
curl -O http://localhost:5000/api/v1/repositories/library/nginx/manifests/v1.0/scan/report
```

### Managing Costs

```bash
# View cost dashboard
curl http://localhost:5000/api/v1/costs/dashboard

# Get zombie images
curl http://localhost:5000/api/v1/costs/zombie-images

# Cleanup zombies (saves storage costs)
curl -X POST http://localhost:5000/api/v1/costs/cleanup-zombies
```

### Signing Images

```bash
# Install Cosign
# Windows: winget install sigstore.cosign
# macOS: brew install cosign
# Linux: wget https://github.com/sigstore/cosign/releases/latest/download/cosign-linux-amd64

# Generate keys (one-time)
cosign generate-key-pair

# Sign image
cosign sign --key cosign.key localhost:5000/library/nginx:v1.0

# Verify signature
cosign verify --key cosign.pub localhost:5000/library/nginx:v1.0
```

---

## 🌐 API Reference

### Authentication

```bash
# Register new user
POST /api/v1/auth/register
{
  "username": "john",
  "email": "john@example.com",
  "password": "SecurePass123!"
}

# Login
POST /api/v1/auth/login
{
  "username": "john",
  "password": "SecurePass123!"
}
# Returns: { "token": "jwt-token", "user": {...} }

# Use token in subsequent requests
Authorization: Bearer {token}
```

### Image Operations

```bash
# List repositories
GET /v2/_catalog

# List tags
GET /v2/{repository}/tags/list

# Get manifest
GET /v2/{repository}/manifests/{tag}

# Delete repository
DELETE /api/v1/repositories/{repository}

# Delete tag
DELETE /api/v1/repositories/{repository}/tags/{tag}
```

### Vulnerability Scanning

```bash
# Get scan status
GET /api/v1/repositories/{name}/manifests/{ref}/scan/status

# Download scan report
GET /api/v1/repositories/{name}/manifests/{ref}/scan/report

# Trigger manual scan
POST /api/v1/repositories/{name}/manifests/{ref}/scan/trigger

# Get prioritized vulnerabilities (EPSS-based)
GET /api/v1/vulnerabilities/prioritized?manifest_id={uuid}
```

### Cost Intelligence

```bash
# Get cost dashboard
GET /api/v1/costs/dashboard

# Get zombie images
GET /api/v1/costs/zombie-images

# Cleanup zombies
POST /api/v1/costs/cleanup-zombies

# Refresh cost calculations
POST /api/v1/costs/refresh
```

### Security Policies

```bash
# Get global security policy
GET /api/v1/system/security/policy

# Update global policy
PUT /api/v1/system/security/policy
{
  "min_severity": "HIGH",
  "max_critical": 0,
  "block_unsigned": true
}

# List repository overrides
GET /api/v1/system/security/policy/overrides

# Create repository override
POST /api/v1/system/security/policy/overrides
{
  "repository": "library/nginx",
  "policy": { "min_severity": "MEDIUM" }
}
```

**Full API documentation**: See [FEATURES_WALKTHROUGH.md](FEATURES_WALKTHROUGH.md#-api-reference)

---

## 🔧 Configuration

### Environment Variables

Edit `.env` file to configure:

```bash
# Database
POSTGRES_PASSWORD=your-secure-password

# Storage
MINIO_ROOT_USER=admin
MINIO_ROOT_PASSWORD=your-minio-password
S3_BUCKET=registryx-data
MINIO_SECURE=false  # Set to 'true' for production with SSL

# Security
JWT_SECRET=your-random-secret-key-change-this
POLICY_ENVIRONMENT=dev  # 'dev' or 'prod'

# Email (for password reset)
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USER=your-email@gmail.com
SMTP_PASS=your-app-password

# Cost Intelligence
STORAGE_COST_PER_GB_MONTH=0.023  # USD
BANDWIDTH_COST_PER_GB=0.09       # USD
```

### Service Ports

| Service | Port | Description |
|---------|------|-------------|
| Frontend | 5173 | Web UI |
| Backend | 5000 | REST API + OCI Registry |
| PostgreSQL | 5432 | Database |
| Redis | 6379 | Cache/Queue |
| MinIO | 9000 | S3 API |
| MinIO Console | 9001 | MinIO Admin UI |

---

## 🚢 Production Deployment

### Docker Compose (Recommended for VPS)

```bash
# 1. Clone repository
git clone https://github.com/ckmine11/registry-x.git
cd registry-x

# 2. Configure environment
cp .env.example .env
nano .env  # Update passwords and secrets

# 3. Start services
docker-compose -f deploy/docker-compose.yml up -d

# 4. Verify deployment
curl http://localhost:5000/api/v1/health-check
```

### Kubernetes

```bash
# 1. Build and push images
docker build -t your-registry/registryx-backend:latest ./backend
docker push your-registry/registryx-backend:latest

docker build -t your-registry/registryx-frontend:latest ./frontend
docker push your-registry/registryx-frontend:latest

# 2. Update image references in k8s manifests
# Edit files in deploy/k8s/

# 3. Deploy to cluster
kubectl apply -f deploy/k8s/

# 4. Verify
kubectl get pods -n registryx
```

### SSL/HTTPS Setup

```bash
# Install Nginx and Certbot
sudo apt install nginx certbot python3-certbot-nginx

# Configure Nginx (see DEPLOYMENT_GUIDE.md)
sudo nano /etc/nginx/sites-available/registryx

# Get SSL certificate
sudo certbot --nginx -d registry.yourdomain.com
```

**Full deployment guide**: See [DEPLOYMENT_GUIDE.md](DEPLOYMENT_GUIDE.md)

---

## 🧪 Testing & Demo

### Included Test Scripts

```powershell
# Push demo images
.\scripts\push_demo_images.ps1

# Push vulnerable images for testing scans
.\scripts\push_vulnerable_images.ps1

# Sign images with Cosign
.\scripts\sign_images.ps1

# Verify vulnerability scanning
.\scripts\verify_scan.ps1

# Verify cost intelligence
.\scripts\verify_costs.ps1

# Test multi-tenant isolation
.\scripts\test_complete_isolation.ps1
```

---

## 📊 Screenshots

### Dashboard
![Dashboard](https://via.placeholder.com/1200x600/1a1a2e/00d4ff?text=Dashboard+with+Real-Time+Statistics)

### Repository Details
![Repository Details](https://via.placeholder.com/1200x600/1a1a2e/00d4ff?text=Repository+Details+with+Vulnerability+Scan)

### Cost Intelligence
![Cost Intelligence](https://via.placeholder.com/1200x600/1a1a2e/00d4ff?text=Cost+Dashboard+with+Zombie+Detection)

### Security Policies
![Security Policies](https://via.placeholder.com/1200x600/1a1a2e/00d4ff?text=Repository-Specific+Security+Policies)

---

## 🔒 Security Features

- ✅ **JWT Authentication** with session management
- ✅ **Multi-Tenant Isolation** at database level
- ✅ **Role-Based Access Control** (Admin/User)
- ✅ **Audit Logging** for all operations
- ✅ **Image Signing** with Cosign
- ✅ **Vulnerability Scanning** with Trivy
- ✅ **EPSS-Based Prioritization** for smart remediation
- ✅ **Repository-Specific Policies** with enforcement
- ✅ **Password Reset** via email
- ✅ **Service Accounts** for API access
- ✅ **Webhook Notifications** for events

---

## 🎓 Use Cases

### 1. Development Teams
- Private registry for internal images
- Automatic vulnerability scanning
- Cost tracking per project
- Easy image sharing

### 2. DevOps/SRE
- CI/CD integration via webhooks
- Policy enforcement for production
- Zombie image cleanup
- Dependency tracking

### 3. Security Teams
- EPSS-based vulnerability prioritization
- Image signing enforcement
- Comprehensive audit logs
- Compliance reporting

### 4. FinOps Teams
- Storage cost optimization
- Bandwidth usage tracking
- Zombie image detection
- Cost allocation per team

---

## 🛠️ Development

### Project Structure

```
registry-x/
├── backend/                 # Go backend
│   ├── cmd/                # Entry points
│   ├── pkg/                # Core packages
│   │   ├── api/           # REST API handlers
│   │   ├── auth/          # Authentication
│   │   ├── scanner/       # Trivy integration
│   │   ├── costs/         # Cost intelligence
│   │   ├── intelligence/  # EPSS prioritization
│   │   └── ...
│   ├── migrations/        # Database migrations
│   └── Dockerfile
├── frontend/               # React frontend
│   ├── src/
│   │   ├── pages/        # UI pages
│   │   ├── components/   # Reusable components
│   │   └── lib/          # Utilities
│   └── Dockerfile
├── deploy/                # Deployment configs
│   ├── docker-compose.yml
│   └── k8s/              # Kubernetes manifests
├── scripts/              # Helper scripts
└── docs/                 # Documentation
```

### Running Locally

```bash
# Backend
cd backend
go run main.go

# Frontend
cd frontend
npm install
npm run dev

# Database
docker run -d -p 5432:5432 -e POSTGRES_PASSWORD=password postgres:16

# Redis
docker run -d -p 6379:6379 redis:7-alpine

# MinIO
docker run -d -p 9000:9000 -p 9001:9001 \
  -e MINIO_ROOT_USER=admin \
  -e MINIO_ROOT_PASSWORD=password \
  minio/minio server /data --console-address ":9001"
```

---

## 🤝 Contributing

Contributions are welcome! Please follow these steps:

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

### Development Guidelines

- Follow Go best practices and conventions
- Write tests for new features
- Update documentation
- Ensure all tests pass
- Follow semantic versioning

---

## 📝 Changelog

### Version 2.4 (Current)
- ✅ Repository-specific security policies
- ✅ Enhanced EPSS intelligence
- ✅ Improved scan reliability
- ✅ Cost optimization features
- ✅ Session management
- ✅ Audit logging

### Version 2.3
- ✅ EPSS-based vulnerability prioritization
- ✅ Cost intelligence dashboard
- ✅ Zombie image detection
- ✅ Dependency graph visualization

### Version 2.2
- ✅ Trivy vulnerability scanning
- ✅ Image signing support
- ✅ Webhook notifications
- ✅ Service accounts

### Version 2.1
- ✅ Multi-tenant isolation
- ✅ User management
- ✅ Password reset flow
- ✅ Audit logging

### Version 2.0
- ✅ React frontend
- ✅ OCI compliance
- ✅ PostgreSQL backend
- ✅ MinIO storage

---

## 🐛 Troubleshooting

### Common Issues

**Scan stuck in "pending"**
```bash
# Check Redis queue
docker exec registryx-redis redis-cli LLEN scan_queue

# Check backend logs
docker logs registryx-backend -f

# Manually trigger scan
curl -X POST http://localhost:5000/api/v1/repositories/{name}/manifests/{ref}/scan/trigger
```

**Cannot login**
```bash
# Check if backend is running
curl http://localhost:5000/api/v1/health-check

# Check database connection
docker logs registryx-backend | grep "database"

# Reset password
# Use "Forgot Password" feature in UI
```

**Images not showing**
```bash
# Check database
docker exec registryx-db psql -U registryx -d registryx -c "SELECT COUNT(*) FROM manifests;"

# Check storage
docker exec registryx-minio mc ls local/registryx-data
```

**More troubleshooting**: See [FEATURES_WALKTHROUGH.md](FEATURES_WALKTHROUGH.md#-troubleshooting)

---

## 📚 Resources

- [OCI Distribution Specification](https://github.com/opencontainers/distribution-spec)
- [Docker Registry API](https://docs.docker.com/registry/spec/api/)
- [Trivy Documentation](https://aquasecurity.github.io/trivy/)
- [EPSS Documentation](https://www.first.org/epss/)
- [Cosign Documentation](https://docs.sigstore.dev/cosign/overview/)

---

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

---

## 🙏 Acknowledgments

- [Trivy](https://github.com/aquasecurity/trivy) for vulnerability scanning
- [MinIO](https://min.io/) for S3-compatible storage
- [FIRST.org](https://www.first.org/) for EPSS data
- [Sigstore](https://www.sigstore.dev/) for Cosign
- [OCI](https://opencontainers.org/) for container standards

---

## 📞 Support

- **Documentation**: [FEATURES_WALKTHROUGH.md](FEATURES_WALKTHROUGH.md)
- **Issues**: [GitHub Issues](https://github.com/ckmine11/registry-x/issues)
- **Email**: support@registryx.io

---

## 🌟 Star History

If you find RegistryX useful, please consider giving it a star! ⭐

---

<div align="center">

**Built with ❤️ by the RegistryX Team**

[Website](https://registryx.io) • [Documentation](FEATURES_WALKTHROUGH.md) • [GitHub](https://github.com/ckmine11/registry-x)

</div>
