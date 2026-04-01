# SchaleDB Student Picker - Work Plan

## TL;DR

> **Quick Summary**: Refactor student-picker TUI app to use SchaleDB as image source with caching, two-tab layout (Browse/Installed), efficient image loading, and LRU cache management.
>
> **Deliverables**:
> - SchaleDB manifest fetching (194 students)
> - Two-tab UI: Browse (search) + Installed (.cache)
> - 5x4 grid layout with icons + portrait preview
> - LRU cache system (configurable 1-N images)
> - Efficient image loading (concurrency control)
> - /tmp preview + clean on exit
>
> **Estimated Effort**: Large
> **Parallel Execution**: YES - 3 waves
> **Critical Path**: Schema/Config → Cache System → UI → Integration → Final Testing

---

## Context

### Original Request
Refactor student-picker TUI to use SchaleDB as image source instead of local files.

### Key Requirements
1. **Manifest from SchaleDB**: Fetch `students.json` → extract `Id`, `FamilyName`, `PersonalName`
2. **Two tabs** (switch with Tab key):
   - Browse (default) - shows all 194 students with search
   - Installed (N) - shows cached students, N = count of cached images
   - **Tab key** switches between tabs
3. **Grid layout**: 5 columns × 4 rows = 20 items visible (TEXT ONLY, no thumbnail images)
4. **Right panel**: Portrait preview (lazy-load on selection)
5. **Grid display**: 
   - Each cell: icon/image on TOP, name BELOW
   - 5 columns × 4 rows = 20 items visible
   - Cursor box highlighting current selection
   - Arrow keys (←↑↓→) for navigation
   - Pagination: 20 items/page
6. **Pagination**: 
   - 20 items per page
   - Keep previous page in memory (up to ~5MB)
   - If page 3+, clear page 1 to save memory if total > 5MB
   - Small images (few KB) - no need to manage, clean on exit only
7. **Preview behavior**:
   - Pre-download lightweight thumbnail to /tmp when selection moves
   - Use icon images for preview (smaller than portrait)
   - Load portrait only when user explicitly selects (Enter)
8. **Cache system**: `.cache/` directory, configurable size (default 5), LRU eviction
9. **Image URLs**: Convert to GitHub raw content
10. **Temp handling**: `/tmp` for previews during run, clean on exit

### Image URL Pattern
```
Icon:     https://raw.githubusercontent.com/SchaleDB/SchaleDB/main/images/student/icon/{Id}.webp
Portrait: https://raw.githubusercontent.com/SchaleDB/SchaleDB/main/images/student/portrait/{Id}.webp
```

### Student Schema (from SchaleDB)
```json
{
  "Id": 10000,
  "FamilyName": "Rikuhachima",
  "PersonalName": "Aru",
  ...
}
```

---

## Work Objectives

### Core Objective
Build a TUI app that lets users browse SchaleDB students, preview portraits, and select one to cache for fastfetch.

### Concrete Deliverables
- [ ] `students.json` manifest cached locally (refreshable)
- [ ] Browse tab with search input + 5x4 grid
- [ ] Installed tab showing cached students
- [ ] Portrait preview on right panel (lazy-loaded)
- [ ] LRU cache in `~/.cache/student-picker/`
- [ ] Configurable cache size (≥1 images)
- [ ] Concurrent image loading with semaphore (max ~5 parallel)
- [ ] Clean `/tmp/*.webp` on exit

### Definition of Done
- [ ] `./student-picker` runs without errors
- [ ] Browse tab shows 194 students in 5x4 grid
- [ ] Search filters students by name in real-time
- [ ] Tab switching works (Browse / Installed)
- [ ] Portrait loads on selection (with loading indicator)
- [ ] Enter caches selected portrait to `.cache/`
- [ ] Installed tab shows cached students
- [ ] Cache respects max size with LRU eviction
- [ ] No zombie processes from image loading

### Must Have
- SchaleDB manifest (194 students)
- 5×4 grid layout (20 visible at once)
- Portrait preview panel
- Search filtering
- Two tabs (Browse/Installed)
- LRU cache with configurable size
- Clean temp files on exit

### Must NOT Have
- Kitty icat (replaced with efficient loader)
- Local image directory dependency
- Memory leaks from goroutines
- Race conditions on cache access

---

## Verification Strategy

### Test Decision
- **Infrastructure exists**: NO (new project structure)
- **Automated tests**: NO (TUI app, manual verification)
- **QA Policy**: Agent-executed QA via PTY

### QA Scenarios

**Scenario: Browse tab displays students**
  Tool: PTY (tmux)
  Steps:
    1. Launch app: `./student-picker`
    2. Verify grid renders with student boxes
    3. Count visible boxes (should be ~20 for 5×4 grid)
  Expected: Grid with names, green highlight on first item
  Evidence: `.sisyphus/evidence/browse-grid.png`

**Scenario: Search filters students**
  Tool: PTY
  Steps:
    1. Type `/` to focus search
    2. Type "Aru"
    3. Verify only matching students shown
  Expected: Filtered grid with "Aru" items only
  Evidence: `.sisyphus/evidence/search-filter.png`

**Scenario: Portrait preview loads**
  Tool: PTY + curl verification
  Steps:
    1. Navigate with j/k to different student
    2. Verify portrait area shows image (or loading)
    3. Check debug log for download activity
  Expected: Image appears in preview panel
  Evidence: `.sisyphus/evidence/portrait-preview.png`

**Scenario: Tab switching**
  Tool: PTY
  Steps:
    1. Press `i` to switch to Installed tab
    2. Press `i` again to return to Browse
  Expected: Tab content switches correctly
  Evidence: `.sisyphus/evidence/tab-switch.png`

**Scenario: Cache selection**
  Tool: PTY + filesystem
  Steps:
    1. Select a student, press Enter
    2. Check `~/.cache/student-picker/`
    3. Verify image file exists
  Expected: Portrait cached to .cache/
  Evidence: `.sisyphus/evidence/cache-created.png`

