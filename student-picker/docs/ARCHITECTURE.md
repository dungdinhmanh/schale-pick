# Project Architecture: Student-Picker

## Module Breakdown
The project is divided into several modules to separate concerns and ensure high performance in a TUI environment.

### 1. Main Entry & Model (`main.go`, `model.go`)
- `main.go`: Initializes the Bubbletea program and sets up logging.
- `model.go`: Contains the `model` struct, which holds the entire application state (manifest, cache, filtering, current selection).

### 2. View Layer (`view.go`, `styles.go`)
- `view.go`: Implements the `View()` function. It uses a composition of multiple rendering functions:
  - `renderGrid()`: Renders the square icons in the browse list.
  - `renderPreview()`: Renders the student details and allocates space for the portrait.
  - `renderGridBoxWithTabs()`: Wraps content in a box with custom borders and embedded tabs.
- `styles.go`: Defines all Lipgloss styles. Using a centralized style file makes it easy to change the theme (e.g., swapping Catppuccin colors).

### 3. Controller & Input (`update_keyboard.go`)
- Handles user inputs (h/j/k/l for navigation, `/` for searching, `Tab` for switching categories).
- Updates the model state and triggers command execution.

### 4. Image Rendering Strategy (`renderer.go`, `termimg_renderer.go`, `terminal.go`)
- `terminal.go`: Detects terminal environment (e.g., `Kitty`, `iTerm`, `Sixel`).
- `renderer.go`: Manages image loading, caching, and conversion.
- `termimg_renderer.go`: A safety wrapper for the `termimg` library to prevent panics in pseudo-terminal environments.

### 5. Data & Networking (`downloader.go`, `url.go`, `config.go`)
- `downloader.go`: Implements an asynchronous downloader with a semaphore to limit concurrent HTTP requests.
- `url.go` & `config.go`: Stores API endpoints and schema definitions.

## Design Patterns
- **The Elm Architecture (TEA)**: Followed strictly with Model, Update, View.
- **Asynchronous Updates**: Heavy tasks (pre-caching, downloading) are performed as `tea.Cmd` to keep the UI frame rate high (60fps target).
- **Graceful Fallback**: The app detects terminal capabilities and falls back to Sixel or simpler rendering if advanced graphics aren't supported.
