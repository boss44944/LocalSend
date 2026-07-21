# LAN Clipboard

<p align="center">
  PC とモバイルブラウザの間でテキスト、画像、ファイルを共有する軽量な LAN クリップボードです。
</p>

<p align="center">
  <a href="README.md">English</a> ·
  <a href="README.zh-CN.md">简体中文</a> ·
  <a href="README.ja.md">日本語</a>
</p>

## 概要

LAN Clipboard は macOS、Windows、Linux の PC を小さなローカル Web サーバーとして動作させます。スマートフォンやタブレットにはアプリのインストールもアカウント登録も不要です。同じ LAN に接続し、QR コードを読み取ってブラウザで開くだけで利用できます。

ブラウザから送信したテキストは PC のシステムクリップボードに自動コピーされます。画像やファイルはローカルに保存され、WebSocket を通じて履歴へリアルタイムに追加されます。

フロントエンドは `go:embed` で実行ファイルに埋め込まれるため、配布物は基本的に 1 つの実行ファイルだけです。実行時データは実行ディレクトリ、または `-data-dir` で指定した場所に保存されます。

## 主な機能

- モバイルまたは PC ブラウザからテキスト、画像、任意ファイルを送信
- 受信したテキストをホスト PC のクリップボードへ自動コピー
- WebSocket によるリアルタイム履歴更新
- ドラッグ＆ドロップアップロードとスクリーンショット貼り付け
- 画像サムネイル、原寸表示、ファイルダウンロード
- LAN IPv4 の自動検出、ブラウザ起動、QR コード表示
- レスポンシブ UI とシステム連動ダークモード
- 最新 100 件の履歴を永続化
- 既定で有効な任意パスワード認証
- アカウント、クラウド、データベース、モバイルアプリ不要
- macOS、Windows、Linux 向け単一バイナリ

## クイックスタート

### Release からダウンロード

1. リポジトリの **Releases** ページを開きます。
2. OS と CPU アーキテクチャに合うアーカイブをダウンロードします。
3. 展開し、`clipboard` を実行します。Windows では `clipboard.exe` です。
4. ターミナルを開いたままにします。LAN アドレス、アクセスパスワード、QR コードが表示されます。
5. スマートフォンを同じ Wi-Fi に接続し、QR コードを読み取ります。

Release アーカイブ名：

```text
lan-clipboard_<version>_<os>_<architecture>
```

一般的なアーキテクチャ：

- `amd64`: 多くの Intel/AMD Windows・Linux PC、および Intel Mac
- `arm64`: Apple Silicon Mac、ARM64 Linux/Windows デバイス

### ソースからビルド

必要環境：

- Go 1.24 以降
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

初回起動時に次のデータが自動作成されます。

```text
uploads/
history.json
```

## 使い方

起動すると次の処理を自動で行います。

1. 利用可能な LAN IPv4 アドレスを検出します。
2. 既定で `0.0.0.0:8000` をリッスンします。
3. ホスト PC で `http://localhost:8000` を開きます。
4. スマートフォンからアクセスできる URL と ASCII QR コードを表示します。
5. パスワード未指定かつ認証有効の場合、6 桁のランダムパスワードを生成します。

同じネットワークの端末で表示された LAN URL を開き、ターミナルに表示されたパスワードを入力してください。その後、テキスト送信やファイルアップロードを利用できます。

### コマンドラインオプション

```text
-port 8000       HTTP ポート
-data-dir .      uploads/ と history.json を保存するディレクトリ
-auth true       パスワード認証を有効にする
-password ""     固定パスワード。空の場合は 6 桁を自動生成
-open true       起動後にローカルブラウザを開く
-max-upload 512  1 ファイルあたりの最大サイズ（MiB）
```

例：

```bash
# 信頼できるプライベートネットワークで認証を無効化
./clipboard -auth=false

# 固定パスワードと別ポートを使用
./clipboard -password=123456 -port=8080

# 実行データを専用ディレクトリへ保存
./clipboard -data-dir="$HOME/.lan-clipboard"

# ブラウザを自動起動しない
./clipboard -open=false
```

