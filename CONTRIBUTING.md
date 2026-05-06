# Contributing to schale-pick

Thank you for taking the time to contribute! Please read this guide before opening issues or pull requests.

---

## Table of Contents

- [Getting Started](#getting-started)
- [Development Workflow](#development-workflow)
- [Commit Convention](#commit-convention)
- [Versioning](#versioning)
- [Pull Request Process](#pull-request-process)
- [Reporting Issues](#reporting-issues)

---

## Getting Started

**Prerequisites**

| Tool | Version |
|------|---------|
| Go | ≥ 1.25 |
| git | any recent |
| `jq` + `imagemagick` | for manual testing |

**Fork & clone**

```bash
git clone https://github.com/dungdinhmanh/schale-pick
cd schale-pick
go build .
```

**Run locally**

```bash
./schale-pick
# or with debug logging
./schale-pick --debug
```

---

## Development Workflow

1. Create a branch from `master`:
   ```bash
   git checkout -b fix/clear-image-on-modal
   git checkout -b feat/pkgbuild-support
   ```
2. Make your changes and ensure the project still builds:
   ```bash
   go build .
   go vet ./...
   ```
3. Commit using [Conventional Commits](#commit-convention).
4. Open a pull request against `master`.

---

## Commit Convention

This project follows [Conventional Commits v1.0.0](https://www.conventionalcommits.org/en/v1.0.0/).

### Format

```
<type>[optional scope]: <description>

[optional body]

[optional footer(s)]
```

### Types

| Type | When to use | Semver bump |
|------|-------------|-------------|
| `feat` | New feature | MINOR |
| `fix` | Bug fix | PATCH |
| `docs` | Documentation only | — |
| `style` | Formatting, whitespace (no logic change) | — |
| `refactor` | Code change that is neither feat nor fix | — |
| `perf` | Performance improvement | PATCH |
| `test` | Adding or updating tests | — |
| `chore` | Build process, dependencies, tooling | — |
| `ci` | CI/CD configuration | — |
| `revert` | Reverts a previous commit | — |

### Breaking Changes

Add `!` after the type, or include `BREAKING CHANGE:` in the commit footer.
Breaking changes trigger a **MAJOR** version bump.

```
feat!: remove --legacy flag

BREAKING CHANGE: --legacy flag has been removed. Use --format=raw instead.
```

### Examples

```bash
feat: add PKGBUILD for AUR packaging
fix: clear kitty image when help modal opens
docs: update install instructions for ARM64
chore: bump golang.org/x/image to v0.29.0
ci: add linux/arm64 build target to release workflow
refactor: extract jq patch logic into updateLogoFields()
perf: cache cell size query to reduce syscalls
```

---

## Versioning

This project uses [Semantic Versioning 2.0.0](https://semver.org/).

```
MAJOR.MINOR.PATCH
```

| Change | Version bump | Example |
|--------|-------------|---------|
| Breaking change (`feat!`, `fix!`, `BREAKING CHANGE:`) | MAJOR | `1.0.0` → `2.0.0` |
| New feature (`feat`) | MINOR | `0.1.0` → `0.2.0` |
| Bug fix (`fix`, `perf`) | PATCH | `0.1.0` → `0.1.1` |
| Docs, style, chore, refactor | none | — |

### Releasing

Releases are automated via GitHub Actions. To trigger a release:

```bash
git tag v0.2.0
git push origin v0.2.0
```

The workflow will build binaries for `linux/amd64` and `linux/arm64`, generate checksums, and publish a GitHub Release automatically.

---

## Pull Request Process

1. Ensure `go build .` and `go vet ./...` pass.
2. Keep PRs focused — one feature or fix per PR.
3. Write a clear PR description explaining *what* and *why*.
4. Reference any related issues: `Closes #42`.
5. A maintainer will review and merge.

---

## Reporting Issues

Use the issue templates:

- **Bug report** — something is broken or behaves unexpectedly.
- **Feature request** — suggest a new feature or improvement.

Please search existing issues before opening a new one.
