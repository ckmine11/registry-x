# Mock Data Cleanup Summary

## Date: 2026-01-18

## Actions Taken:

### 1. Deleted Mock Data Injection Tool
- **Removed**: `backend/cmd/fix_scan/` directory
- **Purpose**: This tool was injecting fake CVE data (CVE-2023-MOCK, CVE-2023-TEST) into the database
- **Impact**: No more artificial vulnerability data will be created

### 2. Cleaned Database
- **Removed**: All vulnerability reports containing mock CVE IDs
  - CVE-2023-MOCK
  - CVE-2023-TEST
- **Reset**: 6 stuck scan statuses that were older than 1 hour
- **Result**: Database now only contains real Trivy scan results

### 3. Updated Code Comments
- **File**: `backend/pkg/api/handlers.go`
- **Change**: Updated StorageUsed field comment from "Mocked for now" to "Calculated from actual blob storage"
- **Reason**: Storage calculation is now based on real blob data, not mock values

## Current State:

### ✅ Real Data Sources:
- **Repositories**: Counted from actual database records
- **Images**: Counted from actual manifests
- **Vulnerabilities**: From real Trivy scans only
- **Storage**: Calculated from actual blob sizes in MinIO
- **Recent Pushes**: From actual push events
- **Severity Counts**: From real vulnerability scan results

### ❌ No Mock Data:
- No fake CVE entries
- No hardcoded test data
- No artificial vulnerability counts
- All data comes from actual registry operations and Trivy scans

## Scripts Available:

### For Testing (Optional):
- `push_demo_images.ps1` - Push real images for testing
- `push_vulnerable_images.ps1` - Push known vulnerable images for real scans
- `test_vulnerable_images.ps1` - Comprehensive test with real vulnerable images

### For Maintenance:
- `clean_mock_data.ps1` - Re-run to clean any future mock data
- `fix_scan_issue.ps1` - Fix scan issues by re-pushing images

## Verification:

To verify all data is real:
1. Check dashboard - all counts should reflect actual registry state
2. Trigger a scan - should show real Trivy results
3. Check vulnerability reports - should only contain real CVE IDs
4. Storage usage - should match actual MinIO blob sizes

## Notes:

- All type safety improvements from previous work are retained
- UI remains clean and simple (post-revert design)
- Backend continues to calculate real statistics
- No functionality was lost, only mock data was removed
