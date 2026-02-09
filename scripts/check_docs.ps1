# Documentation Checker Script
# Checks if documentation is synchronized with code

Write-Host "🔍 RegistryX Documentation Checker" -ForegroundColor Cyan
Write-Host "=================================" -ForegroundColor Cyan
Write-Host ""

# Check if we're in the right directory
if (-not (Test-Path "backend/main.go")) {
    Write-Host "❌ Error: Run this script from the registry-x root directory" -ForegroundColor Red
    exit 1
}

# Initialize counters
$issues = 0

# 1. Check API Endpoints
Write-Host "📡 Checking API Endpoints..." -ForegroundColor Yellow

# Count endpoints in code
$codeEndpoints = @()
$mainGoContent = Get-Content "backend/main.go" -Raw
$matches = [regex]::Matches($mainGoContent, 'HandleFunc\("([^"]+)".*?Methods\("([^"]+)"\)')

foreach ($match in $matches) {
    $path = $match.Groups[1].Value
    $method = $match.Groups[2].Value
    $codeEndpoints += "$method $path"
}

Write-Host "   Found $($codeEndpoints.Count) API endpoints in code" -ForegroundColor Gray

# Count documented endpoints
$docEndpoints = @()
if (Test-Path "FEATURES_WALKTHROUGH.md") {
    $docContent = Get-Content "FEATURES_WALKTHROUGH.md" -Raw
    $docMatches = [regex]::Matches($docContent, '(GET|POST|PUT|DELETE|PATCH)\s+/api/v1/[^\s\n]+')
    
    foreach ($match in $docMatches) {
        $docEndpoints += $match.Value
    }
    
    Write-Host "   Found $($docEndpoints.Count) documented endpoints" -ForegroundColor Gray
    
    if ($docEndpoints.Count -lt $codeEndpoints.Count) {
        Write-Host "   ⚠️  Some endpoints may not be documented!" -ForegroundColor Yellow
        $issues++
    } else {
        Write-Host "   ✅ Endpoint documentation looks good" -ForegroundColor Green
    }
} else {
    Write-Host "   ❌ FEATURES_WALKTHROUGH.md not found!" -ForegroundColor Red
    $issues++
}

Write-Host ""

# 2. Check Database Tables
Write-Host "🗄️  Checking Database Schema..." -ForegroundColor Yellow

$migrationFiles = Get-ChildItem "backend/migrations/*.sql" -ErrorAction SilentlyContinue
if ($migrationFiles) {
    $tables = @()
    foreach ($file in $migrationFiles) {
        $content = Get-Content $file.FullName -Raw
        $tableMatches = [regex]::Matches($content, 'CREATE TABLE\s+(?:IF NOT EXISTS\s+)?(\w+)')
        foreach ($match in $tableMatches) {
            $tables += $match.Groups[1].Value
        }
    }
    
    Write-Host "   Found $($tables.Count) database tables in migrations" -ForegroundColor Gray
    
    # Check if documented in ARCHITECTURE.md
    if (Test-Path "ARCHITECTURE.md") {
        $archContent = Get-Content "ARCHITECTURE.md" -Raw
        $documentedTables = 0
        foreach ($table in $tables) {
            if ($archContent -match $table) {
                $documentedTables++
            }
        }
        
        Write-Host "   Found $documentedTables/$($tables.Count) tables documented in ARCHITECTURE.md" -ForegroundColor Gray
        
        if ($documentedTables -lt $tables.Count) {
            Write-Host "   ⚠️  Some tables may not be documented!" -ForegroundColor Yellow
            $undocumented = $tables | Where-Object { $archContent -notmatch $_ }
            Write-Host "   Missing: $($undocumented -join ', ')" -ForegroundColor Yellow
            $issues++
        } else {
            Write-Host "   ✅ Database schema documentation looks good" -ForegroundColor Green
        }
    }
} else {
    Write-Host "   ⚠️  No migration files found" -ForegroundColor Yellow
}

Write-Host ""

# 3. Check Documentation Files Exist
Write-Host "📚 Checking Documentation Files..." -ForegroundColor Yellow

$requiredDocs = @(
    "README.md",
    "FEATURES_WALKTHROUGH.md",
    "ARCHITECTURE.md",
    "QUICK_REFERENCE.md",
    "DOCUMENTATION_INDEX.md",
    "DEPLOYMENT_GUIDE.md"
)

foreach ($doc in $requiredDocs) {
    if (Test-Path $doc) {
        Write-Host "   ✅ $doc" -ForegroundColor Green
    } else {
        Write-Host "   ❌ $doc missing!" -ForegroundColor Red
        $issues++
    }
}

Write-Host ""

# 4. Check for TODO/FIXME in documentation
Write-Host "🔧 Checking for TODOs in Documentation..." -ForegroundColor Yellow