**Scenario: LRU eviction**
  Tool: PTY + bash
  Steps:
    1. Set cache size to 3 (via settings)
    2. Select 5 different students
    3. Check .cache/ - only 3 newest remain
  Expected: Oldest cached image evicted
  Evidence: `.sisyphus/evidence/lru-eviction.png`

**Scenario: Clean exit**
  Tool: Bash
  Steps:
    1. Run app, select some students
    2. Press q to quit
    3. Check /tmp for leftover files
  Expected: No `sp-*.webp` files in /tmp
  Evidence: `.sisyphus/evidence/clean-exit.png`

---

## Execution Strategy

### Parallel Execution Waves

```
Wave 1 (Foundation - 5 tasks):
├── Task 1: Project scaffolding + go.mod
├── Task 2: Config constants + SchaleDB client
├── Task 3: Student manifest types + fetcher
├── Task 4: LRU cache system
└── Task 5: HTTP image downloader with semaphore

Wave 2 (Core UI - 6 tasks):
├── Task 6: Model struct redesign (tabs, grid, cache)
├── Task 7: Grid renderer (5×4 layout)
├── Task 8: Tab bar + Browse view
├── Task 9: Tab bar + Installed view
├── Task 10: Portrait preview component
└── Task 11: Search input handler

Wave 3 (Integration - 5 tasks):
├── Task 12: Tab switching logic
├── Task 13: Selection → cache workflow
├── Task 14: Settings (cache size config)
├── Task 15: Cleanup on exit (finalize)
└── Task 16: Debug logging

Wave FINAL (4 parallel reviews):
├── F1: Plan compliance audit (oracle)
├── F2: Code quality review (unspecified-high)
├── F3: Real manual QA (unspecified-high)
└── F4: Scope fidelity check (deep)
→ Present results → Get explicit user okay
```

### Dependency Matrix
- **T1-T5**: — — — — — → T6-T11
- **T6**: T3, T4, T5 — → T12-T15
- **T7-T11**: T6 — → T12
- **T12-T15**: T7-T11 — → F1-F4
- **F1-F4**: T12-T15 — → User OK

### Agent Dispatch Summary
- **Wave 1**: 5 tasks → `ultrabrain` (T1), `quick` (T2-T5)
- **Wave 2**: 6 tasks → `ultrabrain` (T6), `visual-engineering` (T7-T10), `quick` (T11)
- **Wave 3**: 5 tasks → `deep` (T12-T15), `quick` (T16)
- **Final**: 4 tasks → `oracle`, `unspecified-high`, `unspecified-high`, `deep`

---

## TODOs

- [ ] 1. Project scaffolding + go.mod

  **What to do**:
  - Initialize Go module: `go mod init student-picker`
  - Create directory structure:
    ```
    student-picker/
    ├── main.go
    ├── go.mod
    ├── cache/
    │   └── .gitkeep
    └── assets/
        └── .gitkeep
    ```
  - Add dependencies: `bubbletea`, `lipgloss`, `github.com/google/wrap`

  **Must NOT do**:
  - Don't create actual image files yet
  - Don't modify existing student-picker binary

  **Recommended Agent Profile**:
  - **Category**: `ultrabrain`
    - Reason: Project setup requires understanding of module structure
  - **Skills**: []
  - **Skills Evaluated but Omitted**:
    - `golang-patterns`: overkill for simple scaffolding

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 1 (with Tasks 2-5)
  - **Blocks**: Tasks 6-11
  - **Blocked By**: None

  **References**:

  **Pattern References** (existing code to follow):
  - `/home/kazukisatou/Downloads/fastfetch-script/main.go:1-40` - Existing init() and constants pattern

  **External References**:
  - Go modules: `github.com/charmbracelet/bubbletea`, `github.com/charmbracelet/lipgloss`

  **Acceptance Criteria**:
  - [ ] `go mod init student-picker` succeeds
  - [ ] Directory structure created
  - [ ] `go mod tidy` succeeds

  **QA Scenarios**:
  ```
  Scenario: Project scaffolding
    Tool: Bash
    Steps:
      1. cd /tmp && rm -rf student-picker-test && mkdir student-picker-test
      2. cd student-picker-test && go mod init student-picker
      3. Verify go.mod exists with module name
    Expected: go.mod with "module student-picker"
    Evidence: .sisyphus/evidence/scaffold-go-mod.png
  ```

  **Commit**: YES
  - Message: `init: scaffold project structure`
  - Files: `go.mod`, `cache/.gitkeep`, `assets/.gitkeep`
  - Pre-commit: `go build ./...`

---

- [ ] 2. Config constants + SchaleDB client

  **What to do**:
  - Define constants:
    ```go
    const (
        SchemaURL    = "https://raw.githubusercontent.com/SchaleDB/SchaleDB/refs/heads/main/data/en/students.json"
        IconBaseURL  = "https://raw.githubusercontent.com/SchaleDB/SchaleDB/main/images/student/icon"
        PortraitBaseURL = "https://raw.githubusercontent.com/SchaleDB/SchaleDB/main/images/student/portrait"
        CacheDir     = "~/.cache/student-picker"
        DefaultCacheSize = 5
        GridCols     = 5
        GridRows     = 4
        PreviewW     = 40
        PreviewH     = 20
    )
    ```
  - Create `schaledb.go` with HTTP client for fetching manifest
  - `FetchManifest()` → downloads JSON, caches to `~/.cache/student-picker/students.json`

  **Must NOT do**:
  - Don't implement caching logic here (Task 4)
  - Don't implement image downloading (Task 5)

  **Recommended Agent Profile**:
  - **Category**: `quick`
    - Reason: Standard Go HTTP client code
  - **Skills**: []
  - **Skills Evaluated but Omitted**:
    - `golang-patterns`: basic HTTP patterns

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 1 (with Tasks 1, 3-5)
  - **Blocks**: Task 6 (model uses these types)
  - **Blocked By**: None

  **References**:

  **Pattern References** (existing code to follow):
  - `/home/kazukisatou/Downloads/fastfetch-script/main.go:28-41` - init() pattern for paths
  - `/home/kazukisatou/Downloads/fastfetch-script/main.go:773-818` - loadStudents() pattern

  **External References**:
  - Go net/http: standard library

  **Acceptance Criteria**:
  - [ ] `FetchManifest()` returns `[]Student` slice
  - [ ] Manifest cached to `~/.cache/student-picker/students.json`
  - [ ] `go build` succeeds

  **QA Scenarios**:
  ```
  Scenario: SchaleDB client fetches manifest
    Tool: Bash
    Preconditions: Network available
    Steps:
      1. Run app
      2. Check ~/.cache/student-picker/students.json exists
      3. Verify JSON has 194 entries with Id, FamilyName, PersonalName
    Expected: Cached manifest with correct schema
    Evidence: .sisyphus/evidence/manifest-cached.png
  ```

  **Commit**: YES
  - Message: `feat: add SchaleDB client and config`
  - Files: `schaledb.go`, `config.go`
  - Pre-commit: `go build ./...`

