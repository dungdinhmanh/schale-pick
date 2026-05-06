# AUR packaging

This directory holds the **source of truth** for the AUR PKGBUILDs. The AUR itself
hosts each package in its own git repo (e.g. `ssh://aur@aur.archlinux.org/schale-pick-bin.git`),
but we keep a copy here so changes are reviewed via PRs and don't get lost.

## Packages

| Package           | Source                                          | Audience                  |
|-------------------|-------------------------------------------------|---------------------------|
| `schale-pick-bin` | Pre-built binary from the GitHub release (`v*`) | End users (recommended)   |

## Prerequisites

```bash
sudo pacman -S --needed base-devel pacman-contrib namcap devtools
```

- `base-devel` — `makepkg`, `fakeroot`, etc.
- `pacman-contrib` — `updpkgsums` for refreshing sha256 sums.
- `namcap` — lints the PKGBUILD and the resulting `.pkg.tar.zst`.
- `devtools` — provides `pkgctl build` for clean-chroot builds (recommended).

You also need an AUR account with an SSH key registered:
<https://aur.archlinux.org/account/>

### One-time SSH + git setup

Generate a dedicated key (do **not** reuse another key — AUR access should be
selectively revocable):

```bash
ssh-keygen -f ~/.ssh/aur
# paste ~/.ssh/aur.pub into https://aur.archlinux.org/account/ → SSH Public Key
```

Add the host entry to `~/.ssh/config`:

```
Host aur.archlinux.org
    IdentityFile ~/.ssh/aur
    User aur
```

AUR rejects pushes from anonymous commits ([FS#45425]), so set the global git
identity if you haven't already:

```bash
git config --global user.name  "dungdinhmanh"
git config --global user.email "dungdinhmanh@users.noreply.github.com"
```

[FS#45425]: https://bugs.archlinux.org/task/45425

## Local test (before publishing)

From `packaging/aur/schale-pick-bin/`:

```bash
# 1. Bake real sha256 sums into the PKGBUILD (replaces the SKIP placeholders).
updpkgsums

# 2. Regenerate .SRCINFO so the AUR web UI shows the correct metadata.
makepkg --printsrcinfo > .SRCINFO

# 3. Lint.
namcap PKGBUILD

# 4a. Build inside a clean chroot (recommended; catches missing makedepends
#     and system-config-induced failures the wiki warns about).
pkgctl build

# 4b. ...or build + install directly into the live system.
makepkg -si

# 5. Smoke test.
schale-pick --version
schale-pick --help
```

`makepkg -si` will pull `fastfetch`, `jq`, and `imagemagick` automatically as runtime
deps, so the user does **not** need the `sudo pacman -S fastfetch jq imagemagick`
line from the README — that line is only for users installing the binary by hand.

## Releasing a new version

### When to bump what

| Change                                                | Bump                |
|-------------------------------------------------------|---------------------|
| New upstream release (`vX.Y.Z`)                       | `pkgver=X.Y.Z`, reset `pkgrel=1` |
| PKGBUILD fix at the same upstream version (deps, build flags, …) | `pkgrel++`          |
| Cosmetic-only edits (comment typo, whitespace)        | **Do not bump** — users won't see it as an update anyway |

### Steps

1. Tag and push `vX.Y.Z` on the main repo. The `Release` workflow uploads
   `schale-pick-linux-amd64` + `schale-pick-linux-arm64` (and their `.sha256` files)
   to the GitHub release.
2. In `packaging/aur/schale-pick-bin/`:
   ```bash
   sed -i "s/^pkgver=.*/pkgver=X.Y.Z/" PKGBUILD
   sed -i "s/^pkgrel=.*/pkgrel=1/"     PKGBUILD
   updpkgsums
   makepkg --printsrcinfo > .SRCINFO
   ```
3. Commit the bump in this repo (review via PR).
4. Push to AUR:
   ```bash
   # one-time clone of the AUR repo, in a sibling directory
   git -c init.defaultBranch=master clone \
       ssh://aur@aur.archlinux.org/schale-pick-bin.git ~/aur/schale-pick-bin

   # sync our PKGBUILD + .SRCINFO + LICENSE into the AUR clone
   cp PKGBUILD .SRCINFO LICENSE ~/aur/schale-pick-bin/

   cd ~/aur/schale-pick-bin
   git add PKGBUILD .SRCINFO LICENSE
   git commit -m "upgpkg: schale-pick-bin X.Y.Z-1"
   git push  # AUR only accepts pushes to master
   ```

> **Note:** the AUR repo should contain `PKGBUILD`, `.SRCINFO`, and a
> `LICENSE` for the build script itself (0BSD recommended by the
> [submission guidelines]). Do **not** push the upstream binary, the built
> `.pkg.tar.zst`, or anything in `pkg/` / `src/` (already covered by AUR's
> default `.gitignore`).
>
> **Forgot `.SRCINFO` in your last commit?** The AUR will reject the push.
> Fix with `git add .SRCINFO && git commit --amend --no-edit && git push`.
>
> [submission guidelines]: https://wiki.archlinux.org/title/AUR_submission_guidelines

## First-time publish (initial AUR submission)

1. Make sure the GitHub release `v0.1.0` exists and has the binaries attached.
2. Run the local-test steps above, confirm the package builds cleanly and the
   binary works.
3. Create the AUR repo (a freshly created AUR repo is empty; the warning
   `cloned an empty repository` is expected):
   ```bash
   git -c init.defaultBranch=master clone \
       ssh://aur@aur.archlinux.org/schale-pick-bin.git ~/aur/schale-pick-bin
   cp PKGBUILD .SRCINFO LICENSE ~/aur/schale-pick-bin/
   cd ~/aur/schale-pick-bin
   git add PKGBUILD .SRCINFO LICENSE
   git commit -m "addpkg: schale-pick-bin 0.1.0-1"
   git push -u origin master
   ```
4. The package will appear at <https://aur.archlinux.org/packages/schale-pick-bin>.

## Ongoing maintenance

- **Watch upstream for non-version changes.** A "minor" upstream release can
  still flip the license, add a runtime dep, or rename a binary. Re-read the
  release notes before each `pkgver` bump — automation cannot replace this.
- **Respond to AUR comments.** Users will report breakage there before they
  open a GitHub issue. Don't post `version X.Y.Z released` comments on every
  bump; that just buries useful feedback.
- **Flag out-of-date** if you spot it before updating; this notifies other
  watchers.
- **Disown rather than abandon.** If you stop maintaining `schale-pick-bin`,
  hit *Disown Package* on the AUR page so someone else can adopt it.
