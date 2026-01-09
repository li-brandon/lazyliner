# Releasing Lazyliner

This document describes how to release new versions of Lazyliner.

## Prerequisites

### 1. Create the Homebrew Tap Repository

Create a new GitHub repository named `homebrew-tap` under your account:

1. Go to https://github.com/new
2. Create repository: `brandonli/homebrew-tap`
3. Initialize with a README
4. Create the `Formula/` directory (can be empty initially)

### 2. Create a Personal Access Token

GoReleaser needs a token to push the Homebrew formula to your tap repository:

1. Go to https://github.com/settings/tokens
2. Generate a new token (classic) with `repo` scope
3. Copy the token

### 3. Add the Token as a Repository Secret

1. Go to your `lazyliner` repository Settings > Secrets and variables > Actions
2. Add a new secret named `HOMEBREW_TAP_GITHUB_TOKEN`
3. Paste your personal access token as the value

## Creating a Release

1. **Update version** (if you have version files to update)

2. **Create and push a tag**:
   ```bash
   git tag -a v1.0.0 -m "Release v1.0.0"
   git push origin v1.0.0
   ```

3. **GitHub Actions will automatically**:
   - Build binaries for all platforms (Linux, macOS, Windows on amd64/arm64)
   - Create a GitHub Release with the binaries
   - Update the Homebrew formula in `brandonli/homebrew-tap`
   - Create deb and rpm packages

## Manual Release (if needed)

If you need to run GoReleaser locally:

```bash
# Install goreleaser
brew install goreleaser

# Dry run (no publishing)
goreleaser release --snapshot --clean

# Actual release (requires GITHUB_TOKEN and HOMEBREW_TAP_GITHUB_TOKEN)
export GITHUB_TOKEN=your_token
export HOMEBREW_TAP_GITHUB_TOKEN=your_homebrew_token
goreleaser release --clean
```

## Installation Verification

After a release, users can install via:

```bash
# Homebrew (macOS)
brew install brandonli/tap/lazyliner

# Verify installation
lazyliner --version
```

## Troubleshooting

### Homebrew formula not updating

- Check that `HOMEBREW_TAP_GITHUB_TOKEN` secret is set correctly
- Verify the token has `repo` scope
- Check the `homebrew-tap` repository exists with a `Formula/` directory

### Build failures

- Check the GitHub Actions logs
- Verify `go.mod` is valid and dependencies are available
- Run `goreleaser release --snapshot --clean` locally to debug
