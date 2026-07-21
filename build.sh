#!/bin/sh
set -eu
APP="LAN Clipboard"
BUNDLE="dist/${APP}.app"
BIN="clipboard"
rm -rf dist
mkdir -p "${BUNDLE}/Contents/MacOS" "${BUNDLE}/Contents/Resources"
CGO_ENABLED=0 GOOS=darwin GOARCH="${GOARCH:-arm64}" go build -trimpath -ldflags="-s -w" -o "${BUNDLE}/Contents/MacOS/${BIN}" .
cat > "${BUNDLE}/Contents/Info.plist" <<PLIST
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0"><dict>
<key>CFBundleName</key><string>${APP}</string>
<key>CFBundleDisplayName</key><string>${APP}</string>
<key>CFBundleIdentifier</key><string>com.local.lanclipboard</string>
<key>CFBundleVersion</key><string>1.0.0</string>
<key>CFBundleShortVersionString</key><string>1.0.0</string>
<key>CFBundleExecutable</key><string>${BIN}</string>
<key>CFBundlePackageType</key><string>APPL</string>
<key>LSMinimumSystemVersion</key><string>11.0</string>
</dict></plist>
PLIST
printf 'Created %s\n' "${BUNDLE}"
