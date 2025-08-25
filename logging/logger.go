package logging

import (
	"fmt"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/Sumire-Labs/Luna/config"
	"github.com/Sumire-Labs/Luna/database"
	"github.com/Sumire-Labs/Luna/embed"
)

type Logger struct {
	session *discordgo.Session
	config  *config.Config
	db      *database.Service
}

type LogEvent string

const (
	EventMessageEdit   LogEvent = "message_edit"
	EventMessageDelete LogEvent = "message_delete"
	EventMemberJoin    LogEvent = "member_join"
	EventMemberLeave   LogEvent = "member_leave"
	EventChannelCreate LogEvent = "channel_create"
	EventChannelDelete LogEvent = "channel_delete"
	EventChannelUpdate LogEvent = "channel_update"
	EventRoleCreate    LogEvent = "role_create"
	EventRoleDelete    LogEvent = "role_delete"
	EventRoleUpdate    LogEvent = "role_update"
	EventMemberBan     LogEvent = "member_ban"
	EventMemberUnban   LogEvent = "member_unban"
	EventMemberKick    LogEvent = "member_kick"
)

func NewLogger(session *discordgo.Session, cfg *config.Config, db *database.Service) *Logger {
	return &Logger{
		session: session,
		config:  cfg,
		db:      db,
	}
}

func (l *Logger) RegisterHandlers() {
	l.session.AddHandler(l.onMessageUpdate)
	l.session.AddHandler(l.onMessageDelete)
	l.session.AddHandler(l.onGuildMemberAdd)
	l.session.AddHandler(l.onGuildMemberRemove)
	l.session.AddHandler(l.onChannelCreate)
	l.session.AddHandler(l.onChannelDelete)
	l.session.AddHandler(l.onChannelUpdate)
	l.session.AddHandler(l.onGuildRoleCreate)
	l.session.AddHandler(l.onGuildRoleDelete)
	l.session.AddHandler(l.onGuildRoleUpdate)
	l.session.AddHandler(l.onGuildBanAdd)
	l.session.AddHandler(l.onGuildBanRemove)
}

func (l *Logger) shouldLog(guildID string, eventType LogEvent) (bool, string) {
	settings, err := l.db.GetGuildSettings(guildID)
	if err != nil || !settings.LoggingEnabled {
		return false, ""
	}

	if settings.LogChannelID == "" {
		return false, ""
	}

	switch eventType {
	case EventMessageEdit:
		return settings.LogMessageEdits, settings.LogChannelID
	case EventMessageDelete:
		return settings.LogMessageDeletes, settings.LogChannelID
	case EventMemberJoin:
		return settings.LogMemberJoins, settings.LogChannelID
	case EventMemberLeave:
		return settings.LogMemberLeaves, settings.LogChannelID
	default:
		return true, settings.LogChannelID
	}
}

func (l *Logger) sendLogMessage(channelID string, embed *discordgo.MessageEmbed) {
	_, err := l.session.ChannelMessageSendEmbed(channelID, embed)
	if err != nil {
		// ログチャンネルへの送信に失敗した場合のエラーハンドリング
		fmt.Printf("Failed to send log message: %v\n", err)
	}
}

