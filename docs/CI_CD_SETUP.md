# CI/CD Setup Guide for aux4 Package Publishing

This guide explains how to set up automated publishing for aux4 packages using GitHub Actions.

## Overview

The CI/CD system enables automatic publishing of aux4 packages to hub.aux4.io when changes are pushed to the main branch. It uses:

- **GitHub Actions** for workflow automation
- **Docker** (aux4/publisher image) for consistent build environments
- **OAuth tokens** for authentication (stored as GitHub Secrets)
- **Semantic versioning** with automatic or manual version bumps

## Prerequisites

- Package repository on GitHub (e.g., aux4/config)
- Access to publish packages on hub.aux4.io
- Docker Hub account (to pull aux4/publisher image)
- GitHub repository admin access (to configure secrets)

## Quick Start

### 1. Create Access Token

Generate an OAuth access token for CI/CD authentication:

```bash
# Login to aux4 (opens browser for OAuth authentication)
aux4 aux4 pkger login

# Extract the access token
cat ~/.aux4.config/credentials | jq -r '.accessToken'
```

**Important:** Copy the token value (64-character hexadecimal string). You'll need it in the next step.

### 2. Store Token in GitHub Secrets

#### Via GitHub Web UI:

1. Go to your package repository on GitHub (e.g., https://github.com/aux4/config)
2. Click **Settings** → **Secrets and variables** → **Actions**
3. Click **New repository secret**
4. Enter:
   - **Name**: `AUX4_ACCESS_TOKEN`
   - **Secret**: Paste the token from step 1
5. Click **Add secret**

#### Via GitHub CLI:

```bash
# Set the secret using gh CLI
gh secret set AUX4_ACCESS_TOKEN --body "$(cat ~/.aux4.config/credentials | jq -r '.accessToken')"

# Verify it was set
gh secret list
```

### 3. Add Workflow to Your Package Repository

Copy the workflow file to your package repository:

```bash
# Create workflows directory
mkdir -p .github/workflows

# Copy the example workflow
curl -o .github/workflows/publish.yml \
  https://raw.githubusercontent.com/aux4/aux4/main/.github/workflows/publish-example.yml.template

# Edit the file and replace YOUR_PACKAGE_NAME with your actual package name
# For example, if your package is "config", replace it with "config"
```

Or manually create `.github/workflows/publish.yml`:

```yaml
name: Publish Package

on:
  push:
    branches:
      - main
    paths:
      - '**'
      - '!.github/**'
      - '!README.md'
      - '!LICENSE'
      - '!.gitignore'

  workflow_dispatch:
    inputs:
      release_level:
        description: 'Release level (patch, minor, major)'
        required: true
        default: 'patch'
        type: choice
        options:
          - patch
          - minor
          - major

concurrency:
  group: publish-package
  cancel-in-progress: false

permissions:
  contents: write
  packages: read

jobs:
  publish:
    runs-on: ubuntu-latest
    steps:
      - name: Checkout repository
        uses: actions/checkout@v4
        with:
          fetch-depth: 0
          token: ${{ secrets.GITHUB_TOKEN }}

      - name: Configure git
        run: |
          git config user.name "github-actions[bot]"
          git config user.email "github-actions[bot]@users.noreply.github.com"

      - name: Pull latest changes
        run: |
          git pull --rebase origin main

      - name: Set release level
        id: release_level
        run: |
          if [ "${{ github.event_name }}" == "workflow_dispatch" ]; then
            echo "level=${{ inputs.release_level }}" >> $GITHUB_OUTPUT
          else
            echo "level=patch" >> $GITHUB_OUTPUT
          fi

      - name: Release package
        env:
          AUX4_ACCESS_TOKEN: ${{ secrets.AUX4_ACCESS_TOKEN }}
        run: |
          docker run --rm \
            -v $(pwd):/workspace \
            -w /workspace \
            -e AUX4_ACCESS_TOKEN="${AUX4_ACCESS_TOKEN}" \
            aux4/publisher:latest \
            "aux4 aux4 releaser release --level ${{ steps.release_level.outputs.level }}"

      - name: Get new version
        id: version
        run: |
          VERSION=$(jq -r '.version' .aux4)
          echo "version=${VERSION}" >> $GITHUB_OUTPUT
          echo "New version: ${VERSION}"

      - name: Push changes and tags
        run: |
          git push origin main
          git push origin --tags

      - name: Create GitHub Release
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
        run: |
          VERSION="${{ steps.version.outputs.version }}"
          PACKAGE_FILE="package/aux4-$(basename $(pwd))-${VERSION}.zip"

          if [ -f "${PACKAGE_FILE}" ]; then
            gh release create "v${VERSION}" \
              "${PACKAGE_FILE}" \
              --title "Release v${VERSION}" \
              --notes "Automated release v${VERSION}"
          else
            echo "Warning: Package file ${PACKAGE_FILE} not found"
            gh release create "v${VERSION}" \
              --title "Release v${VERSION}" \
              --notes "Automated release v${VERSION}"
          fi
```

### 4. Commit and Push

```bash
git add .github/workflows/publish.yml
git commit -m "Add automated publishing workflow"
git push origin main
```

## Usage

### Automatic Publishing (Default)

When you push changes to the `main` branch:

1. GitHub Actions automatically triggers the workflow
2. Version is bumped by **patch** (e.g., 1.0.0 → 1.0.1)
3. Package is built and published to hub.aux4.io
4. Git tag is created (e.g., v1.0.1)
5. GitHub Release is created with the package artifact

### Manual Publishing (Custom Version Bump)

To manually trigger a release with a specific version bump:

1. Go to **Actions** tab in your repository
2. Select **Publish Package** workflow
3. Click **Run workflow**
4. Choose release level:
   - **patch**: 1.0.0 → 1.0.1 (bug fixes)
   - **minor**: 1.0.0 → 1.1.0 (new features, backward compatible)
   - **major**: 1.0.0 → 2.0.0 (breaking changes)
5. Click **Run workflow**

## Security

### Token Security

- Tokens are stored as **GitHub Secrets** (encrypted at rest)
- Secrets are **never logged** or exposed in workflow output
- Only workflows in your repository can access the secret
- Pull requests from forks **cannot** access secrets (security feature)

### Token Expiration

- OAuth tokens expire after **30 days**
- Set a reminder to rotate tokens every **25 days** (5-day buffer)

### Token Rotation

To rotate your token before expiration:

```bash
# 1. Generate new token
aux4 aux4 pkger login

# 2. Extract new token
cat ~/.aux4.config/credentials | jq -r '.accessToken'

# 3. Update GitHub Secret
gh secret set AUX4_ACCESS_TOKEN --body "<new-token-value>"

# Or via web UI: Settings → Secrets → AUX4_ACCESS_TOKEN → Update secret
```

### Recommended Security Settings

1. **Branch Protection** (Settings → Branches → Add rule for `main`):
   - ✅ Require pull request reviews before merging
   - ✅ Require status checks to pass
   - ✅ Restrict who can push to branch

2. **Workflow Permissions** (Settings → Actions → General):
   - ✅ Read and write permissions for GITHUB_TOKEN
   - ✅ Allow GitHub Actions to create and approve pull requests

## Troubleshooting

### Workflow Fails with "Not logged in"

**Cause:** `AUX4_ACCESS_TOKEN` secret is missing or invalid.

**Solution:**
1. Verify secret exists: Repository → Settings → Secrets → Actions
2. Check token hasn't expired (30-day limit)
3. Regenerate and update token (see Token Rotation above)

### Package Not Published to hub.aux4.io

**Cause:** Publishing failed, but git tags were still created.

**Solution:**
1. Check workflow logs for errors (Actions tab → Failed workflow → View logs)
2. Verify package builds locally: `aux4 aux4 releaser build`
3. Test publishing locally: `aux4 aux4 releaser publish <package.zip>`
4. Fix issues and re-run workflow

### Git Push Fails with "Protected Branch"

**Cause:** Branch protection rules prevent github-actions bot from pushing.

**Solution:**
1. Go to Settings → Branches → Edit rule for `main`
2. Under "Allow force pushes", enable "Specify who can force push"
3. Add "github-actions" bot to allowed list
4. Or disable "Require pull request reviews" for the bot

### Version Conflict (Tag Already Exists)

**Cause:** Git tag for the version already exists from a previous run.

**Solution:**
```bash
# Delete the tag locally and remotely
git tag -d v1.0.1
git push origin :refs/tags/v1.0.1

# Re-run the workflow
```

### Docker Image Pull Fails

**Cause:** Docker Hub rate limiting or image doesn't exist.

**Solution:**
1. Verify image exists: `docker pull aux4/publisher:latest`
2. Wait 6 hours if rate limited
3. Or authenticate with Docker Hub in workflow:
   ```yaml
   - name: Login to Docker Hub
     uses: docker/login-action@v3
     with:
       username: ${{ secrets.DOCKERHUB_USERNAME }}
       password: ${{ secrets.DOCKERHUB_TOKEN }}
   ```

## Advanced Configuration

### Custom Publisher Image Version

Use a specific version of the publisher image:

```yaml
- name: Release package
  env:
    AUX4_ACCESS_TOKEN: ${{ secrets.AUX4_ACCESS_TOKEN }}
  run: |
    docker run --rm \
      -v $(pwd):/workspace \
      -w /workspace \
      -e AUX4_ACCESS_TOKEN="${AUX4_ACCESS_TOKEN}" \
      aux4/publisher:1.0.0 \  # Specific version
      "aux4 aux4 releaser release --level ${{ steps.release_level.outputs.level }}"
```

### Skip Specific Paths

Prevent workflow from triggering on certain file changes:

```yaml
on:
  push:
    branches:
      - main
    paths:
      - '**'
      - '!.github/**'
      - '!README.md'
      - '!LICENSE'
      - '!docs/**'  # Add this to skip docs changes
      - '!*.md'     # Skip all markdown files
```

### Conditional Publishing

Only publish when version changes:

```yaml
- name: Check if version changed
  id: version_check
  run: |
    CURRENT_VERSION=$(jq -r '.version' .aux4)
    git fetch --tags
    LATEST_TAG=$(git describe --tags --abbrev=0 2>/dev/null || echo "v0.0.0")
    LATEST_VERSION=${LATEST_TAG#v}

    if [ "$CURRENT_VERSION" == "$LATEST_VERSION" ]; then
      echo "skip=true" >> $GITHUB_OUTPUT
    else
      echo "skip=false" >> $GITHUB_OUTPUT
    fi

- name: Release package
  if: steps.version_check.outputs.skip == 'false'
  # ... rest of release steps
```

## Monitoring

### Check Workflow Runs

- Go to **Actions** tab in your repository
- View all workflow runs, their status, and logs
- Filter by workflow name, branch, or status

### Notifications

Set up notifications for workflow failures:

1. Go to Settings → Notifications
2. Enable "Actions" notifications
3. Choose delivery method (email, Slack, etc.)

### Audit Logs

For organizations:

1. Go to Organization Settings → Audit log
2. Filter by "secret" to see token access history
3. Filter by "workflow_run" to see all workflow executions

## Best Practices

1. **Test Locally First**: Always test `aux4 aux4 releaser build` locally before relying on CI/CD
2. **Use Semantic Versioning**: Choose appropriate version bumps (patch/minor/major)
3. **Rotate Tokens Regularly**: Set calendar reminders for token rotation (every 25 days)
4. **Monitor Workflow Runs**: Check Actions tab regularly for failures
5. **Branch Protection**: Enable branch protection to prevent accidental pushes
6. **Document Changes**: Update README/CHANGELOG when releasing new versions

## FAQ

**Q: Can I use this for private packages?**
A: Yes, as long as your access token has permissions to publish to the package scope.

**Q: Does this work with monorepos?**
A: Yes, but you'll need separate workflows for each package, filtered by path triggers.

**Q: Can I customize the release notes?**
A: Yes, edit the `gh release create` command to use custom notes or read from a CHANGELOG file.

**Q: What if I want to publish to a different registry?**
A: Modify the `aux4 aux4 releaser release` command to target a different registry URL.

**Q: How do I rollback a release?**
A: Delete the git tag and GitHub release, then manually publish the previous version.

## Support

- **Issues**: https://github.com/aux4/aux4/issues
- **Discussions**: https://github.com/aux4/aux4/discussions
- **Documentation**: https://docs.aux4.io

## References

- [GitHub Actions Documentation](https://docs.github.com/en/actions)
- [GitHub Secrets Security](https://docs.github.com/en/actions/security-guides/encrypted-secrets)
- [aux4 CLI Documentation](https://docs.aux4.io/cli)
- [Semantic Versioning](https://semver.org/)
