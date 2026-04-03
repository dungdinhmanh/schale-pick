# UI/UX Improvements: Student-Picker

## Design Principles
The refactored TUI follows a clean, modern aesthetic inspired by the Catppuccin color scheme, prioritizing legibility and visual hierarchy.

## 1. Grid & Navigation
- **Optimized Grid**: The grid size (columns and rows) is dynamically calculated based on terminal width, ensuring content is never truncated accidentally.
- **Visual Feedback**: Selection is clearly highlighted with a rounded border and a distinctive color (`#A6E3A1`).
- **Embedded Tabs**: The `Browse` and `Installed` tabs are integrated into the grid's top border. The `Installed` tab now displays a dynamic count (e.g., `Installed (15)`) to show progress at a glance.

## 2. Preview System
- **Dual-Pane View**: The split-screen layout allows users to see the list on the left and the large portrait on the right simultaneously.
- **Rich Content**: The preview displays the student's full name (Family Name + Personal Name) and high-quality portraits from the SchaleDB archive.
- **Smooth Updates**: Portraits update instantly as the user navigates the grid, with background caching ensuring no UI flickers.

## 3. Search Mode
- **Integrated Search**: Pressing `/` transforms the top border into a search input field (`Search: ...`). 
- **Real-time Filtering**: The list filters instantly as characters are typed, with empty state handling ("No students") when no matches are found.

## 4. Help & Progress Feedback
- **Help Modal**: Pressing `h` triggers a specialized Help Modal that lists all keybinds. This overlay ensures users never feel lost without cluttering the main UI.
- **Progress Bar**: When downloading high-resolution portraits, a visual progress bar `[████░░░░] %` appears in the status line, providing real-time feedback for long-running tasks.

## 5. Graphics Compatibility
- **Sixel & Raw Fallback**: Support for iTerm2, WezTerm, and other Sixel-capable terminals. For non-Kitty terminals, the app uses `raw` logo type in Fastfetch to ensure maximum compatibility.
