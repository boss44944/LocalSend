# LAN Clipboard

LAN Clipboard 是一个轻量的局域网剪贴板。电脑运行一个 Go 可执行文件，手机无需安装 App，只需扫码打开浏览器，即可发送文本、图片和任意文件。收到文本后，程序会自动写入电脑系统剪贴板。

## 功能

- 自动检测局域网 IPv4、打开浏览器并输出终端二维码
- WebSocket 实时同步最近 100 条历史记录
- macOS `pbcopy`、Windows `clip`、Linux `xclip`/`wl-copy`
- 拖拽上传、粘贴截图、图片预览和文件下载
- `go:embed` 嵌入全部静态资源，最终只有一个可执行文件
- 默认生成 6 位访问密码，可通过 Cookie 保持登录
- 响应式 GitHub 风格界面与系统深色模式
- 自动创建 `uploads/` 和 `history.json`

## 构建

需要 Go 1.24 或更高版本：

```bash
go build -o clipboard .
./clipboard
```

启动后会显示局域网地址、访问密码和二维码，并打开 `http://localhost:8000`。

## 参数

```text
-port 8000       HTTP 端口
-data-dir .      uploads 与 history.json 的目录
-auth true       是否启用访问密码
-password ""     固定密码；留空时自动生成
-open true       启动时打开浏览器
-max-upload 512  单文件最大 MiB
```

关闭认证：

```bash
./clipboard -auth=false
```

## macOS App

```bash
chmod +x build.sh
./build.sh
```

默认生成 Apple Silicon 版本 `dist/LAN Clipboard.app`。Intel 版本：

```bash
GOARCH=amd64 ./build.sh
```

## Linux 剪贴板

Linux 需要安装 `xclip` 或 `wl-clipboard`，否则其他功能仍可正常使用，但收到文本时不会自动复制。

## 安全说明

服务默认监听所有网卡，仅建议在可信局域网使用。默认密码可降低酒店或共享网络中的误访问风险，但不要将端口映射到公网。上传文件名会被清洗，上传大小受限制，会话在程序重启后自动失效。

## 目录

```text
.
├── main.go
├── go.mod
├── go.sum
├── build.sh
├── README.md
└── static/
    ├── index.html
    ├── style.css
    └── script.js
```

运行时数据 `uploads/` 和 `history.json` 不提交到 Git。

## License

MIT
