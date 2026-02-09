# Security Hardening Script for RegistryX
# This script helps implement critical security fixes

Write-Host "🔒 RegistryX Security Hardening" -ForegroundColor Cyan
Write-Host "===============================" -ForegroundColor Cyan
Write-Host ""

# Check if we're in the right directory
if (-not (Test-Path ".env")) {
    Write-Host "❌ Error: .env file not found. Run this script from the registry-x root directory" -ForegroundColor Red
    exit 1
}

Write-Host "⚠️  WARNING: This script will modify your configuration for production security" -ForegroundColor Yellow
Write-Host ""
$confirm = Read-Host "Continue? (y/N)"
if ($confirm -ne 'y' -and $confirm -ne 'Y') {
    Write-Host "Aborted." -ForegroundColor Yellow
    exit 0
}

Write-Host ""

# 1. Backup existing .env
Write-Host "📋 Step 1: Backing up existing .env..." -ForegroundColor Yellow
Copy-Item .env .env.backup
Write-Host "   ✅ Backup created: .env.backup" -ForegroundColor Green
Write-Host ""

# 2. Generate strong secrets
Write-Host "🔑 Step 2: Generating strong secrets..." -ForegroundColor Yellow

function Generate-RandomString {
    param([int]$Length = 32)
    $bytes = New-Object byte[] $Length
    $rng = [System.Security.Cryptography.RandomNumberGenerator]::Create()
    $rng.GetBytes($bytes)
    return [Convert]::ToBase64String($bytes).Substring(0, $Length)
}

$jwtSecret = Generate-RandomString -Length 64
$postgresPassword = Generate-RandomString -Length 32
$minioPassword = Generate-RandomString -Length 32

Write-Host "   ✅ Generated JWT_SECRET" -ForegroundColor Green
Write-Host "   ✅ Generated POSTGRES_PASSWORD" -ForegroundColor Green
Write-Host "   ✅ Generated MINIO_ROOT_PASSWORD" -ForegroundColor Green
Write-Host ""

# 3. Get production URLs
Write-Host "🌐 Step 3: Configure production URLs..." -ForegroundColor Yellow
Write-Host ""

$backendUrl = Read-Host "Enter backend URL (e.g., https://registry-api.example.com)"
if ([string]::IsNullOrWhiteSpace($backendUrl)) {
    $backendUrl = "http://localhost:5000"
    Write-Host "   ⚠️  Using default: $backendUrl" -ForegroundColor Yellow
}

$frontendUrl = Read-Host "Enter frontend URL (e.g., https://registry.example.com)"
if ([string]::IsNullOrWhiteSpace($frontendUrl)) {
    $frontendUrl = "http://localhost:5173"
    Write-Host "   ⚠️  Using default: $frontendUrl" -ForegroundColor Yellow
}

Write-Host ""

# 4. Get SMTP credentials
Write-Host "📧 Step 4: Configure SMTP (for password reset)..." -ForegroundColor Yellow
Write-Host ""

$smtpUser = Read-Host "Enter SMTP user (email address)"
if ([string]::IsNullOrWhiteSpace($smtpUser)) {
    $smtpUser = "your-email@gmail.com"
    Write-Host "   ⚠️  Using placeholder: $smtpUser" -ForegroundColor Yellow
}

$smtpPass = Read-Host "Enter SMTP password (app password)" -AsSecureString
$smtpPassPlain = [Runtime.InteropServices.Marshal]::PtrToStringAuto(
    [Runtime.InteropServices.Marshal]::SecureStringToBSTR($smtpPass)
)
if ([string]::IsNullOrWhiteSpace($smtpPassPlain)) {
    $smtpPassPlain = "your-smtp-password"
    Write-Host "   ⚠️  Using placeholder: $smtpPassPlain" -ForegroundColor Yellow
}

Write-Host ""

# 5. Determine environment
Write-Host "🏗️  Step 5: Select environment..." -ForegroundColor Yellow
Write-Host ""
Write-Host "1. Development (HTTP, relaxed security)"
Write-Host "2. Production (HTTPS, strict security)"
Write-Host ""
$envChoice = Read-Host "Select environment (1 or 2)"

$policyEnv = "dev"
$minioSecure = "false"
$corsOrigins = "http://localhost:5173"

if ($envChoice -eq "2") {
    $policyEnv = "prod"
    $minioSecure = "true"
    $corsOrigins = "$frontendUrl,$backendUrl"
    Write-Host "   ✅ Production environment selected" -ForegroundColor Green
}
else {
    Write-Host "   ✅ Development environment selected" -ForegroundColor Green
}

Write-Host ""

# 6. Create new .env file
Write-Host "📝 Step 6: Creating new .env file..." -ForegroundColor Yellow

$envContent = @"
# RegistryX Environment Configuration
# Generated: $(Get-Date -Format 'yyyy-MM-dd HH:mm:ss')
# Environment: $policyEnv

# === SECURITY SECRETS ===
# ⚠️  NEVER commit this file to version control!
JWT_SECRET=$jwtSecret
POSTGRES_PASSWORD=$postgresPassword
MINIO_ROOT_PASSWORD=$minioPassword

# === MINIO CONFIGURATION ===
MINIO_ROOT_USER=admin
S3_BUCKET=registryx-data
MINIO_SECURE=$minioSecure

# === SMTP CONFIGURATION ===
SMTP_USER=$smtpUser
SMTP_PASS=$smtpPassPlain

# === ENVIRONMENT ===
POLICY_ENVIRONMENT=$policyEnv

