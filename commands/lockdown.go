package commands

import (
	"fmt"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/Sumire-Labs/Luna/embed"
)

type LockdownCommand struct{}

func NewLockdownCommand() *LockdownCommand {
	return &LockdownCommand{}
}

func (c *LockdownCommand) Name() string {
	return "lockdown"
}

func (c *LockdownCommand) Description() string {
	return "緊急時にチャンネルをロックダウンします"
}

func (c *LockdownCommand) Usage() string {
	return "/lockdown <action> [target] [reason]"
}

func (c *LockdownCommand) Category() string {
	return "管理"
}

func (c *LockdownCommand) Aliases() []string {
	return []string{"lock", "緊急", "ロック"}
}

func (c *LockdownCommand) Permission() int64 {
	return discordgo.PermissionManageChannels
}

func (c *LockdownCommand) Options() []*discordgo.ApplicationCommandOption {
	return []*discordgo.ApplicationCommandOption{
		{
			Type:        discordgo.ApplicationCommandOptionString,
			Name:        "action",
			Description: "実行するアクション",
			Required:    true,
			Choices: []*discordgo.ApplicationCommandOptionChoice{
				{Name: "🔒 ロック", Value: "lock"},
				{Name: "🔓 ロック解除", Value: "unlock"},
				{Name: "❄️ サーバー凍結", Value: "freeze"},
				{Name: "🔥 凍結解除", Value: "unfreeze"},
			},
		},
		{
			Type:        discordgo.ApplicationCommandOptionString,
			Name:        "target",
			Description: "対象範囲",
			Required:    false,
			Choices: []*discordgo.ApplicationCommandOptionChoice{
				{Name: "📍 このチャンネルのみ", Value: "current"},
				{Name: "📁 このカテゴリー", Value: "category"},
				{Name: "🌐 全チャンネル", Value: "all"},
			},
		},
		{
			Type:        discordgo.ApplicationCommandOptionString,
			Name:        "reason",
			Description: "ロックダウンの理由",
			Required:    false,
			MaxLength:   200,
		},
		{
			Type:        discordgo.ApplicationCommandOptionInteger,
			Name:        "duration",
			Description: "自動解除時間（分）",
			Required:    false,
			MinValue:    func() *float64 { v := 1.0; return &v }(),
			MaxValue:    1440, // 24時間
		},
	}
}

func (c *LockdownCommand) Execute(ctx *Context) error {
	if ctx.GetGuild() == "" {
		return ctx.ReplyEphemeral("❌ このコマンドはサーバー内でのみ使用できます！")
	}

	// 権限チェック
	member, err := ctx.Session.GuildMember(ctx.GetGuild(), ctx.GetUser().ID)
	if err != nil {
		return ctx.ReplyEphemeral("❌ 権限の確認に失敗しました")
	}

	if !c.hasPermission(ctx.Session, ctx.GetGuild(), member) {
		return ctx.ReplyEphemeral("❌ このコマンドを使用するには**チャンネル管理**権限が必要です！")
	}

	action := ctx.GetStringArg("action")
	target := ctx.GetStringArg("target")
	reason := ctx.GetStringArg("reason")
	duration := ctx.GetIntArg("duration")

	if target == "" {
		target = "current"
	}
	if reason == "" {
		reason = "管理者による緊急対応"
	}

	// 処理中メッセージ
	ctx.DeferReply(false)

	switch action {
	case "lock":
		return c.executeLock(ctx, target, reason, int(duration))
	case "unlock":
		return c.executeUnlock(ctx, target, reason)
	case "freeze":
		return c.executeFreeze(ctx, reason, int(duration))
	case "unfreeze":
		return c.executeUnfreeze(ctx, reason)
	default:
		return ctx.EditReply("❌ 不正なアクションです")
	}
}

func (c *LockdownCommand) hasPermission(s *discordgo.Session, guildID string, member *discordgo.Member) bool {
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
		if role.Permissions&discordgo.PermissionManageChannels != 0 || 
		   role.Permissions&discordgo.PermissionAdministrator != 0 {
			return true
		}
	}
	return false
}

