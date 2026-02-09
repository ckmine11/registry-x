---
description: Update documentation when adding new features
---

# Documentation Update Workflow

This workflow ensures all documentation is updated when new features are added to RegistryX.

## When to Use This Workflow

Run this workflow whenever you:
- Add a new feature or API endpoint
- Modify existing functionality
- Add new database tables or schema changes
- Update architecture or dependencies
- Change deployment procedures
- Add new configuration options

## Documentation Update Checklist

### 1. Identify Affected Documentation

Determine which documents need updates based on the change:

| Change Type | Documents to Update |
|-------------|-------------------|
| New API endpoint | README.md, FEATURES_WALKTHROUGH.md, QUICK_REFERENCE.md |
| New feature | README.md, FEATURES_WALKTHROUGH.md, DOCUMENTATION_INDEX.md |
| Database schema | ARCHITECTURE.md, FEATURES_WALKTHROUGH.md |
| Architecture change | ARCHITECTURE.md |
| Deployment change | DEPLOYMENT_GUIDE.md, README.md |
| New script | QUICK_REFERENCE.md |
| Security feature | FEATURES_WALKTHROUGH.md, docs/IMAGE_SIGNING.md or similar |
| Cost feature | FEATURES_WALKTHROUGH.md (Cost Intelligence section) |

### 2. Update README.md

**Sections to check:**
- [ ] Key Features table - Add new feature
- [ ] API Reference - Add new endpoint
- [ ] Usage Examples - Add example if user-facing
- [ ] Configuration - Add new env variables
- [ ] Update version number if applicable

**Template for new feature:**
```markdown
### 🎯 [Feature Name]

**[Feature Description]**

#### Features
- ✅ Feature point 1
- ✅ Feature point 2
- ✅ Feature point 3
```

### 3. Update FEATURES_WALKTHROUGH.md

**Sections to check:**
- [ ] Table of Contents - Add new section
- [ ] Core Features / Advanced Features - Add detailed explanation
- [ ] API Reference - Add complete endpoint documentation
- [ ] Database Schema - Add new tables/columns
- [ ] Usage examples with code

**Template for new API endpoint:**
```markdown
#### [Endpoint Name]

**[Description]**

```bash
[METHOD] /api/v1/[path]

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
  -H "Content-Type: application/json" \
  -d '{"field":"value"}'
```
```

### 4. Update ARCHITECTURE.md

**Sections to check:**
- [ ] High-Level Architecture - If new service/component
- [ ] Data Flow Diagrams - If new flow
- [ ] Database Schema - If new tables
- [ ] Service Layer - If new service
- [ ] Background Workers - If new worker

**Template for new service:**
```markdown
#### [Service Name]

**Purpose**: [What this service does]

**Responsibilities:**
- Responsibility 1
- Responsibility 2

**Dependencies:**
- Dependency 1
- Dependency 2

**API:**
```go
type [Service]Service struct {
    db *sql.DB
    // other fields
}

func New[Service]Service(db *sql.DB) *[Service]Service {
    return &[Service]Service{db: db}
}
```
```

### 5. Update QUICK_REFERENCE.md

**Sections to check:**
- [ ] Common API Calls - Add curl examples
- [ ] Database Commands - Add queries if applicable
- [ ] Testing Scripts - Add new scripts

**Template for new API call:**
```markdown
### [Feature Name]

```bash
# [Description]
curl -X [METHOD] http://localhost:5000/api/v1/[path] \
  -H "Authorization: Bearer {token}" \
  -H "Content-Type: application/json" \
  -d '{"field":"value"}'
```
```

### 6. Update DOCUMENTATION_INDEX.md

**Sections to check:**
- [ ] Feature Documentation Map - Add new feature
- [ ] Technical Documentation Map - If technical change
- [ ] Update documentation statistics

### 7. Update Specialized Docs (if applicable)

- [ ] **DEPLOYMENT_GUIDE.md** - If deployment changes
- [ ] **docs/TRIVY_SCAN_FEATURES.md** - If scanning changes
- [ ] **docs/IMAGE_SIGNING.md** - If signing changes
- [ ] Create new doc in `docs/` if major new feature

### 8. Update Version Numbers

- [ ] Update version in README.md
- [ ] Update "Last Updated" date in DOCUMENTATION_INDEX.md
- [ ] Update version in backend/main.go if applicable

## Step-by-Step Process

### Step 1: Analyze the Change

```bash
# Review what was changed
git diff

# List new files
git status

# Check new API endpoints in code
grep -r "HandleFunc\|Handle" backend/main.go
```

### Step 2: Update Documentation Files

// turbo
```bash
# Open all documentation files
code README.md FEATURES_WALKTHROUGH.md ARCHITECTURE.md QUICK_REFERENCE.md DOCUMENTATION_INDEX.md
```

