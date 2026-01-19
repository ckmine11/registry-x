# RegistryX Image Signing with Cosign
# This script signs container images using Cosign

Write-Host "=== RegistryX Image Signing Setup ===" -ForegroundColor Cyan
Write-Host ""

# Check if cosign is installed
Write-Host "1. Checking for Cosign..." -ForegroundColor Yellow
$cosignExists = Get-Command cosign -ErrorAction SilentlyContinue
if (-not $cosignExists) {
    Write-Host "   Cosign not found! Installing..." -ForegroundColor Red
    Write-Host ""
    Write-Host "   Please install Cosign from: https://github.com/sigstore/cosign/releases" -ForegroundColor Yellow
    Write-Host "   Or use: winget install sigstore.cosign" -ForegroundColor Yellow
    Write-Host ""
    Write-Host "   After installation, run this script again." -ForegroundColor Cyan
    exit 1
}
Write-Host "   Cosign found: $($cosignExists.Source)" -ForegroundColor Green
Write-Host ""

# Generate key pair if it doesn't exist
Write-Host "2. Checking for signing keys..." -ForegroundColor Yellow
if (-not (Test-Path "cosign.key") -or -not (Test-Path "cosign.pub")) {
    Write-Host "   Generating new key pair..." -ForegroundColor Cyan
    Write-Host "   You will be prompted to enter a password for the private key." -ForegroundColor Yellow
    Write-Host ""
    
    # Generate keys (user will be prompted for password)
    cosign generate-key-pair
    
    if ($LASTEXITCODE -ne 0) {
        Write-Host "   Failed to generate keys!" -ForegroundColor Red
        exit 1
    }
    
    Write-Host ""
    Write-Host "   Keys generated successfully!" -ForegroundColor Green
    Write-Host "   - Private key: cosign.key (keep this secret!)" -ForegroundColor Yellow
    Write-Host "   - Public key: cosign.pub (share this for verification)" -ForegroundColor Yellow
}
else {
    Write-Host "   Using existing keys:" -ForegroundColor Green
    Write-Host "   - cosign.key" -ForegroundColor Gray
    Write-Host "   - cosign.pub" -ForegroundColor Gray
}
Write-Host ""

# List of images to sign
$images = @(
    "localhost:5000/demo/alpine:latest",
    "localhost:5000/demo/nginx:alpine",
    "localhost:5000/vulnerable/node:14.17.0",
    "localhost:5000/test/hello:v1"
)

Write-Host "3. Signing images..." -ForegroundColor Yellow
Write-Host ""

foreach ($image in $images) {
    Write-Host "   Signing: $image" -ForegroundColor Cyan
    
    # Check if image exists
    $imageExists = docker image inspect $image 2>$null
    if (-not $imageExists) {
        Write-Host "   Image not found locally, pulling..." -ForegroundColor Yellow
        docker pull $image 2>$null
        if ($LASTEXITCODE -ne 0) {
            Write-Host "   Failed to pull image, skipping..." -ForegroundColor Red
            continue
        }
    }
    
    # Sign the image
    # Using --yes to skip confirmation and COSIGN_PASSWORD env var for automation
    $env:COSIGN_PASSWORD = ""  # Empty password for demo, or set your password
    cosign sign --key cosign.key --yes $image 2>&1 | Out-Null
    
    if ($LASTEXITCODE -eq 0) {
        Write-Host "   ✓ Signed successfully!" -ForegroundColor Green
    }
    else {
        Write-Host "   ✗ Failed to sign" -ForegroundColor Red
    }
}

Write-Host ""
Write-Host "4. Verifying signatures..." -ForegroundColor Yellow
Write-Host ""

foreach ($image in $images) {
    Write-Host "   Verifying: $image" -ForegroundColor Cyan
    
    cosign verify --key cosign.pub $image 2>&1 | Out-Null
    
    if ($LASTEXITCODE -eq 0) {
        Write-Host "   ✓ Signature valid!" -ForegroundColor Green
    }
    else {
        Write-Host "   ✗ No valid signature found" -ForegroundColor Yellow
    }
}

Write-Host ""
Write-Host "=== Image Signing Complete ===" -ForegroundColor Green
Write-Host ""
Write-Host "Next steps:" -ForegroundColor Cyan
Write-Host "1. Refresh the RegistryX UI to see 'Signed' status" -ForegroundColor White
Write-Host "2. Keep cosign.key secure - it's your private signing key" -ForegroundColor White
Write-Host "3. Share cosign.pub with users who need to verify signatures" -ForegroundColor White
Write-Host ""