func (c *LockdownCommand) executeLock(ctx *Context, target, reason string, duration int) error {
	channels, err := c.getTargetChannels(ctx, target)
	if err != nil {
		return ctx.EditReply(fmt.Sprintf("❌ チャンネル取得に失敗: %v", err))
	}

	successCount := 0
	failedChannels := []string{}

	for _, channel := range channels {
		if c.lockChannel(ctx.Session, channel, reason) {
			successCount++
		} else {
			failedChannels = append(failedChannels, channel.Name)
		}
	}

	// 結果メッセージ
	resultEmbed := embed.New().
		SetTitle("🔒 ロックダウン実行完了").
		SetColor(embed.M3Colors.Error).
		SetTimestamp()

	if successCount > 0 {
		resultEmbed.AddField("✅ ロック成功", fmt.Sprintf("%d個のチャンネル", successCount), true)
	}

	if len(failedChannels) > 0 {
		failedList := strings.Join(failedChannels, ", ")
		if len(failedList) > 1000 {
			failedList = failedList[:1000] + "..."
		}
		resultEmbed.AddField("❌ ロック失敗", failedList, true)
	}

	resultEmbed.AddField("📋 理由", reason, false)

	if duration > 0 {
		unlockTime := time.Now().Add(time.Duration(duration) * time.Minute)
		resultEmbed.AddField("⏰ 自動解除", fmt.Sprintf("<t:%d:R>", unlockTime.Unix()), true)
		
		// 自動解除をスケジュール
		go c.scheduleUnlock(ctx, target, duration)
	}

	resultEmbed.SetFooter(fmt.Sprintf("実行者: %s", ctx.GetUser().Username), ctx.GetUser().AvatarURL("64"))

	return ctx.EditReplyEmbed(resultEmbed.Build())
}

func (c *LockdownCommand) executeUnlock(ctx *Context, target, reason string) error {
	channels, err := c.getTargetChannels(ctx, target)
	if err != nil {
		return ctx.EditReply(fmt.Sprintf("❌ チャンネル取得に失敗: %v", err))
	}

	successCount := 0
	failedChannels := []string{}

	for _, channel := range channels {
		if c.unlockChannel(ctx.Session, channel) {
			successCount++
		} else {
			failedChannels = append(failedChannels, channel.Name)
		}
	}

	// 結果メッセージ
	resultEmbed := embed.New().
		SetTitle("🔓 ロック解除完了").
		SetColor(embed.M3Colors.Success).
		SetTimestamp()

	if successCount > 0 {
		resultEmbed.AddField("✅ 解除成功", fmt.Sprintf("%d個のチャンネル", successCount), true)
	}

	if len(failedChannels) > 0 {
		failedList := strings.Join(failedChannels, ", ")
		if len(failedList) > 1000 {
			failedList = failedList[:1000] + "..."
		}
		resultEmbed.AddField("❌ 解除失敗", failedList, true)
	}

	resultEmbed.AddField("📋 理由", reason, false)
	resultEmbed.SetFooter(fmt.Sprintf("実行者: %s", ctx.GetUser().Username), ctx.GetUser().AvatarURL("64"))

	return ctx.EditReplyEmbed(resultEmbed.Build())
}

func (c *LockdownCommand) executeFreeze(ctx *Context, reason string, duration int) error {
	guild, err := ctx.Session.Guild(ctx.GetGuild())
	if err != nil {
		return ctx.EditReply("❌ サーバー情報の取得に失敗しました")
	}

	// @everyone ロールの権限を変更してサーバーを凍結
	everyoneRole := c.findEveryoneRole(guild.Roles)
	if everyoneRole == nil {
		return ctx.EditReply("❌ @everyone ロールが見つかりません")
	}

	// 権限を大幅に制限
	newPermissions := everyoneRole.Permissions
	newPermissions &^= discordgo.PermissionSendMessages
	newPermissions &^= discordgo.PermissionAddReactions

	_, err = ctx.Session.GuildRoleEdit(ctx.GetGuild(), everyoneRole.ID, &discordgo.RoleParams{
		Permissions: &newPermissions,
	})

	if err != nil {
		return ctx.EditReply(fmt.Sprintf("❌ サーバー凍結に失敗: %v", err))
	}

	// 結果メッセージ
	resultEmbed := embed.New().
		SetTitle("❄️ サーバー凍結完了").
		SetDescription("サーバーが緊急凍結されました。一般メンバーの発言・リアクションが制限されています。").
		SetColor(embed.M3Colors.Error).
		AddField("📋 理由", reason, false).
		SetTimestamp()

	if duration > 0 {
		unfreezeTime := time.Now().Add(time.Duration(duration) * time.Minute)
		resultEmbed.AddField("⏰ 自動解除", fmt.Sprintf("<t:%d:R>", unfreezeTime.Unix()), true)
		
		// 自動解除をスケジュール
		go c.scheduleUnfreeze(ctx, duration, everyoneRole.Permissions)
	}

	resultEmbed.SetFooter(fmt.Sprintf("実行者: %s", ctx.GetUser().Username), ctx.GetUser().AvatarURL("64"))

	return ctx.EditReplyEmbed(resultEmbed.Build())
}

