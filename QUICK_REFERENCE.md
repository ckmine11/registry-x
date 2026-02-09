# RegistryX - Quick Reference Guide

## 🚀 Quick Commands

### Starting RegistryX

```bash
# Windows
.\scripts\start_production.ps1

# Linux/Mac
docker-compose -f deploy/docker-compose.yml up -d

# Stop
docker-compose -f deploy/docker-compose.yml down

# View logs
docker-compose -f deploy/docker-compose.yml logs -f
```

---

## 🔑 Default Credentials

```
URL: http://localhost:5173
Username: admin
Password: password123
```

⚠️ **Change immediately in production!**

---

## 🐳 Docker Commands

### Push/Pull Images

```bash
# Login
docker login localhost:5000
Username: admin
Password: password123

# Tag image
docker tag nginx:latest localhost:5000/library/nginx:v1.0

# Push image
docker push localhost:5000/library/nginx:v1.0

# Pull image
docker pull localhost:5000/library/nginx:v1.0

# List repositories
curl http://localhost:5000/v2/_catalog

# List tags
curl http://localhost:5000/v2/library/nginx/tags/list
```

---

## 📡 Common API Calls

### Authentication

```bash
# Register
curl -X POST http://localhost:5000/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"username":"john","email":"john@example.com","password":"SecurePass123!"}'

# Login
curl -X POST http://localhost:5000/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"john","password":"SecurePass123!"}'

# Response: {"token":"jwt-token","user":{...}}
# Use token in subsequent requests:
# -H "Authorization: Bearer {token}"
```

### Vulnerability Scanning

```bash
# Get scan status
curl http://localhost:5000/api/v1/repositories/library/nginx/manifests/v1.0/scan/status

# Download scan report
curl -O http://localhost:5000/api/v1/repositories/library/nginx/manifests/v1.0/scan/report

# Trigger manual scan
curl -X POST http://localhost:5000/api/v1/repositories/library/nginx/manifests/v1.0/scan/trigger

# Get prioritized vulnerabilities
curl http://localhost:5000/api/v1/vulnerabilities/prioritized?manifest_id={uuid}
```

### Cost Intelligence

```bash
# Get cost dashboard
curl http://localhost:5000/api/v1/costs/dashboard

# Get zombie images
curl http://localhost:5000/api/v1/costs/zombie-images

# Cleanup zombies
curl -X POST http://localhost:5000/api/v1/costs/cleanup-zombies

# Refresh costs
curl -X POST http://localhost:5000/api/v1/costs/refresh
```

### Security Policies

```bash
# Get global policy
curl http://localhost:5000/api/v1/system/security/policy

# Update global policy
curl -X PUT http://localhost:5000/api/v1/system/security/policy \
  -H "Content-Type: application/json" \
  -d '{"min_severity":"HIGH","max_critical":0,"block_unsigned":true}'

# List repository overrides
curl http://localhost:5000/api/v1/system/security/policy/overrides

# Create override
curl -X POST http://localhost:5000/api/v1/system/security/policy/overrides \
  -H "Content-Type: application/json" \
  -d '{"repository":"library/nginx","policy":{"min_severity":"MEDIUM"}}'
```

---

## 🔐 Image Signing (Cosign)

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

# Sign without password prompt (CI/CD)
export COSIGN_PASSWORD="your-password"
cosign sign --key cosign.key --yes localhost:5000/library/nginx:v1.0
```

---

## 🗄️ Database Commands

```bash
# Connect to PostgreSQL
docker exec -it registryx-db psql -U registryx -d registryx

# Common queries
SELECT COUNT(*) FROM manifests;
SELECT COUNT(*) FROM vulnerability_reports;
SELECT * FROM users;
SELECT * FROM namespaces;
SELECT * FROM repositories;

# Check scan status
SELECT manifest_id, status, scanned_at, critical_count, high_count 
FROM vulnerability_reports 
ORDER BY scanned_at DESC 
LIMIT 10;

# Check costs
SELECT * FROM storage_costs ORDER BY total_cost_usd DESC LIMIT 10;

# Check zombie images
SELECT * FROM zombie_images ORDER BY days_since_last_pull DESC;
```

---

## 📊 Redis Commands

```bash
# Connect to Redis
docker exec -it registryx-redis redis-cli

# Check scan queue
LLEN scan_queue

# View queue items
LRANGE scan_queue 0 -1

# Check sessions
KEYS session:*

# Clear all sessions (force logout all users)
FLUSHDB
```

---

## 🪣 MinIO Commands

```bash
# Access MinIO Console
# URL: http://localhost:9001
# Username: admin (from .env MINIO_ROOT_USER)
# Password: (from .env MINIO_ROOT_PASSWORD)

# List buckets (via mc CLI)
docker exec registryx-minio mc ls local/

# List objects in bucket
docker exec registryx-minio mc ls local/registryx-data/

# Check bucket size
docker exec registryx-minio mc du local/registryx-data/
```

---

## 🧪 Testing Scripts

```powershell
# Push demo images
.\scripts\push_demo_images.ps1

# Push vulnerable images
.\scripts\push_vulnerable_images.ps1

# Sign images
.\scripts\sign_images.ps1

# Verify scans
.\scripts\verify_scan.ps1

