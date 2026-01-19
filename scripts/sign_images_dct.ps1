# Alternative: Sign images using Docker Content Trust (DCT)
# This is simpler than Cosign and built into Docker

Write-Host "=== RegistryX Image Signing with Docker Content Trust ===" -ForegroundColor Cyan
Write-Host ""

Write-Host "Note: Docker Content Trust requires a Notary server." -ForegroundColor Yellow
Write-Host "For local testing, we'll use Cosign instead (see sign_images.ps1)" -ForegroundColor Yellow
Write-Host ""

Write-Host "Quick Start with Cosign:" -ForegroundColor Cyan
Write-Host "1. Install Cosign: winget install sigstore.cosign" -ForegroundColor White
Write-Host "2. Run: .\sign_images.ps1" -ForegroundColor White
Write-Host ""

Write-Host "Manual signing example:" -ForegroundColor Cyan
Write-Host "  cosign generate-key-pair" -ForegroundColor Gray
Write-Host "  cosign sign --key cosign.key localhost:5000/demo/alpine:latest" -ForegroundColor Gray
Write-Host ""
