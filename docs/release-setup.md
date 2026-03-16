# Release Pipeline Setup Guide

## 1. Create the Homebrew Tap Repository

On GitHub, create a **public** repository named exactly:

```
homebrew-tools
```

under the same owner/org as `ctxlog`. Homebrew requires the `homebrew-` prefix — this lets users install with:

```bash
brew install <owner>/tools/ctxlog
```

The repository needs at least one commit (initialize with a README). GoReleaser will create the `Formula/` directory and push the formula automatically.

## 2. Generate a Personal Access Token (PAT)

GoReleaser needs a token with permission to push to the `homebrew-tools` repo (which is separate from the repo that triggers the workflow).

1. Go to **GitHub > Settings > Developer settings > Personal access tokens > Fine-grained tokens**.
2. Click **Generate new token**.
3. Configure:
   - **Token name:** `goreleaser-tap`
   - **Expiration:** choose an appropriate lifetime (e.g., 90 days; set a calendar reminder to rotate).
   - **Repository access:** select **Only select repositories**, then pick `homebrew-tools`.
   - **Permissions > Repository permissions:**
     - **Contents:** Read and write (required to push the formula file).
   - No other permissions are needed.
4. Click **Generate token** and copy the value immediately.

> **Why fine-grained over classic?** Fine-grained tokens are scoped to specific repos and permissions, following the principle of least privilege.

## 3. Add the Secret to the ctxlog Repository

1. Go to the `ctxlog` repository on GitHub.
2. Navigate to **Settings > Secrets and variables > Actions**.
3. Click **New repository secret**.
4. Set:
   - **Name:** `TAP_GITHUB_TOKEN`
   - **Value:** paste the PAT from the previous step.
5. Click **Add secret**.

## 4. Create a Release

Tag and push to trigger the pipeline:

```bash
git tag v0.1.0
git push origin v0.1.0
```

The workflow will:
1. Build static binaries for `darwin/amd64`, `darwin/arm64`, `linux/amd64`, `linux/arm64`.
2. Create a GitHub Release with the archives and checksums.
3. Push the Homebrew formula to `homebrew-tools/Formula/ctxlog.rb`.

## 5. Verify

```bash
brew tap <owner>/tools
brew install ctxlog
ctxlog read -shard=test
```
