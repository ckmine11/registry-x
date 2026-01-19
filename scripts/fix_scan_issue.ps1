Write-Host "=== Fixing Scan Issue by Re-pushing Images ===" -ForegroundColor Cyan
Write-Host ""

# Step 1: Clear blob cache in database to force re-upload
Write-Host "1. Clearing blob metadata from database..." -ForegroundColor Yellow
docker exec -i registryx-db psql -U registryx -d registryx -c "DELETE FROM blobs;"
Write-Host "   Done!" -ForegroundColor Green
Write-Host ""

# Step 2: Pull fresh images
Write-Host "2. Pulling fresh images from Docker Hub..." -ForegroundColor Yellow
docker pull alpine:latest
docker pull nginx:alpine  
docker pull node:14.17.0
Write-Host "   Done!" -ForegroundColor Green
Write-Host ""

# Step 3: Tag for local registry
Write-Host "3. Tagging images for local registry..." -ForegroundColor Yellow
docker tag alpine:latest localhost:5000/demo/alpine:latest
docker tag nginx:alpine localhost:5000/demo/nginx:alpine
docker tag node:14.17.0 localhost:5000/vulnerable/node:14.17.0
Write-Host "   Done!" -ForegroundColor Green
Write-Host ""

# Step 4: Push to local registry
Write-Host "4. Pushing images to local registry..." -ForegroundColor Yellow
Write-Host "   Pushing alpine..." -ForegroundColor Cyan
docker push localhost:5000/demo/alpine:latest

Write-Host "   Pushing nginx..." -ForegroundColor Cyan
docker push localhost:5000/demo/nginx:alpine

Write-Host "   Pushing node..." -ForegroundColor Cyan
docker push localhost:5000/vulnerable/node:14.17.0

Write-Host "   Done!" -ForegroundColor Green
Write-Host ""

Write-Host "=== All images re-pushed! Scans should now work. ===" -ForegroundColor Green
Write-Host "You can verify by triggering a manual scan in the UI." -ForegroundColor Cyan