---

- [ ] 3. Student manifest types + fetcher

  **What to do**:
  - Define types:
    ```go
    type Student struct {
        Id           int    // e.g., 10000
        FamilyName   string // e.g., "Rikuhachima"
        PersonalName string // e.g., "Aru"
    }
    
    type Manifest struct {
        Students []Student
        FetchedAt time.Time
    }
    ```
  - `FetchManifest() (Manifest, error)` - downloads + parses JSON
  - `GetStudentIconURL(id int) string` - returns icon URL
  - `GetStudentPortraitURL(id int) string` - returns portrait URL

  **Must NOT do**:
  - Don't implement image download (Task 5)
  - Don't implement cache storage (Task 4)

  **Recommended Agent Profile**:
  - **Category**: `quick`
    - Reason: Type definitions and URL helpers
  - **Skills**: []
  - **Skills Evaluated but Omitted**:
    - `golang-patterns`: simple struct definitions

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 1 (with Tasks 1-2, 4-5)
  - **Blocks**: Task 6
  - **Blocked By**: None

  **References**:

  **API/Type References** (contracts to implement against):
  - SchaleDB JSON: `{"Id": 10000, "FamilyName": "Rikuhachima", "PersonalName": "Aru"}`

  **Acceptance Criteria**:
  - [ ] `Student` struct has Id, FamilyName, PersonalName
  - [ ] `GetStudentIconURL(10000)` returns correct URL
  - [ ] `GetStudentPortraitURL(10000)` returns correct URL
  - [ ] `go build` succeeds

  **QA Scenarios**:
  ```
  Scenario: Student types and URL generation
    Tool: Bash
    Steps:
      1. go run . -test-urls 2>&1 | head -10
    Expected: URLs contain correct SchaleDB paths
    Evidence: .sisyphus/evidence/url-generation.png
  ```

  **Commit**: YES
  - Message: `feat: add Student types and URL helpers`
  - Files: `student.go`, `url.go`
  - Pre-commit: `go build ./...`

---

- [ ] 4. LRU cache system

  **What to do**:
  - `cache.go` with LRU implementation:
    ```go
    type Cache struct {
        dir       string
        maxSize   int
        items     map[string]CachedImage  // key = studentId
        order     []string                // ordered list (newest last)
        mu        sync.Mutex
    }
    
    type CachedImage struct {
        StudentId int
        IconPath  string  // ~/.cache/student-picker/icon/{Id}.webp
        PortraitPath string
        CachedAt  time.Time
    }
    ```
  - `NewCache(dir string, maxSize int) *Cache`
  - `Get(studentId int) (CachedImage, bool)`
  - `Put(studentId int, iconData, portraitData []byte) error`
  - `EvictOldest()` when over maxSize
  - `List() []CachedImage` - returns all cached (for Installed tab)
  - `Clear()` - removes all cached files

  **Must NOT do**:
  - Don't implement HTTP downloads (Task 5)
  - Don't implement UI (Tasks 6-15)

  **Recommended Agent Profile**:
  - **Category**: `quick`
    - Reason: Standard LRU cache implementation
  - **Skills**: []
  - **Skills Evaluated but Omitted**:
    - `golang-patterns`: data structures and sync.Mutex

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 1 (with Tasks 1-3, 5)
  - **Blocks**: Tasks 12-15 (selection workflow needs cache)
  - **Blocked By**: None

  **References**:

  **Pattern References** (existing code to follow):
  - `/home/kazukisatou/Downloads/fastfetch-script/main.go:829-841` - cleanup pattern

  **External References**:
  - Go sync.Mutex for thread safety
  - Go os for file operations

  **Acceptance Criteria**:
  - [ ] Cache respects maxSize
  - [ ] LRU eviction removes oldest when over limit
  - [ ] `List()` returns cached students in LRU order (newest first)
  - [ ] Thread-safe with mutex

  **QA Scenarios**:
  ```
  Scenario: LRU eviction works
    Tool: Bash
    Steps:
      1. Create cache with maxSize=3
      2. Put students 1, 2, 3, 4
      3. Verify only 2, 3, 4 remain (1 evicted)
      4. Verify order is correct
    Expected: Student 1 evicted, others in LRU order
    Evidence: .sisyphus/evidence/lru-test.png
  ```

  **Commit**: YES
  - Message: `feat: implement LRU cache system`
  - Files: `cache.go`
  - Pre-commit: `go build ./...`

---

