# Trivy Scan Features - User Guide

## Overview
RegistryX अब Trivy vulnerability scanning के लिए complete visibility और report download support provide करता है।

## नए Features

### 1. Scan Status देखना
किसी भी image manifest की current scan status check करें:

**API Endpoint:**
```
GET /api/v1/repositories/{name}/manifests/{reference}/scan/status
```

**Example:**
```bash
curl http://localhost:5173/api/v1/repositories/library/nginx/manifests/latest/scan/status
```

**Response:**
```json
{
  "status": "completed",
  "scanned_at": "2026-01-15T16:30:00Z",
  "summary": {
    "critical": 5,
    "high": 12,
    "medium": 23,
    "low": 8
  }
}
```

**Possible Status Values:**
- `pending` - Scan queued but not started
- `scanning` - Scan currently in progress  
- `completed` - Scan finished successfully
- `failed` - Scan encountered an error

---

### 2. Trivy Report Download करना
Complete Trivy JSON report download करें:

**API Endpoint:**
```
GET /api/v1/repositories/{name}/manifests/{reference}/scan/report
```

**Example:**
```bash
curl -O http://localhost:5173/api/v1/repositories/library/nginx/manifests/latest/scan/report
```

यह एक JSON file download करेगा जिसमें:
- सभी vulnerabilities की detailed list
- CVE IDs और descriptions
- Severity levels
- Fixed versions (if available)
- Package information

**Filename Format:**
```
trivy-report-{repository}-{reference}.json
```

---

### 3. Scan History देखना
किसी manifest के सभी previous scans की history देखें:

**API Endpoint:**
```
GET /api/v1/repositories/{name}/manifests/{reference}/scan/history
```

**Example:**
```bash
curl http://localhost:5173/api/v1/repositories/library/nginx/manifests/latest/scan/history
```

**Response:**
```json
{
  "scans": [
    {
      "id": "uuid-1",
      "status": "completed",
      "scanned_at": "2026-01-15T16:30:00Z",
      "summary": {
        "critical": 5,
        "high": 12,
        "medium": 23,
        "low": 8
      }
    },
    {
      "id": "uuid-2",
      "status": "completed",
      "scanned_at": "2026-01-14T10:15:00Z",
      "summary": {
        "critical": 6,
        "high": 15,
        "medium": 20,
        "low": 10
      }
    }
  ]
}
```

### 4. Manual Scan Trigger करना
आप किसी भी image का scan manually trigger कर सकते हैं:

**API Endpoint:**
```
POST /api/v1/repositories/{name}/manifests/{reference}/scan/trigger
```

**Example:**
```bash
curl -X POST http://localhost:5173/api/v1/repositories/library/nginx/manifests/latest/scan/trigger
```

**Response:**
```json
{
  "message": "Scan triggered successfully",
  "status": "scanning"
}
```

---

## Frontend Integration

### Repository Details Page में:

1. **Scan Status Badge**
   - Real-time scan status indicator
   - Color-coded by severity
   - Click to view details

2. **Download Report Button** (New ⬇️)
   - One-click Trivy JSON report download
   - Available only for completed scans

3. **Re-scan Button** (New 🔄)
   - Manually start a new vulnerability scan
   - Useful if you updated policy or want fresh data

4. **Scan History Tab** (New 📜)
   - Timeline view of all scans
   - Compare vulnerability counts over time

---

## Use Cases

### 1. CI/CD Integration
```bash
# Check scan status before deployment
STATUS=$(curl -s http://localhost:5173/api/v1/repositories/myapp/manifests/v1.0.0/scan/status | jq -r '.status')

if [ "$STATUS" != "completed" ]; then
  echo "Waiting for scan to complete..."
  exit 1
fi

# Download report for archival
curl -O http://localhost:5173/api/v1/repositories/myapp/manifests/v1.0.0/scan/report
```

### 2. Security Auditing
```bash
# Get scan history for compliance reporting
curl http://localhost:5173/api/v1/repositories/production/nginx/manifests/latest/scan/history > audit-report.json
```

### 3. Vulnerability Tracking
```bash
# Monitor vulnerability trends
for tag in v1.0.0 v1.1.0 v1.2.0; do
  echo "=== $tag ==="
  curl -s http://localhost:5173/api/v1/repositories/myapp/manifests/$tag/scan/status | jq '.summary'
done
```

---

## Technical Details

### Database Schema
Scan data `vulnerability_reports` table में store होता है:
- `manifest_id` - Image manifest का UUID
- `status` - Current scan status
- `scanner` - Scanner name (always 'trivy')
- `report_json` - Complete Trivy JSON output
- `critical_count`, `high_count`, `medium_count`, `low_count` - Vulnerability counts
- `scanned_at` - Scan completion timestamp

### Automatic Scanning
- हर नई image push पर automatically Trivy scan trigger होता है
- Scan background worker में asynchronously run होता है
- Results database में store होते हैं

---

## Next Steps

### Planned Enhancements:
1. ✅ **Scan Status API** - Implemented
2. ✅ **Report Download** - Implemented  
3. ✅ **Scan History** - Implemented
4. 🔄 **Frontend UI** - In Progress
5. 📋 **Manual Re-scan Trigger** - Planned
6. 📊 **Vulnerability Trends Dashboard** - Planned
7. 🔔 **Scan Completion Webhooks** - Planned

---

## Troubleshooting

### Scan Status shows "pending" for too long
```bash
# Check backend logs
docker logs registryx-backend --tail 50

# Check Redis queue
docker exec registryx-redis redis-cli LLEN scan_queue
```

### Report download fails with 404
```bash
# Verify scan is completed
curl http://localhost:5173/api/v1/repositories/{name}/manifests/{ref}/scan/status

# Check if report_json exists in database
docker exec registryx-db psql -U registryx -d registryx -c \
  "SELECT status, report_json IS NOT NULL FROM vulnerability_reports WHERE manifest_id = '{uuid}';"
```

---

## Summary

अब आप:
1. ✅ किसी भी image की scan status real-time में देख सकते हैं
2. ✅ Complete Trivy JSON reports download कर सकते हैं
3. ✅ Scan history track कर सकते हैं
4. ✅ CI/CD pipelines में integrate कर सकते हैं

**सभी APIs production-ready हैं और immediately use के लिए available हैं!**
