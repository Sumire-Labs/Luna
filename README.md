# 🌙 Luna - Version 0.1.7

<div align="center">

[![Go Version](https://img.shields.io/badge/Go-1.24.4-00ADD8.svg)](https://golang.org/)
[![Discord API](https://img.shields.io/badge/Discord%20API-v10-5865F2.svg)](https://discord.com/developers/docs/)
[![License](https://img.shields.io/badge/License-LGPL--3.0-blue.svg)](LICENSE.md)

**高性能でモダンな Discord ボット - Luna AI、チケット管理、包括的ログ機能、ゲーム機能を搭載**

[✨ 機能](#-機能) • [🚀 セットアップ](#-セットアップ) • [📝 設定](#-設定) • [💡 使用方法](#-使用方法) • [🤝 貢献](#-貢献)

</div>

---

## ✨ 機能

### 🤖 Luna AI
- **自然な日本語対話**: Gemini 2.5 を活用したインテリジェントな AI アシスタント
- **画像生成**: Imagen 4 による高品質な画像生成 (`/imagine`)
- **画像解析**: OCR とコンテンツ分析 (`/ocr`)
- **多言語翻訳**: リアルタイム翻訳機能 (`/translate`)

### 🎮 ゲーム機能
- **BR ルーレット**: War Thunder の BR（Battle Rating）をランダム選択 (`/br`)
  - 🛩️ 空軍・🚗 陸軍・🚢 海軍 の3つのモード
  - インタラクティブなボタン操作
  - BR 除外設定機能（モーダル入力）
  - 3秒間のルーレットアニメーション
- **かっこ使用量ランキング**: メンバーのかっこ使用状況を統計・ランキング表示 (`/brackets`)
  - リアルタイムでかっこ `()（）` をカウント
  - TOP10 ランキング表示（🥇🥈🥉）
  - 個別ユーザー統計とバランスチェック
  - 楽しいコメント機能（マスター称号など）

### 🎫 チケットシステム
- **ワンクリック設定**: `/config` からモーダルで簡単設定
- **自動チャンネル作成**: 権限付きプライベートチャンネル
- **管理者通知**: チケット作成・クローズ時の自動通知
- **優先度システム**: 緊急度に応じたチケット管理

### 📊 包括的ログ機能
- **メッセージ監視**: 編集・削除の詳細ログ（編集前後の内容保存）
- **メンバー管理**: 参加・退出・権限変更の追跡
- **チャンネル管理**: 作成・削除・設定変更の記録
- **権限チェック**: ボット権限の自動確認と分かりやすいエラー表示

### 🎨 Material Design 3 UI
- **Luna Material 1**: カスタム Material Design テーマ
- **日本語完全対応**: すべての UI とメッセージが日本語
- **アクセシビリティ**: カラーコントラストと読みやすさを重視

---

## 🚀 セットアップ

### 📋 前提条件

- **Go 1.24.4+** - [ダウンロード](https://golang.org/dl/)
- **Discord Developer Portal** - [ボット作成](https://discord.com/developers/applications)
- **Google Cloud** (オプション) - AI機能に必要

### ⚡ クイックスタート

```bash
# 1. リポジトリをクローン
git clone https://github.com/Sumire-Labs/Luna.git
cd Luna

# 2. 依存関係をインストール
go mod tidy

# 3. 設定ファイルを作成
cp config.toml.example config.toml

# 4. 設定ファイルを編集
# Discord Token を設定してください
nano config.toml

# 5. ビルドと実行
go build -o luna cmd/bot/main.go
./luna
```

### 🐳 Docker で実行

```bash
# イメージをビルド
docker build -t luna-bot .

# コンテナを起動
docker run -d --name luna-bot \
  -v $(pwd)/config.toml:/app/config.toml \
  -v $(pwd)/data:/app/data \
  luna-bot
```

---

## 📝 設定

### 🔧 基本設定

`config.toml` を編集してボットを設定します：

```toml
[discord]
token = "YOUR_DISCORD_BOT_TOKEN"  # 必須
app_id = "YOUR_APPLICATION_ID"

[bot]
status_message = "Luna AI でサポート中"
debug = false
owners = ["YOUR_USER_ID"]  # 管理者権限

# AI機能（オプション）
[google_cloud]
use_studio_api = true
studio_api_key = "YOUR_GOOGLE_AI_STUDIO_KEY"
```

### 🤖 Luna AI の設定方法

#### Google AI Studio (無料・簡単)
1. [Google AI Studio](https://aistudio.google.com/) でAPIキーを取得
2. `config.toml` に設定：
```toml
[google_cloud]
use_studio_api = true
studio_api_key = "YOUR_API_KEY"
```

#### Vertex AI (高性能)
1. Google Cloud プロジェクトを作成
2. Vertex AI API を有効化
3. サービスアカウントキーを作成
4. `config.toml` に設定：
```toml
[google_cloud]
project_id = "your-project-id"
credentials_path = "path/to/service-account-key.json"
```

詳細は [CONFIG.md](CONFIG.md) をご覧ください。

---

## 💡 使用方法

### 🤖 Luna AI コマンド

```bash
/ask 今日の天気はどう？          # Luna AI との対話
/imagine 美しい夕焼けの風景      # AI 画像生成
/ocr                           # 画像のテキスト抽出
/translate こんにちは english   # 翻訳機能
```

### 🎮 ゲームコマンド

```bash
/br                           # War Thunder BR ルーレット
                              # - ゲームモード選択ボタン
                              # - BR 除外設定モーダル
                              # - アニメーション付きルーレット

/brackets                     # かっこ使用量ランキング TOP10
/brackets @user               # 特定ユーザーの詳細統計
```

### ⚙️ 管理コマンド

```bash
/config                        # 統合設定パネル
/ping                         # ボットの応答速度確認
/avatar @user                 # ユーザー情報表示
/purge 10                     # メッセージ一括削除
/lockdown                     # チャンネルロック
/activity                     # アクティビティ設定
```

### 🎫 チケットシステム

1. `/config` → 🎫 チケット設定 をクリック
2. 必要な情報を入力：
   - **カテゴリ**: チケット用カテゴリID
   - **サポートロール**: 対応スタッフのロール
   - **管理者ロール**: 管理権限を持つロール
3. パネルを作成してユーザーがチケットを作成できるように

### 📝 ログシステム

1. `/config` → 📝 ログ設定 をクリック
2. **ログチャンネル** を指定
3. すべてのイベントが自動で有効化されます

### 📊 統計機能

**かっこ使用量追跡**
- メッセージ送信時に自動的にかっこ `()（）` をカウント
- サーバー内でのランキング表示
- 個別ユーザー統計（バランスチェック付き）
- 楽しいコメントとマスター称号システム

---

## 🏗️ アーキテクチャ

```
Luna/
├── 🏠 cmd/bot/main.go           # エントリーポイント
├── 🤖 ai/                       # AI サービス層
│   ├── vertex_gemini.go         # 新 Vertex AI API
│   ├── gemini_studio.go         # Google AI Studio
│   ├── multimodal.go            # マルチモーダル処理
│   └── service.go               # 旧 Vertex AI (Imagen用)
├── 📝 commands/                 # コマンド実装
│   ├── ai.go                    # Luna AI コマンド
│   ├── config.go                # 設定コマンド
│   ├── brackets.go              # かっこ統計コマンド
│   ├── wt_simple.go             # BR ルーレット（旧 wt）
│   ├── interactions.go          # モーダル・ボタン処理
│   └── ...
├── 🎮 services/                 # ゲームサービス
│   └── warthunder_simple.go     # War Thunder BR 管理
├── 🗄️ database/                # データベース層
│   ├── database.go              # 接続管理
│   └── service.go               # CRUD 操作 + 統計機能
├── 📊 logging/                  # ログシステム
├── 🎨 embed/                    # Material Design UI
├── ⚙️ config/                   # 設定管理
└── 🧩 di/                       # 依存性注入
```

### 🛠️ 技術スタック

- **言語**: Go 1.24.4
- **Discord**: DiscordGo v0.29.0
- **AI**: Vertex AI , Google AI Studio, Imagen 4
- **データベース**: SQLite (modernc.org/sqlite) + 統計テーブル
- **設定**: Viper (TOML)
- **アーキテクチャ**: Clean Architecture + Dependency Injection

### 🎮 新機能の技術仕様

**BR ルーレットシステム**
- 正確な War Thunder BR 進行パターン (1.0 → 1.3 → 1.7 → 2.0...)
- モード別 BR 範囲管理（空軍: 14.0、陸軍: 12.0、海軍: 8.7）
- インタラクティブ UI（ボタン、モーダル）

**統計システム**
- リアルタイム文字カウント（半角・全角対応）
- 効率的なデータベース集計（UPSERT, インデックス最適化）
- ランキング生成とユーザー統計分析

---

## 📊 パフォーマンス

### ⚡ 最適化機能
- **並行処理**: ゴルーチンによる非同期処理
- **メッセージキャッシュ**: 編集・削除ログ用の効率的キャッシュ
- **権限チェック**: 事前権限確認でエラー防止
- **データベース**: WAL モード、プリペアドステートメント
- **統計処理**: インデックス最適化とバッチ処理

### 📈 監視機能
- **レスポンス時間**: `/ping` でリアルタイム確認
- **エラー処理**: 詳細なエラーログと回復処理
- **リソース管理**: 自動クリーンアップとメモリ最適化
- **統計データ**: 使用量追跡とパフォーマンス分析

---

## 🎮 ゲーム機能詳細

### 🎯 BR ルーレット (`/br`)
War Thunder プレイヤー向けの BR 決定ツール

**機能:**
- **モード選択**: 空軍 🛩️ / 陸軍 🚗 / 海軍 🚢
- **除外設定**: 特定 BR をモーダルで除外可能
- **アニメーション**: 3秒間のルーレット演出
- **リピート機能**: 「もう一回」「メニューに戻る」ボタン

### 📊 かっこ統計 (`/brackets`)
コミュニティの面白い統計機能

**追跡対象:**
- 開きかっこ: `(` `（`
- 閉じかっこ: `)` `）`

**表示内容:**
- **TOP10 ランキング**: 🥇🥈🥉 メダル付き
- **バランスチェック**: 開き/閉じの差分表示
- **個別統計**: 順位、詳細、楽しいコメント
- **マスター称号**: 使用回数に応じた称号システム

---

## 🤝 貢献

### 🐛 バグ報告・機能要望
[GitHub Issues](https://github.com/Sumire-Labs/Luna/issues) でお気軽にご報告ください。

### 📝 開発コマンド

```bash
# ビルド
go build -o luna cmd/bot/main.go

# 開発用実行（デバッグモード）
go run cmd/bot/main.go

# テスト実行
go test ./...

# リリース作成（"bump to x.x.x" コミットで自動リリース）
git commit -m "bump to v1.0.0"
git push origin main
```

---

## 📄 ライセンス

このプロジェクトは LGPL-3.0 ライセンスの下で公開されています。
詳細は [LICENSE.md](LICENSE.md) をご覧ください。

---

## 🌟 サポート

### 💬 コミュニティ
- **Discord サーバー**: [参加する](https://discord.gg/H8eh2hR79e)

### 📚 ドキュメント
- [📝 設定ガイド](CONFIG.md)
- [🏗️ アーキテクチャ](ARCHITECTURE.md)
- [🔌 API リファレンス](API.md)

---

<div align="center">

**🌙 Made with ❤️ for the Discord community**

[⭐ Star](https://github.com/Sumire-Labs/Luna) • [🐛 Report Bug](https://github.com/Sumire-Labs/Luna/issues) • [💡 Request Feature](https://github.com/Sumire-Labs/Luna/issues)

</div>