- [ ] 5. HTTP image downloader with semaphore

  **What to do**:
  - `downloader.go`:
    ```go
    type Downloader struct {
        client    *http.Client
        semaphore chan struct{}  // max concurrent downloads
    }
    
    func NewDownloader(maxConcurrent int) *Downloader
    func (d *Downloader) DownloadImage(url string) ([]byte, error)
    func (d *Downloader) DownloadPortrait(id int) ([]byte, error)
    func (d *Downloader) DownloadIcon(id int) ([]byte, error)
    ```
  - Max ~5 concurrent downloads to avoid overwhelming system
  - Timeout: 30 seconds per image
  - Retry: 1 retry on failure
  - `DownloadWithProgress(url string, progress func(downloaded int))` for UI feedback

  **Must NOT do**:
  - Don't implement UI progress display (Task 10)
  - Don't call from main goroutine (causes blocking)

  **Recommended Agent Profile**:
  - **Category**: `quick`
    - Reason: Standard concurrent HTTP patterns
  - **Skills**: []
  - **Skills Evaluated but Omitted**:
    - `golang-patterns`: concurrency control

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 1 (with Tasks 1-4)
  - **Blocks**: Tasks 10, 13 (preview and cache need downloader)
  - **Blocked By**: None

  **References**:

  **Pattern References** (existing code to follow):
  - `/home/kazukisatou/Downloads/fastfetch-script/main.go:501-586` - existing kitty rendering logic

  **External References**:
  - Go net/http with timeout context

  **Acceptance Criteria**:
  - [ ] Max 5 concurrent downloads
  - [ ] Timeout after 30 seconds
  - [ ] Returns []byte on success
  - [ ] Returns error on failure

  **QA Scenarios**:
  ```
  Scenario: Concurrent download limit
    Tool: Bash
    Preconditions: Network available
    Steps:
      1. Spawn 10 download goroutines
      2. Use semaphore to limit to 5 concurrent
      3. Verify only 5 active at once
    Expected: Semaphore controls concurrency
    Evidence: .sisyphus/evidence/concurrency-test.png
  ```

  **Commit**: YES
  - Message: `feat: add HTTP downloader with semaphore`
  - Files: `downloader.go`
  - Pre-commit: `go build ./...`

---

- [ ] 6. Model struct redesign (tabs, grid, cache)

  **What to do**:
  - Redefine model:
    ```go
    type Tab int
    const (
        TabBrowse Tab = iota
        TabInstalled
    )
    
    type model struct {
        // Layout
        width    int
        height   int
        
        // Data
        manifest []Student       // all students from SchaleDB
        filtered []Student       // filtered by search (or all if no search)
        
        // UI State
        tab      Tab
        gridIdx  int             // current selection index in filtered
        
        // Search
        searchQuery string
        searchInput textinput.Model
        
        // Cache
        cache    *Cache
        cacheSize int           // configured max size
        
        // Preview (loaded async)
        previewURL string
        previewLoading bool
        previewError string
        
        // Settings
        settingsVisible bool
        settingsList list.Model
    }
    ```
  - `newModel() model` - initializes with empty manifest
  - `Init()` - fetches manifest in background, shows loading state

  **Must NOT do**:
  - Don't implement View() (Tasks 7-10)
  - Don't implement Update() (Tasks 11-14)

  **Recommended Agent Profile**:
  - **Category**: `ultrabrain`
    - Reason: Core data structure design, affects everything
  - **Skills**: []
  - **Skills Evaluated but Omitted**:
    - `golang-patterns`: type design

  **Parallelization**:
  - **Can Run In Parallel**: NO (sequential after T5)
  - **Parallel Group**: Wave 2 (with Tasks 7-11)
  - **Blocks**: Tasks 7-11
  - **Blocked By**: Tasks 1-5

  **References**:

  **Pattern References** (existing code to follow):
  - `/home/kazukisatou/Downloads/fastfetch-script/main.go:139-151` - existing model struct

  **API/Type References** (contracts to implement against):
  - `Student` from Task 3
  - `Cache` from Task 4

  **Acceptance Criteria**:
  - [ ] model has all required fields
  - [ ] `newModel()` initializes correctly
  - [ ] `Init()` starts manifest fetch

  **QA Scenarios**:
  ```
  Scenario: Model initializes with empty state
    Tool: Bash
    Steps:
      1. go run . (without network)
      2. Verify app starts (may show error if no cache)
    Expected: App launches without crash
    Evidence: .sisyphus/evidence/model-init.png
  ```

  **Commit**: YES
  - Message: `refactor: redesign model for SchaleDB`
  - Files: `model.go`
  - Pre-commit: `go build ./...`

---

- [ ] 7. Grid renderer (5×4 layout, icon + name)

  **What to do**:
  - `renderGrid(students []Student, selectedIdx int, gridCols int) string`
  - Constants:
    ```go
    cellW = 18   // cell width (characters per column)
    cellH = 6    // cell height (lines per row - icon + name)
    gridCols = 5
    gridRows = 4
    visibleCount = gridCols * gridRows  // 20
    ```
  - Grid layout (single box cursor - ONLY selected cell has border):
    <pre>
    ┌────┐
    │ [img] │  [img]    [img]    [img]    [img]  
    │ Aru   │  Hina     Mika     Ayumu    Ayaka  
    └────┘
    </pre>
  - Each cell: icon on top, name below
  - **NO borders on unselected cells** - just spaces
  - **Selected cell: has box border** (cursor box style)
  - Pagination indicator if more students than visible

  **Must NOT do**:
  - Don't implement image loading (Task 5 handles download)
  - Don't implement scroll (just paginate)

  **Recommended Agent Profile**:
  - **Category**: `visual-engineering`
    - Reason: Grid layout with images, lipgloss styling
  - **Skills**: []
  - **Skills Evaluated but Omitted**:
    - `frontend-philosophy`: lipgloss styling

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 2 (with Tasks 6, 8-11)
  - **Blocks**: Task 12 (grid display)
  - **Blocked By**: Task 6

  **References**:

  **Pattern References** (existing code to follow):
  - `/home/kazukisatou/Downloads/fastfetch-script/main.go:410-466` - existing grid rendering
  - `/home/kazukisatou/Downloads/fastfetch-script/main.go:501-586` - kitty image rendering
  - `/home/kazukisatou/Downloads/fastfetch-script/main.go:92-137` - lipgloss styles

  **External References**:
  - lipgloss.Reverse for cursor box highlighting
  - kitty icat for image rendering

  **Acceptance Criteria**:
  - [ ] 5 columns × 4 rows = 20 visible
  - [ ] Each cell: icon on top, name below
  - [ ] ONLY selected cell has box border
  - [ ] Unselected cells have no border
  - [ ] Pagination if >20 students

  **QA Scenarios**:
  ```
  Scenario: Grid renders 20 students with icons
    Tool: PTY
    Steps:
      1. Launch app
      2. Verify grid shows 5×4 cells
      3. Verify each cell has icon on top, name below
      4. Verify selected cell has cursor box
    Expected: 4 rows of 5 cells with icons + names
    Evidence: .sisyphus/evidence/grid-icon-name.png
  ```

  **Commit**: YES
  - Message: `feat: implement 5x4 grid renderer (icon + name)`
  - Files: `grid.go`
  - Pre-commit: `go build ./...`

