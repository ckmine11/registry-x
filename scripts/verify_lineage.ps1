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

# 2. Get Dependencies (Lineage)
Write-Host "Fetching Dependency Graph..."
$headers = @{
    "Authorization" = "Bearer $token"
    "Content-Type"  = "application/json"
}

try {
    $graph = Invoke-RestMethod -Uri "$baseUrl/api/v1/dependencies" -Headers $headers -Method Get
    Write-Host "Graph Response:"
    # Only show if not empty
    if ($graph.nodes.Count -gt 0) {
        $graph | ConvertTo-Json -Depth 5
        Write-Host "SUCCESS: Dependency graph returned data." -ForegroundColor Green
    }
    else {
        Write-Host "Graph is empty (No dependencies detected yet)." -ForegroundColor Cyan
        Write-Host "This is expected if 'vulnerable-app' has no parent or dependency detection hasn't needed to link it."
        Write-Host "SUCCESS: Endpoint is accessible." -ForegroundColor Green
    }
}
catch {
    Write-Host "Failed to get dependencies: $_" -ForegroundColor Red
    exit 1
}