# === URLS (Production) ===
BACKEND_URL=$backendUrl
FRONTEND_URL=$frontendUrl

# === CORS ===
CORS_ALLOWED_ORIGINS=$corsOrigins

# === COST INTELLIGENCE ===
STORAGE_COST_PER_GB_MONTH=0.023
BANDWIDTH_COST_PER_GB=0.09
"@

$envContent | Out-File -FilePath .env -Encoding UTF8
Write-Host "   ✅ New .env file created" -ForegroundColor Green
Write-Host ""

# 7. Update .gitignore
Write-Host "🔒 Step 7: Updating .gitignore..." -ForegroundColor Yellow

if (Test-Path .gitignore) {
    $gitignoreContent = Get-Content .gitignore -Raw
    if ($gitignoreContent -notmatch "\.env$") {
        Add-Content .gitignore "`n# Environment variables (contains secrets)`n.env"
        Write-Host "   ✅ Added .env to .gitignore" -ForegroundColor Green
    }
    else {
        Write-Host "   ✅ .env already in .gitignore" -ForegroundColor Green
    }
}
else {
    ".env" | Out-File -FilePath .gitignore -Encoding UTF8
    Write-Host "   ✅ Created .gitignore with .env" -ForegroundColor Green
}

Write-Host ""

# 8. Create .env.example
Write-Host "📄 Step 8: Creating .env.example template..." -ForegroundColor Yellow

$envExampleContent = @"
# RegistryX Environment Configuration Template
# Copy this file to .env and fill in your values

# === SECURITY SECRETS ===
# Generate with: openssl rand -hex 32
JWT_SECRET=your-random-secret-key-here
POSTGRES_PASSWORD=your-secure-password
MINIO_ROOT_PASSWORD=your-minio-password

# === MINIO CONFIGURATION ===
MINIO_ROOT_USER=admin
S3_BUCKET=registryx-data
MINIO_SECURE=false  # Set to true for production with HTTPS

# === SMTP CONFIGURATION ===
# For Gmail: Use App Password (https://support.google.com/accounts/answer/185833)
SMTP_USER=your-email@gmail.com
SMTP_PASS=your-smtp-app-password

# === ENVIRONMENT ===
POLICY_ENVIRONMENT=dev  # dev or prod

# === URLS (Production) ===
BACKEND_URL=http://localhost:5000
FRONTEND_URL=http://localhost:5173

# === CORS ===
CORS_ALLOWED_ORIGINS=http://localhost:5173

# === COST INTELLIGENCE ===
STORAGE_COST_PER_GB_MONTH=0.023
BANDWIDTH_COST_PER_GB=0.09
"@

$envExampleContent | Out-File -FilePath .env.example -Encoding UTF8
Write-Host "   ✅ Created .env.example" -ForegroundColor Green
Write-Host ""

# 9. Display secrets (one-time only)
Write-Host "🔑 Step 9: Generated Secrets (SAVE THESE SECURELY!)" -ForegroundColor Yellow
Write-Host "=================================================" -ForegroundColor Yellow
Write-Host ""
Write-Host "JWT_SECRET:" -ForegroundColor Cyan
Write-Host "  $jwtSecret" -ForegroundColor White
Write-Host ""
Write-Host "POSTGRES_PASSWORD:" -ForegroundColor Cyan
Write-Host "  $postgresPassword" -ForegroundColor White
Write-Host ""
Write-Host "MINIO_ROOT_PASSWORD:" -ForegroundColor Cyan
Write-Host "  $minioPassword" -ForegroundColor White
Write-Host ""
Write-Host "⚠️  IMPORTANT: Save these secrets in a secure password manager!" -ForegroundColor Yellow
Write-Host "⚠️  These secrets will NOT be shown again!" -ForegroundColor Yellow
Write-Host ""

# 10. Security checklist
Write-Host "✅ Security Hardening Complete!" -ForegroundColor Green
Write-Host "===============================" -ForegroundColor Green
Write-Host ""
Write-Host "📋 Next Steps:" -ForegroundColor Cyan
Write-Host ""
Write-Host "1. Review the new .env file" -ForegroundColor White
Write-Host "2. Save the secrets in a password manager" -ForegroundColor White
Write-Host "3. Delete .env.backup after verification" -ForegroundColor White
Write-Host ""

if ($policyEnv -eq "prod") {
    Write-Host "🔒 Production Security Checklist:" -ForegroundColor Yellow
    Write-Host ""
    Write-Host "  [ ] Configure HTTPS/SSL (Nginx + Let's Encrypt)" -ForegroundColor White
    Write-Host "  [ ] Update DATABASE_URL with sslmode=require" -ForegroundColor White
    Write-Host "  [ ] Configure firewall rules" -ForegroundColor White
    Write-Host "  [ ] Set up monitoring and alerting" -ForegroundColor White
    Write-Host "  [ ] Implement rate limiting" -ForegroundColor White
    Write-Host "  [ ] Review SECURITY_AUDIT_REPORT.md" -ForegroundColor White
    Write-Host ""
    Write-Host "📚 See: DEPLOYMENT_GUIDE.md for full production setup" -ForegroundColor Cyan
}
else {
    Write-Host "💡 Development Mode:" -ForegroundColor Yellow
    Write-Host "  - HTTP is enabled (not secure for production)" -ForegroundColor White
    Write-Host "  - Use this for local development only" -ForegroundColor White
    Write-Host "  - Run security hardening again with 'Production' for deployment" -ForegroundColor White
}

Write-Host ""
Write-Host "🎉 Done! Your RegistryX is now more secure!" -ForegroundColor Green
Write-Host ""