---

- [ ] 8. Tab bar + Browse view

  **What to do**:
  - Tab bar component:
    ```
    [ Browse ]  [ Installed (3) ]
    ```
  - Active tab: bright green text
  - Inactive tab: dim gray text
  - Installed tab shows count of cached images in parentheses
  - Browse view structure (single box cursor - ONLY selected has border):
    <pre>
    ┌─ Browse ────Installed (3) ──────────────┬─ Preview ─────────┐
    │ / search...                              │                    │
    ├──────────────────────────────────────────┤                    │
    │  [img]    [img]    [img]    [img]    [img]                    │
    │  Aru      Hina     Mika     Ayumu    Ayaka                   │
    │                                                        │
    │  [img]    [img]    [img]    [img]    [img]                │
    │  Chinatsu Haruka   Iroha    J尾      ...                    │
    │                                                        │
    │ (page 1/10)                               │                   │
    ├──────────────────────────────────────────┴───────────────────┤
    │ ←↑↓→: move | Enter: Use | Tab: switch tab | q: quit           │
    └──────────────────────────────────────────────────────────────┘
    </pre>
    Note: Grid cells have NO borders. Selected cell has a box border cursor.
  - Search input at top of grid panel
  - Page indicator at bottom left
  - **Each grid cell**: icon on top, name below

  **Must NOT do**:
  - Don't implement search logic (Task 11)
  - Don't implement portrait loading (Task 10)

  **Recommended Agent Profile**:
  - **Category**: `visual-engineering`
    - Reason: Complex layout with tabs and panels
  - **Skills**: []
  - **Skills Evaluated but Omitted**:
    - `frontend-philosophy`: panel layout

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 2 (with Tasks 6-7, 9-11)
  - **Blocks**: Task 12 (tab switching)
  - **Blocked By**: Task 6

  **References**:

  **Pattern References** (existing code to follow):
  - `/home/kazukisatou/Downloads/fastfetch-script/main.go:383-408` - viewMain() layout
  - `/home/kazukisatou/Downloads/fastfetch-script/main.go:92-137` - lipgloss styles

  **Acceptance Criteria**:
  - [ ] Tab bar shows both tabs with active indicator
  - [ ] Installed tab shows cached count
  - [ ] Browse view matches layout above
  - [ ] Preview panel on right side
  - [ ] Grid cells: icon on top, name below

  **QA Scenarios**:
  ```
  Scenario: Browse tab renders correctly
    Tool: PTY
    Steps:
      1. Launch app (Browse is default)
      2. Verify tab bar shows [ Browse ] active
      3. Verify search input visible
      4. Verify 5×4 grid with icon+name cells
      5. Verify preview panel on right
    Expected: Browse tab fully rendered with correct layout
    Evidence: .sisyphus/evidence/browse-tab-icon-name.png
  ```

  **Commit**: YES
  - Message: `feat: implement Browse tab view`
  - Files: `view_browse.go`
  - Pre-commit: `go build ./...`

---

- [ ] 9. Tab bar + Installed view

  **What to do**:
  - Same tab bar as Browse
  - Installed view shows cached students (single box cursor):
    <pre>
    ┌─ Browse ────Installed (3) ──────────────┬─ Preview ─────────┐
    │ [ Browse ]  [ Installed (3) ]            │                   │
    ├──────────────────────────────────────────┤                   │
    │  [img]    [img]    [img]                                  │
    │  Aru      Hina     Mika                                    │
    │                                                        │
    │ (3 cached)                              │                   │
    ├──────────────────────────────────────────┴───────────────────┤
    │ ←↑↓→: move | Enter: Use | Tab: switch tab | q: quit           │
    └──────────────────────────────────────────────────────────────┘
    </pre>
    Note: Grid cells have NO borders. Selected cell has box cursor.
  - Grid shows cached students (from Cache.List())
  - Same 5×4 layout with icon+name cells
  - Tab bar shows "Installed (N)" where N = count of cached images
  - If no cached students, show "No cached students" message
  - **Tab key** switches between Browse/Installed tabs

  **Must NOT do**:
  - Don't implement cache selection (Task 13)
  - Don't implement fastfetch integration (Task 13)

  **Recommended Agent Profile**:
  - **Category**: `visual-engineering`
    - Reason: Same complexity as Browse tab
  - **Skills**: []
  - **Skills Evaluated but Omitted**:
    - `frontend-philosophy`: panel layout

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 2 (with Tasks 6-8, 10-11)
  - **Blocks**: Task 12 (tab switching)
  - **Blocked By**: Task 6

  **References**:

  **Pattern References** (existing code to follow):
  - `/home/kazukisatou/Downloads/fastfetch-script/main.go:614-698` - settings view pattern

  **Acceptance Criteria**:
  - [ ] Tab bar shows "Installed (N)" where N = cached count
  - [ ] Shows cached students in 5×4 grid (icon + name)
  - [ ] Shows "No cached students" if empty
  - [ ] "i" key opens settings

  **QA Scenarios**:
  ```
  Scenario: Installed tab with cached students
    Tool: PTY
    Preconditions: 3 students cached
    Steps:
      1. Press i (settings) to go back to Browse
      2. Verify grid shows cached students with icons
    Expected: Cached students visible with icon+name layout
    Evidence: .sisyphus/evidence/installed-tab-icon-name.png
  ```

  **Commit**: YES
  - Message: `feat: implement Installed tab view`
  - Files: `view_installed.go`
  - Pre-commit: `go build ./...`