func (c *LockdownCommand) executeUnfreeze(ctx *Context, reason string) error {
	guild, err := ctx.Session.Guild(ctx.GetGuild())
	if err != nil {
		return ctx.EditReply("❌ サーバー情報の取得に失敗しました")
	}

	everyoneRole := c.findEveryoneRole(guild.Roles)
	if everyoneRole == nil {
		return ctx.EditReply("❌ @everyone ロールが見つかりません")
	}

	// 権限を復元
	newPermissions := everyoneRole.Permissions
	newPermissions |= discordgo.PermissionSendMessages
	newPermissions |= discordgo.PermissionAddReactions

	_, err = ctx.Session.GuildRoleEdit(ctx.GetGuild(), everyoneRole.ID, &discordgo.RoleParams{
		Permissions: &newPermissions,
	})

	if err != nil {
		return ctx.EditReply(fmt.Sprintf("❌ サーバー凍結解除に失敗: %v", err))
	}

	resultEmbed := embed.New().
		SetTitle("🔥 サーバー凍結解除完了").
		SetDescription("サーバーの凍結が解除されました。通常の活動が再開できます。").
		SetColor(embed.M3Colors.Success).
		AddField("📋 理由", reason, false).
		SetTimestamp().
		SetFooter(fmt.Sprintf("実行者: %s", ctx.GetUser().Username), ctx.GetUser().AvatarURL("64"))

	return ctx.EditReplyEmbed(resultEmbed.Build())
}

func (c *LockdownCommand) getTargetChannels(ctx *Context, target string) ([]*discordgo.Channel, error) {
	switch target {
	case "current":
		channel, err := ctx.Session.Channel(ctx.GetChannel())
		if err != nil {
			return nil, err
		}
		return []*discordgo.Channel{channel}, nil

	case "category":
		currentChannel, err := ctx.Session.Channel(ctx.GetChannel())
		if err != nil {
			return nil, err
		}
		
		channels, err := ctx.Session.GuildChannels(ctx.GetGuild())
		if err != nil {
			return nil, err
		}
		
		var categoryChannels []*discordgo.Channel
		for _, ch := range channels {
			if ch.ParentID == currentChannel.ParentID && ch.Type == discordgo.ChannelTypeGuildText {
				categoryChannels = append(categoryChannels, ch)
			}
		}
		return categoryChannels, nil

	case "all":
		channels, err := ctx.Session.GuildChannels(ctx.GetGuild())
		if err != nil {
			return nil, err
		}
		
		var textChannels []*discordgo.Channel
		for _, ch := range channels {
			if ch.Type == discordgo.ChannelTypeGuildText {
				textChannels = append(textChannels, ch)
			}
		}
		return textChannels, nil

	default:
		return nil, fmt.Errorf("不正なターゲット: %s", target)
	}
}

func (c *LockdownCommand) lockChannel(s *discordgo.Session, channel *discordgo.Channel, reason string) bool {
	// @everyone の送信権限を拒否
	err := s.ChannelPermissionSet(channel.ID, channel.GuildID, discordgo.PermissionOverwriteTypeRole, 0, discordgo.PermissionSendMessages)
	return err == nil
}

func (c *LockdownCommand) unlockChannel(s *discordgo.Session, channel *discordgo.Channel) bool {
	// @everyone の送信権限拒否を削除
	err := s.ChannelPermissionDelete(channel.ID, channel.GuildID)
	return err == nil
}

func (c *LockdownCommand) findEveryoneRole(roles []*discordgo.Role) *discordgo.Role {
	for _, role := range roles {
		if role.Name == "@everyone" {
			return role
		}
	}
	return nil
}

func (c *LockdownCommand) scheduleUnlock(ctx *Context, target string, duration int) {
	time.Sleep(time.Duration(duration) * time.Minute)
	
	// 自動解除を実行
	c.executeUnlock(ctx, target, "自動解除（スケジュール）")
}

func (c *LockdownCommand) scheduleUnfreeze(ctx *Context, duration int, originalPermissions int64) {
	time.Sleep(time.Duration(duration) * time.Minute)
	
	// 自動解除を実行
	c.executeUnfreeze(ctx, "自動解除（スケジュール）")
}