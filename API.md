# 🔌 Luna Bot API リファレンス

<div align="center">

**Luna Bot の内部 API とコマンドシステムの完全ガイド**

[💬 コマンド API](#-コマンド-api) • [🗄️ データベース API](#️-データベース-api) • [🤖 AI サービス API](#-ai-サービス-api) • [📊 ログ API](#-ログ-api)

</div>

---

## 💬 コマンド API

### 🏗️ コマンドインターフェース

すべてのコマンドは `Command` インターフェースを実装します：

```go
type Command interface {
    Name() string                                              // コマンド名
    Description() string                                       // コマンド説明
    Usage() string                                            // 使用方法
    Category() string                                         // カテゴリ
    Aliases() []string                                        // エイリアス
    Permission() int64                                        // 必要権限
    Options() []*discordgo.ApplicationCommandOption          // オプション定義
    Execute(ctx *Context) error                               // 実行ロジック
}
```

### 📝 利用可能なコマンド

#### 🤖 AI コマンド

##### `/ask` - Luna AI との対話
```go
Name: "ask"
Description: "Luna AIに質問して回答を得ます"
Options: [
    {
        Type: ApplicationCommandOptionString
        Name: "question"
        Description: "Luna AIに聞きたい質問"
        Required: true
    }
]
```

**使用例:**
```bash
/ask 今日の天気はどう？
/ask Go言語の特徴を教えて
/ask 面白いジョークを言って
```

**レスポンス:**
```json
{
    "title": "🌙 Luna AI の回答",
    "fields": [
        {
            "name": "💬 質問",
            "value": "今日の天気はどう？"
        },
        {
            "name": "📝 回答", 
            "value": "申し訳ありませんが、私はリアルタイムの天気情報にアクセスできません..."
        }
    ],
    "footer": "回答者: ユーザー名 • Powered by Luna AI"
}
```

##### `/imagine` - AI 画像生成
```go
Name: "imagine"
Description: "Luna AIで画像を生成します"
Options: [
    {
        Type: ApplicationCommandOptionString
        Name: "prompt"
        Description: "生成したい画像の説明"
        Required: true
    },
    {
        Type: ApplicationCommandOptionString
        Name: "style"
        Description: "画像のスタイル"
        Required: false
        Choices: [
            {Name: "🎨 アート", Value: "artistic"},
            {Name: "📷 写実的", Value: "photorealistic"},
            {Name: "🖼️ アニメ", Value: "anime"},
            {Name: "🎮 ゲーム", Value: "game"},
            {Name: "✏️ スケッチ", Value: "sketch"}
        ]
    }
]
```

##### `/ocr` - 画像テキスト抽出
```go
Name: "ocr"
Description: "Luna AIを使って画像からテキストを抽出・分析します"
Options: [
    {
        Type: ApplicationCommandOptionString
        Name: "image_url"
        Description: "画像のURL（添付も可）"
        Required: false
    },
    {
        Type: ApplicationCommandOptionString
        Name: "mode"
        Description: "解析モード"
        Required: false
        Choices: [
            {Name: "📖 テキスト抽出", Value: "extract"},
            {Name: "🔍 詳細分析", Value: "analyze"},
            {Name: "🌐 翻訳", Value: "translate"}
        ]
    }
]
```

##### `/translate` - 多言語翻訳
```go
Name: "translate"
Description: "テキストを指定した言語に翻訳します"
Options: [
    {
        Type: ApplicationCommandOptionString
        Name: "text"
        Description: "翻訳するテキスト"
        Required: true
    },
    {
        Type: ApplicationCommandOptionString
        Name: "target_language"
        Description: "翻訳先の言語"
        Required: false
        Choices: [
            {Name: "🇺🇸 English", Value: "english"},
            {Name: "🇯🇵 日本語", Value: "japanese"},
            {Name: "🇰🇷 한국어", Value: "korean"},
            {Name: "🇨🇳 中文", Value: "chinese"}
        ]
    }
]
```

#### ⚙️ 管理コマンド

##### `/config` - 統合設定パネル
```go
Name: "config"
Description: "サーバーの設定を管理します"
Permission: Discordgo.PermissionManageGuild
```

**インタラクションボタン:**
- `🎫 チケット設定` - チケットシステム設定
- `📝 ログ設定` - ログシステム設定
- `🔄 設定リセット` - 全設定初期化

##### `/ping` - 応答速度確認
```go
Name: "ping"
Description: "ボットの応答速度を確認します"
```

**レスポンス例:**
```json
{
    "title": "🏓 Pong!",
    "fields": [
        {
            "name": "⏱️ レイテンシ",
            "value": "45ms"
        },
        {
            "name": "🌐 API遅延",
            "value": "120ms"
        },
        {
            "name": "📊 ステータス", 
            "value": "🟢 正常"
        }
    ]
}
```

##### `/avatar` - ユーザー情報表示
```go
Name: "avatar"
Description: "ユーザーのアバターとバナーを表示します"
Options: [
    {
        Type: ApplicationCommandOptionUser
        Name: "user"
        Description: "アバターを表示するユーザー"
        Required: false
    },
    {
        Type: ApplicationCommandOptionBoolean
        Name: "show_banner"
        Description: "ユーザーのバナーも表示する"
        Required: false
    }
]
```

---

## 🗄️ データベース API

### 📊 データベーススキーマ

#### `guilds` テーブル
```sql
CREATE TABLE guilds (
    id TEXT PRIMARY KEY,                    -- Discord Guild ID
    name TEXT NOT NULL,                     -- サーバー名
    joined_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
```

#### `guild_settings` テーブル
```sql
CREATE TABLE guild_settings (
    guild_id TEXT PRIMARY KEY,
    
    -- チケットシステム設定
    ticket_enabled BOOLEAN DEFAULT FALSE,
    ticket_category_id TEXT,
    ticket_support_role_id TEXT,
    ticket_admin_role_id TEXT,
    ticket_log_channel_id TEXT,
    ticket_auto_close_hours INTEGER DEFAULT 24,
    
    -- ログシステム設定
    logging_enabled BOOLEAN DEFAULT FALSE,
    log_channel_id TEXT,
    log_message_edits BOOLEAN DEFAULT TRUE,
    log_message_deletes BOOLEAN DEFAULT TRUE,
    log_member_joins BOOLEAN DEFAULT TRUE,
    log_member_leaves BOOLEAN DEFAULT TRUE,
    log_channel_events BOOLEAN DEFAULT FALSE,
    log_role_events BOOLEAN DEFAULT FALSE,
    log_voice_events BOOLEAN DEFAULT FALSE,
    log_moderation_events BOOLEAN DEFAULT FALSE,
    log_server_events BOOLEAN DEFAULT FALSE,
    log_nickname_changes BOOLEAN DEFAULT FALSE,
    
    -- その他設定
    prefix TEXT DEFAULT '/',
    language TEXT DEFAULT 'ja',
    timezone TEXT DEFAULT 'Asia/Tokyo',
    
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    
    FOREIGN KEY (guild_id) REFERENCES guilds(id)
);
```

#### `command_usage` テーブル
```sql
CREATE TABLE command_usage (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    guild_id TEXT NOT NULL,
    user_id TEXT NOT NULL,
    command_name TEXT NOT NULL,
    execution_time_ms INTEGER,
    success BOOLEAN,
    error_message TEXT,
    used_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    
    FOREIGN KEY (guild_id) REFERENCES guilds(id)
);
```

### 🔧 データベースサービス API

```go
type Service interface {
    // Guild 管理
    UpsertGuild(guildID, name, prefix string) error
    GetGuild(guildID string) (*Guild, error)
    
    // Guild Settings 管理
    UpsertGuildSettings(settings *GuildSettings) error
    GetGuildSettings(guildID string) (*GuildSettings, error)
    
    // コマンド使用統計
    LogCommandUsage(guildID, userID, command string, duration time.Duration, success bool, errorMsg string) error
    GetCommandStats(guildID string) (*CommandStats, error)
    
    // データベース管理
    Migrate() error
    Close() error
}
```

**使用例:**
```go
// ギルド設定の更新
settings := &database.GuildSettings{
    GuildID:              "123456789012345678",
    TicketEnabled:        true,
    TicketCategoryID:     "987654321098765432",
    TicketSupportRoleID:  "111222333444555666",
    LoggingEnabled:       true,
    LogChannelID:         "777888999000111222",
}

err := db.UpsertGuildSettings(settings)
if err != nil {
    return fmt.Errorf("failed to update settings: %w", err)
}
```

---

## 🤖 AI サービス API

### 🧠 AI サービスインターフェース

```go
type AIService interface {
    AskGemini(ctx context.Context, question string, userID string) (string, error)
    GenerateImage(ctx context.Context, prompt string, userID string) ([]byte, error)
    AnalyzeImage(ctx context.Context, imageData []byte, mimeType string, question string) (string, error)
    TranslateText(ctx context.Context, text string, targetLang string) (string, error)
    Close() error
}
```

### 🎯 Vertex AI Gemini サービス

```go
type VertexGeminiService struct {
    client    *genai.Client
    model     *genai.GenerativeModel
    projectID string
    location  string
}

// Luna AI との対話
func (s *VertexGeminiService) AskGemini(ctx context.Context, question, userID string) (string, error) {
    prompt := fmt.Sprintf(`あなたは「Luna AI」です。Discord ボット「Luna」に統合された高性能AIアシスタントとして動作しています。

以下のガイドラインに従って回答してください：
- 親切で知識豊富なLuna AIとして振る舞う
- 日本語で丁寧に回答する
- Discord用に最適化された回答（2000文字以内）
- 最新の情報（2025年）に基づいて回答
- 必要に応じて絵文字を使って親しみやすく
- 自分を「Luna AI」または「私」と呼ぶ
- Geminiという名前は一切使わない

ユーザーID: %s
ユーザーの質問: %s`, userID, question)

    resp, err := s.model.GenerateContent(ctx, genai.Text(prompt))
    if err != nil {
        return "", fmt.Errorf("Gemini APIの呼び出しに失敗しました: %w", err)
    }
    
    // レスポンス処理...
    return result, nil
}
```

### 🖼️ 画像生成サービス

```go
// Imagen 4 による画像生成
func (s *Service) GenerateImage(ctx context.Context, prompt string, userID string) ([]byte, error) {
    endpoint := fmt.Sprintf("projects/%s/locations/%s/publishers/google/models/%s",
        s.projectID, s.location, s.config.ImagenModel)

    request := &aiplatformpb.PredictRequest{
        Endpoint: endpoint,
        Instances: []*structpb.Value{
            {
                Kind: &structpb.Value_StructValue{
                    StructValue: &structpb.Struct{
                        Fields: map[string]*structpb.Value{
                            "prompt": {
                                Kind: &structpb.Value_StringValue{
                                    StringValue: prompt,
                                },
                            },
                        },
                    },
                },
            },
        },
    }

    resp, err := s.predictionClient.Predict(ctx, request)
    // 画像データ処理...
    return imageData, nil
}
```

---

## 📊 ログ API

### 📋 ログシステムアーキテクチャ

```go
type Logger struct {
    session      *discordgo.Session
    config       *config.Config
    db           *database.Service
    messageCache *MessageCache
}

type MessageCache struct {
    mu       sync.RWMutex
    messages map[string]*CachedMessage
}

type CachedMessage struct {
    Content     string
    AuthorID    string
    AuthorName  string
    Attachments []string
    Embeds      int
    Timestamp   time.Time
}
```

### 🔍 ログイベントタイプ

```go
type LogEvent string

const (
    EventMessageEdit           LogEvent = "message_edit"
    EventMessageDelete         LogEvent = "message_delete"
    EventMemberJoin           LogEvent = "member_join"
    EventMemberLeave          LogEvent = "member_leave"
    EventChannelCreate        LogEvent = "channel_create"
    EventChannelDelete        LogEvent = "channel_delete"
    EventChannelUpdate        LogEvent = "channel_update"
    EventRoleCreate           LogEvent = "role_create"
    EventRoleDelete           LogEvent = "role_delete"
    EventRoleUpdate           LogEvent = "role_update"
    EventVoiceStateUpdate     LogEvent = "voice_state_update"
    EventModerationBan        LogEvent = "moderation_ban"
    EventModerationUnban      LogEvent = "moderation_unban"
    EventServerUpdate         LogEvent = "server_update"
    EventNicknameChange       LogEvent = "nickname_change"
)
```

### 📝 ログメソッド

```go
// メッセージ編集ログ
func (l *Logger) onMessageUpdate(s *discordgo.Session, m *discordgo.MessageUpdate) {
    // 権限チェック
    shouldLog, channelID := l.shouldLog(m.GuildID, EventMessageEdit)
    if !shouldLog {
        return
    }
    
    // キャッシュから編集前の内容を取得
    var oldContent string
    l.messageCache.mu.RLock()
    if cached, exists := l.messageCache.messages[m.ID]; exists {
        oldContent = cached.Content
    }
    l.messageCache.mu.RUnlock()
    
    // ログ埋め込み作成
    embedBuilder := embed.New().
        SetTitle("📝 メッセージが編集されました").
        SetColor(embed.M3Colors.Warning).
        AddField("👤 ユーザー", fmt.Sprintf("<@%s>", m.Author.ID), true).
        AddField("📍 チャンネル", fmt.Sprintf("<#%s>", m.ChannelID), true).
        AddField("🕐 編集時刻", fmt.Sprintf("<t:%d:F>", time.Now().Unix()), true)
    
    if oldContent != "" {
        if len(oldContent) > 1000 {
            oldContent = oldContent[:1000] + "..."
        }
        embedBuilder.AddField("📜 編集前", oldContent, false)
    }
    
    newContent := m.Content
    if len(newContent) > 1000 {
        newContent = newContent[:1000] + "..."
    }
    embedBuilder.AddField("📝 編集後", newContent, false)
    
    l.sendLogMessage(channelID, embedBuilder.Build())
}
```

---

## 🎨 UI API (Embed Builder)

### 🌈 Material Design 3 カラーパレット

```go
var M3Colors = struct {
    Primary   int
    Secondary int
    Tertiary  int
    Error     int
    Success   int
    Warning   int
    Info      int
    Surface   int
}{
    Primary:   0x6750A4,  // メインブランドカラー
    Secondary: 0x625B71,  // セカンダリカラー
    Tertiary:  0x7D5260,  // アクセントカラー
    Error:     0xBA1A1A,  // エラー表示
    Success:   0x4CAF50,  // 成功表示
    Warning:   0xFF9800,  // 警告表示
    Info:      0x2196F3,  // 情報表示
    Surface:   0x1C1B1F,  // サーフェスカラー
}
```

### 🏗️ Embed Builder API

```go
type Builder struct {
    embed *discordgo.MessageEmbed
}

func New() *Builder {
    return &Builder{
        embed: &discordgo.MessageEmbed{
            Timestamp: time.Now().Format(time.RFC3339),
        },
    }
}

// チェーンメソッド
func (b *Builder) SetTitle(title string) *Builder
func (b *Builder) SetDescription(description string) *Builder
func (b *Builder) SetColor(color int) *Builder
func (b *Builder) SetThumbnail(url string) *Builder
func (b *Builder) SetImage(url string) *Builder
func (b *Builder) AddField(name, value string, inline bool) *Builder
func (b *Builder) SetFooter(text, iconURL string) *Builder
func (b *Builder) Build() *discordgo.MessageEmbed
```

**使用例:**
```go
embed := embed.New().
    SetTitle("🌙 Luna AI の回答").
    SetColor(embed.M3Colors.Primary).
    AddField("💬 質問", question, false).
    AddField("📝 回答", answer, false).
    SetFooter("Powered by Luna AI", "").
    Build()

ctx.EditReplyEmbed(embed)
```

---

## 🔧 エラーハンドリング API

### ⚠️ カスタムエラータイプ

```go
type LunaError struct {
    Code    ErrorCode `json:"code"`
    Message string    `json:"message"`
    Details string    `json:"details,omitempty"`
}

type ErrorCode int

const (
    ErrUnknown ErrorCode = iota
    ErrInvalidInput
    ErrPermissionDenied
    ErrRateLimit
    ErrExternalAPI
    ErrDatabase
    ErrConfiguration
)

func (e *LunaError) Error() string {
    return fmt.Sprintf("[%d] %s: %s", e.Code, e.Message, e.Details)
}
```

### 📋 エラーレスポンス

```go
func handleError(ctx *Context, err error) error {
    var lunaErr *LunaError
    if errors.As(err, &lunaErr) {
        embed := embed.New().
            SetTitle("❌ エラーが発生しました").
            SetDescription(lunaErr.Message).
            SetColor(embed.M3Colors.Error)
        
        if lunaErr.Details != "" {
            embed.AddField("詳細", lunaErr.Details, false)
        }
        
        return ctx.ReplyEphemeral(embed.Build())
    }
    
    // 一般的なエラー
    return ctx.ReplyEphemeral("予期しないエラーが発生しました。")
}
```

---

## 📊 メトリクス API

### 📈 パフォーマンス監視

```go
type Metrics struct {
    CommandLatency   map[string]time.Duration `json:"command_latency"`
    AIRequestCount   int64                    `json:"ai_request_count"`
    DatabaseQueries  int64                    `json:"database_queries"`
    ActiveUsers      int64                    `json:"active_users"`
    ErrorRate        float64                  `json:"error_rate"`
    UptimeSeconds    int64                    `json:"uptime_seconds"`
}

func (m *Metrics) RecordCommandExecution(command string, duration time.Duration) {
    m.CommandLatency[command] = duration
}

func (m *Metrics) IncrementAIRequest() {
    atomic.AddInt64(&m.AIRequestCount, 1)
}
```

---

## 🚀 拡張 API

### 🧩 プラグインインターフェース

```go
type Plugin interface {
    Name() string
    Version() string
    Description() string
    Initialize(container *di.Container) error
    Commands() []Command
    EventHandlers() []EventHandler
    Shutdown() error
}

type EventHandler interface {
    Event() string
    Handle(session *discordgo.Session, event interface{}) error
}
```

### 📡 Webhook API

```go
type WebhookManager struct {
    webhooks map[string]*discordgo.Webhook
}

func (w *WebhookManager) SendMessage(channelID string, content *WebhookContent) error {
    webhook, exists := w.webhooks[channelID]
    if !exists {
        // Webhook作成
        webhook, err := w.createWebhook(channelID)
        if err != nil {
            return err
        }
        w.webhooks[channelID] = webhook
    }
    
    return w.executeWebhook(webhook, content)
}
```

---

## 📞 サポート

### 🐛 API に関する問題

- **GitHub Issues**: [API関連のバグ報告](https://github.com/yourusername/luna-bot/issues/new?labels=api)
- **Discord サーバー**: [開発者チャンネル](https://discord.gg/H8eh2hR79e)

### 📚 関連ドキュメント

- [README.md](README.md) - プロジェクト概要
- [CONFIG.md](CONFIG.md) - 設定ガイド
- [ARCHITECTURE.md](ARCHITECTURE.md) - アーキテクチャ詳細