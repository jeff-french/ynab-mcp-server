# GitHub Actions Workflows

This directory contains automated workflows for the YNAB MCP Server project.

## Workflows

### sync-agents.yml

**Purpose**: Automatically syncs agent instruction files from `agents/` to `.github/agents/` for GitHub Copilot compatibility.

**Trigger**: Runs when files in the `agents/` directory are modified and pushed to the `main` branch.

**What it does**:
1. Detects changes to any files in `agents/**`
2. Removes the symlink at `.github/agents` (if present)
3. Copies all files from `agents/` to `.github/agents/` using `rsync`
4. Commits and pushes the synced files back to the repository
5. Uses Conventional Commits format: `chore(agents): sync agent files to .github/agents`

**Key features**:
- **Infinite loop prevention**: Skips execution if the commit was made by `github-actions[bot]`
- **Efficient syncing**: Uses `rsync --delete` to mirror the source directory (adds, updates, and removes files)
- **No-op handling**: Exits gracefully if there are no changes to commit
- **Proper git configuration**: Uses GitHub Actions bot identity for commits

**Testing locally**:
```bash
# Simulate what the workflow does
rm -rf .github/agents
mkdir -p .github/agents
rsync -av --delete agents/ .github/agents/
git status .github/agents/
```

**Manual trigger**: This workflow only triggers on push to main. To test:
1. Make a change to any file in `agents/`
2. Commit and push to `main` branch
3. Check the Actions tab to see the workflow run
4. Verify that `.github/agents/` contains the updated files

### test.yml

**Purpose**: Runs automated tests across multiple platforms and Go versions.

**Trigger**: Push to `main` or pull requests targeting `main`.

**Jobs**:
- `test`: Runs tests on Linux, macOS, and Windows with race detection
- `lint`: Runs golangci-lint on Linux

### release.yml

**Purpose**: Automated release process using GoReleaser.

**Trigger**: Push of version tags (e.g., `v1.0.0`)

**What it does**:
1. Runs tests across all platforms
2. Builds binaries for all supported platforms
3. Creates Docker images
4. Publishes to GitHub Container Registry
5. Creates GitHub release with changelog
6. Updates Homebrew tap (if configured)

## Workflow Dependencies

### Required Permissions

All workflows require specific permissions defined in the workflow file:
- `contents: write` - For committing and pushing changes (sync-agents.yml)
- `contents: read` - For checking out code (test.yml)
- `packages: write` - For publishing Docker images (release.yml)

### Required Secrets

- `GITHUB_TOKEN` - Automatically provided by GitHub Actions
- `HOMEBREW_TAP_TOKEN` - (Optional) For release.yml to update Homebrew tap

## Troubleshooting

### Sync workflow not triggering

1. Check that changes are being pushed to `main` branch
2. Verify that changed files are in the `agents/` directory
3. Ensure the commit wasn't made by `github-actions[bot]`

### Commit loop detected

If you see repeated commits, check:
1. The `if: github.event.head_commit.author.name != 'github-actions[bot]'` condition is present
2. Git user.name is set to exactly `github-actions[bot]`

### Permission denied errors

Ensure the workflow has `contents: write` permission in the workflow file.

## Best Practices

1. **Test workflows locally first**: Use the testing commands provided in each workflow's documentation
2. **Review workflow runs**: Check the Actions tab after changes to ensure workflows complete successfully
3. **Monitor for failures**: Set up notifications for workflow failures
4. **Keep workflows simple**: Each workflow should have a single, clear purpose
5. **Use specific action versions**: Pin actions to specific versions (e.g., `@v4`) for stability
