# student-picker

TUI chọn ảnh học sinh cho fastfetch, render preview bằng Kitty Graphics Protocol.

## Cấu trúc thư mục

```
~/students/
├── students.txt     # danh sách (tùy chọn)
├── HS001.jpg
├── HS002.jpg
├── HS003.png
└── ...
```

### Format students.txt

```
# Dòng bắt đầu bằng # là comment
HS001|Nguyễn Văn A
HS002|Trần Thị B
HS003|Lê Văn C
```

Nếu không có file này, app tự scan thư mục và dùng mã số làm tên.

## Build & Cài đặt

```bash
git clone <repo>
cd student-picker
go mod tidy
go build -o student-picker .

# Cài vào PATH
sudo mv student-picker /usr/local/bin/
# hoặc
mv student-picker ~/.local/bin/
```

## Cấu hình

Sửa các hằng số đầu file `main.go` nếu cần:

```go
const (
    imageDir        = "$HOME/students"           // thư mục ảnh
    studentListFile = "$HOME/students/students.txt"
    fastfetchConfig = "$HOME/.config/fastfetch/config.jsonc"
    previewW        = 38  // chiều rộng preview (cell)
    previewH        = 22  // chiều cao preview (cell)
)
```

## Yêu cầu

- **Kitty terminal** (máy phụ CachyOS đã có ✓)
- Go ≥ 1.22
- fastfetch với config.jsonc có key `logo.source`

## Phím tắt

| Phím | Chức năng |
|------|-----------|
| ↑ / ↓ / k / j | Di chuyển danh sách |
| ← / → / h / l | Di chuyển ngang trong Grid |
| / | Tìm kiếm theo tên / mã số |
| Tab | Chuyển đổi Browse / Installed |
| h | Hiện bảng trợ giúp (Help Modal) |
| i | Xem thông tin chi tiết / Cài đặt |
| Enter | Chọn & lưu vào fastfetch config |
| q / Ctrl+C | Thoát không lưu |

## Tính năng nổi bật

- **Offline Mode**: Tự động lưu cache và mapping thông tin học sinh để dùng khi không có mạng.
- **2-Pane Layout**: Hiển thị song song Grid danh sách và Preview chân dung học sinh.
- **Progress Bar**: Theo dõi tiến trình tải ảnh chân dung thời gian thực.
- **Auto Fallback**: Tự động phát hiện terminal và chọn phương thức render ảnh (Kitty/Sixel/Raw) phù hợp nhất.

## Fastfetch config mẫu

```jsonc
{
  // config.jsonc
  "logo": {
    "source": "/home/user/students/HS001.jpg",
    "type": "kitty",
    "width": 30,
    "height": 15
  },
  "modules": [
    "title",
    "os",
    "kernel",
    "uptime"
  ]
}
```

## Lưu ý Kitty rendering

App dùng Kitty Graphics Protocol với `t=f` (file path transfer):
- Kitty đọc file ảnh trực tiếp — không cần base64 encode toàn bộ data
- Ảnh xóa tự động khi thoát (`a=d,d=A`)
- Nếu ảnh không hiển thị: kiểm tra `$TERM` = `xterm-kitty`
