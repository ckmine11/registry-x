
# Script to populate all users with a vulnerable image
$Users = @(
    @{ Name = "admin"; Pass = "password123" },
    @{ Name = "joy"; Pass = "password123" },
    @{ Name = "testcost"; Pass = "password123" },
    @{ Name = "cost_user"; Pass = "password123" },
    @{ Name = "dashboard_user"; Pass = "password123" }
)

$VulnerableImage = "vulnerables/web-dvwa" # A known vulnerable image
# Or use python:2.7-slim (has many CVEs)
$BaseImage = "python:2.7-slim"

Write-Host "Pulling base vulnerable image ($BaseImage)..." -ForegroundColor Cyan
docker pull $BaseImage

foreach ($User in $Users) {
    $Username = $User.Name
    $Password = $User.Pass
    
    Write-Host "Processing User: $Username" -ForegroundColor Yellow
    
    # 1. Login
    echo $Password | docker login localhost:5000 -u $Username --password-stdin
    if ($LASTEXITCODE -ne 0) {
        Write-Host "Failed to login as $Username - Skipping" -ForegroundColor Red
        continue
    }
    
    # 2. Tag
    $TargetTag = "localhost:5000/$Username/vulnerable-app:latest"
    docker tag $BaseImage $TargetTag
    
    # 3. Push
    Write-Host "Pushing to $TargetTag..." -ForegroundColor Cyan
    docker push $TargetTag
    
    if ($LASTEXITCODE -eq 0) {
        Write-Host "SUCCESS: Pushed to $Username" -ForegroundColor Green
    }
    else {
        Write-Host "ERROR: Failed to push to $Username" -ForegroundColor Red
    }
}
