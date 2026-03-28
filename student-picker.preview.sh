#!/usr/bin/env bash

img="$1"

[ -z "$img" ] && exit 1
[ ! -f "$img" ] && exit 1

dims=$(magick identify -format "%w %h" "$img" 2>/dev/null)
read -r w h <<< "$dims"
echo "File: $(basename "$img")"
echo "Size: ${w}x${h}"
echo ""

if command -v chafa >/dev/null 2>&1; then
    chafa --size=50x20 "$img" 2>/dev/null || kitty +kitten icat --scale-up "$img"
elif command -v img2txt >/dev/null 2>&1; then
    img2txt -W 50 -H 20 "$img"
elif command -v convert >/dev/null 2>&1; then
    convert "$img" -resize 100x40 -characters 80 jpg:- 2>/dev/null | strings | head -20
else
    echo "(Install chafa or img2txt for ASCII preview)"
fi
