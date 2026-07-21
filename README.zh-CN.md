# LAN Clipboard

<p align="center">
  一个用于在电脑与手机浏览器之间传输文字、图片和文件的轻量级局域网剪贴板。
</p>

<p align="center">
  <a href="README.md">English</a> ·
  <a href="README.zh-CN.md">简体中文</a> ·
  <a href="README.ja.md">日本語</a>
</p>

## 项目简介

LAN Clipboard 会把 macOS、Windows 或 Linux 电脑变成一个小型局域网 Web 服务。手机和平板无需安装 App、无需注册账号，只要与电脑连接同一个局域网，扫描二维码后使用浏览器即可访问。

手机发送的文字会自动复制到电脑系统剪贴板。图片和文件会保存在电脑本地，并通过 WebSocket 实时显示在历史记录中，无需刷新页面。

前端资源通过 `go:embed` 嵌入程序，正式发布时每个平台只需要一个可执行文件。运行数据默认保存在程序当前目录，也可以通过 `-data-dir` 指定其他目录。

## 主要功能

- 从手机或电脑浏览器发送文字、图片和任意文件
- 收到文字后自动复制到电脑剪贴板
- WebSocket 实时同步历史记录，无需刷新
- 支持拖拽上传和直接粘贴截图
- 支持图片缩略图、原图预览和文件下载
- 自动检测局域网 IPv4、打开浏览器并显示二维码
- 响应式页面、手机优先并自动跟随系统深色模式
- 持久化最近 100 条历史记录
- 默认启用可选密码保护
- 无账号、无云服务、无数据库、无需手机 App
- macOS、Windows、Linux 单文件运行

## 快速开始

### 下载 Release

1. 打开仓库的 **Releases** 页面。
2. 根据操作系统和 CPU 架构下载对应压缩包。
3. 解压后运行 `clipboard`，Windows 运行 `clipboard.exe`。
4. 保持终端窗口开启，终端会显示局域网地址、访问密码和二维码。
5. 手机连接同一个 Wi-Fi 后扫描二维码。

Release 文件名格式：

```text
lan-clipboard_<版本>_<操作系统>_<架构>
```

常见架构：

- `amd64`：大多数 Intel/AMD Windows、Linux 电脑和 Intel Mac
- `arm64`：Apple Silicon Mac 以及 ARM64 Windows/Linux 设备

### 从源码构建

环境要求：

- Go 1.24 或更高版本
- Git

```bash
git clone https://github.com/boss44944/LocalSend.git
cd LocalSend
go build -trimpath -o clipboard .
./clipboard
```

Windows PowerShell：

```powershell
go build -trimpath -o clipboard.exe .
.\clipboard.exe
```

首次启动会自动创建：

```text
uploads/
history.json
```

## 使用说明

程序启动后会自动：

1. 检测可用的局域网 IPv4 地址。
2. 默认监听 `0.0.0.0:8000`。
3. 在电脑上打开 `http://localhost:8000`。
4. 在终端打印手机可访问的地址和 ASCII 二维码。
5. 在未指定密码且未关闭认证时，生成一个随机 6 位密码。

在同一网络中的手机浏览器打开终端显示的局域网地址，输入终端中的密码后，即可发送文字或上传文件。

### 命令行参数

```text
-port 8000       HTTP 监听端口
-data-dir .      uploads/ 和 history.json 所在目录
-auth true       是否启用密码认证
-password ""     固定密码；留空时生成随机 6 位密码
-open true       启动后是否自动打开本机浏览器
-max-upload 512  单个上传文件最大大小，单位 MiB
```

使用示例：

```bash
# 在可信私有网络中关闭认证
./clipboard -auth=false

# 使用固定密码并修改端口
./clipboard -password=123456 -port=8080

# 把运行数据保存在独立目录
./clipboard -data-dir="$HOME/.lan-clipboard"

# 启动时不自动打开浏览器
./clipboard -open=false
```

## 剪贴板支持

| 平台 | 使用命令 | 额外要求 |
| --- | --- | --- |
| macOS | `pbcopy` | 无 |
| Windows | `clip` | 无 |
| Linux/X11 | `xclip` | 安装 `xclip` |
| Linux/Wayland | `wl-copy` | 安装 `wl-clipboard` |

Linux 没有安装以上命令时，文字和文件传输仍可正常使用，但收到文字后不会自动复制。

## macOS App 打包

仓库中的 `build.sh` 可以生成可双击运行的 `.app`：

```bash
chmod +x build.sh
./build.sh
```

默认构建 Apple Silicon 版本，输出位置：

```text
dist/LAN Clipboard.app
```

构建 Intel Mac 版本：

```bash
GOARCH=amd64 ./build.sh
```

生成的 App 没有进行 Apple 代码签名和公证。首次运行时，可能需要右键或按住 Control 点击应用，然后选择“打开”。

## 安全说明

为了让手机可以访问，LAN Clipboard 会监听所有网络接口。请仅在可信网络中使用。

- 默认启用密码认证。
- 不要通过路由器端口转发、公共反向代理等方式暴露到公网。
- 上传文件名会被清理，单个文件大小受限制。
- 登录 Cookie 和历史数据都只保存在本机。
- 上传文件会一直保留在 `uploads/` 中，需要使用者手动删除。
- 局域网通信使用 HTTP，不提供 HTTPS 传输加密。

在酒店、公司、共享 Wi-Fi 等网络中，请保持认证开启，并在使用完成后关闭程序。

## 数据与备份

以下运行数据不会提交到 Git：

```text
uploads/       上传的图片和文件
history.json   最近的历史记录
```

迁移数据时，请先关闭程序，然后同时复制这两个项目。

## GitHub Actions 与自动发布

仓库包含两个工作流：

- **CI**：在 Push 和 Pull Request 时执行格式检查、`go vet`、测试和编译。
- **Release**：为 macOS、Windows、Linux 的 `amd64` 和 `arm64` 架构构建压缩包，并发布到 GitHub Release。

发布新版本：

```bash
git tag v1.0.0
git push origin v1.0.0
```

推送符合 `v*` 格式的标签后，会自动触发发布流程。也可以进入 **Actions → Release → Run workflow**，输入例如 `v1.0.0` 的版本号手动发布。

## 常见问题

### 手机无法打开页面

- 确认手机和电脑连接的是同一个 Wi-Fi 或局域网。
- 如果手机优先使用移动网络，可以暂时关闭移动数据。
- 检查电脑防火墙是否允许 `clipboard` 访问网络。
- 检查 Wi-Fi 是否启用了访客隔离、客户端隔离或 AP Isolation。
- 不扫描二维码，直接手动输入终端中显示的 IP 地址尝试访问。

### Linux 收到文字后没有自动复制

安装以下任意工具：

```bash
# Debian/Ubuntu X11
sudo apt install xclip

# Debian/Ubuntu Wayland
sudo apt install wl-clipboard
```

### macOS 阻止运行下载的程序

可以在“系统设置 → 隐私与安全性”中允许该程序，或右键/Control 点击后选择“打开”。请只运行来自可信 Release 的文件。

## 项目结构

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

## 参与开发

欢迎提交 Issue 和 Pull Request。项目以轻量为原则：优先使用 Go 标准库，不引入前端框架，在项目目标没有发生明显变化前不引入数据库。

提交 PR 前建议执行：

```bash
gofmt -w main.go
go vet ./...
go test ./...
go build ./...
```

## 开源协议

LAN Clipboard 使用 [MIT License](LICENSE)。