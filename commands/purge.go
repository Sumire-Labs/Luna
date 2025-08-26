package commands

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/Sumire-Labs/Luna/embed"
)

type PurgeCommand struct{}

func NewPurgeCommand() *PurgeCommand {
	return &PurgeCommand{}
}

func (c *PurgeCommand) Name() string {
	return "purge"
}

func (c *PurgeCommand) Description() string {
	return "メッセージを一括削除します"
}

func (c *PurgeCommand) Usage() string {
	return "/purge <amount> [user] [filter]"
}

func (c *PurgeCommand) Category() string {
	return "管理"
}

func (c *PurgeCommand) Aliases() []string {
	return []string{"clear", "delete", "削除"}
}

func (c *PurgeCommand) Permission() int64 {
	return discordgo.PermissionManageMessages
}

func (c *PurgeCommand) Options() []*discordgo.ApplicationCommandOption {
	return []*discordgo.ApplicationCommandOption{
		{
			Type:        discordgo.ApplicationCommandOptionInteger,
			Name:        "amount",
			Description: "削除するメッセージの数（1-100）",
			Required:    true,
			MinValue:    func() *float64 { v := 1.0; return &v }(),
			MaxValue:    100,
		},
		{
			Type:        discordgo.ApplicationCommandOptionUser,
			Name:        "user",
			Description: "特定のユーザーのメッセージのみを削除",
			Required:    false,
		},
		{
			Type:        discordgo.ApplicationCommandOptionString,
			Name:        "filter",
			Description: "削除フィルター",
			Required:    false,
			Choices: []*discordgo.ApplicationCommandOptionChoice{
				{Name: "🤖 BOTのメッセージのみ", Value: "bots"},
				{Name: "👤 人間のメッセージのみ", Value: "humans"},
				{Name: "🔗 リンク付きメッセージ", Value: "links"},
				{Name: "📎 添付ファイル付き", Value: "attachments"},
				{Name: "💬 埋め込み付き", Value: "embeds"},
				{Name: "📌 ピン留め以外", Value: "not-pinned"},
			},
		},
		{
			Type:        discordgo.ApplicationCommandOptionString,
			Name:        "contains",
			Description: "特定の文字列を含むメッセージのみを削除",
			Required:    false,
			MaxLength:   100,
		},
		{
			Type:        discordgo.ApplicationCommandOptionBoolean,
			Name:        "include_pinned",
			Description: "ピン留めメッセージも削除するか（デフォルト: false）",
			Required:    false,
		},
	}
}

func (c *PurgeCommand) Execute(ctx *Context) error {
	if ctx.GetGuild() == "" {
		return ctx.ReplyEphemeral("❌ このコマンドはサーバー内でのみ使用できます！")
	}

	// 権限チェック
	member, err := ctx.Session.GuildMember(ctx.GetGuild(), ctx.GetUser().ID)
	if err != nil {
		return ctx.ReplyEphemeral("❌ 権限の確認に失敗しました")
	}

	if !c.hasPermission(ctx.Session, ctx.GetGuild(), member) {
		return ctx.ReplyEphemeral("❌ このコマンドを使用するには**メッセージ管理**権限が必要です！")
	}

	amount := int(ctx.GetIntArg("amount"))
	targetUser := ctx.GetUserArg("user")
	filter := ctx.GetStringArg("filter")
	contains := ctx.GetStringArg("contains")
	includePinned := ctx.GetBoolArg("include_pinned")

	// 処理中メッセージ（すぐに削除される可能性があるので短めに）
	ctx.DeferReply(true)

	// メッセージを取得
	messages, err := ctx.Session.ChannelMessages(ctx.GetChannel(), amount+50, "", "", "")
	if err != nil {
		return ctx.EditReply("❌ メッセージの取得に失敗しました")
	}

	// フィルタリング
	var messagesToDelete []*discordgo.Message
	for _, msg := range messages {
		if len(messagesToDelete) >= amount {
			break
		}

		// 14日以上前のメッセージはスキップ（Discord制限）
		if time.Since(msg.Timestamp) > 14*24*time.Hour {
			continue
		}

		// フィルタリング条件をチェック
		if !c.matchesFilter(msg, targetUser, filter, contains, includePinned) {
			continue
		}

		messagesToDelete = append(messagesToDelete, msg)
	}

	if len(messagesToDelete) == 0 {
		return ctx.EditReply("❌ 削除条件に一致するメッセージが見つかりませんでした")
	}

	// バルク削除の実行
	deletedCount, err := c.bulkDeleteMessages(ctx.Session, ctx.GetChannel(), messagesToDelete)
	if err != nil {
		return ctx.EditReply(fmt.Sprintf("❌ メッセージ削除中にエラーが発生しました: %v", err))
	}

	// 削除完了メッセージ
	resultEmbed := embed.New().
		SetTitle("🧹 メッセージ削除完了").
		SetColor(embed.M3Colors.Success).
		AddField("🗑️ 削除数", fmt.Sprintf("%d件のメッセージ", deletedCount), true).
		SetTimestamp()

	// フィルター情報
	filterInfo := c.getFilterDescription(targetUser, filter, contains, includePinned)
	if filterInfo != "" {
		resultEmbed.AddField("🔍 削除条件", filterInfo, false)
	}

	resultEmbed.SetFooter(fmt.Sprintf("実行者: %s", ctx.GetUser().Username), ctx.GetUser().AvatarURL("64"))

	// 5秒後に結果メッセージも削除
	go func() {
		time.Sleep(5 * time.Second)
		ctx.Session.ChannelMessageDelete(ctx.GetChannel(), "")
	}()

	return ctx.EditReplyEmbed(resultEmbed.Build())
}

