# Releasing shdoc-ng

## How it works

Releases are fully automated via [GoReleaser](https://goreleaser.com/) and GitHub Actions. Pushing a semver tag triggers the release pipeline, which:

1. Runs the full CI suite (tests, vet, lint) — **release is blocked if CI fails**
2. Builds binaries for all supported platforms
3. Generates a changelog from conventional commit messages
4. Publishes a GitHub Release with archives, checksums, and Linux packages (deb/rpm/apk)
5. Updates the Homebrew formula in [jdevera/homebrew-tap](https://github.com/jdevera/homebrew-tap)
6. Uploads RPM packages to [COPR](https://copr.fedorainfracloud.org/coprs/jdevera/shdoc-ng/) for Fedora users

## Cutting a release

```bash
# Ensure you're on main and up to date
git checkout main
git pull

# Tag the release (vMAJOR.MINOR.PATCH)
git tag v0.2.0

# Push the tag to trigger the release
git push origin v0.2.0
```

The release workflow runs automatically. Monitor it at: **Actions > Release** in GitHub.

## Versioning

This project follows [Semantic Versioning](https://semver.org/):

- **MAJOR** — breaking changes to CLI interface, output format, or documented behavior
- **MINOR** — new features, new tags, new output formats
- **PATCH** — bug fixes, documentation, internal improvements

The version is injected at build time via ldflags (`-X main.version=...`). Local builds show `dev`.

## Commit messages

We use [Conventional Commits](https://www.conventionalcommits.org/). The changelog is auto-generated from these:

| Prefix     | Appears in changelog |
|------------|---------------------|
| `feat:`    | Yes (Features)      |
| `fix:`     | Yes (Bug Fixes)     |
| `docs:`    | No                  |
| `test:`    | No                  |
| `ci:`      | No                  |
| `chore:`   | No                  |
| `refactor:`| No                  |

## Supported platforms

| OS      | Architectures  | Status     |
|---------|---------------|------------|
| Linux   | amd64, arm64  | Tested     |
| macOS   | amd64, arm64  | Tested     |
| Windows | amd64, arm64  | Builds, untested |

## Distribution channels

### GitHub Releases

Binaries, archives, and SHA256 checksums are published automatically to the [Releases page](https://github.com/jdevera/shdoc-ng/releases).

### go install

```bash
go install github.com/jdevera/shdoc-ng/cmd/shdoc-ng@latest
```

### Homebrew

```bash
brew install jdevera/tap/shdoc-ng
```

The formula in [jdevera/homebrew-tap](https://github.com/jdevera/homebrew-tap) is updated automatically on each release by GoReleaser.

### Fedora (COPR)

```bash
sudo dnf copr enable jdevera/shdoc-ng
sudo dnf install shdoc-ng
```

RPM packages are automatically uploaded to [COPR](https://copr.fedorainfracloud.org/coprs/jdevera/shdoc-ng/) after each release. Supported chroots: Fedora 42, 43, 44, and Rawhide (x86_64 and aarch64).

### Linux packages (deb/rpm/apk)

`.deb`, `.rpm`, and `.apk` packages are attached to each GitHub Release, built automatically by GoReleaser's nfpms integration.

```bash
# Debian/Ubuntu
sudo dpkg -i shdoc-ng_0.2.0_amd64.deb

# Fedora/RHEL (manual install without COPR)
sudo rpm -i shdoc-ng_0.2.0_amd64.rpm

# Alpine
apk add --allow-untrusted shdoc-ng_0.2.0_amd64.apk
```

Note: Alpine requires `--allow-untrusted` for packages not from a signed repository. This is typical for packages distributed via GitHub Releases.

## Required secrets

The following secrets must be configured in the GitHub repository settings:

| Secret                      | Purpose                                          |
|-----------------------------|--------------------------------------------------|
| `GITHUB_TOKEN`              | Built-in. Used for creating releases.            |
| `HOMEBREW_TAP_TOKEN` | PAT with `repo` scope on `jdevera/homebrew-tap`. |
| `COPR_LOGIN`          | COPR API login (from https://copr.fedorainfracloud.org/api/). |
| `COPR_USERNAME`       | COPR username.                                   |
| `COPR_TOKEN`          | COPR API token.                                  |

### Setting up the COPR token

1. Log in to [COPR](https://copr.fedorainfracloud.org) with your Fedora account
2. Go to the [API page](https://copr.fedorainfracloud.org/api/) to get your login, username, and token
3. Add them as repository secrets named `COPR_LOGIN`, `COPR_USERNAME`, and `COPR_TOKEN`

Note: COPR API tokens expire periodically. If the COPR upload step fails, regenerate the token and update the secrets.

### Setting up the Homebrew tap token

1. Create a [fine-grained personal access token](https://github.com/settings/tokens?type=beta) with:
   - Repository access: `jdevera/homebrew-tap` only
   - Permissions: Contents (read and write)
2. Add it as a repository secret named `HOMEBREW_TAP_TOKEN` in the shdoc-ng repo settings.