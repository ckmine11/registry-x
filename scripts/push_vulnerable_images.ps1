Write-Host "=== PUSHING 3 VULNERABLE IMAGES ==="

# Image 1: Old Node.js (known vulnerabilities)
Write-Host "`n1. Pulling node:14.17.0 (vulnerable)..."
docker pull node:14.17.0
docker tag node:14.17.0 localhost:5000/vulnerable/node:14.17.0
Write-Host "   Pushing to registry..."
docker push localhost:5000/vulnerable/node:14.17.0

# Image 2: Old Nginx (known vulnerabilities)
Write-Host "`n2. Pulling nginx:1.18.0 (vulnerable)..."
docker pull nginx:1.18.0
docker tag nginx:1.18.0 localhost:5000/vulnerable/nginx:1.18.0
Write-Host "   Pushing to registry..."
docker push localhost:5000/vulnerable/nginx:1.18.0

# Image 3: Old Python (known vulnerabilities)
Write-Host "`n3. Pulling python:3.8.0 (vulnerable)..."
docker pull python:3.8.0
docker tag python:3.8.0 localhost:5000/vulnerable/python:3.8.0
Write-Host "   Pushing to registry..."
docker push localhost:5000/vulnerable/python:3.8.0

Write-Host "`n=== ALL 3 VULNERABLE IMAGES PUSHED ==="
Write-Host "`nVerifying images in registry..."
Invoke-RestMethod -Method Get -Uri "http://localhost:5173/v2/vulnerable/node/tags/list"
Invoke-RestMethod -Method Get -Uri "http://localhost:5173/v2/vulnerable/nginx/tags/list"
Invoke-RestMethod -Method Get -Uri "http://localhost:5173/v2/vulnerable/python/tags/list"
