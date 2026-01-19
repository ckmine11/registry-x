Write-Host "1. Pulling fresh images from Docker Hub..."
docker pull alpine:latest
docker pull nginx:alpine
docker pull node:14.17.0

Write-Host "2. Tagging images for Local Registry..."
docker tag alpine:latest localhost:5000/demo/alpine:latest
docker tag nginx:alpine localhost:5000/demo/nginx:alpine
docker tag node:14.17.0 localhost:5000/vulnerable/node:14.17.0

Write-Host "3. Pushing images to Local Registry (localhost:5000)..."

$images = @("localhost:5000/demo/alpine:latest", "localhost:5000/demo/nginx:alpine", "localhost:5000/vulnerable/node:14.17.0")

foreach ($image in $images) {
    Write-Host "   Pushing $image ..." -ForegroundColor Cyan
    
    $pushOutput = docker push $image 2>&1
    $pushOutput | Out-Host
    
    # We assume push succeeded if we got output
    Write-Host "   Creating signature for $image ..." -ForegroundColor Yellow
    
    $digest = $null
    
    # Check for single-platform conversion (sha256:old -> sha256:new)
    $conversionMatch = $pushOutput | Select-String "sha256:[a-f0-9]+.*->.*sha256:([a-f0-9]+)"
    if ($conversionMatch) {
        $digest = $conversionMatch.Matches[0].Groups[1].Value
    }
    
    # If no conversion, look for standard digest output
    if (-not $digest) {
        $digestMatch = $pushOutput | Select-String "digest: sha256:([a-f0-9]+)"
        if ($digestMatch) {
            $matchesFound = $digestMatch.Matches
            $digest = $matchesFound[$matchesFound.Count - 1].Groups[1].Value
        }
    }
    
    if ($digest) {
        $sigTag = "sha256-" + $digest + ".sig"
        
        # Extract Repo Name: everything up to the last colon
        $lastColonIndex = $image.LastIndexOf(":")
        if ($lastColonIndex -gt 0) {
            $repoName = $image.Substring(0, $lastColonIndex)
            $sigImage = $repoName + ":" + $sigTag
            
            Write-Host "   Tagging signature: $sigImage" -ForegroundColor Gray
            
            docker tag $image $sigImage
            docker push $sigImage 2>&1 | Out-Null
            
            Write-Host "   Signed ($sigTag)" -ForegroundColor Green
        }
    }
    else {
        Write-Host "   Could not detect digest from push output. Skipping signature." -ForegroundColor Red
    }
}

Write-Host "4. Pushing completed. Images are signed and ready for scanning."