## クリップボード対応

| プラットフォーム | 使用コマンド | 追加設定 |
| --- | --- | --- |
| macOS | `pbcopy` | 不要 |
| Windows | `clip` | 不要 |
| Linux/X11 | `xclip` | `xclip` をインストール |
| Linux/Wayland | `wl-copy` | `wl-clipboard` をインストール |

Linux で対応コマンドがない場合でも、共有やアップロードは利用できます。自動コピーだけが無効になります。

## macOS App バンドル

`build.sh` でダブルクリック可能な `.app` を作成できます。

```bash
chmod +x build.sh
./build.sh
```

既定では Apple Silicon 向けで、出力先は次の通りです。

```text
dist/LAN Clipboard.app
```

Intel Mac 向け：

```bash
GOARCH=amd64 ./build.sh
```

生成された App はコード署名・公証されていません。初回起動時は **Control キーを押しながらクリック → 開く** が必要になる場合があります。

## セキュリティ

スマートフォンからアクセスできるよう、LAN Clipboard はすべてのネットワークインターフェースをリッスンします。信頼できるネットワークでのみ使用してください。

- パスワード認証は既定で有効です。
- ルーターのポート転送、公開リバースプロキシ、インターネット公開は行わないでください。
- アップロードファイル名は無害化され、1 ファイルのサイズも制限されます。
- 認証 Cookie と履歴データはホスト PC 内に保存されます。
- アップロード済みファイルは手動削除するまで `uploads/` に残ります。
- LAN 通信は HTTP であり、HTTPS 暗号化は提供しません。

ホテル、会社、共有 Wi-Fi では認証を有効のままにし、利用後はプログラムを終了してください。

## データとバックアップ

次の実行データは Git にコミットされません。

```text
uploads/       アップロード済み画像・ファイル
history.json   最新の履歴
```

別 PC へ移行する場合はプログラムを終了し、この 2 つを一緒にコピーしてください。

## GitHub Actions とリリース

リポジトリには 2 つのワークフローがあります。

- **CI**: Push と Pull Request でフォーマット確認、`go vet`、テスト、ビルドを実行します。
- **Release**: macOS、Windows、Linux の `amd64` と `arm64` をビルドし、GitHub Release に公開します。

新しいバージョンを公開するには：

```bash
git tag v1.0.0
git push origin v1.0.0
```

`v*` 形式のタグを Push すると自動で Release ワークフローが起動します。**Actions → Release → Run workflow** から `v1.0.0` のようなバージョンを入力して手動実行することもできます。

## トラブルシューティング

### スマートフォンからページを開けない

- PC とスマートフォンが同じ Wi-Fi または LAN に接続されているか確認します。
- スマートフォンがモバイル回線を優先する場合は、一時的にモバイルデータを無効にします。
- ホストのファイアウォールで `clipboard` を許可します。
- Wi-Fi のクライアント分離、ゲスト分離、AP Isolation を確認します。
- QR コードではなく、表示された IP アドレスを直接入力します。

### Linux でテキストが自動コピーされない

対応ツールをインストールしてください。

```bash
# Debian/Ubuntu X11
sudo apt install xclip

# Debian/Ubuntu Wayland
sudo apt install wl-clipboard
```

### macOS がダウンロードしたバイナリをブロックする

**システム設定 → プライバシーとセキュリティ** から許可するか、Control キーを押しながらクリックして「開く」を選択してください。信頼できる Release のみ利用してください。

## プロジェクト構成

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

## コントリビュート

Issue と Pull Request を歓迎します。軽量性を維持するため、Go 標準ライブラリを優先し、フロントエンドフレームワークを避け、プロジェクトの目的が大きく変わらない限りデータベースを導入しない方針です。

PR 前の確認：

```bash
gofmt -w main.go
go vet ./...
go test ./...
go build ./...
```

## ライセンス

LAN Clipboard は [MIT License](LICENSE) で公開されています。