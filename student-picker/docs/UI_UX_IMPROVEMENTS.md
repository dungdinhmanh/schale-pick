# UI/UX Improvements: Student-Picker

## Design Principles
The refactored TUI follows a clean, modern aesthetic inspired by the Catppuccin color scheme, prioritizing legibility and visual hierarchy.

## 1. Grid & Navigation
- **Optimized Grid**: The grid size (columns and rows) is dynamically calculated based on terminal width, ensuring content is never truncated.
- **Master Box Layout**: The entire interface is contained within a unified Master Box, providing a consistent border and structure across different terminal sizes.
- **Visual Feedback**: Selection is clearly highlighted with a rounded border and a distinctive color (`#A6E3A1`).
- **Embedded Tabs**: The `Browse` and `Installed` tabs are integrated into the grid's top border. The `Installed` tab displays a dynamic count (e.g., `Installed (15)`).

## 2. Preview System
- **Dual-Pane View**: The split-screen layout allows users to see the list on the left and the large portrait on the right simultaneously.
- **Rich Content**: Displays student's full name and high-quality portraits from SchaleDB.
- **Vertical Alignment**: The student's name is moved to the absolute bottom of the preview panel for better visual balance.
- **Performance**: Native `termimg` rendering ensures no UI flicker during navigation.

## 3. Search Mode
- **Integrated Search**: Pressing `/` transforms the top border into a search input field (`Search: ...`). 
- **Real-time Filtering**: The list filters instantly as characters are typed, with empty state handling ("No students") when no matches are found.

## 4. Help & Progress Feedback
- **Help Modal**: Pressing `h` triggers a specialized Help Modal that lists all keybinds. This overlay ensures users never feel lost without cluttering the main UI.
- **Progress Bar**: When downloading high-resolution portraits, a visual progress bar `[████░░░░] %` appears in the status line, providing real-time feedback for long-running tasks.

## 5. Interactive Settings
- **Real Menu**: The previous static info screen is now a fully interactive menu (key `i`).
- **Live Configuration**: Users can navigate settings rows and toggle values (like Fastfetch Logo Format) in real-time.
- **Instant Persistence**: Changes are applied immediately to the session.

## 6. Graphics Compatibility
- **termimg Integration**: Universal image rendering support.
- **Sixel & Raw Fallback**: Support for iTerm2, WezTerm, and other Sixel-capable terminals. For non-Kitty terminals, the app automatically uses `raw` logo type in Fastfetch.
