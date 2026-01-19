# RegistryX Manual Testing Guide

Follow these steps to verify the newly implemented features.
**Prerequisite:** Ensure Docker Desktop is running.

## 1. Setup (Register a User)
First, register a test user (e.g., `tester`) via the API.
```bash
# Windows PowerShell
Invoke-RestMethod -Method Post -Uri "http://localhost:5000/api/v1/auth/register" -Body '{"username":"tester", "email":"test@example.com", "password":"password123"}' -ContentType "application/json"
```

## 2. Test User Isolation (Namespaces)
**Goal:** Verify `tester` can push to `tester/...` but NOT `admin/...` or `library/...`.

1. **Login with Docker:**
   ```bash
   docker login localhost:5000 -u tester -p password123
   ```
2. **Pull a small image to use:**
   ```bash
   docker pull alpine
   ```
3. **Tag for YOUR namespace (Should Work):**
   ```bash
   docker tag alpine localhost:5000/tester/my-app:v1
   docker push localhost:5000/tester/my-app:v1
   # Result: Success
   ```
4. **Tag for ANOTHER namespace (Should Fail):**
   ```bash
   docker tag alpine localhost:5000/admin/forbidden-app:v1
   docker push localhost:5000/admin/forbidden-app:v1
   # Result: "denied: requested access to the resource is denied" or "403 Forbidden"
   ```

## 3. Test Audit Logs
**Goal:** Verify the "PUSH" and "LOGIN" actions are recorded.

1. **Get your Authentication Token** (from Login step or re-login):
   ```bash
   $resp = Invoke-RestMethod -Method Post -Uri "http://localhost:5000/api/v1/auth/login" -Body '{"username":"tester", "password":"password123"}' -ContentType "application/json"
   $token = $resp.token
   ```
2. **Fetch Logs:**
   ```bash
   Invoke-RestMethod -Method Get -Uri "http://localhost:5000/api/v1/user/audit-logs" -Headers @{"Authorization"="Bearer $token"}
   ```
   **Expected Output:** JSON list showing `LOGIN` and `PUSH` events.

## 4. Test Auto-Cleanup (Garbage Collection)
**Goal:** Verify untagged manifests are deleted.

1. **Create an Untagged Image:** (Overwrite an existing tag)
   ```bash
   # Push v1 (Already done above)
   # Pull a DIFFERENT image (e.g. busybox) and tag it as the SAME v1
   docker pull busybox
   docker tag busybox localhost:5000/tester/my-app:v1
   docker push localhost:5000/tester/my-app:v1
   ```
   *The previous `alpine` image is now "untagged" (orphaned).*

2. **Run GC (Requires Admin):**
   *(We must use the admin user created during setup, usually `admin` / `admin123` or similar. If you don't have one, use the `tester` user if they are the first user (unlikely admin) OR check logs).*
   *Assuming `tester` is just a user, this might fail with 401/403. You need an Admin token.*
   
   **Trick:** Database Hack to make `tester` an admin for testing:
   ```bash
   docker exec registryx-db psql -U registryx -d registryx -c "UPDATE users SET role='admin' WHERE username='tester';"
   ```

   **Now Run GC:**
   ```bash
   Invoke-RestMethod -Method Post -Uri "http://localhost:5000/api/v1/system/gc" -Headers @{"Authorization"="Bearer $token"}
   ```
   **Expected Output:** `{"blobsDeleted": X, "manifestsDeleted": 1, ...}`

## 5. Test Storage Quota
**Goal:** Verify quota enforcement.

1. **Lower Quota Manually:** Set `tester` quota to 1KB.
   ```bash
   docker exec registryx-db psql -U registryx -d registryx -c "UPDATE namespaces SET quota_bytes=1024 WHERE name='tester';"
   ```
2. **Try to Push:**
   ```bash
   docker push localhost:5000/tester/my-app:v1
   ```
   **Expected Output:** `denied: quota exceeded` or `500 Internal Server Error (Quota exceeded)`

3. **Reset Quota:**
   ```bash
   docker exec registryx-db psql -U registryx -d registryx -c "UPDATE namespaces SET quota_bytes=5368709120 WHERE name='tester';"
   ```
