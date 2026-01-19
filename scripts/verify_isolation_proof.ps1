$baseUrl = "http://localhost:5000"

function Get-Token {
    param ($u, $p)
    $tokenUrl = "$baseUrl/auth/token?service=registry&scope=repository:*:pull,push"
    $headers = @{ "Authorization" = "Basic " + [Convert]::ToBase64String([Text.Encoding]::ASCII.GetBytes("$($u):$($p)")) }
    return (Invoke-RestMethod -Uri $tokenUrl -Headers $headers -Method Get).token
}

# 1. Setup: Mine creates a PRIVATE repo
Write-Host "1. 'mine' user creating PRIVATE repository 'mine/secret-project'..."
$mineToken = Get-Token "mine" "123456"
# We can't actually "push" easily without docker client, but we can try to RegisterManifest explicitly via API or create a placeholder.
# Actually, the quickest way to create a repo entry is to simulate a push or use the CreateRepository endpoint if it exists.
# Checking main.go: apiV1.Handle("/repositories", ...).Methods("POST") -> CreateRepository.
$headersMine = @{ "Authorization" = "Bearer $mineToken"; "Content-Type" = "application/json" }

try {
    $body = @{ name = "mine/secret-project" } | ConvertTo-Json
    Invoke-RestMethod -Uri "$baseUrl/api/v1/repositories" -Headers $headersMine -Method Post -Body $body
    Write-Host "   -> Success: Private repo 'mine/secret-project' created." -ForegroundColor Green
}
catch {
    Write-Host "   -> Failed to create private repo (might already exist): $_" -ForegroundColor Yellow
}

# 2. Check: Joy checks catalog
Write-Host "`n2. 'joy' user checking Catalog..."
$joyToken = Get-Token "joy" "123456"
$headersJoy = @{ "Authorization" = "Bearer $joyToken" }
$catalog = Invoke-RestMethod -Uri "$baseUrl/v2/_catalog" -Headers $headersJoy -Method Get

Write-Host "   -> Joy sees the following repositories:"
$foundPrivate = $false
foreach ($repo in $catalog.repositories) {
    Write-Host "      - $repo"
    if ($repo -eq "mine/secret-project") { $foundPrivate = $true }
}

if ($foundPrivate) {
    Write-Host "`nFAILURE: Joy user CAN see 'mine/secret-project'! Isolation is broken." -ForegroundColor Red
}
else {
    Write-Host "`nSUCCESS: Joy user CANNOT see 'mine/secret-project'. Isolation is working." -ForegroundColor Green
    Write-Host "         (Joy only sees 'library/...' because 'library' is public)" -ForegroundColor Cyan
}