func (c *PurgeCommand) hasPermission(s *discordgo.Session, guildID string, member *discordgo.Member) bool {
	// オーナーかどうかチェック
	guild, err := s.Guild(guildID)
	if err == nil && guild.OwnerID == member.User.ID {
		return true
	}

	// 権限チェック
	for _, roleID := range member.Roles {
		role, err := s.State.Role(guildID, roleID)
		if err != nil {
			continue
		}
		if role.Permissions&discordgo.PermissionManageMessages != 0 || 
		   role.Permissions&discordgo.PermissionAdministrator != 0 {
			return true
		}
	}
	return false
}

func (c *PurgeCommand) matchesFilter(msg *discordgo.Message, targetUser *discordgo.User, filter, contains string, includePinned bool) bool {
	// ピン留めメッセージのチェック
	if msg.Pinned && !includePinned {
		return false
	}

	// 特定ユーザーのフィルター
	if targetUser != nil && msg.Author.ID != targetUser.ID {
		return false
	}

	// テキスト検索フィルター
	if contains != "" && !strings.Contains(strings.ToLower(msg.Content), strings.ToLower(contains)) {
		return false
	}

	// カテゴリーフィルター
	switch filter {
	case "bots":
		return msg.Author.Bot
	case "humans":
		return !msg.Author.Bot
	case "links":
		return strings.Contains(msg.Content, "http://") || 
			   strings.Contains(msg.Content, "https://") ||
			   strings.Contains(msg.Content, "discord.gg/")
	case "attachments":
		return len(msg.Attachments) > 0
	case "embeds":
		return len(msg.Embeds) > 0
	case "not-pinned":
		return !msg.Pinned
	}

	return true
}

func (c *PurgeCommand) bulkDeleteMessages(s *discordgo.Session, channelID string, messages []*discordgo.Message) (int, error) {
	if len(messages) == 0 {
		return 0, nil
	}

	var messageIDs []string
	var singleDeletes []*discordgo.Message

	for _, msg := range messages {
		// 2週間以内のメッセージは bulk delete
		if time.Since(msg.Timestamp) < 14*24*time.Hour {
			messageIDs = append(messageIDs, msg.ID)
		} else {
			singleDeletes = append(singleDeletes, msg)
		}
	}

	deletedCount := 0

	// Bulk delete (2-100件)
	if len(messageIDs) >= 2 {
		// Discord APIの制限で一度に100件まで
		for i := 0; i < len(messageIDs); i += 100 {
			end := i + 100
			if end > len(messageIDs) {
				end = len(messageIDs)
			}

			batch := messageIDs[i:end]
			if len(batch) >= 2 {
				err := s.ChannelMessagesBulkDelete(channelID, batch)
				if err == nil {
					deletedCount += len(batch)
				}
			} else if len(batch) == 1 {
				// 1件の場合は個別削除
				err := s.ChannelMessageDelete(channelID, batch[0])
				if err == nil {
					deletedCount++
				}
			}
		}
	} else if len(messageIDs) == 1 {
		// 1件の場合は個別削除
		err := s.ChannelMessageDelete(channelID, messageIDs[0])
		if err == nil {
			deletedCount++
		}
	}

	// 古いメッセージは個別削除
	for _, msg := range singleDeletes {
		err := s.ChannelMessageDelete(channelID, msg.ID)
		if err == nil {
			deletedCount++
		}
		// レート制限を避けるために少し待機
		time.Sleep(100 * time.Millisecond)
	}

	return deletedCount, nil
}

func (c *PurgeCommand) getFilterDescription(targetUser *discordgo.User, filter, contains string, includePinned bool) string {
	var conditions []string

	if targetUser != nil {
		conditions = append(conditions, fmt.Sprintf("👤 ユーザー: %s", targetUser.Username))
	}

	switch filter {
	case "bots":
		conditions = append(conditions, "🤖 BOTのみ")
	case "humans":
		conditions = append(conditions, "👤 人間のみ")
	case "links":
		conditions = append(conditions, "🔗 リンク付き")
	case "attachments":
		conditions = append(conditions, "📎 添付ファイル付き")
	case "embeds":
		conditions = append(conditions, "💬 埋め込み付き")
	case "not-pinned":
		conditions = append(conditions, "📌 ピン留め以外")
	}

	if contains != "" {
		conditions = append(conditions, fmt.Sprintf("🔍 内容: '%s'", contains))
	}

	if includePinned {
		conditions = append(conditions, "📌 ピン留め含む")
	}

	if len(conditions) == 0 {
		return ""
	}

	return strings.Join(conditions, " • ")
}