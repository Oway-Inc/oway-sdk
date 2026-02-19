# Release Guide

How to create and publish releases for Oway SDK.

---

## Release Process (Based on Stripe's Pattern)

### TypeScript SDK Release

**1. Update version and changelog:**
```bash
cd packages/typescript

# Update CHANGELOG.md with changes
# Move [Unreleased] items to new version section

# Bump version (updates package.json)
npm version patch  # 0.1.0 → 0.1.1
# or
npm version minor  # 0.1.0 → 0.2.0
# or
npm version major  # 0.1.0 → 1.0.0
```

**2. Commit and push:**
```bash
git add packages/typescript/CHANGELOG.md packages/typescript/package.json
git commit -m "Release TypeScript SDK v0.1.1"
git push
```

**3. Create and push tag:**
```bash
git tag v0.1.1
git push origin v0.1.1
```

**4. Automated (GitHub Actions):**
- ✅ Runs tests
- ✅ Builds package
- ✅ Publishes to npm
- ✅ Creates GitHub release

**5. Verify:**
```bash
npm view @oway/sdk version
# Should show: 0.1.1
```

---

### Go SDK Release

**1. Update version and changelog:**
```bash
cd packages/go

# Update CHANGELOG.md with changes
# Go uses git tags for versions (no package.json)
```

**2. Commit and push:**
```bash
git add packages/go/CHANGELOG.md
git commit -m "Release Go SDK v0.1.1"
git push
```

**3. Create and push tag:**
```bash
git tag packages/go/v0.1.1
git push origin packages/go/v0.1.1
```

**4. Automated (GitHub Actions):**
- ✅ Runs tests
- ✅ Verifies build
- ✅ Creates GitHub release
- ✅ pkg.go.dev indexes automatically (~15 min)

**5. Verify:**
```bash
# Wait 15 minutes, then:
open https://pkg.go.dev/github.com/Oway-Inc/oway-sdk/packages/go@v0.1.1
```

---

## Semantic Versioning

### When to bump MAJOR (1.0.0 → 2.0.0)

**Breaking changes that affect existing code:**
- Removed methods
- Changed method signatures
- Removed or renamed types
- Changed auth requirements

**Example:**
```typescript
// Before:
oway.quotes.create({ origin, destination })

// After (breaking):
oway.quotes.create({ origin, destination, requiredNewField })
```

### When to bump MINOR (0.1.0 → 0.2.0)

**New features, backwards-compatible:**
- Added new methods
- Added new optional parameters
- Added new resources
- New functionality

**Example:**
```typescript
// Added (doesn't break existing code):
oway.carriers.getJobs()
```

### When to bump PATCH (0.1.0 → 0.1.1)

**Bug fixes, no API changes:**
- Fixed error handling
- Improved retry logic
- Fixed TypeScript types
- Performance improvements

---

## Changelog Format

### Template

```markdown
## [0.2.0] - 2026-03-15

### Added
- New `carriers` resource for carrier API
- Support for GPS data submission
- Webhook signature verification helper

### Changed
- Improved error messages for auth failures

### Fixed
- Token refresh race condition in high-concurrency scenarios
- TypeScript types for optional fields

### Security
- Enhanced credential sanitization in logs
```

### Categories

- **Added**: New features
- **Changed**: Changes in existing functionality
- **Deprecated**: Soon-to-be removed features
- **Removed**: Removed features
- **Fixed**: Bug fixes
- **Security**: Security fixes

---

## GitHub Release Notes

### Auto-Generated from Changelog

```bash
# Extract version section from CHANGELOG.md
VERSION=0.1.1
sed -n "/## \[$VERSION\]/,/## \[/p" CHANGELOG.md | head -n -1
```

### Example Release Note

```markdown
## TypeScript SDK v0.1.1

### Installation
\`\`\`bash
npm install @oway/sdk@0.1.1
\`\`\`

### What's Changed
- Fixed token refresh race condition
- Improved error messages
- Updated TypeScript types

### Documentation
- [README](./packages/typescript/README.md)
- [API Docs](https://docs.shipoway.com)

**Full Changelog**: v0.1.0...v0.1.1
```

