#!/usr/bin/env bash

IMAGE_DIR="${HOME}/students"
CONFIG_FILE="${HOME}/.config/fastfetch/config.jsonc"
BACKUP_SUFFIX=".bak"

[ ! -d "$IMAGE_DIR" ] && echo "Error: $IMAGE_DIR not found" && exit 1

get_image() {
    local id="$1"
    for ext in webp png jpg jpeg; do
        [ -f "$IMAGE_DIR/${id}.${ext}" ] && echo "$IMAGE_DIR/${id}.${ext}" && return
    done
}

list_students() {
    if [ -f "$IMAGE_DIR/students.txt" ]; then
        while IFS='|' read -r id name; do
            [[ "$id" =~ ^# ]] && continue
            [ -z "$id" ] && continue
            img=$(get_image "$id")
            [ -n "$img" ] && echo -e "$id\t$name\t$img"
        done < "$IMAGE_DIR/students.txt"
    else
        for img in "$IMAGE_DIR"/*.{webp,png,jpg,jpeg}; do
            [ -f "$img" ] || continue
            id=$(basename "$img" | sed 's/\.[^.]*$//')
            echo -e "$id\t$id\t$img"
        done
    fi
}

preview_image() {
    local img="$1"
    [ -z "$img" ] || [ ! -f "$img" ] && return
    
    local dim="${FZF_PREVIEW_COLUMNS}x${FZF_PREVIEW_LINES}"
    [ "$dim" = "x" ] && dim="60x25"
    
    if [ -n "$KITTY_WINDOW_ID" ] && command -v kitten >/dev/null 2>&1; then
        kitten icat --clear --transfer-mode=memory --place="${dim}@0x0" "$img" 2>/dev/null
    elif command -v chafa >/dev/null 2>&1; then
        chafa -s "$dim" "$img" 2>/dev/null
    else
        dims=$(magick identify -format "%w x %h" "$img" 2>/dev/null)
        echo ""
        echo "  [Preview]"
        echo "  $(basename "$img")"
        echo "  $dims"
        echo ""
        echo "  (Install chafa or use Kitty for image preview)"
    fi
}

export -f preview_image

selected=$(list_students | fzf \
    --delimiter=$'\t' \
    --with-nth=2 \
    --prompt="Select: " \
    --preview="preview_image {3}" \
    --preview-window="right:50%,border-sharp" \
    --height=50% \
    --margin=1 \
    --border=rounded \
    --ansi)

[ -z "$selected" ] && exit 0

IFS=$'\t' read -r id name img <<< "$selected"

echo ""
echo "ID:    $id"
echo "Name:  $name"
echo "Image: $img"
echo ""

update_config() {
    [ ! -f "$CONFIG_FILE" ] && echo "Error: $CONFIG_FILE not found" && return 1
    [ -f "$CONFIG_FILE$BACKUP_SUFFIX" ] || cp "$CONFIG_FILE" "$CONFIG_FILE$BACKUP_SUFFIX"
    
    read -r w h < <(magick identify -format "%w %h" "$img" 2>/dev/null)
    [ -z "$w" ] || [ -z "$h" ] && { w=30; h=20; }
    
    aspect=$(awk "BEGIN {printf \"%.4f\", $h / $w}")
    
    if awk "BEGIN {exit !($aspect >= 1.0)}"; then
        new_w=45
    else
        new_w=35
    fi
    
    sed -i "s|\"source\": \".*\"|\"source\": \"$img\"|" "$CONFIG_FILE"
    sed -i "s|\"type\": \"auto\"|\"type\": \"kitty\"|" "$CONFIG_FILE"
    sed -i "s|\"width\": null|\"width\": $new_w|" "$CONFIG_FILE"
    sed -i "s|\"height\": null|\"height\": null|" "$CONFIG_FILE"
    sed -i "s|\"preserveAspectRatio\": false|\"preserveAspectRatio\": true|" "$CONFIG_FILE"
    sed -i "s|\"top\": 0|\"top\": 2|" "$CONFIG_FILE"
    sed -i "s|\"left\": 0|\"left\": 2|" "$CONFIG_FILE"
    sed -i "s|\"right\": 4|\"right\": 2|" "$CONFIG_FILE"
    
    echo "Updated: width=$new_w (height auto)"
    echo "Original: ${w}x${h} (aspect: ${aspect})"
}

PS3="Choice? "
select opt in "Update fastfetch" "View config" "Cancel"; do
    case $opt in
        "Update fastfetch") update_config; break ;;
        "View config") [ -f "$CONFIG_FILE" ] && cat "$CONFIG_FILE"; break ;;
        *) echo "Cancelled"; break ;;
    esac
done