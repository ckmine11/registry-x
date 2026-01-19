# Quick Demo: Mark Images as Signed (Without Cosign)
# This creates dummy signature tags to demonstrate the "Signed" status in RegistryX

Write-Host "=== RegistryX Signature Demo ===" -ForegroundColor Cyan
Write-Host ""
Write-Host "This script creates signature tags to demonstrate the 'Signed' status." -ForegroundColor Yellow
Write-Host ""

# List of images to mark as signed
$images = @(
    @{name = "demo/alpine"; tag = "latest" },
    @{name = "demo/nginx"; tag = "alpine" }
)

Write-Host "Creating signature tags..." -ForegroundColor Cyan
Write-Host ""

foreach ($img in $images) {
    $imageName = $img.name
    $imageTag = $img.tag
    $fullImage = "localhost:5000/$imageName" + ":" + "$imageTag"
    
    Write-Host "Processing: $fullImage" -ForegroundColor White
    
    # Get the image digest
    $digestOutput = docker inspect $fullImage 2>$null | ConvertFrom-Json
    if (-not $digestOutput) {
        Write-Host "  Image not found, skipping..." -ForegroundColor Red
        continue
    }
    
    # Extract digest from RepoDigests
    $repoDigest = $digestOutput[0].RepoDigests[0]
    if (-not $repoDigest) {
        Write-Host "  No digest found, skipping..." -ForegroundColor Red
        continue
    }
    
    # Extract just the sha256 part
    $digest = $repoDigest.Split('@')[1]
    Write-Host "  Digest: $digest" -ForegroundColor Gray
    
    # Create Cosign-style signature tag
    $sigTag = $digest.Replace('sha256:', 'sha256-') + '.sig'
    Write-Host "  Signature tag: $sigTag" -ForegroundColor Gray
    
    # Tag and push
    $sigImage = "localhost:5000/$imageName" + ":" + "$sigTag"
    docker tag $fullImage $sigImage 2>$null
    
    if ($LASTEXITCODE -eq 0) {
        docker push $sigImage 2>&1 | Out-Null
        
        if ($LASTEXITCODE -eq 0) {
            Write-Host "  Signature tag created!" -ForegroundColor Green
        }
        else {
            Write-Host "  Failed to push signature tag" -ForegroundColor Red
        }
    }
    else {
        Write-Host "  Failed to create signature tag" -ForegroundColor Red
    }
    
    Write-Host ""
}

Write-Host "=== Demo Complete ===" -ForegroundColor Green
Write-Host ""
Write-Host "Refresh the RegistryX UI to see images marked as 'Signed'" -ForegroundColor Cyan
Write-Host ""
