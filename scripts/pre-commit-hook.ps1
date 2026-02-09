#!/usr/bin/env pwsh
# Git Pre-Commit Hook for Documentation Updates
# Place this in .git/hooks/pre-commit (remove .ps1 extension)

Write-Host ""
Write-Host "🔍 Checking for documentation updates..." -ForegroundColor Cyan

# Get list of staged files
$stagedFiles = git diff --cached --name-only

# Check if code files are being committed
$codeChanged = $false
$docsChanged = $false

foreach ($file in $stagedFiles) {
    if ($file -match '\.(go|ts|tsx|sql)$') {
        $codeChanged = $true
    }
    if ($file -match '\.md$') {
        $docsChanged = $true
    }
}

# If code changed but no docs changed, warn the user
if ($codeChanged -and -not $docsChanged) {
    Write-Host ""
    Write-Host "⚠️  WARNING: Code changes detected but no documentation updates!" -ForegroundColor Yellow
    Write-Host ""
    Write-Host "You are committing code changes without updating documentation." -ForegroundColor Yellow
    Write-Host ""
    Write-Host "Please consider updating:" -ForegroundColor White
    Write-Host "  • README.md - If adding new features" -ForegroundColor Gray
    Write-Host "  • FEATURES_WALKTHROUGH.md - For detailed feature docs" -ForegroundColor Gray
    Write-Host "  • ARCHITECTURE.md - For architecture changes" -ForegroundColor Gray
    Write-Host "  • QUICK_REFERENCE.md - For new commands/APIs" -ForegroundColor Gray
    Write-Host ""
    Write-Host "Run documentation checker:" -ForegroundColor Cyan
    Write-Host "  .\scripts\check_docs.ps1" -ForegroundColor White
    Write-Host ""
    Write-Host "Generate API docs:" -ForegroundColor Cyan
    Write-Host "  .\scripts\generate_api_docs.ps1" -ForegroundColor White
    Write-Host ""
    Write-Host "See workflow:" -ForegroundColor Cyan
    Write-Host "  .\.agent\workflows\update_documentation.md" -ForegroundColor White
    Write-Host ""
    
    # Ask user if they want to continue
    $response = Read-Host "Continue with commit anyway? (y/N)"
    
    if ($response -ne 'y' -and $response -ne 'Y') {
        Write-Host ""
        Write-Host "❌ Commit aborted. Please update documentation first." -ForegroundColor Red
        Write-Host ""
        exit 1
    }
}

# Check for specific patterns that definitely need docs
$needsDocs = $false
$reasons = @()

foreach ($file in $stagedFiles) {
    if ($file -match '\.go$') {
        $diff = git diff --cached $file
        
        # Check for new API endpoints
        if ($diff -match '\+.*HandleFunc.*Methods') {
            $needsDocs = $true
            $reasons += "New API endpoint detected in $file"
        }
        
        # Check for new database tables
        if ($file -match 'migrations.*\.sql$' -and $diff -match '\+.*CREATE TABLE') {
            $needsDocs = $true
            $reasons += "New database table detected in $file"
        }
        
        # Check for new services
        if ($diff -match '\+.*type.*Service struct') {
            $needsDocs = $true
            $reasons += "New service detected in $file"
        }
    }
}

if ($needsDocs) {
    Write-Host ""
    Write-Host "🚨 DOCUMENTATION UPDATE REQUIRED!" -ForegroundColor Red
    Write-Host ""
    Write-Host "The following changes require documentation updates:" -ForegroundColor Yellow
    foreach ($reason in $reasons) {
        Write-Host "  • $reason" -ForegroundColor Yellow
    }
    Write-Host ""
    Write-Host "Please update documentation before committing." -ForegroundColor White
    Write-Host ""
    Write-Host "Run: .\scripts\generate_api_docs.ps1" -ForegroundColor Cyan
    Write-Host ""
    
    $response = Read-Host "Continue anyway? (y/N)"
    
    if ($response -ne 'y' -and $response -ne 'Y') {
        Write-Host ""
        Write-Host "❌ Commit aborted." -ForegroundColor Red
        Write-Host ""
        exit 1
    }
}

# If docs were updated, show a nice message
if ($docsChanged) {
    Write-Host ""
    Write-Host "✅ Documentation updates detected. Great job!" -ForegroundColor Green
    Write-Host ""
}

# Allow commit to proceed
exit 0
