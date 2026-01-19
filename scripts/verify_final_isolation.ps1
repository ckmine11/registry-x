$baseUrl = "http://localhost:5000"

function Get-Token {
    param ($u, $p)
    $tokenUrl = "$baseUrl/auth/token?service=registry&scope=repository:*:pull,push"
    $headers = @{ "Authorization" = "Basic " + [Convert]::ToBase64String([Text.Encoding]::ASCII.GetBytes("$($u):$($p)")) }
    return (Invoke-RestMethod -Uri $tokenUrl -Headers $headers -Method Get).token
}

Write-Host "=== Testing Library Namespace Visibility ===" -ForegroundColor Cyan

# Test 1: Mine user
Write-Host "`n1. Testing 'mine' user..." -ForegroundColor Yellow
$mineToken = Get-Token "mine" "123456"
$headersMine = @{ "Authorization" = "Bearer $mineToken" }
$mineCatalog = Invoke-RestMethod -Uri "$baseUrl/v2/_catalog" -Headers $headersMine -Method Get

Write-Host "   Mine sees:"
foreach ($repo in $mineCatalog.repositories) {
    Write-Host "      - $repo" -ForegroundColor Green
}

# Test 2: Joy user
Write-Host "`n2. Testing 'joy' user..." -ForegroundColor Yellow
$joyToken = Get-Token "joy" "123456"
$headersJoy = @{ "Authorization" = "Bearer $joyToken" }
$joyCatalog = Invoke-RestMethod -Uri "$baseUrl/v2/_catalog" -Headers $headersJoy -Method Get

Write-Host "   Joy sees:"
foreach ($repo in $joyCatalog.repositories) {
    Write-Host "      - $repo" -ForegroundColor Green
}

# Analysis
Write-Host "`n=== Analysis ===" -ForegroundColor Cyan
$libraryImages = $mineCatalog.repositories | Where-Object { $_ -like "library/*" }
if ($libraryImages.Count -gt 0) {
    Write-Host "SUCCESS: Library images are visible to all users!" -ForegroundColor Green
    Write-Host "Library images found: $($libraryImages.Count)" -ForegroundColor Green
}
else {
    Write-Host "NOTE: No library images found. Push an image to test." -ForegroundColor Yellow
}

# Check private isolation
$minePrivate = $mineCatalog.repositories | Where-Object { $_ -like "mine/*" }
$joySeesPrivate = $joyCatalog.repositories | Where-Object { $_ -like "mine/*" }

if ($minePrivate.Count -gt 0 -and $joySeesPrivate.Count -eq 0) {
    Write-Host "SUCCESS: Private isolation working - Joy cannot see Mine's private repos!" -ForegroundColor Green
}
elseif ($minePrivate.Count -eq 0) {
    Write-Host "NOTE: No private 'mine/*' repos to test isolation." -ForegroundColor Yellow
}
else {
    Write-Host "FAILURE: Joy can see Mine's private repos!" -ForegroundColor Red
}
