# 🌙 Luna Discord Bot

[![Go Version](https://img.shields.io/badge/Go-1.24.4-00ADD8.svg)](https://golang.org/)
[![Discord API](https://img.shields.io/badge/Discord%20API-v10-5865F2.svg)](https://discord.com/developers/docs/)
[![Material Design](https://img.shields.io/badge/Material%20Design-3-6750A4.svg)](https://m3.material.io/)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Language](https://img.shields.io/badge/Language-Japanese-red.svg)](README.md)

**Luna** は Go で開発された現代的な Discord ボットです。Material Design 3 のスタイリング、日本語完全対応、モジュラーアーキテクチャ、依存性注入を特徴とし、SQLite データベースと包括的なログ機能を搭載しています。

## ✨ 主要機能

### 🎯 コマンド
- **`/ping`** - ボットの応答速度とパフォーマンス監視
- **`/avatar`** - ユーザーのアバターと情報表示（高解像度対応）
- **`/config`** - サーバー管理者向け統合設定パネル

### 🛠️ 管理機能
- **🎫 チケットシステム設定** - サポート体制の完全セットアップ
- **📝 ログシステム設定** - メッセージ、メンバー、ロール変更の監視
- **⚙️ インタラクティブ設定** - モーダルベースの直感的UI

### 📊 ログ機能
- **メッセージイベント**: 編集・削除の追跡
- **メンバーイベント**: 参加・退出の監視、新規アカウント検出
- **チャンネル管理**: 作成・削除・変更の記録
- **ロール管理**: 権限変更とロール操作の監視
- **モデレーション**: BAN・UNBAN等の管理ログ

### 🎨 デザインシステム
- **Luna Material 1**
- **日本語完全対応**

## 🏗️ アーキテクチャ

### 📁 プロジェクト構造
```
Luna/
├── 📄 main.go              # アプリケーションエントリーポイント
├── 📁 bot/                 # Discord ボットコア
├── 📁 commands/            # スラッシュコマンド実装
│   ├── ping.go            # レイテンシ監視
│   ├── avatar.go          # ユーザー情報表示
│   ├── config.go          # 管理設定
│   └── interactions.go    # インタラクション処理
├── 📁 database/           # データベース層
│   ├── service.go         # データベースサービス
│   └── migrations.sql     # スキーマ定義
├── 📁 logging/            # 包括的ログシステム
├── 📁 embed/              # Material Design 3 UI
├── 📁 config/             # 設定管理
├── 📁 di/                 # 依存性注入コンテナ
├── 📄 go.mod              # Go モジュール定義
└── 📄 Makefile            # ビルドタスク
```

### 🔧 技術スタック
- **言語**: Go 1.24.4
- **Discord API**: DiscordGo v0.29.0
- **データベース**: SQLite (modernc.org/sqlite v1.38.2)
- **設定管理**: Viper v1.20.1
- **アーキテクチャ**: Clean Architecture + DI Pattern

## 🚀 セットアップ

### 📋 前提条件
- Go 1.24.4 以上
- Discord Developer Portal でのボット作成
- 必要な権限: `bot`, `applications.commands`

### ⚡ クイックスタート

1. **リポジトリのクローン**
```bash
git clone https://github.com/yourusername/luna-bot.git
cd luna-bot
```

2. **依存関係のインストール**
```bash
make deps
```

3. **環境設定**
```bash
cp .env.example .env
# .env ファイルを編集してトークンを設定
```

4. **設定ファイル作成**
```yaml
# configs/config.yaml
discord:
  token: "YOUR_BOT_TOKEN"
  app_id: "YOUR_APPLICATION_ID"
  guild_id: "YOUR_GUILD_ID"  # 開発用（オプション）

database:
  path: "./data/luna.db"
  max_connections: 10

bot:
  prefix: "!"
  status: "🌙 Luna Bot"
  activity_type: 0  # Playing
  debug: false
  owners: ["YOUR_USER_ID"]
```

5. **ビルドと実行**
```bash
make build
make run
```

### 🐳 Docker での実行
```bash
# Docker イメージのビルド
docker build -t luna-bot .

# コンテナの起動
docker run -d --name luna-bot \
  -v $(pwd)/configs:/app/configs \
  -v $(pwd)/data:/app/data \
  luna-bot
```

## 📚 使用方法

### 🎫 チケットシステム設定
1. `/config` コマンドを実行
2. 「🎫 チケット設定」ボタンをクリック
3. 必要な情報を入力:
   - **サポートカテゴリ**: チケットが作成されるカテゴリ
   - **サポートロール**: チケットに対応するロール
   - **管理者ロール**: チケット管理権限を持つロール
   - **ログチャンネル**: チケット活動のログ先
   - **自動クローズ**: 非活性チケットの自動クローズ時間

### 📝 ログシステム設定
1. `/config` コマンドを実行
2. 「📝 ログ設定」ボタンをクリック
3. ログ設定を構成:
   - **ログチャンネル**: ログメッセージの送信先
   - **監視イベント**: 記録したいイベントを選択
     - メッセージ編集・削除
     - メンバー参加・退出
     - チャンネル・ロール変更

### 👤 ユーザー情報表示
```
/avatar @user          # 指定ユーザーのアバター表示
/avatar @user true     # バナーも含めて表示
/ping                  # ボットの応答速度確認
```

## 🎨 Luna Material 1

### 🌈 カラーパレット
```go
Primary:   0x6750A4  // メインブランドカラー
Secondary: 0x625B71  // セカンダリカラー
Tertiary:  0x7D5260  // アクセントカラー
Error:     0xBA1A1A  // エラー表示
Success:   0x4CAF50  // 成功表示
Warning:   0xFF9800  // 警告表示
Info:      0x2196F3  // 情報表示
Surface:   0x1C1B1F  // サーフェスカラー
```

### 🖼️ 埋め込みスタイル
- 統一されたビジュアル体験
- 自動タイムスタンプ
- レスポンシブレイアウト
- アクセシビリティ対応

## 🗃️ データベーススキーマ

### 📊 主要テーブル

**guilds** - サーバー情報
```sql
CREATE TABLE guilds (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    joined_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
```

**guild_settings** - サーバー設定
```sql
CREATE TABLE guild_settings (
    guild_id TEXT PRIMARY KEY,
    -- チケットシステム設定
    ticket_enabled BOOLEAN DEFAULT FALSE,
    ticket_category_id TEXT,
    ticket_support_role_id TEXT,
    -- ログシステム設定
    log_enabled BOOLEAN DEFAULT FALSE,
    log_channel_id TEXT,
    log_message_edits BOOLEAN DEFAULT TRUE,
    -- その他設定...
);
```

**command_usage** - コマンド使用統計
```sql
CREATE TABLE command_usage (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    guild_id TEXT NOT NULL,
    user_id TEXT NOT NULL,
    command_name TEXT NOT NULL,
    used_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
```

## 🛡️ セキュリティ

### 🔐 権限システム
- ロールベースアクセス制御
- ギルド固有の権限検証
- オーナー限定コマンドサポート
- 自動権限検証

### 🔒 データ保護
- SQLite 外部キー制約
- プリペアドステートメント（SQLインジェクション対策）
- 入力値検証
- 環境変数による安全なトークン管理

## 🚧 開発中機能

### 📋 予定機能
- **🛡️ モデレーション**: 自動コンテンツフィルタリング
- **👋 ウェルカムシステム**: カスタム歓迎メッセージ
- **📊 統計ダッシュボード**: サーバー分析機能
- **🎵 音楽機能**: 音声チャンネル再生機能

## 🔧 開発

### 📖 Make コマンド
```bash
make build       # ビルド
make run         # 実行
make test        # テスト実行
make build-all   # クロスプラットフォームビルド
make deps        # 依存関係更新
make fmt         # コードフォーマット
make lint        # コード品質チェック
make clean       # クリーンアップ
```

### 🧪 テスト
```bash
# 全テスト実行
make test

# カバレッジレポート
go test -race -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### 🎯 コードスタイル
- Clean Architecture
- 依存性注入パターン
- インターフェースベース設計
- エラーハンドリング

## 📈 パフォーマンス

### ⚡ 最適化
- **データベース**: WAL モード、コネクションプーリング
- **メモリ**: 効率的なキャッシュ戦略
- **ネットワーク**: 非同期処理、レート制限対応
- **CPU**: 並行処理の活用

### 📊 監視
- コマンド実行統計
- レスポンス時間監視
- エラー率追跡
- リソース使用量監視

## 🤝 コントリビューション

### 📋 ガイドライン
1. **Issue 作成**: バグ報告や機能要望
2. **Fork**: リポジトリをフォーク
3. **Branch 作成**: feature/your-feature-name
4. **実装**: コードの実装とテスト
5. **Pull Request**: 詳細な説明と共に提出

### 🎨 コーディング規約
- Go の標準規約に準拠
- 日本語コメント推奨
- テストカバレッジ 80% 以上
- Luna Material 1 ガイドライン遵守

## 📄 ライセンス

LGPL-3.0 License - 詳細は [LICENSE.md](LICENSE.md) ファイルを参照してください。

## 🙋‍♂️ サポート

### 📞 お問い合わせ
- **Issues**: [GitHub Issues](https://github.com/yourusername/luna-bot/issues)
- **Discord**: [サポートサーバー](https://discord.gg/H8eh2hR79e)

### 📚 ドキュメント
- [API リファレンス](docs/api.md)
- [設定ガイド](docs/configuration.md)


---

<div align="center">

**🌙 Made with ❤️ for the Discord community**

[⭐ Star this project](https://github.com/yourusername/luna-bot) • [🐛 Report Bug](https://github.com/yourusername/luna-bot/issues) • [💡 Request Feature](https://github.com/yourusername/luna-bot/issues)

</div>