---

## Pre-Release Checklist

Before creating a release:

- [ ] All tests passing locally
- [ ] CHANGELOG.md updated
- [ ] Version bumped in package.json (TypeScript)
- [ ] No internal references in code/docs
- [ ] Generated code is committed
- [ ] README.md reviewed
- [ ] Examples tested

---

## Emergency Hotfix Process

**If critical bug in production:**

```bash
# 1. Create hotfix branch from tag
git checkout v0.1.0
git checkout -b hotfix/0.1.1

# 2. Fix bug
git commit -m "Fix critical auth issue"

# 3. Bump patch version
cd packages/typescript
npm version patch

# 4. Merge to main
git checkout main
git merge hotfix/0.1.1

# 5. Tag and push
git tag v0.1.1
git push origin main v0.1.1

# 6. GitHub Actions auto-publishes
```

---

## Release Cadence

**Recommended schedule:**

| Type | Frequency | Trigger |
|------|-----------|---------|
| **Patch** | As needed | Bug fixes |
| **Minor** | Monthly | New features from API updates |
| **Major** | Quarterly | Breaking changes (rare) |

---

## What Happens When You Push a Tag

### TypeScript (tag: `v0.1.1`)

```
1. Push tag
   ↓
2. GitHub Actions triggers
   ↓
3. Run tests (Node 18, 20, 22)
   ↓
4. Build package
   ↓
5. Publish to npm
   ↓
6. Create GitHub release
   ↓
7. Done! ✅
```

**Visible at:**
- npm: https://www.npmjs.com/package/@oway/sdk
- GitHub: https://github.com/Oway-Inc/oway-sdk/releases

### Go (tag: `packages/go/v0.1.1`)

```
1. Push tag
   ↓
2. GitHub Actions triggers
   ↓
3. Run tests (Go 1.21, 1.22)
   ↓
4. Verify build
   ↓
5. Create GitHub release
   ↓
6. pkg.go.dev auto-indexes (15 min)
   ↓
7. Done! ✅
```

**Visible at:**
- pkg.go.dev: https://pkg.go.dev/github.com/Oway-Inc/oway-sdk/packages/go
- GitHub: https://github.com/Oway-Inc/oway-sdk/releases

---

## Secrets Required

### GitHub Repository Secrets

**Required:**
```
NPM_TOKEN
• Description: Automation token for @oway scope
• How to get: npmjs.com → Access Tokens → Generate (Automation)
• Used by: publish workflow
```

**Optional:**
```
CODECOV_TOKEN
• Description: For code coverage reports
• How to get: codecov.io
• Used by: ci workflow
```

**NOT needed:**
- ❌ Oway API credentials (not used in CI)
- ❌ AWS keys (no deployment)
- ❌ Database credentials (no internal systems)

---

## First Release

### TypeScript v0.1.0

```bash
# 1. Finalize CHANGELOG
# 2. Set version
cd packages/typescript
npm version 0.1.0

# 3. Commit and tag
git add .
git commit -m "Release TypeScript SDK v0.1.0"
git tag v0.1.0
git push origin main v0.1.0

# 4. Watch GitHub Actions
# 5. Verify on npm
```

### Go v0.1.0

```bash
# 1. Finalize CHANGELOG
# 2. Tag
git tag packages/go/v0.1.0
git push origin packages/go/v0.1.0

# 3. Watch GitHub Actions
# 4. Wait 15 min for pkg.go.dev indexing
```

---

## Summary

**Releases are:**
- ✅ Automated via GitHub Actions
- ✅ Triggered by git tags
- ✅ Safe for public repos (no internal references)
- ✅ Follow Stripe's patterns
- ✅ Create changelog-driven release notes

**External developers see:**
- Test results (builds trust)
- Release automation (professional)
- No internal Oway infrastructure (secure)