# Verify costs
.\scripts\verify_costs.ps1

# Test isolation
.\scripts\test_complete_isolation.ps1
```

---

## 🔍 Troubleshooting

### Check Service Health

```bash
# Health check
curl http://localhost:5000/api/v1/health-check

# Check all containers
docker-compose -f deploy/docker-compose.yml ps

# View logs
docker logs registryx-backend -f
docker logs registryx-frontend -f
docker logs registryx-db -f
docker logs registryx-redis -f
docker logs registryx-minio -f
```

### Restart Services

```bash
# Restart specific service
docker-compose -f deploy/docker-compose.yml restart backend

# Restart all services
docker-compose -f deploy/docker-compose.yml restart

# Rebuild and restart
docker-compose -f deploy/docker-compose.yml up -d --build
```

### Clear Data (Reset)

```bash
# Stop services
docker-compose -f deploy/docker-compose.yml down

# Remove volumes (WARNING: Deletes all data!)
docker-compose -f deploy/docker-compose.yml down -v

# Start fresh
docker-compose -f deploy/docker-compose.yml up -d
```

---

## 📁 Important Files

| File | Purpose |
|------|---------|
| `.env` | Environment configuration |
| `deploy/docker-compose.yml` | Docker Compose config |
| `backend/main.go` | Backend entry point |
| `frontend/src/App.tsx` | Frontend entry point |
| `backend/migrations/` | Database migrations |
| `scripts/` | Helper scripts |

---

## 🌐 Service URLs

| Service | URL | Credentials |
|---------|-----|-------------|
| Frontend | http://localhost:5173 | admin / password123 |
| Backend API | http://localhost:5000 | - |
| MinIO Console | http://localhost:9001 | admin / (from .env) |
| PostgreSQL | localhost:5432 | registryx / (from .env) |
| Redis | localhost:6379 | - |

---

## 🔧 Environment Variables

### Essential Variables (.env)

```bash
# Security
JWT_SECRET=your-random-secret-key
POSTGRES_PASSWORD=your-db-password
MINIO_ROOT_PASSWORD=your-minio-password

# Email (for password reset)
SMTP_USER=your-email@gmail.com
SMTP_PASS=your-app-password

# Storage
S3_BUCKET=registryx-data
MINIO_SECURE=false  # true for production

# Policy
POLICY_ENVIRONMENT=dev  # prod for production
```

---

## 📊 Monitoring

### Check System Status

```bash
# Container stats
docker stats

# Disk usage
docker system df

# Database size
docker exec registryx-db psql -U registryx -d registryx -c "SELECT pg_size_pretty(pg_database_size('registryx'));"

# Storage usage
docker exec registryx-minio mc du local/registryx-data/
```

---

## 🚨 Common Issues

### Scan Stuck in "pending"

```bash
# Check Redis queue
docker exec registryx-redis redis-cli LLEN scan_queue

# Check worker logs
docker logs registryx-backend | grep "Worker:"

# Restart backend
docker-compose -f deploy/docker-compose.yml restart backend
```

### Cannot Login

```bash
# Check backend logs
docker logs registryx-backend -f

# Verify database connection
docker exec registryx-db psql -U registryx -d registryx -c "SELECT COUNT(*) FROM users;"

# Reset admin password (in database)
docker exec registryx-db psql -U registryx -d registryx -c "UPDATE users SET password_hash='$2a$10$...' WHERE username='admin';"
```

### Images Not Showing

```bash
# Check manifests table
docker exec registryx-db psql -U registryx -d registryx -c "SELECT COUNT(*) FROM manifests;"

# Check storage
docker exec registryx-minio mc ls local/registryx-data/

# Check user namespace
docker exec registryx-db psql -U registryx -d registryx -c "SELECT * FROM namespaces;"
```

---

## 📚 Documentation Links

- [Complete Features Walkthrough](FEATURES_WALKTHROUGH.md)
- [Deployment Guide](DEPLOYMENT_GUIDE.md)
- [Trivy Scan Features](docs/TRIVY_SCAN_FEATURES.md)
- [Image Signing Guide](docs/IMAGE_SIGNING.md)
- [Manual Testing](MANUAL_TESTING.md)

---

## 🎯 Quick Tips

1. **Always use HTTPS in production** - Set up Nginx with SSL
2. **Change default passwords** - Update `.env` before deploying
3. **Enable image signing** - Use Cosign for supply chain security
4. **Monitor costs** - Check zombie images regularly
5. **Review audit logs** - Track user activity for compliance
6. **Set up webhooks** - Integrate with CI/CD pipelines
7. **Configure SMTP** - Enable password reset functionality
8. **Backup regularly** - Export PostgreSQL and MinIO data
9. **Use service accounts** - For CI/CD authentication
10. **Review security policies** - Adjust per repository as needed

---

## 🆘 Getting Help

1. Check [FEATURES_WALKTHROUGH.md](FEATURES_WALKTHROUGH.md) for detailed documentation
2. Review logs: `docker logs registryx-backend -f`
3. Check health: `curl http://localhost:5000/api/v1/health-check`
4. Open an issue on GitHub
5. Contact support: support@registryx.io

---

**Happy Registrying! 🚀**
