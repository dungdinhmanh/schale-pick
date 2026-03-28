---
session: ses_2cca
updated: 2026-03-28T07:37:52.139Z
---

# Session Summary

## Goal
Explore the Kitty source code at ~/Downloads/kitty to thoroughly understand the Kitty Graphics Protocol implementation, finding exact file paths and line numbers for: x=/y= pixel position parsing, image placement/rendering, composition mode (C=1), pixel vs cell coordinate conversion, and action types (a=T/t/p).

## Constraints & Preferences
- Use grep with pattern matching and read relevant files to understand implementation
- Return exact file paths and line numbers for all relevant code sections
- Focus on C code files (graphics.c, screen.c, shaders.c, parse-graphics-command.h)

## Progress
### Done
- [x] Explored Kitty source directory structure at ~/Downloads/kitty/kitty/
- [x] Found GraphicsCommand struct definition in `graphics.h` (lines 11-26)
- [x] Located command parsing in `parse-graphics-command.h` (lines 1-390)
- [x] Found `screen_handle_graphics_command()` in `screen.c` line 1533
- [x] Found `grman_handle_command()` in `graphics.c` line 2195 - handles action routing
- [x] Found `handle_add_command()` in `graphics.c` lines 712-785 - handles a=T, a=t, a=q
- [x] Found `handle_put_command()` in `graphics.c` lines 1063-1153 - handles a=p
- [x] Found `handle_delete_command()` in `graphics.c` lines 2093-2161 - handles a=d
- [x] Found `grman_update_layers()` in `graphics.c` lines 1206-1312 - cell-to-pixel conversion
- [x] Found z-index classification in `graphics.c` lines 1267-1272
- [x] Found rendering order in `shaders.c` lines 1142-1157
- [x] Reviewed full protocol documentation in `docs/graphics-protocol.rst` (lines 1-1099)

### In Progress
- (none - research phase complete)

### Blocked
- (none)

## Key Decisions
- **x=/y= are source coordinates, not screen position**: Confirmed in `handle_put_command()` (graphics.c:1120) where `ref->src_x = g->x_offset; ref->src_y = g->y_offset;`
- **Screen placement uses cursor position**: `start_row = c->y; start_column = c->x` (graphics.c:1124) plus cell_x_offset/cell_y_offset pixel offsets
- **C=1 for frame composition disables alpha blending**: Line 1595 in graphics.c shows `.alpha_blend = g->compose_mode != 1 && !load_data->is_opaque`
- **Z-index layering**: z < INT32_MIN/2 → under cells; z < 0 → under text; z >= 0 → over text

## Next Steps
1. The research phase is complete - all 5 areas have been thoroughly explored
2. The implementation details are ready to be used for any specific task the user wants to accomplish

## Critical Context
- **Key insight on x=/y=**: The protocol spec says x=/y= are "source rectangle" coordinates (pixels) for which part of the image to display, NOT screen positioning. Screen positioning comes from cursor position in cells plus X=/Y= (cell offsets in pixels)
- **C parameter confusion**: The `C` parameter in `cursor_movement` (graphics.h:15) controls cursor movement during placement (C=1 = don't move cursor). The `compose_mode` union field controls alpha blending during frame composition
- **Action routing**: All graphics commands flow through `grman_handle_command()` which routes to specific handlers based on the `action` field
- **Z-index special value**: INT32_MIN/2 (-1,073,741,824) is a threshold for being drawn even under cells with custom background colors

## File Operations
### Read
- `/home/kazukisatou/Downloads/kitty/kitty/parse-graphics-command.h` (390 lines) - Command parsing, x_offset='x', y_offset='y' definitions
- `/home/kazukisatou/Downloads/kitty/kitty/graphics.h` (222 lines) - GraphicsCommand struct, ImageRef struct, ImageRenderData
- `/home/kazukisatou/Downloads/kitty/kitty/graphics.c` (2481 lines) - Main implementation: handle_add_command (712), handle_put_command (1063), grman_update_layers (1206), handle_delete_command (2093), grman_handle_command (2195)
- `/home/kazukisatou/Downloads/kitty/kitty/screen.c` (6141 lines) - screen_handle_graphics_command (1533)
- `/home/kazukisatou/Downloads/kitty/kitty/shaders.c` (1710 lines) - Rendering order with layers (1142-1157)
- `/home/kazukisatou/Downloads/kitty/docs/graphics-protocol.rst` (1143 lines) - Full protocol specification

### Modified
- (none) - This was a read-only research session