// メッセージ編集ログ
func (l *Logger) onMessageUpdate(s *discordgo.Session, m *discordgo.MessageUpdate) {
	if m.GuildID == "" || m.Author == nil || m.Author.Bot {
		return
	}

	shouldLog, channelID := l.shouldLog(m.GuildID, EventMessageEdit)
	if !shouldLog {
		return
	}

	// 元のメッセージを取得（可能であれば）
	var oldContent string
	if m.BeforeUpdate != nil {
		oldContent = m.BeforeUpdate.Content
	}

	if oldContent == m.Content {
		return // 内容に変更がない場合はログしない
	}

	embedBuilder := embed.New().
		SetTitle("📝 メッセージが編集されました").
		SetColor(embed.M3Colors.Warning).
		AddField("👤 ユーザー", fmt.Sprintf("<@%s>", m.Author.ID), true).
		AddField("📍 チャンネル", fmt.Sprintf("<#%s>", m.ChannelID), true).
		AddField("🕐 編集時刻", fmt.Sprintf("<t:%d:F>", time.Now().Unix()), true)

	if oldContent != "" {
		// 内容が長い場合は切り詰める
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

	if m.ID != "" {
		embedBuilder.AddField("🔗 ジャンプ", fmt.Sprintf("[メッセージに移動](https://discord.com/channels/%s/%s/%s)", m.GuildID, m.ChannelID, m.ID), false)
	}

	embedBuilder.SetFooter(fmt.Sprintf("メッセージID: %s", m.ID), "")

	l.sendLogMessage(channelID, embedBuilder.Build())
}

// メッセージ削除ログ
func (l *Logger) onMessageDelete(s *discordgo.Session, m *discordgo.MessageDelete) {
	if m.GuildID == "" {
		return
	}

	shouldLog, channelID := l.shouldLog(m.GuildID, EventMessageDelete)
	if !shouldLog {
		return
	}

	embedBuilder := embed.New().
		SetTitle("🗑️ メッセージが削除されました").
		SetColor(embed.M3Colors.Error).
		AddField("📍 チャンネル", fmt.Sprintf("<#%s>", m.ChannelID), true).
		AddField("🕐 削除時刻", fmt.Sprintf("<t:%d:F>", time.Now().Unix()), true)

	// メッセージの情報が利用可能な場合
	if m.BeforeDelete != nil {
		msg := m.BeforeDelete
		if msg.Author != nil && !msg.Author.Bot {
			embedBuilder.AddField("👤 作成者", fmt.Sprintf("<@%s>", msg.Author.ID), true)
		}

		if msg.Content != "" {
			content := msg.Content
			if len(content) > 1000 {
				content = content[:1000] + "..."
			}
			embedBuilder.AddField("📜 削除されたメッセージ", content, false)
		}

		if len(msg.Attachments) > 0 {
			attachmentList := make([]string, len(msg.Attachments))
			for i, att := range msg.Attachments {
				attachmentList[i] = fmt.Sprintf("• %s", att.Filename)
			}
			embedBuilder.AddField("📎 添付ファイル", strings.Join(attachmentList, "\n"), false)
		}

		if len(msg.Embeds) > 0 {
			embedBuilder.AddField("🖼️ Embed", fmt.Sprintf("%d個のEmbedが含まれていました", len(msg.Embeds)), false)
		}
	}

	embedBuilder.SetFooter(fmt.Sprintf("メッセージID: %s", m.ID), "")

	l.sendLogMessage(channelID, embedBuilder.Build())
}

// メンバー参加ログ
func (l *Logger) onGuildMemberAdd(s *discordgo.Session, m *discordgo.GuildMemberAdd) {
	shouldLog, channelID := l.shouldLog(m.GuildID, EventMemberJoin)
	if !shouldLog {
		return
	}

	// アカウント作成日を計算
	userID := m.User.ID
	snowflake, _ := discordgo.SnowflakeTimestamp(userID)
	accountAge := time.Since(snowflake)

	embedBuilder := embed.New().
		SetTitle("📥 新しいメンバーが参加しました").
		SetColor(embed.M3Colors.Success).
		AddField("👤 ユーザー", fmt.Sprintf("<@%s>", m.User.ID), true).
		AddField("🆔 ユーザーID", m.User.ID, true).
		AddField("🕐 参加時刻", fmt.Sprintf("<t:%d:F>", time.Now().Unix()), true).
		AddField("📅 アカウント作成日", fmt.Sprintf("<t:%d:F>", snowflake.Unix()), true).
		AddField("⏰ アカウント年数", fmt.Sprintf("%.0f日前", accountAge.Hours()/24), true)

	if m.User.Avatar != "" {
		embedBuilder.SetThumbnail(m.User.AvatarURL("256"))
	}

	// 新規アカウントの場合は警告
	if accountAge < time.Hour*24*7 {
		embedBuilder.AddField("⚠️ 注意", "新規作成アカウント（7日以内）", false)
		embedBuilder.SetColor(embed.M3Colors.Warning)
	}

	l.sendLogMessage(channelID, embedBuilder.Build())
}

// メンバー退出ログ
func (l *Logger) onGuildMemberRemove(s *discordgo.Session, m *discordgo.GuildMemberRemove) {
	shouldLog, channelID := l.shouldLog(m.GuildID, EventMemberLeave)
	if !shouldLog {
		return
	}

	embedBuilder := embed.New().
		SetTitle("📤 メンバーが退出しました").
		SetColor(embed.M3Colors.Error).
		AddField("👤 ユーザー", fmt.Sprintf("<@%s>", m.User.ID), true).
		AddField("🆔 ユーザーID", m.User.ID, true).
		AddField("🕐 退出時刻", fmt.Sprintf("<t:%d:F>", time.Now().Unix()), true)

	if m.User.Avatar != "" {
		embedBuilder.SetThumbnail(m.User.AvatarURL("256"))
	}

	// 参加期間を計算（可能であれば）
	guild, err := s.State.Guild(m.GuildID)
	if err == nil {
		for _, member := range guild.Members {
			if member.User.ID == m.User.ID && !member.JoinedAt.IsZero() {
				joinTime := member.JoinedAt
				err := (error)(nil)
				if err == nil {
					duration := time.Since(joinTime)
					embedBuilder.AddField("⏱️ 参加期間", fmt.Sprintf("%.0f日間", duration.Hours()/24), true)
				}
				break
			}
		}
	}

	l.sendLogMessage(channelID, embedBuilder.Build())
}

// チャンネル作成ログ
func (l *Logger) onChannelCreate(s *discordgo.Session, c *discordgo.ChannelCreate) {
	if c.GuildID == "" {
		return
	}

	shouldLog, channelID := l.shouldLog(c.GuildID, EventChannelCreate)
	if !shouldLog {
		return
	}

	channelType := l.getChannelTypeString(c.Type)

	embedBuilder := embed.New().
		SetTitle("➕ チャンネルが作成されました").
		SetColor(embed.M3Colors.Success).
		AddField("📍 チャンネル", fmt.Sprintf("<#%s>", c.ID), true).
		AddField("📝 チャンネル名", c.Name, true).
		AddField("🔖 タイプ", channelType, true).
		AddField("🕐 作成時刻", fmt.Sprintf("<t:%d:F>", time.Now().Unix()), true)

	if c.Topic != "" {
		embedBuilder.AddField("📜 トピック", c.Topic, false)
	}

	embedBuilder.SetFooter(fmt.Sprintf("チャンネルID: %s", c.ID), "")

	l.sendLogMessage(channelID, embedBuilder.Build())
}

// チャンネル削除ログ
func (l *Logger) onChannelDelete(s *discordgo.Session, c *discordgo.ChannelDelete) {
	if c.GuildID == "" {
		return
	}

	shouldLog, channelID := l.shouldLog(c.GuildID, EventChannelDelete)
	if !shouldLog {
		return
	}

	channelType := l.getChannelTypeString(c.Type)

	embedBuilder := embed.New().
		SetTitle("➖ チャンネルが削除されました").
		SetColor(embed.M3Colors.Error).
		AddField("📝 チャンネル名", c.Name, true).
		AddField("🔖 タイプ", channelType, true).
		AddField("🕐 削除時刻", fmt.Sprintf("<t:%d:F>", time.Now().Unix()), true)

	if c.Topic != "" {
		embedBuilder.AddField("📜 トピック", c.Topic, false)
	}

	embedBuilder.SetFooter(fmt.Sprintf("チャンネルID: %s", c.ID), "")

	l.sendLogMessage(channelID, embedBuilder.Build())
}

// チャンネル更新ログ
func (l *Logger) onChannelUpdate(s *discordgo.Session, c *discordgo.ChannelUpdate) {
	if c.GuildID == "" {
		return
	}

	shouldLog, channelID := l.shouldLog(c.GuildID, EventChannelUpdate)
	if !shouldLog {
		return
	}

	// 変更点を検出
	changes := l.detectChannelChanges(c.BeforeUpdate, c.Channel)
	if len(changes) == 0 {
		return
	}

	embedBuilder := embed.New().
		SetTitle("📝 チャンネルが更新されました").
		SetColor(embed.M3Colors.Warning).
		AddField("📍 チャンネル", fmt.Sprintf("<#%s>", c.ID), true).
		AddField("🕐 更新時刻", fmt.Sprintf("<t:%d:F>", time.Now().Unix()), true)

	for _, change := range changes {
		embedBuilder.AddField(change.Field, change.Description, false)
	}

	embedBuilder.SetFooter(fmt.Sprintf("チャンネルID: %s", c.ID), "")

	l.sendLogMessage(channelID, embedBuilder.Build())
}

// ロール作成ログ
func (l *Logger) onGuildRoleCreate(s *discordgo.Session, r *discordgo.GuildRoleCreate) {
	shouldLog, channelID := l.shouldLog(r.GuildID, EventRoleCreate)
	if !shouldLog {
		return
	}

	embedBuilder := embed.New().
		SetTitle("➕ ロールが作成されました").
		SetColor(embed.M3Colors.Success).
		AddField("🎭 ロール", fmt.Sprintf("<@&%s>", r.Role.ID), true).
		AddField("📝 ロール名", r.Role.Name, true).
		AddField("🎨 カラー", fmt.Sprintf("#%06x", r.Role.Color), true).
		AddField("🕐 作成時刻", fmt.Sprintf("<t:%d:F>", time.Now().Unix()), true).
		AddField("📍 位置", fmt.Sprintf("%d", r.Role.Position), true).
		AddField("🔒 管理者権限", l.getBoolString(r.Role.Permissions&discordgo.PermissionAdministrator != 0), true)

	embedBuilder.SetFooter(fmt.Sprintf("ロールID: %s", r.Role.ID), "")

	l.sendLogMessage(channelID, embedBuilder.Build())
}

// ロール削除ログ
func (l *Logger) onGuildRoleDelete(s *discordgo.Session, r *discordgo.GuildRoleDelete) {
	shouldLog, channelID := l.shouldLog(r.GuildID, EventRoleDelete)
	if !shouldLog {
		return
	}

	embedBuilder := embed.New().
		SetTitle("➖ ロールが削除されました").
		SetColor(embed.M3Colors.Error).
		AddField("🕐 削除時刻", fmt.Sprintf("<t:%d:F>", time.Now().Unix()), true)

	embedBuilder.SetFooter(fmt.Sprintf("ロールID: %s", r.RoleID), "")

	l.sendLogMessage(channelID, embedBuilder.Build())
}

// ロール更新ログ
func (l *Logger) onGuildRoleUpdate(s *discordgo.Session, r *discordgo.GuildRoleUpdate) {
	shouldLog, channelID := l.shouldLog(r.GuildID, EventRoleUpdate)
	if !shouldLog {
		return
	}

	// ロール更新を記録（変更前の情報は利用できないため、現在の状態のみ記録）

	embedBuilder := embed.New().
		SetTitle("📝 ロールが更新されました").
		SetColor(embed.M3Colors.Warning).
		AddField("🎭 ロール", fmt.Sprintf("<@&%s>", r.Role.ID), true).
		AddField("🏷️ ロール名", r.Role.Name, true).
		AddField("🎨 カラー", fmt.Sprintf("#%06X", r.Role.Color), true).
		AddField("🕐 更新時刻", fmt.Sprintf("<t:%d:F>", time.Now().Unix()), true).
		SetFooter(fmt.Sprintf("ロールID: %s", r.Role.ID), "")

	l.sendLogMessage(channelID, embedBuilder.Build())
}

// BANログ
func (l *Logger) onGuildBanAdd(s *discordgo.Session, b *discordgo.GuildBanAdd) {
	shouldLog, channelID := l.shouldLog(b.GuildID, EventMemberBan)
	if !shouldLog {
		return
	}

	embedBuilder := embed.New().
		SetTitle("🔨 メンバーがBANされました").
		SetColor(embed.M3Colors.Error).
		AddField("👤 ユーザー", fmt.Sprintf("<@%s>", b.User.ID), true).
		AddField("🆔 ユーザーID", b.User.ID, true).
		AddField("🕐 BAN時刻", fmt.Sprintf("<t:%d:F>", time.Now().Unix()), true)

	if b.User.Avatar != "" {
		embedBuilder.SetThumbnail(b.User.AvatarURL("256"))
	}

	l.sendLogMessage(channelID, embedBuilder.Build())
}

// BAN解除ログ
func (l *Logger) onGuildBanRemove(s *discordgo.Session, b *discordgo.GuildBanRemove) {
	shouldLog, channelID := l.shouldLog(b.GuildID, EventMemberUnban)
	if !shouldLog {
		return
	}

	embedBuilder := embed.New().
		SetTitle("🔓 BANが解除されました").
		SetColor(embed.M3Colors.Success).
		AddField("👤 ユーザー", fmt.Sprintf("<@%s>", b.User.ID), true).
		AddField("🆔 ユーザーID", b.User.ID, true).
		AddField("🕐 解除時刻", fmt.Sprintf("<t:%d:F>", time.Now().Unix()), true)

	if b.User.Avatar != "" {
		embedBuilder.SetThumbnail(b.User.AvatarURL("256"))
	}

	l.sendLogMessage(channelID, embedBuilder.Build())
}

// ヘルパー関数

type ChangeInfo struct {
	Field       string
	Description string
}

func (l *Logger) getChannelTypeString(channelType discordgo.ChannelType) string {
	switch channelType {
	case discordgo.ChannelTypeGuildText:
		return "📝 テキスト"
	case discordgo.ChannelTypeGuildVoice:
		return "🔊 ボイス"
	case discordgo.ChannelTypeGuildCategory:
		return "📁 カテゴリ"
	case discordgo.ChannelTypeGuildNews:
		return "📰 ニュース"
	case discordgo.ChannelTypeGuildStore:
		return "🛒 ストア"
	case discordgo.ChannelTypeGuildNewsThread:
		return "🧵 ニューススレッド"
	case discordgo.ChannelTypeGuildPublicThread:
		return "🧵 パブリックスレッド"
	case discordgo.ChannelTypeGuildPrivateThread:
		return "🧵 プライベートスレッド"
	case discordgo.ChannelTypeGuildStageVoice:
		return "🎤 ステージ"
	default:
		return "❓ 不明"
	}
}

func (l *Logger) detectChannelChanges(before, after *discordgo.Channel) []ChangeInfo {
	var changes []ChangeInfo

	if before == nil {
		return changes
	}

	if before.Name != after.Name {
		changes = append(changes, ChangeInfo{
			Field:       "📝 チャンネル名",
			Description: fmt.Sprintf("`%s` → `%s`", before.Name, after.Name),
		})
	}

	if before.Topic != after.Topic {
		beforeTopic := before.Topic
		if beforeTopic == "" {
			beforeTopic = "未設定"
		}
		afterTopic := after.Topic
		if afterTopic == "" {
			afterTopic = "未設定"
		}
		changes = append(changes, ChangeInfo{
			Field:       "📜 トピック",
			Description: fmt.Sprintf("`%s` → `%s`", beforeTopic, afterTopic),
		})
	}

	if before.NSFW != after.NSFW {
		changes = append(changes, ChangeInfo{
			Field:       "🔞 NSFW",
			Description: fmt.Sprintf("`%s` → `%s`", l.getBoolString(before.NSFW), l.getBoolString(after.NSFW)),
		})
	}

	return changes
}

func (l *Logger) detectRoleChanges(before, after *discordgo.Role) []ChangeInfo {
	var changes []ChangeInfo

	if before == nil {
		return changes
	}

	if before.Name != after.Name {
		changes = append(changes, ChangeInfo{
			Field:       "📝 ロール名",
			Description: fmt.Sprintf("`%s` → `%s`", before.Name, after.Name),
		})
	}

	if before.Color != after.Color {
		changes = append(changes, ChangeInfo{
			Field:       "🎨 カラー",
			Description: fmt.Sprintf("`#%06x` → `#%06x`", before.Color, after.Color),
		})
	}

	if before.Hoist != after.Hoist {
		changes = append(changes, ChangeInfo{
			Field:       "📍 別表示",
			Description: fmt.Sprintf("`%s` → `%s`", l.getBoolString(before.Hoist), l.getBoolString(after.Hoist)),
		})
	}

	if before.Mentionable != after.Mentionable {
		changes = append(changes, ChangeInfo{
			Field:       "💬 メンション可能",
			Description: fmt.Sprintf("`%s` → `%s`", l.getBoolString(before.Mentionable), l.getBoolString(after.Mentionable)),
		})
	}

	return changes
}

func (l *Logger) getBoolString(b bool) string {
	if b {
		return "有効"
	}
	return "無効"
}