---

- [ ] 10. Portrait preview component

  **What to do**:
  - `renderPreview(studentId int, loading bool, err string) string`
  - Portrait area structure:
    ```
    ┌────────────────────────────────────┐
    │  Loading...                       │
    │  or                               │
    │  [portrait image]                 │
    │  or                               │
    │  Error: failed to load            │
    ├────────────────────────────────────┤
    │  Student Name                     │
    │  FamilyName, PersonalName         │
    └────────────────────────────────────┘
    ```
  - Loading state: spinner/progress indicator
  - Error state: red error message
  - Success state: portrait image + name info
  - Uses async download via Task 5 downloader
  - Download triggered on gridIdx change (debounced ~200ms)

  **Must NOT do**:
  - Don't implement debounce logic (Task 11)
  - Don't implement downloader (Task 5)

  **Recommended Agent Profile**:
  - **Category**: `visual-engineering`
    - Reason: Visual component with states
  - **Skills**: []
  - **Skills Evaluated but Omitted**:
    - `frontend-philosophy`: component states

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 2 (with Tasks 6-9, 11)
  - **Blocks**: Task 12 (preview display)
  - **Blocked By**: Task 6

  **References**:

  **Pattern References** (existing code to follow):
  - `/home/kazukisatou/Downloads/fastfetch-script/main.go:468-499` - existing preview render

  **Acceptance Criteria**:
  - [ ] Shows loading state
  - [ ] Shows portrait when loaded
  - [ ] Shows error if failed
  - [ ] Shows student name info

  **QA Scenarios**:
  ```
  Scenario: Portrait preview shows after selection
    Tool: PTY
    Preconditions: Network available
    Steps:
      1. Navigate to different student
      2. Wait for portrait to load
      3. Verify image appears
    Expected: Portrait displays correctly
    Evidence: .sisyphus/evidence/portrait-load.png
  ```

  **Commit**: YES
  - Message: `feat: implement portrait preview component`
  - Files: `view_preview.go`
  - Pre-commit: `go build ./...`

---

- [ ] 11. Search input handler

  **What to do**:
  - Search logic in Update():
    ```go
    case "/":
        // Focus search input
        m.searchInput.Focus()
        return m, nil
        
    case tea.KeyMsg:
        if m.searchInput.Focused() {
            m.searchInput, cmd = m.searchInput.Update(msg)
            m.filtered = filterStudents(m.manifest, m.searchInput.Value())
            // Clamp gridIdx to valid range
            if m.gridIdx >= len(m.filtered) {
                m.gridIdx = max(0, len(m.filtered)-1)
            }
            return m, cmd
        }
    }
    ```
  - `filterStudents(students []Student, query string) []Student`:
    - Case-insensitive match
    - Matches FamilyName OR PersonalName OR Id (as string)
    - Returns all if query empty
  - Escape key clears search and unfocuses

  **Must NOT do**:
  - Don't implement filter UI (Tasks 8-9)
  - Don't implement grid rendering (Task 7)

  **Recommended Agent Profile**:
  - **Category**: `quick`
    - Reason: Standard text filtering logic
  - **Skills**: []
  - **Skills Evaluated but Omitted**:
    - `golang-patterns`: string matching

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 2 (with Tasks 6-10)
  - **Blocks**: Task 12 (search works)
  - **Blocked By**: Task 6

  **References**:

  **Pattern References** (existing code to follow):
  - `/home/kazukisatou/Downloads/fastfetch-script/main.go:773-818` - loadStudents filtering

  **External References**:
  - charmbracelet/textinput

  **Acceptance Criteria**:
  - [ ] `/` focuses search
  - [ ] Typing filters students
  - [ ] Escape clears and unfocuses
  - [ ] Empty query shows all

  **QA Scenarios**:
  ```
  Scenario: Search filters students
    Tool: PTY
    Steps:
      1. Press /
      2. Type "Aru"
      3. Verify only Aru shown
      4. Press Escape
      5. Verify all students shown again
    Expected: Filter works correctly
    Evidence: .sisyphus/evidence/search-test.png
  ```

  **Commit**: YES
  - Message: `feat: implement search filtering`
  - Files: `search.go`
  - Pre-commit: `go build ./...`

---

- [ ] 12. Tab switching logic

  **What to do**:
  - Update() handles tab switching:
    ```go
    case "i":
        switch m.tab {
        case TabBrowse:
            m.tab = TabInstalled
        case TabInstalled:
            m.tab = TabBrowse
        }
        m.gridIdx = 0  // Reset selection
        m.searchInput.Reset()  // Clear search
        return m, nil
    }
    ```
  - Tab state affects:
    - Which view renders (Browse vs Installed)
    - Which student list is used (manifest vs cache)
    - Help text changes slightly

  **Must NOT do**:
  - Don't implement cache listing (Task 4)
  - Don't implement fastfetch integration (Task 13)

  **Recommended Agent Profile**:
  - **Category**: `deep`
    - Reason: State management across views
  - **Skills**: []
  - **Skills Evaluated but Omitted**:
    - `golang-patterns`: state machine

  **Parallelization**:
  - **Can Run In Parallel**: NO
  - **Parallel Group**: Wave 3 (with Tasks 13-16)
  - **Blocks**: Final integration
  - **Blocked By**: Tasks 7-11

  **References**:

  **Pattern References** (existing code to follow):
  - `/home/kazukisatou/Downloads/fastfetch-script/main.go:224-230` - existing screen switching

  **Acceptance Criteria**:
  - [ ] `i` key toggles tabs
  - [ ] Selection resets on tab switch
  - [ ] Search clears on tab switch

  **QA Scenarios**:
  ```
  Scenario: Tab switching works
    Tool: PTY
    Steps:
      1. Press i, verify Installed tab
      2. Press i again, verify Browse tab
      3. Repeat multiple times
    Expected: Tabs switch correctly
    Evidence: .sisyphus/evidence/tab-switch.png
  ```

  **Commit**: YES
  - Message: `feat: implement tab switching`
  - Files: `tabs.go`
  - Pre-commit: `go build ./...`

