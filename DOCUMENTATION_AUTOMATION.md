# Documentation Automation Setup Guide

This guide explains how to set up automatic documentation updates for RegistryX.

---

## 🎯 Overview

The documentation automation system includes:

1. **Workflow Guide** - Step-by-step process for updating docs
2. **Documentation Checker** - Validates docs are up-to-date
3. **API Documentation Generator** - Auto-generates API docs from code
4. **Pre-Commit Hook** - Reminds you to update docs before committing

---

## 📁 Files Created

```
registry-x/
├── .agent/
│   └── workflows/
│       └── update_documentation.md    # Workflow guide
├── scripts/
│   ├── check_docs.ps1                 # Documentation checker
│   ├── generate_api_docs.ps1          # API doc generator
│   └── pre-commit-hook.ps1            # Git pre-commit hook
└── DOCUMENTATION_AUTOMATION.md        # This file
```

---

## 🚀 Quick Start

### 1. Install Git Pre-Commit Hook (Recommended)

This will remind you to update documentation when committing code changes.

**Windows (PowerShell):**
```powershell
# Copy pre-commit hook to git hooks directory
Copy-Item scripts\pre-commit-hook.ps1 .git\hooks\pre-commit

# Make it executable (if on WSL/Git Bash)
chmod +x .git/hooks/pre-commit
```

**Linux/Mac:**
```bash
# Copy and make executable
cp scripts/pre-commit-hook.ps1 .git/hooks/pre-commit
chmod +x .git/hooks/pre-commit
```

### 2. Run Documentation Checker

Check if your documentation is up-to-date:

```powershell
.\scripts\check_docs.ps1
```

**What it checks:**
- ✅ API endpoints (code vs docs)
- ✅ Database tables (migrations vs docs)
- ✅ Required documentation files exist
- ✅ TODOs in documentation
- ✅ Version number consistency
- ✅ Broken internal links
- ✅ Documentation freshness

### 3. Generate API Documentation

Auto-generate API documentation from code:

```powershell
.\scripts\generate_api_docs.ps1
```

**Output:**
- Creates `API_ENDPOINTS.md` with all endpoints
- Categorizes endpoints (Auth, Repos, Scanning, etc.)
- Includes curl examples
- Shows summary statistics

---

## 📋 Workflow for Adding New Features

### Step-by-Step Process

**1. Develop Your Feature**
```powershell
# Write code as usual
code backend/pkg/myfeature/service.go
```

**2. Before Committing - Generate API Docs**
```powershell
# Generate API documentation
.\scripts\generate_api_docs.ps1

# This creates API_ENDPOINTS.md
```

**3. Check Documentation Status**
```powershell
# Run documentation checker
.\scripts\check_docs.ps1

# Review any warnings
```

**4. Update Documentation**

Follow the workflow guide:
```powershell
# Open workflow guide
code .\.agent\workflows\update_documentation.md
```

**Update these files based on your changes:**

| Change Type | Files to Update |
|-------------|----------------|
| New API endpoint | README.md, FEATURES_WALKTHROUGH.md, QUICK_REFERENCE.md |
| New feature | README.md, FEATURES_WALKTHROUGH.md, DOCUMENTATION_INDEX.md |
| Database change | ARCHITECTURE.md, FEATURES_WALKTHROUGH.md |
| Architecture change | ARCHITECTURE.md |
| Deployment change | DEPLOYMENT_GUIDE.md, README.md |

**5. Verify Documentation**
```powershell
# Run checker again
.\scripts\check_docs.ps1

# Should show no issues
```

**6. Commit Changes**
```powershell
# Stage code and documentation
git add .

# Commit (pre-commit hook will run)
git commit -m "feat: Add new feature with documentation"
```

---

## 🔧 Detailed Tool Usage

### Documentation Checker (`check_docs.ps1`)

**Usage:**
```powershell
.\scripts\check_docs.ps1
```

