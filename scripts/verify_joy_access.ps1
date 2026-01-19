$username = "joy"
$password = "123456"
$baseUrl = "http://localhost:5000"

# 1. Get Token for JOY
Write-Host "Obtaining auth token for user '$username'..."
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
    Write-Host "Failed to get token for $username. Does the user exist?" -ForegroundColor Red
    # Register joy if not exists (Testing fallback)
    try {
        Write-Host "Attempting to register $username..."
        $regBody = @{ username = $username; password = $password; email = "$username@example.com" } | ConvertTo-Json
        Invoke-RestMethod -Uri "$baseUrl/api/v1/user/register" -Headers @{"Content-Type" = "application/json" } -Method Post -Body $regBody
        # Retry token
        $tokenResponse = Invoke-RestMethod -Uri $tokenUrl -Headers $headers -Method Get
        $token = $tokenResponse.token
        Write-Host "Token obtained after registration."
    }
    catch {
        Write-Host "Registration failed or user already exists but creds wrong." -ForegroundColor Red
        exit 1
    }
}

# 2. Check Catalog Visibility
Write-Host "Checking Catalog visibility for $username..."
$headers = @{ "Authorization" = "Bearer $token" }
try {
    $catalog = Invoke-RestMethod -Uri "$baseUrl/v2/_catalog" -Headers $headers -Method Get
    Write-Host "Visible Repositories:"
    $catalog.repositories | ForEach-Object { Write-Host " - $_" }
    
    if ($catalog.repositories -contains "library/vulnerable-app") {
        Write-Host "NOTE: User sees 'library/vulnerable-app'. This is PUBLIC data." -ForegroundColor Yellow
    }
}
catch {
    Write-Host "Failed to get catalog: $_" -ForegroundColor Red
}

# 3. Check Dashboard Stats
Write-Host "`nChecking Dashboard Stats for $username..."
try {
    $stats = Invoke-RestMethod -Uri "$baseUrl/api/v1/stats" -Headers $headers -Method Get
    Write-Host "Stats: Repos=$($stats.repository_count), Storage=$($stats.storage_used_bytes)"
}
catch {
    Write-Host "Failed to get stats: $_" -ForegroundColor Red
}
