$user = "mine"
$pass = "123456"
$authEndpoint = "http://localhost:5000/auth/token?service=registry&scope=registry:catalog:*"

# 1. Get Token
$tokenResponse = curl.exe -u "$($user):$($pass)" $authEndpoint 2>$null
$tokenObj = $tokenResponse | ConvertFrom-Json
$token = $tokenObj.token

if (!$token) {
    Write-Host "Failed to get token."
    exit 1
}

Write-Host "Token obtained."

# 2. Get Catalog
$catalogResponse = curl.exe -H "Authorization: Bearer $token" http://localhost:5000/v2/_catalog 2>$null
Write-Host "Catalog Response:"
Write-Host $catalogResponse

# 3. Check for vulnerable-app
if ($catalogResponse -match "vulnerable-app") {
    Write-Host "SUCCESS: vulnerable-app found in catalog."
}
else {
    Write-Host "FAILURE: vulnerable-app NOT found in catalog."
}
