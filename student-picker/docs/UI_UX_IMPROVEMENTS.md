# UI/UX Improvements: Student-Picker

## Design Principles
The refactored TUI follows a clean, modern aesthetic inspired by the Catppuccin color scheme, prioritizing legibility and visual hierarchy.

## 1. Grid & Navigation
- **Optimized Grid**: The grid size (columns and rows) is dynamically calculated based on terminal width, ensuring content is never truncated accidentally.
- **Visual Feedback**: Selection is clearly highlighted with a rounded border and a distinctive color (`#A6E3A1`), providing an immediate sense of focus.
- **Embedded Tabs**: The `Browse` and `Installed` tabs are integrated into the grid's top border. This minimizes vertical space usage and results in a more cohesive "application" feel compared to floating labels.

## 2. Preview System
- **Dual-Pane View**: The split-screen layout allows users to see the list on the left and the large portrait on the right simultaneously.
- **Rich Content**: The preview displays the student's full name (Family Name + Personal Name) and high-quality portraits from the SchaleDB archive.
- **Smooth Updates**: Portraits update instantly as the user navigates the grid, with background caching ensuring no UI flickers.

## 3. Search Mode
- **Integrated Search**: Pressing `/` transforms the top border into a search input field (`Search: ...`). 
- **Real-time Filtering**: The list filters instantly as characters are typed, with empty state handling ("No students") when no matches are found.

## 4. Graphics Compatibility
- **Sixel Support**: Users on iTerm2, WezTerm, or other Sixel-capable terminals can now see graphics even if they don't support the Kitty graphics protocol.
- **Automatic Detection**: The app transparently detects the best possible rendering method on startup.