$todoCount = 0
foreach ($doc in $requiredDocs) {
    if (Test-Path $doc) {
        $content = Get-Content $doc -Raw
        $todos = [regex]::Matches($content, 'TODO|FIXME|XXX|\[TBD\]')
        if ($todos.Count -gt 0) {
            Write-Host "   ⚠️  Found $($todos.Count) TODOs in $doc" -ForegroundColor Yellow
            $todoCount += $todos.Count
        }
    }
}

if ($todoCount -eq 0) {
    Write-Host "   ✅ No TODOs found" -ForegroundColor Green
} else {
    Write-Host "   ⚠️  Total TODOs: $todoCount" -ForegroundColor Yellow
}

Write-Host ""

# 5. Check Version Consistency
Write-Host "🔢 Checking Version Numbers..." -ForegroundColor Yellow

$versions = @{}

# Check backend version
if (Test-Path "backend/main.go") {
    $mainContent = Get-Content "backend/main.go" -Raw
    if ($mainContent -match 'VERSION\s+([0-9.]+)') {
        $versions['backend'] = $matches[1]
        Write-Host "   Backend version: $($versions['backend'])" -ForegroundColor Gray
    }
}

# Check README version
if (Test-Path "README.md") {
    $readmeContent = Get-Content "README.md" -Raw
    if ($readmeContent -match 'Version\s+([0-9.]+)') {
        $versions['readme'] = $matches[1]
        Write-Host "   README version: $($versions['readme'])" -ForegroundColor Gray
    }
}

if ($versions.Count -gt 1) {
    $uniqueVersions = $versions.Values | Select-Object -Unique
    if ($uniqueVersions.Count -eq 1) {
        Write-Host "   ✅ Version numbers are consistent" -ForegroundColor Green
    } else {
        Write-Host "   ⚠️  Version numbers are inconsistent!" -ForegroundColor Yellow
        $issues++
    }
}

Write-Host ""

# 6. Check for Broken Links (basic check)
Write-Host "🔗 Checking Internal Links..." -ForegroundColor Yellow

$brokenLinks = 0
foreach ($doc in $requiredDocs) {
    if (Test-Path $doc) {
        $content = Get-Content $doc -Raw
        $links = [regex]::Matches($content, '\[([^\]]+)\]\(([^\)]+)\)')
        
        foreach ($link in $links) {
            $target = $link.Groups[2].Value
            # Check if it's a local file link
            if ($target -match '^[^http].*\.md') {
                if (-not (Test-Path $target)) {
                    Write-Host "   ⚠️  Broken link in $doc : $target" -ForegroundColor Yellow
                    $brokenLinks++
                }
            }
        }
    }
}

if ($brokenLinks -eq 0) {
    Write-Host "   ✅ No broken internal links found" -ForegroundColor Green
} else {
    Write-Host "   ⚠️  Found $brokenLinks broken links" -ForegroundColor Yellow
    $issues++
}

Write-Host ""

# 7. Check Documentation Freshness
Write-Host "📅 Checking Documentation Freshness..." -ForegroundColor Yellow

$codeLastModified = (Get-ChildItem -Path "backend" -Recurse -Filter "*.go" | 
    Sort-Object LastWriteTime -Descending | 
    Select-Object -First 1).LastWriteTime

$docLastModified = (Get-ChildItem -Path "*.md" | 
    Sort-Object LastWriteTime -Descending | 
    Select-Object -First 1).LastWriteTime

Write-Host "   Last code change: $($codeLastModified.ToString('yyyy-MM-dd HH:mm'))" -ForegroundColor Gray
Write-Host "   Last doc update: $($docLastModified.ToString('yyyy-MM-dd HH:mm'))" -ForegroundColor Gray

$daysDiff = ($codeLastModified - $docLastModified).Days
if ($daysDiff -gt 7) {
    Write-Host "   ⚠️  Documentation may be outdated (${daysDiff} days old)" -ForegroundColor Yellow
    $issues++
} else {
    Write-Host "   ✅ Documentation is up to date" -ForegroundColor Green
}

Write-Host ""

# Summary
Write-Host "=================================" -ForegroundColor Cyan
if ($issues -eq 0) {
    Write-Host "✅ Documentation check passed!" -ForegroundColor Green
    Write-Host "   No issues found." -ForegroundColor Green
} else {
    Write-Host "⚠️  Documentation check completed with $issues issue(s)" -ForegroundColor Yellow
    Write-Host "   Please review and update documentation as needed." -ForegroundColor Yellow
    Write-Host "   Run: Get-Help .\.agent\workflows\update_documentation.md" -ForegroundColor Cyan
}
Write-Host ""

# Exit with appropriate code
if ($issues -gt 0) {
    exit 1
} else {
    exit 0
}
