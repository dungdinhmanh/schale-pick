# Graph Report - .  (2026-04-07)

## Corpus Check
- 48 files · ~80,328 words
- Verdict: corpus is large enough that graph structure adds value.

## Summary
- 234 nodes · 304 edges · 35 communities detected
- Extraction: 64% EXTRACTED · 36% INFERRED · 0% AMBIGUOUS · INFERRED: 108 edges (avg confidence: 0.51)
- Token cost: 0 input · 0 output

## God Nodes (most connected - your core abstractions)
1. `model` - 17 edges
2. `model` - 16 edges
3. `model` - 12 edges
4. `fileExists()` - 8 edges
5. `Cache` - 7 edges
6. `updateFastfetchImage()` - 7 edges
7. `newModel()` - 5 edges
8. `updateFastfetchImage()` - 5 edges
9. `4 Golden Rules for TUI Layout` - 5 edges
10. `menuItem` - 4 edges

## Surprising Connections (you probably didn't know these)
- `main()` --calls--> `newModel()`  [INFERRED]
  student-picker/main.go → main.go
- `main()` --calls--> `loadStudents()`  [INFERRED]
  student-picker/main.go → main.go
- `Bubbletea TUI Skill` --references--> `4 Golden Rules for TUI Layout`  [EXTRACTED]
  SKILL.md → references/golden-rules.md
- `student-picker TUI Application` --implements--> `Bubbletea TUI Skill`  [EXTRACTED]
  README.md → SKILL.md

## Hyperedges (group relationships)
- **TUI Layout Golden Rules System** — golden_rules_rule1_borders, golden_rules_rule2_autowrap, golden_rules_rule3_mouse, golden_rules_rule4_weights [EXTRACTED 0.95]

## Communities

### Community 0 - "Main TUI Application"
Cohesion: 0.1
Nodes (21): calculateFastfetchSize(), cleanupTempFiles(), copyFile(), debug(), fileExists(), findImageExt(), formatFileSize(), getImageDimensions() (+13 more)

### Community 1 - "Bubbletea Model Core"
Cohesion: 0.08
Nodes (28): cacheErrorMsg, cacheSuccessMsg, checkMagick(), checkMagickCmd(), copyFile(), downloadIconCmd(), downloadProgressMsg, fastfetchConfigPath() (+20 more)

### Community 2 - "View Rendering"
Cohesion: 0.21
Nodes (4): model, placeOverlay(), renderBorderTitle(), renderProgressBar()

### Community 3 - "SchaleDB Tests"
Cohesion: 0.14
Nodes (0): 

### Community 4 - "Model Methods"
Cohesion: 0.3
Nodes (1): model

### Community 5 - "Termimg Tests"
Cohesion: 0.14
Nodes (0): 

### Community 6 - "LRU Cache"
Cohesion: 0.24
Nodes (2): Cache, CachedImage

### Community 7 - "Cache Tests"
Cohesion: 0.2
Nodes (0): 

### Community 8 - "Termimg Renderer"
Cohesion: 0.33
Nodes (8): cleanupTermimg(), ClearImages(), clearImagesTermimg(), CloseTerminal(), DrawImage(), DrawImageFile(), GetTerminal(), renderImageTermimg()

### Community 9 - "Image Downloader"
Cohesion: 0.32
Nodes (1): Downloader

### Community 10 - "Icon/Portrait Renderer"
Cohesion: 0.43
Nodes (7): convertWebpToPng(), ensureIconCached(), ensurePngIcon(), ensurePortraitCached(), getIconPath(), getPngIconPath(), getPortraitPath()

### Community 11 - "TUI Golden Rules"
Cohesion: 0.29
Nodes (7): 4 Golden Rules for TUI Layout, Rule 1: Always Account for Borders, Rule 2: Never Auto-Wrap in Bordered Panels, Rule 3: Match Mouse Detection to Layout, Rule 4: Use Weights Not Pixels, Bubbletea TUI Skill, student-picker TUI Application

### Community 12 - "URL Tests"
Cohesion: 0.33
Nodes (0): 

### Community 13 - "Keyboard Input Handler"
Cohesion: 0.4
Nodes (1): model

### Community 14 - "Fetch Test Utility"
Cohesion: 0.67
Nodes (3): fetchAndDraw(), imgJob, main()

### Community 15 - "URL Utilities"
Cohesion: 0.67
Nodes (0): 

### Community 16 - "Terminal Detection"
Cohesion: 0.67
Nodes (0): 

### Community 17 - "Mika Portrait Assets"
Cohesion: 0.67
Nodes (3): Mika Portrait, Mika, Student Portrait

### Community 18 - "Test Helper"
Cohesion: 1.0
Nodes (0): 

### Community 19 - "Configuration"
Cohesion: 1.0
Nodes (0): 

### Community 20 - "SchaleDB Integration"
Cohesion: 1.0
Nodes (0): 

### Community 21 - "Student Struct"
Cohesion: 1.0
Nodes (1): Student

### Community 22 - "Hanako Portrait Assets"
Cohesion: 1.0
Nodes (2): Hanako, Hanako (Gym Portrait)

### Community 23 - "Keyboard Update Entry"
Cohesion: 1.0
Nodes (0): 

### Community 24 - "Styles Module"
Cohesion: 1.0
Nodes (0): 

