# LAN Clipboard

<p align="center">
  A lightweight, private LAN clipboard for sharing text, images, and files between a computer and mobile browsers.
</p>

<p align="center">
  <a href="README.md">English</a> ·
  <a href="README.zh-CN.md">简体中文</a> ·
  <a href="README.ja.md">日本語</a>
</p>

## Overview

LAN Clipboard turns a macOS, Windows, or Linux computer into a small local web server. Phones and tablets do not need an application or account: connect them to the same local network, scan the QR code, and open the page in a browser.

Text sent from the browser is copied automatically to the computer clipboard. Images and files are saved locally and appear immediately in the shared history through WebSocket updates.

The frontend is embedded with `go:embed`, so the release consists of a single executable. Runtime data is stored beside the executable, or in a directory selected with `-data-dir`.

## Features

- Send text, images, and arbitrary files from a mobile or desktop browser
- Automatically copy received text to the host clipboard
- Live history updates over WebSocket without refreshing the page
- Drag-and-drop uploads and pasted screenshots
- Image thumbnails, full-size previews, and file downloads
- Automatic private IPv4 detection, browser launch, and QR-code display
- Embedded responsive interface with automatic dark mode
- Persistent history limited to the latest 100 records
- Optional password protection, enabled by default
- No account, cloud service, database, or mobile application
- Single-file binaries for macOS, Windows, and Linux

## Quick start

### Download a release

1. Open the repository's **Releases** page.
2. Download the archive for your operating system and CPU architecture.
3. Extract it and run `clipboard` (`clipboard.exe` on Windows).
4. Keep the terminal window open. It displays the LAN address, access password, and QR code.
5. Connect the phone to the same Wi-Fi network and scan the QR code.

Release archive names use the following format:

```text
lan-clipboard_<version>_<os>_<architecture>
```

Common architectures:

- `amd64`: most Intel/AMD Windows and Linux computers, and Intel Macs
- `arm64`: Apple Silicon Macs and ARM64 Linux/Windows devices

### Build from source

Requirements:

- Go 1.24 or newer
- Git

```bash
git clone https://github.com/boss44944/LocalSend.git
cd LocalSend
go build -trimpath -o clipboard .
./clipboard
```

On Windows PowerShell:

```powershell
go build -trimpath -o clipboard.exe .
.\clipboard.exe
```

At first startup, LAN Clipboard creates:

```text
uploads/
history.json
```

## Usage

When the program starts, it:

1. Detects a private LAN IPv4 address.
2. Listens on `0.0.0.0:8000` by default.
3. Opens `http://localhost:8000` on the host computer.
4. Prints the phone-accessible URL and an ASCII QR code.
5. Generates a random six-digit password unless a password was supplied or authentication was disabled.

Open the displayed LAN URL from another device on the same network. Enter the terminal password, then use the page to send text or upload files.

### Command-line options

```text
-port 8000       HTTP listening port
-data-dir .      Directory containing uploads/ and history.json
-auth true       Enable password authentication
-password ""     Fixed password; empty generates a random six-digit password
-open true       Open the local browser after startup
-max-upload 512  Maximum size of one uploaded file in MiB
```

Examples:

```bash
# Disable authentication on a trusted private network
./clipboard -auth=false

# Use a fixed password and another port
./clipboard -password=123456 -port=8080

# Keep runtime data in a dedicated directory
./clipboard -data-dir="$HOME/.lan-clipboard"

# Do not open a browser automatically
./clipboard -open=false
```

## Clipboard support

| Platform | Command used | Additional setup |
| --- | --- | --- |
| macOS | `pbcopy` | None |
| Windows | `clip` | None |
| Linux/X11 | `xclip` | Install `xclip` |
| Linux/Wayland | `wl-copy` | Install `wl-clipboard` |

If no supported Linux clipboard command is installed, sharing and uploads still work; only automatic text copying is unavailable.

## macOS application bundle

The repository includes `build.sh`, which creates a double-clickable `.app` bundle:

```bash
chmod +x build.sh
./build.sh
```

The default target is Apple Silicon and the output is:

```text
dist/LAN Clipboard.app
```

Build an Intel version with:

```bash
GOARCH=amd64 ./build.sh
```

The generated application is not code-signed or notarized. On first launch, macOS may require using **Control-click → Open**.

## Security notes

LAN Clipboard listens on all network interfaces so phones can reach it. Use it only on networks you trust.

- Password authentication is enabled by default.
- Do not expose the port through router forwarding, a public reverse proxy, or the public internet.
- Uploaded filenames are sanitized and individual upload sizes are limited.
- Authentication cookies and history are local to the host computer.
- Uploaded files remain in `uploads/` until manually deleted.
- LAN Clipboard does not provide transport encryption. Traffic on the LAN uses HTTP, not HTTPS.

For hotel, office, or shared Wi-Fi, keep authentication enabled and close the program when finished.

## Data and backup

Runtime files are intentionally excluded from Git:

```text
uploads/       Uploaded images and files
history.json   Latest history records
```

To move the data to another computer, stop LAN Clipboard and copy these two items together.

## GitHub Actions and releases

Two workflows are included:

- **CI** runs formatting checks, `go vet`, tests, and builds on pushes and pull requests.
- **Release** builds archives for macOS, Windows, and Linux on `amd64` and `arm64`, then publishes them to a GitHub Release.

To publish a version:

```bash
git tag v1.0.0
git push origin v1.0.0
```

A tag matching `v*` starts the release workflow automatically. The workflow can also be started manually from **Actions → Release → Run workflow** by entering a version such as `v1.0.0`.

## Troubleshooting

### The phone cannot open the page

- Confirm both devices are connected to the same Wi-Fi or LAN.
- Disable cellular data temporarily if the phone prefers it over Wi-Fi.
- Allow `clipboard` through the host firewall.
- Check whether the Wi-Fi enables client isolation, guest isolation, or AP isolation.
- Try the printed IP address manually instead of scanning the QR code.

### Text is not copied on Linux

Install one of the supported tools:

```bash
# Debian/Ubuntu X11
sudo apt install xclip

# Debian/Ubuntu Wayland
sudo apt install wl-clipboard
```

### macOS blocks the downloaded binary

Open **System Settings → Privacy & Security** and allow the application, or use **Control-click → Open**. Only download binaries from releases you trust.

## Project structure

```text
.
├── main.go
├── go.mod
├── go.sum
├── build.sh
├── README.md
├── README.zh-CN.md
├── README.ja.md
├── LICENSE
├── .github/workflows/
│   ├── ci.yml
│   └── release.yml
└── static/
    ├── index.html
    ├── style.css
    ├── script.js
    └── favicon.ico
```

## Contributing

Issues and pull requests are welcome. Keep the project lightweight: prefer the Go standard library, avoid frontend frameworks, and do not introduce a database unless the project's scope changes substantially.

Before submitting a pull request, run:

```bash
gofmt -w main.go
go vet ./...
go test ./...
go build ./...
```

## License

LAN Clipboard is released under the [MIT License](LICENSE).