**Example Output:**
```
🔍 RegistryX Documentation Checker
=================================

📡 Checking API Endpoints...
   Found 45 API endpoints in code
   Found 45 documented endpoints
   ✅ Endpoint documentation looks good

🗄️  Checking Database Schema...
   Found 15 database tables in migrations
   Found 15/15 tables documented in ARCHITECTURE.md
   ✅ Database schema documentation looks good

📚 Checking Documentation Files...
   ✅ README.md
   ✅ FEATURES_WALKTHROUGH.md
   ✅ ARCHITECTURE.md
   ✅ QUICK_REFERENCE.md
   ✅ DOCUMENTATION_INDEX.md
   ✅ DEPLOYMENT_GUIDE.md

🔧 Checking for TODOs in Documentation...
   ✅ No TODOs found

🔢 Checking Version Numbers...
   Backend version: 2.4
   README version: 2.4
   ✅ Version numbers are consistent

🔗 Checking Internal Links...
   ✅ No broken internal links found

📅 Checking Documentation Freshness...
   Last code change: 2026-02-07 22:30
   Last doc update: 2026-02-07 22:45
   ✅ Documentation is up to date

=================================
✅ Documentation check passed!
   No issues found.
```

**Exit Codes:**
- `0` - All checks passed
- `1` - Issues found (warnings)

---

### API Documentation Generator (`generate_api_docs.ps1`)

**Usage:**
```powershell
.\scripts\generate_api_docs.ps1
```

**Example Output:**
```
📝 RegistryX API Documentation Generator
========================================

🔍 Scanning backend/main.go for API endpoints...

Found 45 API endpoints

✅ Generated documentation: API_ENDPOINTS.md

📊 Summary:
   Total endpoints: 45
   Authentication: 7
   Repositories: 8
   Scanning: 6
   Security: 5
   Costs: 4
   Users: 4
   System: 5
   Webhooks: 4
   OCI Registry: 2

💡 Next steps:
   1. Review API_ENDPOINTS.md
   2. Copy relevant sections to FEATURES_WALKTHROUGH.md
   3. Add detailed descriptions and examples
   4. Update QUICK_REFERENCE.md with common endpoints
```

**Generated File Structure:**
```markdown
# API Endpoints Reference

## Authentication

### POST /api/v1/auth/login
**Handler:** `Login`
```bash
POST http://localhost:5000/api/v1/auth/login
Content-Type: application/json

Request:
{
  "username": "admin",
  "password": "password123"
}
```

**Example:**
```bash
curl -X POST http://localhost:5000/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"password123"}'
```

---

[... more endpoints ...]
```

---

### Pre-Commit Hook (`pre-commit-hook.ps1`)

**What it does:**
- Runs automatically before each commit
- Checks if code changed without docs
- Warns about specific changes (API endpoints, DB tables, services)
- Allows you to abort commit to update docs first

**Example Interaction:**
```
🔍 Checking for documentation updates...

⚠️  WARNING: Code changes detected but no documentation updates!

You are committing code changes without updating documentation.

Please consider updating:
  • README.md - If adding new features
  • FEATURES_WALKTHROUGH.md - For detailed feature docs
  • ARCHITECTURE.md - For architecture changes
  • QUICK_REFERENCE.md - For new commands/APIs

Run documentation checker:
  .\scripts\check_docs.ps1

Generate API docs:
  .\scripts\generate_api_docs.ps1

See workflow:
  .\.agent\workflows\update_documentation.md

Continue with commit anyway? (y/N): n

❌ Commit aborted. Please update documentation first.
```

**Bypass Hook (Not Recommended):**
```powershell
# Skip pre-commit hook (emergency only)
git commit --no-verify -m "message"
```

---

## 📝 Documentation Templates

### New Feature Template

When adding a new feature, use this template in `FEATURES_WALKTHROUGH.md`:

```markdown
## [Feature Name]

### Overview
[Brief description of what this feature does]

### Key Capabilities
- ✅ Capability 1
- ✅ Capability 2
- ✅ Capability 3

### How It Works
[Detailed explanation with diagrams if needed]

### API Endpoints

**[Endpoint Name]**
```bash
[METHOD] /api/v1/[path]
Authorization: Bearer {token}
Content-Type: application/json

Request:
{
  "field": "value"
}

Response:
{
  "result": "data"
}
```

**Example:**
```bash
curl -X [METHOD] http://localhost:5000/api/v1/[path] \
  -H "Authorization: Bearer {token}" \
  -H "Content-Type: application/json" \
  -d '{"field":"value"}'
```

### Database Schema
```sql
CREATE TABLE [table_name] (
    id UUID PRIMARY KEY,
    -- fields
);
```

### Usage Example
[Step-by-step usage example]

### Configuration
[Any environment variables or configuration needed]
```

---

## 🔄 CI/CD Integration