### Step 3: Follow Templates

Use the templates above for each document type.

### Step 4: Verify Examples

// turbo
```bash
# Test API examples work
curl http://localhost:5000/api/v1/health-check

# Verify commands in QUICK_REFERENCE.md
# Test each curl command
```

### Step 5: Update Cross-References

Ensure all documents link to each other correctly:
- README.md links to detailed docs
- FEATURES_WALKTHROUGH.md references ARCHITECTURE.md
- DOCUMENTATION_INDEX.md has updated map

### Step 6: Commit Documentation

// turbo
```bash
git add *.md docs/*.md
git commit -m "docs: Update documentation for [feature name]"
```

## Automation Scripts

### Script 1: Documentation Checker

Create `scripts/check_docs.ps1`:

```powershell
# Check if documentation is up to date
$apiEndpoints = Select-String -Path "backend/main.go" -Pattern "HandleFunc|Handle" | Measure-Object | Select-Object -ExpandProperty Count
$documentedEndpoints = Select-String -Path "FEATURES_WALKTHROUGH.md" -Pattern "GET |POST |PUT |DELETE " | Measure-Object | Select-Object -ExpandProperty Count

Write-Host "API Endpoints in code: $apiEndpoints"
Write-Host "Documented endpoints: $documentedEndpoints"

if ($documentedEndpoints -lt $apiEndpoints) {
    Write-Host "⚠️  WARNING: Some API endpoints may not be documented!" -ForegroundColor Yellow
    Write-Host "Please update FEATURES_WALKTHROUGH.md"
}
```

### Script 2: Generate API Documentation

Create `scripts/generate_api_docs.ps1`:

```powershell
# Extract API endpoints from code and generate documentation template
$endpoints = Select-String -Path "backend/main.go" -Pattern 'HandleFunc\("([^"]+)".*Methods\("([^"]+)"\)' -AllMatches

Write-Host "# API Endpoints Found" -ForegroundColor Green
Write-Host ""

foreach ($match in $endpoints) {
    $path = $match.Matches.Groups[1].Value
    $method = $match.Matches.Groups[2].Value
    
    Write-Host "## $method $path"
    Write-Host ""
    Write-Host '```bash'
    Write-Host "$method /api/v1$path"
    Write-Host '```'
    Write-Host ""
}
```

## Documentation Templates

### New Feature Template

```markdown
## [Feature Name]

### Overview
[Brief description of what this feature does]

### Key Capabilities
- ✅ Capability 1
- ✅ Capability 2
- ✅ Capability 3

### How It Works
[Detailed explanation]

### API Endpoints

**[Endpoint Name]**
```bash
[METHOD] /api/v1/[path]
```

**Example:**
```bash
curl -X [METHOD] http://localhost:5000/api/v1/[path]
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

### New API Endpoint Template

```markdown
**[Endpoint Description]**
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
```

## Validation Checklist

Before committing documentation updates:

- [ ] All code examples tested and working
- [ ] All curl commands return expected results
- [ ] Cross-references between docs are correct
- [ ] Table of contents updated
- [ ] Version numbers updated
- [ ] "Last Updated" dates updated
- [ ] No broken links
- [ ] Consistent formatting (markdown)
- [ ] Screenshots updated if UI changed
- [ ] Examples use realistic data

## Quick Reference

### Most Common Updates

**New REST API Endpoint:**
1. README.md → API Reference section
2. FEATURES_WALKTHROUGH.md → API Reference section
3. QUICK_REFERENCE.md → Common API Calls

**New Database Table:**
1. ARCHITECTURE.md → Database Schema section
2. FEATURES_WALKTHROUGH.md → Database Schema section

**New Feature:**
1. README.md → Key Features
2. FEATURES_WALKTHROUGH.md → New section
3. DOCUMENTATION_INDEX.md → Feature map

**New Configuration:**
1. README.md → Configuration section
2. FEATURES_WALKTHROUGH.md → Configuration section
3. QUICK_REFERENCE.md → Environment Variables

## Maintenance Schedule

- **After every feature**: Update relevant docs
- **Weekly**: Run documentation checker script
- **Monthly**: Review all docs for accuracy
- **Before release**: Complete documentation audit

## Tips for Good Documentation

1. **Be Specific**: Use exact commands, not placeholders
2. **Test Everything**: Every example should work
3. **Use Examples**: Show, don't just tell
4. **Keep Updated**: Documentation debt compounds
5. **Cross-Reference**: Link related sections
6. **Version Control**: Track doc changes with code
7. **User Perspective**: Write for the reader, not yourself

## Contact

If you're unsure which documentation to update, refer to DOCUMENTATION_INDEX.md or ask the team.
