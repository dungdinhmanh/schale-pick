
## 2026-03-29: termimg Research + Portrait Preview Implementation

### termimg Library Findings
- termimg supports 15+ drawing protocols (Sixel, Kitty, iTerm2, etc.)
- Auto-detects terminal and selects best protocol
- Has TUI integrations for bubbletea (experimental)
- Key issue: TTY multiplexing problem - both bubbletea and termimg need to read from /dev/tty
- For Kitty-only use case, native Kitty protocol is simpler

### Architecture Decision
- **Left grid**: Text-only (names in bordered boxes) - icons would require complex TTY mux
- **Right panel**: Portrait image rendered via Kitty Graphics Protocol
- `renderKittyImage()` now uses `ensurePortraitCached()` instead of `ensureIconCached()`
- Portrait path: `~/.cache/student-picker/portrait/{id}.webp`
- Icon path (temp): `/tmp/sp-icon-{id}.webp`

### Files Modified
- renderer.go: Added `ensurePortraitCached()` function
- model.go: Changed renderKittyImage to use portraits, renderPreview shows name text

### Build Status: ✅ SUCCESS
