#!/usr/bin/env bash
set -euo pipefail
mkdir -p work/icon.iconset
python3 - <<'PY'
import base64
raw=open('assets/app-icon.png.b64','rb').read()
open('work/icon.png','wb').write(base64.b64decode(raw))
PY
brew list imagemagick >/dev/null 2>&1 || brew install imagemagick
magick work/icon.png -define icon:auto-resize=256,128,64,48,32,16 work/app.ico
for n in 16 32 128 256 512; do
  sips -z "$n" "$n" work/icon.png --out "work/icon.iconset/icon_${n}x${n}.png" >/dev/null
  d=$((n*2))
  sips -z "$d" "$d" work/icon.png --out "work/icon.iconset/icon_${n}x${n}@2x.png" >/dev/null
done
iconutil -c icns work/icon.iconset -o work/AppIcon.icns
go install github.com/akavel/rsrc@latest