---

- [ ] 13. Selection → cache workflow

  **What to do**:
  - Enter key behavior:
    1. Get current student from gridIdx
    2. Download portrait if not cached
    3. Save to cache via Cache.Put()
    4. If cache exceeds maxSize, evict oldest
    5. Show success message
    6. If TabInstalled, refresh cache list
  - For fastfetch integration:
    - On TabInstalled, Enter copies cached image to:
      `~/.cache/student-picker/fastfetch-portrait.webp`
    - Updates fastfetch config to point to this file

  **Must NOT do**:
  - Don't implement fastfetch config writing (keep simple, copy to cache only)
  - Don't implement downloader (Task 5)

  **Recommended Agent Profile**:
  - **Category**: `deep`
    - Reason: Multi-step workflow with async operations
  - **Skills**: []
  - **Skills Evaluated but Omitted**:
    - `golang-patterns`: async workflow

  **Parallelization**:
  - **Can Run In Parallel**: NO
  - **Parallel Group**: Wave 3 (with Tasks 12, 14-16)
  - **Blocks**: Final integration
  - **Blocked By**: Tasks 4, 5, 12

  **References**:

  **Pattern References** (existing code to follow):
  - `/home/kazukisatou/Downloads/fastfetch-script/main.go:253-265` - existing Enter handling
  - `/home/kazukisatou/Downloads/fastfetch-script/main.go:696-732` - updateFastfetchImage

  **Acceptance Criteria**:
  - [ ] Enter caches current student portrait
  - [ ] Cache respects max size with LRU
  - [ ] Success message shown

  **QA Scenarios**:
  ```
  Scenario: Cache selection
    Tool: PTY + bash
    Preconditions: Network available
    Steps:
      1. Select a student in Browse tab
      2. Press Enter
      3. Check ~/.cache/student-picker/
      4. Verify portrait exists
    Expected: Portrait cached to .cache/
    Evidence: .sisyphus/evidence/cache-select.png
  ```

  **Commit**: YES
  - Message: `feat: implement selection to cache workflow`
  - Files: `selection.go`
  - Pre-commit: `go build ./...`

---

- [ ] 14. Settings (cache size config)

  **What to do**:
  - Settings accessed via `s` key from any tab
  - Settings overlay shows:
    ```
    ┌─────────────────────────────────────┐
    │  Settings                      [X] │
    ├─────────────────────────────────────┤
    │  Cache Size: [  5  ] +/- buttons   │
    │                                     │
    │  [  5  ] = keep 5 cached images    │
    │                                     │
    │  Current cache: 3 images           │
    │                                     │
    │  [Clear Cache]                      │
    │                                     │
    │  q/b/esc: close                     │
    └─────────────────────────────────────┘
    ```
  - Cache size: min 1, max 50
  - `+/-` keys adjust cache size
  - "Clear Cache" removes all cached files
  - Settings stored in `~/.config/student-picker/settings.json`

  **Must NOT do**:
  - Don't implement full settings system (keep simple)
  - Don't persist settings between runs (only cache size)

  **Recommended Agent Profile**:
  - **Category**: `quick`
    - Reason: Simple settings UI
  - **Skills**: []
  - **Skills Evaluated but Omitted**:
    - `golang-patterns`: simple config

  **Parallelization**:
  - **Can Run In Parallel**: NO
  - **Parallel Group**: Wave 3 (with Tasks 12-13, 15-16)
  - **Blocks**: Final integration
  - **Blocked By**: Task 12

  **References**:

  **Pattern References** (existing code to follow):
  - `/home/kazukisatou/Downloads/fastfetch-script/main.go:321-369` - existing settings handling
  - `/home/kazukisatou/Downloads/fastfetch-script/main.go:614-698` - viewSettings

  **Acceptance Criteria**:
  - [ ] `s` key opens settings
  - [ ] +/- changes cache size
  - [ ] Clear Cache removes all cached files
  - [ ] Settings persist on restart

  **QA Scenarios**:
  ```
  Scenario: Settings cache size
    Tool: PTY
    Steps:
      1. Press s to open settings
      2. Press + to increase cache size
      3. Press - to decrease
      4. Press q to close
    Expected: Cache size changes, settings saved
    Evidence: .sisyphus/evidence/settings.png
  ```

  **Commit**: YES
  - Message: `feat: implement settings for cache size`
  - Files: `settings.go`
  - Pre-commit: `go build ./...`

---

