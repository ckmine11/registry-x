Write-Host "Starting RegistryX in Production Mode..." -ForegroundColor Cyan

# 1. Stop existing instances
docker-compose -f deploy/docker-compose.yml --env-file .env down

# 2. Start in Detached Mode (Build ensures latest code)
docker-compose -f deploy/docker-compose.yml --env-file .env up -d --build

# 3. Health Check
Write-Host "Waiting for services to initialize (10s)..." -ForegroundColor Yellow
Start-Sleep -Seconds 10

try {
    $response = Invoke-RestMethod -Uri "http://localhost:5000/api/v1/health-check" -ErrorAction Stop
    if ($response.status -eq "ok") {
        Write-Host "SUCCESS: RegistryX is ONLINE at http://localhost:5000" -ForegroundColor Green
        Write-Host "Version: $($response.version)" -ForegroundColor Gray
    }
    else {
        Write-Host "WARNING: Service is up but reported status: $($response.status)" -ForegroundColor Yellow
    }
}
catch {
    Write-Host "ERROR: Health check failed. Check docker logs." -ForegroundColor Red
    docker logs registryx-backend --tail 20
}