### Community 25 - "Kitty Debug Session"
Cohesion: 1.0
Nodes (1): Kitty a=d Delete After Issue

### Community 26 - "Termimg Singleton"
Cohesion: 1.0
Nodes (1): termimg Terminal Singleton Pattern

### Community 27 - "Architecture Docs"
Cohesion: 1.0
Nodes (1): Student-Picker Architecture

### Community 28 - "Emoji Width Fix"
Cohesion: 1.0
Nodes (1): Emoji Width Alignment Fix for Terminal UIs

### Community 29 - "Components Catalog"
Cohesion: 1.0
Nodes (1): Bubbletea Components Catalog

### Community 30 - "Student 10081 Icon"
Cohesion: 1.0
Nodes (1): Student Portrait 10081

### Community 31 - "Hanako Portrait"
Cohesion: 1.0
Nodes (1): Hanako

### Community 32 - "Mika Sleeveless"
Cohesion: 1.0
Nodes (1): Mika Sleeveless Portrait

### Community 33 - "Nonomi Portrait"
Cohesion: 1.0
Nodes (1): Nonomi

### Community 34 - "Student 10081 PNG"
Cohesion: 1.0
Nodes (1): Student Portrait 10081

## Knowledge Gaps
- **37 isolated node(s):** `screen`, `settingsActionMsg`, `imgJob`, `Student`, `CachedImage` (+32 more)
  These have ≤1 connection - possible missing edges or undocumented components.
- **Thin community `Test Helper`** (2 nodes): `test.go`, `main()`
  Too small to be a meaningful cluster - may be noise or needs more connections extracted.
- **Thin community `Configuration`** (2 nodes): `config.go`, `CacheDir()`
  Too small to be a meaningful cluster - may be noise or needs more connections extracted.
- **Thin community `SchaleDB Integration`** (2 nodes): `schaledb.go`, `FetchManifest()`
  Too small to be a meaningful cluster - may be noise or needs more connections extracted.
- **Thin community `Student Struct`** (2 nodes): `student.go`, `Student`
  Too small to be a meaningful cluster - may be noise or needs more connections extracted.
- **Thin community `Hanako Portrait Assets`** (2 nodes): `Hanako`, `Hanako (Gym Portrait)`
  Too small to be a meaningful cluster - may be noise or needs more connections extracted.
- **Thin community `Keyboard Update Entry`** (1 nodes): `update_keyboard.go`
  Too small to be a meaningful cluster - may be noise or needs more connections extracted.
- **Thin community `Styles Module`** (1 nodes): `styles.go`
  Too small to be a meaningful cluster - may be noise or needs more connections extracted.
- **Thin community `Kitty Debug Session`** (1 nodes): `Kitty a=d Delete After Issue`
  Too small to be a meaningful cluster - may be noise or needs more connections extracted.
- **Thin community `Termimg Singleton`** (1 nodes): `termimg Terminal Singleton Pattern`
  Too small to be a meaningful cluster - may be noise or needs more connections extracted.
- **Thin community `Architecture Docs`** (1 nodes): `Student-Picker Architecture`
  Too small to be a meaningful cluster - may be noise or needs more connections extracted.
- **Thin community `Emoji Width Fix`** (1 nodes): `Emoji Width Alignment Fix for Terminal UIs`
  Too small to be a meaningful cluster - may be noise or needs more connections extracted.
- **Thin community `Components Catalog`** (1 nodes): `Bubbletea Components Catalog`
  Too small to be a meaningful cluster - may be noise or needs more connections extracted.
- **Thin community `Student 10081 Icon`** (1 nodes): `Student Portrait 10081`
  Too small to be a meaningful cluster - may be noise or needs more connections extracted.
- **Thin community `Hanako Portrait`** (1 nodes): `Hanako`
  Too small to be a meaningful cluster - may be noise or needs more connections extracted.
- **Thin community `Mika Sleeveless`** (1 nodes): `Mika Sleeveless Portrait`
  Too small to be a meaningful cluster - may be noise or needs more connections extracted.
- **Thin community `Nonomi Portrait`** (1 nodes): `Nonomi`
  Too small to be a meaningful cluster - may be noise or needs more connections extracted.
- **Thin community `Student 10081 PNG`** (1 nodes): `Student Portrait 10081`
  Too small to be a meaningful cluster - may be noise or needs more connections extracted.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `model` connect `Model Methods` to `Bubbletea Model Core`?**
  _High betweenness centrality (0.016) - this node is a cross-community bridge._
- **Are the 7 inferred relationships involving `fileExists()` (e.g. with `newSettingsList()` and `.renderPreview()`) actually correct?**
  _`fileExists()` has 7 INFERRED edges - model-reasoned connections that need verification._
- **What connects `screen`, `settingsActionMsg`, `imgJob` to the rest of the system?**
  _37 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `Main TUI Application` be split into smaller, more focused modules?**
  _Cohesion score 0.1 - nodes in this community are weakly interconnected._
- **Should `Bubbletea Model Core` be split into smaller, more focused modules?**
  _Cohesion score 0.08 - nodes in this community are weakly interconnected._
- **Should `SchaleDB Tests` be split into smaller, more focused modules?**
  _Cohesion score 0.14 - nodes in this community are weakly interconnected._
- **Should `Termimg Tests` be split into smaller, more focused modules?**
  _Cohesion score 0.14 - nodes in this community are weakly interconnected._