- [ ] 15. Cleanup on exit (finalize)

  **What to do**:
  - On app exit (q, Ctrl+C, or error):
    1. Stop all pending downloads (close channels)
    2. Remove all `/tmp/sp-*.webp` temp files
    3. Remove `/tmp/student-picker.log`
  - Cleanup function:
    ```go
    func cleanup() {
        // Stop downloader
        downloader.Close()
        
        // Clean temp files
        entries, _ := os.ReadDir(os.TempDir())
        for _, e := range entries {
            if strings.HasPrefix(e.Name(), "sp-") {
                os.Remove(filepath.Join(os.TempDir(), e.Name()))
            }
        }
        
        // Remove log
        os.Remove("/tmp/student-picker.log")
    }
    ```
  - Called via `defer cleanup()` in main()

  **Must NOT do**:
  - Don't clear cache (leave cached images)
  - Don't modify fastfetch config

  **Recommended Agent Profile**:
  - **Category**: `quick`
    - Reason: Cleanup function
  - **Skills**: []
  - **Skills Evaluated but Omitted**:
    - `golang-patterns`: defer pattern

  **Parallelization**:
  - **Can Run In Parallel**: NO
  - **Parallel Group**: Wave 3 (with Tasks 12-14, 16)
  - **Blocks**: Final integration
  - **Blocked By**: Task 14

  **References**:

  **Pattern References** (existing code to follow):
  - `/home/kazukisatou/Downloads/fastfetch-script/main.go:829-841` - existing cleanup

  **Acceptance Criteria**:
  - [ ] Temp files removed on exit
  - [ ] No zombie processes
  - [ ] Log file removed

  **QA Scenarios**:
  ```
  Scenario: Clean exit
    Tool: Bash
    Preconditions: App ran previously
    Steps:
      1. ls /tmp/sp-* 2>/dev/null | wc -l
      2. Run app and quit with q
      3. ls /tmp/sp-* 2>/dev/null | wc -l
    Expected: Second count is 0
    Evidence: .sisyphus/evidence/clean-exit.png
  ```

  **Commit**: YES
  - Message: `feat: implement cleanup on exit`
  - Files: `cleanup.go`
  - Pre-commit: `go build ./...`

---

- [ ] 16. Debug logging

  **What to do**:
  - Debug log to `/tmp/student-picker.log`:
    ```go
    func debug(format string, args ...interface{}) {
        if debugLog != nil {
            msg := fmt.Sprintf(format, args...)
            debugLog.WriteString(msg + "\n")
            debugLog.Sync()
        }
    }
    ```
  - Log events:
    - App start/exit
    - Manifest fetch (success/failure)
    - Image download start/complete/fail
    - Cache put/evict
    - Tab switch
    - Search query
  - Enable via `-debug` flag or `DEBUG=1` env var

  **Must NOT do**:
  - Don't log sensitive data
  - Don't leave debug enabled by default

  **Recommended Agent Profile**:
  - **Category**: `quick`
    - Reason: Simple logging utility
  - **Skills**: []
  - **Skills Evaluated but Omitted**:
    - `golang-patterns`: logging pattern

  **Parallelization**:
  - **Can Run In Parallel**: NO
  - **Parallel Group**: Wave 3 (with Tasks 12-15)
  - **Blocks**: None
  - **Blocked By**: Task 12

  **References**:

  **Pattern References** (existing code to follow):
  - `/home/kazukisatou/Downloads/fastfetch-script/main.go:68-84` - existing debug logging

  **Acceptance Criteria**:
  - [ ] Debug log created with -debug flag
  - [ ] Key events logged
  - [ ] No sensitive data logged

  **QA Scenarios**:
  ```
  Scenario: Debug logging works
    Tool: Bash
    Steps:
      1. DEBUG=1 ./student-picker &
      2. sleep 2
      3. kill %1 2>/dev/null
      4. cat /tmp/student-picker.log
    Expected: Log shows startup events
    Evidence: .sisyphus/evidence/debug-log.png
  ```

  **Commit**: YES
  - Message: `feat: add debug logging`
  - Files: `debug.go`
  - Pre-commit: `go build ./...`

---

## Final Verification Wave

- [ ] F1. **Plan Compliance Audit** — `oracle`
  Read the plan end-to-end. For each "Must Have": verify implementation exists. For each "Must NOT Have": search for forbidden patterns.
  Output: `Must Have [N/N] | Must NOT Have [N/N] | Tasks [N/N] | VERDICT`

- [ ] F2. **Code Quality Review** — `unspecified-high`
  Run `go vet`, linter, `go build`. Review for: `as any`/`@ts-ignore`, empty catches, console.log, unused imports. Check AI slop patterns.
  Output: `Build [PASS/FAIL] | Vet [PASS/FAIL] | Issues [N] | VERDICT`

- [ ] F3. **Real Manual QA** — `unspecified-high`
  Execute EVERY QA scenario from EVERY task. Save evidence. Test integration across tasks.
  Output: `Scenarios [N/N pass] | Integration [N/N] | VERDICT`

- [ ] F4. **Scope Fidelity Check** — `deep`
  For each task: read "What to do", read actual diff. Verify 1:1. Check "Must NOT do" compliance.
  Output: `Tasks [N/N compliant] | Contamination [CLEAN/N issues] | VERDICT`

---

## Commit Strategy

- **Wave 1**: `feat: add foundation (config, schema, cache, downloader)`
- **Wave 2**: `feat: add UI layer (model, grid, tabs, preview)`
- **Wave 3**: `feat: add integration (tab switching, selection, settings)`
- **Final**: `feat: add polish (cleanup, debug logging)` + `fix: review feedback`

---

## Success Criteria

### Verification Commands
```bash
# Build
go build -o student-picker .

# Run
./student-picker

# Check cache
ls ~/.cache/student-picker/

# Check temp clean
ls /tmp/sp-*.webp 2>/dev/null | wc -l  # Should be 0

# Check manifest cached
cat ~/.cache/student-picker/students.json | python3 -c "import json,sys; print(len(json.load(sys.stdin)))"  # 194
```

### Final Checklist
- [ ] All "Must Have" present
- [ ] All "Must NOT Have" absent
- [ ] No race conditions
- [ ] No memory leaks
- [ ] Clean temp files on exit
- [ ] 194 students displayed
- [ ] 5×4 grid layout
- [ ] Browse/Installed tabs work
- [ ] Search filters correctly
- [ ] Cache respects max size
- [ ] LRU eviction works
