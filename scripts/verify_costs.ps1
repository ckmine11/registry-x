$username = "mine"
$password = "123456"
$baseUrl = "http://localhost:5000"

# 1. Get Token
Write-Host "Obtaining auth token..."
$tokenUrl = "$baseUrl/auth/token?service=registry&scope=repository:*:pull,push"
$headers = @{
    "Authorization" = "Basic " + [Convert]::ToBase64String([Text.Encoding]::ASCII.GetBytes("$($username):$($password)"))
}

try {
    $tokenResponse = Invoke-RestMethod -Uri $tokenUrl -Headers $headers -Method Get
    $token = $tokenResponse.token
    Write-Host "Token obtained."
}
catch {
    Write-Host "Failed to get token." -ForegroundColor Red
    exit 1
}

# 2. Trigger Refresh Costs
Write-Host "Triggering Cost Refresh..."
$headers = @{
    "Authorization" = "Bearer $token"
    "Content-Type"  = "application/json"
}

try {
    Invoke-RestMethod -Uri "$baseUrl/api/v1/costs/refresh" -Headers $headers -Method Post
    Write-Host "Costs refreshed."
}
catch {
    Write-Host "Failed to refresh costs: $_" -ForegroundColor Red
    # Continue anyway to check dashboard
}

# 3. Get Cost Dashboard
Write-Host "Fetching Cost Dashboard..."
try {
    $dashboard = Invoke-RestMethod -Uri "$baseUrl/api/v1/costs/dashboard" -Headers $headers -Method Get
    Write-Host "Dashboard Response:"
    $dashboard | ConvertTo-Json -Depth 5

    # Check for library/vulnerable-app in TopExpensiveImages
    $found = $false
    if ($dashboard.top_expensive_images) {
        foreach ($img in $dashboard.top_expensive_images) {
            if ($img.repository -like "*vulnerable-app*") {
                $found = $true
                break
            }
        }
    }

    if ($found) {
        Write-Host "SUCCESS: vulnerable-app found in Cost Dashboard." -ForegroundColor Green
    }
    else {
        Write-Host "FAILURE: vulnerable-app NOT found in Cost Dashboard." -ForegroundColor Red
        # Also check Total Images
        if ($dashboard.total_images -gt 0) {
            Write-Host "NOTE: Statistics show images ($($dashboard.total_images)), so system is working but maybe not sorting this one first?" -ForegroundColor Yellow
        }
        else {
            Write-Host "FAILURE: Total images is 0." -ForegroundColor Red
        }
    }

}
catch {
    Write-Host "Failed to get dashboard: $_" -ForegroundColor Red
    exit 1
}

# 4. Get Zombie Images
Write-Host "`nFetching Zombie Images..."
try {
    $zombies = Invoke-RestMethod -Uri "$baseUrl/api/v1/costs/zombie-images" -Headers $headers -Method Get
    Write-Host "Zombie Response:"
    $zombies | ConvertTo-Json -Depth 5
}
catch {
    Write-Host "Failed to get zombies: $_" -ForegroundColor Red
}
