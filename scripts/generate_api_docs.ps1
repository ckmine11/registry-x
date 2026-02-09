# API Documentation Generator
# Extracts API endpoints from code and generates documentation templates

Write-Host "📝 RegistryX API Documentation Generator" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

# Check if we're in the right directory
if (-not (Test-Path "backend/main.go")) {
    Write-Host "❌ Error: Run this script from the registry-x root directory" -ForegroundColor Red
    exit 1
}

# Parse main.go for API endpoints
Write-Host "🔍 Scanning backend/main.go for API endpoints..." -ForegroundColor Yellow
Write-Host ""

$mainGoContent = Get-Content "backend/main.go" -Raw

# Extract API endpoints
$endpoints = @()
$pattern = 'apiV1\.(?:Handle|HandleFunc)\("([^"]+)",\s*(?:authMiddleware\()?(?:http\.HandlerFunc\()?(?:\w+\.)?(\w+)\)?.*?Methods\("([^"]+)"\)'
$endpointMatches = [regex]::Matches($mainGoContent, $pattern)

foreach ($match in $endpointMatches) {
    $path = $match.Groups[1].Value
    $handler = $match.Groups[2].Value
    $method = $match.Groups[3].Value
    
    $endpoints += [PSCustomObject]@{
        Method  = $method
        Path    = "/api/v1$path"
        Handler = $handler
    }
}

# Also check for v2 endpoints (OCI registry)
$v2Pattern = 'v2\.(?:Handle|HandleFunc)\("([^"]+)",\s*(?:authMiddleware\()?(?:http\.HandlerFunc\()?(?:\w+\.)?(\w+)\)?.*?Methods\("([^"]+)"\)'
$v2Matches = [regex]::Matches($mainGoContent, $v2Pattern)

foreach ($match in $v2Matches) {
    $path = $match.Groups[1].Value
    $handler = $match.Groups[2].Value
    $method = $match.Groups[3].Value
    
    $endpoints += [PSCustomObject]@{
        Method  = $method
        Path    = "/v2$path"
        Handler = $handler
    }
}

Write-Host "Found $($endpoints.Count) API endpoints" -ForegroundColor Green
Write-Host ""

# Group by category
$categories = @{
    'Authentication' = @()
    'Repositories'   = @()
    'Scanning'       = @()
    'Security'       = @()
    'Costs'          = @()
    'Users'          = @()
    'System'         = @()
    'Webhooks'       = @()
    'OCI Registry'   = @()
    'Other'          = @()
}

foreach ($endpoint in $endpoints) {
    $path = $endpoint.Path
    
    if ($path -match '/auth/') {
        $categories['Authentication'] += $endpoint
    }
    elseif ($path -match '/repositories') {
        $categories['Repositories'] += $endpoint
    }
    elseif ($path -match '/scan|/vulnerabilities') {
        $categories['Scanning'] += $endpoint
    }
    elseif ($path -match '/security|/policy') {
        $categories['Security'] += $endpoint
    }
    elseif ($path -match '/costs') {
        $categories['Costs'] += $endpoint
    }
    elseif ($path -match '/users') {
        $categories['Users'] += $endpoint
    }
    elseif ($path -match '/system|/health|/config') {
        $categories['System'] += $endpoint
    }
    elseif ($path -match '/webhooks') {
        $categories['Webhooks'] += $endpoint
    }
    elseif ($path -match '^/v2/') {
        $categories['OCI Registry'] += $endpoint
    }
    else {
        $categories['Other'] += $endpoint
    }
}

# Generate markdown documentation
$output = @"
# API Endpoints Reference

**Auto-generated on $(Get-Date -Format 'yyyy-MM-dd HH:mm:ss')**

This document lists all API endpoints found in the codebase.

---

"@

foreach ($category in $categories.Keys | Sort-Object) {
    $categoryEndpoints = $categories[$category]
    
    if ($categoryEndpoints.Count -gt 0) {
        $output += "`n## $category`n`n"
        
        foreach ($endpoint in ($categoryEndpoints | Sort-Object Method, Path)) {
            $method = $endpoint.Method
            $path = $endpoint.Path
            $handler = $endpoint.Handler
            
            $output += "### $method $path`n`n"
            $output += "**Handler:** ``$handler```n`n"
            $output += '```bash' + "`n"
            $output += "$method http://localhost:5000$path`n"
            
            # Add Authorization header for most endpoints
            if ($path -notmatch '/auth/login|/auth/register|/health|/v2/$') {
                $output += "Authorization: Bearer {token}`n"
            }
            
            # Add example request body for POST/PUT
            if ($method -match 'POST|PUT') {
                $output += "Content-Type: application/json`n`n"
                $output += "Request:`n"
                $output += "{`n"
                $output += '  "field": "value"' + "`n"
                $output += "}`n"
            }
            
            $output += '```' + "`n`n"
            
            # Add example curl command
            $output += "**Example:**`n"
            $output += '```bash' + "`n"
            $output += "curl -X $method http://localhost:5000$path"
            
            if ($path -notmatch '/auth/login|/auth/register|/health|/v2/$') {
                $output += " \```n  -H `"Authorization: Bearer {token}`""
            }
            
            if ($method -match 'POST|PUT') {
                $output += " \```n  -H `"Content-Type: application/json`" \```n  -d '{`"field`":`"value`"}'"
            }
            
            $output += "`n" + '```' + "`n`n"
            $output += "---`n`n"
        }
    }
}

# Add summary
$output += "`n## Summary`n`n"
$output += "**Total Endpoints:** $($endpoints.Count)`n`n"
$output += "**By Category:**`n"
foreach ($category in $categories.Keys | Sort-Object) {
    $count = $categories[$category].Count
    if ($count -gt 0) {
        $output += "- $category`: $count`n"
    }
}

$output += "`n**By Method:**`n"
$methods = $endpoints | Group-Object Method | Sort-Object Name
foreach ($method in $methods) {
    $output += "- $($method.Name): $($method.Count)`n"
}

# Save to file
$outputFile = "API_ENDPOINTS.md"
$output | Out-File -FilePath $outputFile -Encoding UTF8

Write-Host "✅ Generated documentation: $outputFile" -ForegroundColor Green
Write-Host ""
Write-Host "📊 Summary:" -ForegroundColor Cyan
Write-Host "   Total endpoints: $($endpoints.Count)" -ForegroundColor Gray
foreach ($category in $categories.Keys | Sort-Object) {
    $count = $categories[$category].Count
    if ($count -gt 0) {
        Write-Host "   $category`: $count" -ForegroundColor Gray
    }
}
Write-Host ""
Write-Host "💡 Next steps:" -ForegroundColor Yellow
Write-Host "   1. Review $outputFile" -ForegroundColor Gray
Write-Host "   2. Copy relevant sections to FEATURES_WALKTHROUGH.md" -ForegroundColor Gray
Write-Host "   3. Add detailed descriptions and examples" -ForegroundColor Gray
Write-Host "   4. Update QUICK_REFERENCE.md with common endpoints" -ForegroundColor Gray
Write-Host ""
