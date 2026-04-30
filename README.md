# student-picker

A terminal UI for swapping the [fastfetch](https://github.com/fastfetch-cli/fastfetch) logo
with a Blue Archive student portrait. Browse the full SchaleDB roster, preview portraits
inline using the Kitty Graphics Protocol, and rewrite your fastfetch `config.jsonc` with one
keystroke.

```
┌─────────── Browse / Installed ───────────┐  ┌── Preview ──┐
│  Aru     Akari    Akane    Aki    Ako    │  │             │
│  Arisu   Asuna    Atsuko   Ayane  Azusa  │  │   [image]   │
│ ▶Cherino  Chihiro  Chinatsu Chise Chiyo  │  │             │
│  ...                                     │  │  Cherino    │
└──────────────────────────────────────────┘  └─────────────┘
```

## Features

- **Live preview** via Kitty Graphics Protocol (also degrades to `raw` for non-Kitty terminals).
- **Two-pane layout**: searchable grid on the left, portrait preview on the right.
- **Online + offline**: portraits are downloaded once from SchaleDB and cached locally.
- **LRU cache** with a configurable size limit (no unbounded disk growth).
- **Auto-backup**: your existing `config.jsonc` is copied to `.bak` on first write.
- **Surgical config patches**: only `.logo.source`, `.logo.type`, `.logo.width`, and
  `.logo.padding.left` are modified — the rest of your fastfetch config is left intact.
- **Settings screen**: change cache size, toggle auto-backup, switch logo format on the fly.

## Requirements

| Tool         | Why                                                          |
|--------------|--------------------------------------------------------------|
| Go ≥ 1.25    | Build the binary.                                            |
| `fastfetch`  | The whole point — must already have a `config.jsonc` with a `logo` block. |
| `jq`         | Used to surgically patch `~/.config/fastfetch/config.jsonc`. |
| `magick`     | (ImageMagick) Used to read portrait dimensions for sizing.   |
| Kitty / WezTerm / Konsole / Ghostty | Recommended for inline preview rendering. Other terminals fall back gracefully. |

> **Note:** `~/.config/fastfetch/config.jsonc` must already exist and contain a `"logo": {}` block.
> Run `fastfetch --gen-config` once if you haven't already.

On Arch / CachyOS:

```bash
sudo pacman -S fastfetch jq imagemagick
```

On Debian / Ubuntu:

```bash
sudo apt install fastfetch jq imagemagick
```

## Install

**From source** (replace `YOUR_USERNAME` with your GitHub username after publishing):

```bash
git clone https://github.com/YOUR_USERNAME/student-picker
cd student-picker
go build -o student-picker .
sudo install -m 0755 student-picker /usr/local/bin/
```

Or via `go install` once the module path is set:

```bash
go install github.com/YOUR_USERNAME/student-picker@latest
```

## Usage

```bash
student-picker
```

| Key           | Action                                            |
|---------------|---------------------------------------------------|
| `←`/`→`/`h`/`l` | Move cursor left/right                          |
| `↑`/`↓`/`k`/`j` | Move cursor up/down                             |
| `Tab`           | Switch between **Browse** and **Installed** tabs |
| `/`             | Search by name                                  |
| `Enter`         | Select student → patch fastfetch config         |
| `i`             | Open settings                                   |
| `?`             | Help                                            |
| `q` / `Ctrl+C`  | Quit                                            |

### CLI flags

```
student-picker [flags]

  --version       print version and exit
  --clear-cache   delete all cached portraits and exit
  --debug         write debug log to /tmp/student-picker-debug.log
  --help          show this help
```

## How it patches your config

When you press `Enter`, the app runs:

```bash
jq --arg source <portrait-path> \
   --arg itype  <kitty|raw> \
   --argjson width <calculated> \
   '.logo.source   = $source
  | .logo.type     = $itype
  | .logo.width    = $width
  | .logo.padding.left = 2' \
   ~/.config/fastfetch/config.jsonc
```

Only those four fields change. Your modules, separators, colors, and any other custom
fields are preserved. A `.bak` is written on first run if auto-backup is enabled
(default: on).

## Cache layout

```
~/.cache/student-picker/
├── students.json          # SchaleDB roster (fetched on startup, used offline as fallback)
├── meta.json              # student ID → name mapping for offline use
└── portrait/
    └── <student-id>.webp  # cached portraits (one per student)
```

Default cache size is 5 portraits; configurable from the settings screen (1–20).

## Settings

| Setting       | Values        | Default | Notes |
|---------------|---------------|---------|-------|
| Logo Format   | `kitty`/`raw` | auto    | `kitty` uses graphics protocol; `raw` falls back to ANSI block art. |
| Cache Size    | 1–20          | 5       | LRU eviction once limit is hit. |
| Auto Backup   | on / off      | on      | When on, backs up silently on first write. When off, asks for confirmation first. |

## Contributing

Issues and PRs welcome. The codebase is small (~3k LOC) and uses
[Bubble Tea v2](https://github.com/charmbracelet/bubbletea) +
[Lipgloss v2](https://github.com/charmbracelet/lipgloss).

## License

MIT — see [LICENSE](LICENSE).

Portrait artwork is © Yostar / Nexon and served via [SchaleDB](https://schale.gg/).
This project only links to those images; it does not redistribute them.