### GitHub Actions Example

Add to `.github/workflows/docs-check.yml`:

```yaml
name: Documentation Check

on:
  pull_request:
    branches: [main]

jobs:
  check-docs:
    runs-on: windows-latest
    steps:
      - uses: actions/checkout@v3
      
      - name: Run Documentation Checker
        shell: pwsh
        run: |
          .\scripts\check_docs.ps1
      
      - name: Comment on PR
        if: failure()
        uses: actions/github-script@v6
        with:
          script: |
            github.rest.issues.createComment({
              issue_number: context.issue.number,
              owner: context.repo.owner,
              repo: context.repo.repo,
              body: '⚠️ Documentation check failed! Please update documentation for your changes.'
            })
```

---

## 📊 Maintenance Schedule

### Daily
- Pre-commit hook runs automatically
- Developers update docs with code changes

### Weekly
```powershell
# Run documentation checker
.\scripts\check_docs.ps1

# Generate fresh API docs
.\scripts\generate_api_docs.ps1

# Review and update as needed
```

### Monthly
- Complete documentation review
- Update screenshots if UI changed
- Review and close documentation TODOs
- Update version numbers

### Before Release
```powershell
# Full documentation audit
.\scripts\check_docs.ps1

# Generate API docs
.\scripts\generate_api_docs.ps1

# Manual review of all docs
# Update CHANGELOG
# Update version numbers
```

---

## 🎯 Best Practices

### 1. Update Docs with Code
- **Don't wait** - Update docs in the same commit as code
- **Use templates** - Follow the templates provided
- **Test examples** - Verify all curl commands work

### 2. Run Checker Regularly
```powershell
# Before committing
.\scripts\check_docs.ps1

# Before pushing
.\scripts\check_docs.ps1

# Before creating PR
.\scripts\check_docs.ps1
```

### 3. Use API Generator
```powershell
# After adding new endpoints
.\scripts\generate_api_docs.ps1

# Copy to FEATURES_WALKTHROUGH.md
# Add descriptions and examples
```

### 4. Keep Docs DRY (Don't Repeat Yourself)
- Use cross-references between docs
- Link to detailed docs from README
- Use DOCUMENTATION_INDEX.md for navigation

### 5. Version Everything
- Update version numbers together
- Keep CHANGELOG up to date
- Tag releases with matching docs

---

## 🆘 Troubleshooting

### Pre-Commit Hook Not Running

**Problem:** Hook doesn't execute on commit

**Solution:**
```powershell
# Check if hook exists
ls .git\hooks\pre-commit

# Make executable (Linux/Mac)
chmod +x .git/hooks/pre-commit

# Verify it's the right file
cat .git/hooks/pre-commit
```

### Documentation Checker Fails

**Problem:** Checker reports false positives

**Solution:**
```powershell
# Review the specific warnings
.\scripts\check_docs.ps1

# Check if endpoints are actually documented
# May need to update regex patterns in checker
```

### API Generator Misses Endpoints

**Problem:** Some endpoints not detected

**Solution:**
```powershell
# Check endpoint pattern in code
# Generator looks for specific patterns
# May need to update regex in generate_api_docs.ps1
```

---

## 📚 Additional Resources

- **Workflow Guide**: `.agent/workflows/update_documentation.md`
- **Documentation Index**: `DOCUMENTATION_INDEX.md`
- **Feature Walkthrough**: `FEATURES_WALKTHROUGH.md`
- **Architecture**: `ARCHITECTURE.md`

---

## ✅ Summary

With this automation system:

✅ **Pre-commit hook** reminds you to update docs  
✅ **Documentation checker** validates completeness  
✅ **API generator** creates doc templates automatically  
✅ **Workflow guide** provides step-by-step process  
✅ **Templates** ensure consistency  
✅ **CI/CD integration** enforces documentation  

**Your documentation will always stay up-to-date!** 🚀

---

## 🎓 Quick Reference

```powershell
# Check documentation
.\scripts\check_docs.ps1

# Generate API docs
.\scripts\generate_api_docs.ps1

# Install pre-commit hook
Copy-Item scripts\pre-commit-hook.ps1 .git\hooks\pre-commit

# View workflow
code .\.agent\workflows\update_documentation.md

# Update docs (manual)
code README.md FEATURES_WALKTHROUGH.md ARCHITECTURE.md
```

**Happy documenting!** 📝
