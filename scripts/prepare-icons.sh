#!/usr/bin/env bash
set -euo pipefail

ICON_SOURCE="assets/app-icon.png"
if [ ! -f "$ICON_SOURCE" ]; then
  echo "Missing $ICON_SOURCE. Upload the selected square PNG icon to this path." >&2
  exit 1
fi

mkdir -p work/icon.iconset
brew list imagemagick >/dev/null 2>&1 || brew install imagemagick
magick "$ICON_SOURCE" -alpha on -define icon:auto-resize=256,128,64,48,32,16 work/app.ico

for n in 16 32 128 256 512; do
  sips -z "$n" "$n" "$ICON_SOURCE" --out "work/icon.iconset/icon_${n}x${n}.png" >/dev/null
  d=$((n*2))
  sips -z "$d" "$d" "$ICON_SOURCE" --out "work/icon.iconset/icon_${n}x${n}@2x.png" >/dev/null
done

iconutil -c icns work/icon.iconset -o work/AppIcon.icns
go install github.com/akavel/rsrc@latest
