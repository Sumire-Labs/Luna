# 📝 Luna Bot 設定ガイド

<div align="center">

**Luna Bot の設定を簡単に理解して、すぐに使い始めましょう**

[🚀 クイック設定](#-クイック設定) • [📋 設定項目](#-設定項目) • [🤖 AI設定](#-ai設定) • [🔧 高度な設定](#-高度な設定)

</div>

---

## 🚀 クイック設定

### ⚡ 最速セットアップ（3ステップ）

```bash
# 1. 設定ファイルを作成
cp config.toml.example config.toml

# 2. Discord トークンを設定
nano config.toml  # [discord] token = "YOUR_TOKEN"

# 3. 起動
./luna
```

これだけで基本機能が使用できます！

---

## 📋 設定項目

### 🎯 必須設定

| 項目 | 説明 | 例 |
|------|------|-----|
| `discord.token` | Discord ボットトークン | `"MTA...xyz"` |

### ⚙️ 基本設定

```toml
[discord]
token = "YOUR_DISCORD_BOT_TOKEN"          # 必須
app_id = "1234567890123456789"            # アプリケーションID
guild_id = ""                             # 空＝グローバル、指定＝テスト用
permissions = 8                           # ボット権限 (8 = 管理者)

[bot]
prefix = "/"                              # コマンドプレフィックス
status_message = "Luna AI でサポート中"     # ステータス表示
activity_type = 0                         # 0:Playing 1:Streaming 2:Listening 3:Watching
debug = false                             # デバッグモード
owners = ["123456789012345678"]           # ボット管理者のユーザーID

[database]
path = "./data/luna.db"                   # データベースファイル
max_connections = 10                      # 最大DB接続数
```

---

## 🤖 AI設定

### 🆓 Google AI Studio（推奨・無料）

```toml
[google_cloud]
use_studio_api = true
studio_api_key = "AIzaSyD..."
```

**設定手順:**
1. [Google AI Studio](https://aistudio.google.com/) にアクセス
2. 「Get API Key」をクリック
3. 新しいプロジェクトを作成または選択
4. APIキーをコピーして `studio_api_key` に設定

### 🚀 Vertex AI（高性能・有料）

```toml
[google_cloud]
use_studio_api = false
project_id = "your-gcp-project-id"
location = "us-central1"
credentials_path = "./service-account-key.json"
gemini_model = "gemini-2.5-flash-lite"
imagen_model = "imagen-4.0-fast-generate-preview-06-06"
```

**設定手順:**
1. [Google Cloud Console](https://console.cloud.google.com/) でプロジェクト作成
2. Vertex AI API を有効化
3. サービスアカウントを作成し、JSONキーをダウンロード
4. 設定ファイルに項目を追加

### 🎨 利用可能な AI機能

| 機能 | Google AI Studio | Vertex AI |
|------|:---------------:|:---------:|
| 💬 テキスト対話 (`/ask`) | ✅ | ✅ |
| 🖼️ 画像生成 (`/imagine`) | ❌ | ✅ |
| 📖 OCR (`/ocr`) | ✅ | ✅ |
| 🌐 翻訳 (`/translate`) | ✅ | ✅ |

---

## 🔧 高度な設定

### 📊 ログ設定

```toml
[logging]
level = "info"                            # debug, info, warn, error
format = "text"                           # text, json
output = "console"                        # console, file, both
```

### 🎛️ 機能フラグ

```toml
[features]
enable_ai = true                          # Luna AI 機能
enable_logging = true                     # Discord ログ機能
enable_tickets = true                     # チケットシステム
enable_moderation = false                 # モデレーション（開発中）
enable_music = false                      # 音楽機能（開発中）
```

---

## 🔍 設定の確認方法

### ✅ 設定検証

Luna Bot 起動時に以下が表示されれば設定完了です：

```bash
Using config file: /path/to/config.toml
Luna Bot is now running. Press CTRL+C to exit.
```

### 🐛 よくあるエラー

| エラー | 原因 | 解決方法 |
|--------|------|----------|
| `discord.token is required` | トークン未設定 | `[discord] token = "..."` を追加 |
| `config.toml not found` | 設定ファイル無し | `cp config.toml.example config.toml` |
| `HTTP 401 Unauthorized` | 無効なトークン | Discord Developer Portal でトークン確認 |
| `HTTP 403 Forbidden` | 権限不足 | ボットをサーバーに再招待 |

---

## 🌍 環境変数オーバーライド

設定ファイルの値は環境変数で上書きできます：

```bash
# 環境変数名は大文字 + アンダースコア
export DISCORD_TOKEN="your_token"
export GOOGLE_CLOUD_PROJECT_ID="your_project"
export BOT_DEBUG="true"

# または .env ファイル
echo "DISCORD_TOKEN=your_token" > .env
```

### 📋 環境変数一覧

| 設定項目 | 環境変数 |
|----------|----------|
| `discord.token` | `DISCORD_TOKEN` |
| `discord.app_id` | `DISCORD_APP_ID` |
| `bot.debug` | `BOT_DEBUG` |
| `google_cloud.project_id` | `GOOGLE_CLOUD_PROJECT_ID` |
| `google_cloud.studio_api_key` | `GOOGLE_CLOUD_STUDIO_API_KEY` |

---

## 🔒 セキュリティのベストプラクティス

### 🛡️ トークン保護

```toml
# ❌ 悪い例（リポジトリにコミットしてしまう）
[discord]
token = "MTA1234567890.ABCDEF.xyz-secret-token"

# ✅ 良い例（環境変数を使用）
[discord]
token = "${DISCORD_TOKEN}"  # 環境変数から取得
```

### 📁 ファイル権限

```bash
# 設定ファイルの権限を制限
chmod 600 config.toml
chmod 600 .env
chmod 600 service-account-key.json
```

### 🚫 .gitignore

```gitignore
# 秘密情報をコミットしない
config.toml
.env
*.json
data/
```

---

## 🎯 用途別設定例

### 🧪 開発・テスト環境

```toml
[discord]
token = "YOUR_DEV_TOKEN"
guild_id = "123456789012345678"           # テスト用サーバーID

[bot]
debug = true                              # 詳細ログ
status_message = "🔧 開発中"

[features]
enable_ai = false                         # AI機能を無効化
```

### 🚀 本番環境

```toml
[discord]
token = "${DISCORD_TOKEN}"                # 環境変数
guild_id = ""                             # グローバルコマンド

[bot]
debug = false
status_message = "Luna AI でサポート中"

[google_cloud]
use_studio_api = true
studio_api_key = "${GOOGLE_AI_STUDIO_KEY}"

[logging]
level = "warn"                            # 重要なログのみ
format = "json"                           # 構造化ログ
```

---

## 📞 サポート

### 🔧 設定でお困りの場合

1. **GitHub Issues**: [設定に関する質問](https://github.com/yourusername/luna-bot/issues/new?template=config.md)
2. **Discord サーバー**: [サポートチャンネル](https://discord.gg/H8eh2hR79e)
3. **ドキュメント**: [README.md](README.md) | [ARCHITECTURE.md](ARCHITECTURE.md)

### 📝 設定テンプレート

用途別の設定テンプレートは [examples/](examples/) ディレクトリにあります：
- `config.minimal.toml` - 最小構成
- `config.ai-enabled.toml` - AI機能有効
- `config.production.toml` - 本番環境用