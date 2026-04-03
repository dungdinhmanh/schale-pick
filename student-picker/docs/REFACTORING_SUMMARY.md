# Refactoring Summary: Student-Picker TUI

## Overview
This document summarizes the changes made to the `student-picker` project to improve maintainability, performance, and user experience.

## 1. Modularization (Architecture)
The original `model.go` (a single file with over 1100 lines) was refactored into multiple smaller, focused files to comply with Bubbletea best practices and maintain readability.

- **model.go**: Core application logic, initialization, and update loops.
- **view.go**: UI rendering logic using Lipgloss, including the new grid and preview layout.
- **styles.go**: Centralized design system with color palettes and styling tokens.
- **update_keyboard.go**: Specialized keyboard event handling.
- **terminal.go**: Terminal diagnostics and capability detection.
- **renderer.go**: Image rendering management with Sixel/Kitty fallback.
- **downloader.go**: Asynchronous image fetching with concurrency control.
- **url.go**: Standardized URL generation for student assets.
- **config.go**: Centralized configuration and path management.

## 2. Critical Bug Fixes
- **UI Blocking**: Chuyển `preCachePortraits` sang chạy bất đồng bộ bằng Goroutines, loại bỏ tình trạng treo ứng dụng khi mới khởi động hoặc chuyển danh sách.
- **Crash/Panic Handling**: Thêm cơ chế `recover()` cho các tác vụ vẽ ảnh đồ họa, đảm bảo ứng dụng không bị panic khi chạy trong môi trường pseudo-TTY không hỗ trợ hoặc lỗi hệ thống.
- **Event Handling**: Sửa lỗi bỏ sót `showModalMsg`, giúp các hộp thoại thiết bị và thông báo lỗi hiển thị chính xác.
- **Path Portability**: Thay thế các đường dẫn cứng (hardcoded) bằng việc sử dụng `os.UserHomeDir()`, cho phép ứng dụng chạy trên bất kỳ hệ máy nào.

## 3. UI/UX Enhancements
- **Dual-Pane Layout**: Hệ thống layout 2 cột (Box Grid bên trái, Box Preview bên phải) chuyên nghiệp hơn.
- **Tabbed Box**: Các nhãn tab `Browse` / `Installed` được nhúng trực tiếp lên đường viền trên của box chứa Grid.
- **Sixel Fallback**: Hỗ trợ hiển thị ảnh qua chuẩn Sixel cho các terminal cũ hoặc iTerm2, WezTerm thay vì chỉ hỗ trợ Kitty Graphics.
- **Portrait Integration**: Tự động tải và hiển thị ảnh Portrait (chân dung) lớn ở cột bên phải khi chọn học sinh.

## 4. Maintenance
- Tích hợp `tui-devtools` cho việc debug và kiểm tra giao diện trực quan.
- Tối ưu hóa dung lượng code (xóa tool check Magick dư thừa, tinh gọn CSS-in-Go logic).

- **Enhanced Reliability**: Automatic fallback to `raw` logo type for compatibility across all Sixel-capable and standard terminals.

## 6. V6: Master Box Layout & High-Performance Rendering (Latest)
- **Master Box**: Unified UI container for Grid, Preview, Status, and Help.
- **termimg Native Driver**: Switched from external `icat` to native Go `termimg` for rendering 10-30 images simultaneously without performance degradation.
- **Responsive Grid**: Recalculated column logic to fit terminal resizing perfectly.
- **Interactive Settings**: Transformed static info screen into a navigable settings menu.
- **UI Alignment**: Fixed top/bottom bar offsets and right-side overflow issues.
