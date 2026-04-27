#!/bin/bash
# Integration test script for student-picker image rendering in Kitty terminal
# Run this script inside Kitty terminal: ./test_kitty_render.sh

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BUILD_DIR="$SCRIPT_DIR"

echo "=== Student-Picker Kitty Integration Test ==="
echo ""

# Check if running in Kitty
if [ -z "$KITTY_WINDOW_ID" ] && [ "$TERM" != "xterm-kitty" ]; then
    echo "⚠️  WARNING: Not running in Kitty terminal"
    echo "   Current TERM: $TERM"
    echo "   Image rendering will fall back to Sixel or fail"
    echo ""
fi

# Build the application
echo "[1/4] Building student-picker..."
cd "$BUILD_DIR"
go build -o student-picker . || {
    echo "❌ Build failed"
    exit 1
}
echo "✓ Build successful"
echo ""

# Create test data directory
TEST_DIR=$(mktemp -d)
echo "[2/4] Creating test data in $TEST_DIR..."

# Create students.txt
cat > "$TEST_DIR/students.txt" << 'EOF'
# Test students list
TEST001|Nguyễn Văn A
TEST002|Trần Thị B
TEST003|Lê Văn C
TEST004|Phạm Minh D
TEST005|Hoàng Thị E
EOF

# Create placeholder images (simple colored rectangles using ImageMagick if available)
if command -v convert &> /dev/null; then
    for i in 1 2 3 4 5; do
        convert -size 100x150 xc:"$(shuf -e red blue green yellow purple -n1)" \
            -gravity center -pointsize 16 -fill white -annotate 0 "Student $i" \
            "$TEST_DIR/TEST00$i.jpg" 2>/dev/null || true
    done
    echo "✓ Created test images with ImageMagick"
else
    echo "⚠️  ImageMagick not installed - skipping test image creation"
    echo "   You can manually add images to: $TEST_DIR"
fi
echo ""

# Set environment for test
export STUDENT_PICKER_TEST_DIR="$TEST_DIR"
export STUDENT_PICKER_IMAGE_DIR="$TEST_DIR"
export STUDENT_PICKER_LIST_FILE="$TEST_DIR/students.txt"

echo "[3/4] Environment:"
echo "   TERM: $TERM"
echo "   KITTY_WINDOW_ID: ${KITTY_WINDOW_ID:-not set}"
echo "   Test dir: $TEST_DIR"
echo ""

echo "[4/4] Starting application..."
echo ""
echo "=== INSTRUCTIONS ==="
echo "1. Use Arrow keys or j/k to navigate"
echo "2. Press Enter to select a student"
echo "3. Press q to quit"
echo ""
echo "Press any key to start..."
read -n1 -s

# Run the application
"$BUILD_DIR/student-picker" 2>&1 || true

# Cleanup
echo ""
echo "=== Test Complete ==="
echo "Test data preserved at: $TEST_DIR"
echo "To clean up: rm -rf $TEST_DIR"