# Test Script: Complete Private Isolation
# This script demonstrates that each user can only see their own images

$baseUrl = "http://localhost:5000"

Write-Host "=== Complete Private Isolation Test ===" -ForegroundColor Cyan
Write-Host "This will demonstrate that each user can ONLY see their own images.`n"

# Step 1: Mine user pushes an image
Write-Host "Step 1: User 'mine' will push an image..." -ForegroundColor Yellow
Write-Host "   Run this command in another terminal:"
Write-Host "   docker login localhost:5000 -u mine -p 123456" -ForegroundColor Green
Write-Host "   docker tag alpine:latest localhost:5000/test-mine:v1" -ForegroundColor Green
Write-Host "   docker push localhost:5000/test-mine:v1" -ForegroundColor Green
Write-Host ""

# Step 2: Joy user pushes an image
Write-Host "Step 2: User 'joy' will push an image..." -ForegroundColor Yellow
Write-Host "   Run this command in another terminal:"
Write-Host "   docker login localhost:5000 -u joy -p 123456" -ForegroundColor Green
Write-Host "   docker tag alpine:latest localhost:5000/test-joy:v1" -ForegroundColor Green
Write-Host "   docker push localhost:5000/test-joy:v1" -ForegroundColor Green
Write-Host ""

Write-Host "After pushing, press Enter to verify isolation..." -ForegroundColor Cyan
Read-Host

# Verification
function Get-Token {
    param ($u, $p)
    $tokenUrl = "$baseUrl/auth/token?service=registry&scope=repository:*:pull,push"
    $headers = @{ "Authorization" = "Basic " + [Convert]::ToBase64String([Text.Encoding]::ASCII.GetBytes("$($u):$($p)")) }
    return (Invoke-RestMethod -Uri $tokenUrl -Headers $headers -Method Get).token
}

Write-Host "`n=== Verification ===" -ForegroundColor Cyan

# Mine's view
Write-Host "`n1. What 'mine' user sees:" -ForegroundColor Yellow
$mineToken = Get-Token "mine" "123456"
$mineCatalog = Invoke-RestMethod -Uri "$baseUrl/v2/_catalog" -Headers @{ "Authorization" = "Bearer $mineToken" } -Method Get
if ($mineCatalog.repositories.Count -gt 0) {
    foreach ($repo in $mineCatalog.repositories) {
        Write-Host "   - $repo" -ForegroundColor Green
    }
}
else {
    Write-Host "   (No images)" -ForegroundColor Gray
}

# Joy's view
Write-Host "`n2. What 'joy' user sees:" -ForegroundColor Yellow
$joyToken = Get-Token "joy" "123456"
$joyCatalog = Invoke-RestMethod -Uri "$baseUrl/v2/_catalog" -Headers @{ "Authorization" = "Bearer $joyToken" } -Method Get
if ($joyCatalog.repositories.Count -gt 0) {
    foreach ($repo in $joyCatalog.repositories) {
        Write-Host "   - $repo" -ForegroundColor Green
    }
}
else {
    Write-Host "   (No images)" -ForegroundColor Gray
}

# Analysis
Write-Host "`n=== Result ===" -ForegroundColor Cyan
$mineCanSeeJoy = $mineCatalog.repositories | Where-Object { $_ -like "*joy*" }
$joyCanSeeMine = $joyCatalog.repositories | Where-Object { $_ -like "*mine*" }

if ($mineCanSeeJoy.Count -eq 0 -and $joyCanSeeMine.Count -eq 0) {
    Write-Host "SUCCESS: Complete isolation working!" -ForegroundColor Green
    Write-Host "  - Mine cannot see Joy's images" -ForegroundColor Green
    Write-Host "  - Joy cannot see Mine's images" -ForegroundColor Green
}
else {
    Write-Host "FAILURE: Isolation broken!" -ForegroundColor Red
    if ($mineCanSeeJoy.Count -gt 0) {
        Write-Host "  - Mine can see Joy's images: $mineCanSeeJoy" -ForegroundColor Red
    }
    if ($joyCanSeeMine.Count -gt 0) {
        Write-Host "  - Joy can see Mine's images: $joyCanSeeMine" -ForegroundColor Red
    }
}
