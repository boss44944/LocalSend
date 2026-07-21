#!/usr/bin/env bash
set -euo pipefail
VERSION=${1:?version required}
rm -rf dist work/LAN\ Clipboard.app
mkdir -p dist
make_mac() {
  arch=$1; label=$2; app='work/LAN Clipboard.app'
  rm -rf "$app"; mkdir -p "$app/Contents/MacOS" "$app/Contents/Resources"
  GOOS=darwin GOARCH="$arch" CGO_ENABLED=0 go build -trimpath -ldflags='-s -w' -o "$app/Contents/MacOS/clipboard" .
  cp work/AppIcon.icns "$app/Contents/Resources/AppIcon.icns"
  cat > "$app/Contents/Info.plist" <<EOF
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0"><dict>
<key>CFBundleName</key><string>LAN Clipboard</string>
<key>CFBundleIdentifier</key><string>com.localsend.lanclipboard</string>
<key>CFBundleExecutable</key><string>clipboard</string>
<key>CFBundleIconFile</key><string>AppIcon</string>
<key>CFBundlePackageType</key><string>APPL</string>
<key>CFBundleShortVersionString</key><string>${VERSION#v}</string>
<key>CFBundleVersion</key><string>${VERSION#v}</string>
<key>LSMinimumSystemVersion</key><string>11.0</string>
</dict></plist>
EOF
  chmod +x "$app/Contents/MacOS/clipboard"
  codesign --force --deep --sign - "$app"
  ditto -c -k --sequesterRsrc --keepParent "$app" "dist/lan-clipboard-${VERSION}-macos-${label}.zip"
}
make_mac amd64 intel
make_mac arm64 apple-silicon
"$(go env GOPATH)/bin/rsrc" -ico work/app.ico -o rsrc_windows_amd64.syso
GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -trimpath -ldflags='-s -w -H=windowsgui' -o work/LAN-Clipboard.exe .
rm rsrc_windows_amd64.syso
(cd work && zip -q "../dist/lan-clipboard-${VERSION}-windows-amd64.zip" LAN-Clipboard.exe)
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -trimpath -ldflags='-s -w' -o work/lan-clipboard .
tar -C work -czf "dist/lan-clipboard-${VERSION}-linux-amd64.tar.gz" lan-clipboard
(cd dist && shasum -a 256 * > SHA256SUMS.txt)
