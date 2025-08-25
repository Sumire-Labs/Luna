package commands

import (
	"github.com/bwmarrin/discordgo"
	"github.com/Sumire-Labs/Luna/embed"
)

type ConfigCommand struct{}

func NewConfigCommand() *ConfigCommand {
	return &ConfigCommand{}
}

func (c *ConfigCommand) Name() string {
	return "config"
}

func (c *ConfigCommand) Description() string {
	return "サーバーのボット機能を設定します"
}

func (c *ConfigCommand) Usage() string {
	return "/config"
}

func (c *ConfigCommand) Category() string {
	return "管理"
}

func (c *ConfigCommand) Aliases() []string {
	return []string{"設定", "セットアップ"}
}

func (c *ConfigCommand) Permission() int64 {
	return discordgo.PermissionManageGuild
}

func (c *ConfigCommand) Options() []*discordgo.ApplicationCommandOption {
	return []*discordgo.ApplicationCommandOption{}
}

func (c *ConfigCommand) Execute(ctx *Context) error {
	if ctx.GetGuild() == "" {
		return ctx.ReplyEphemeral("❌ このコマンドはサーバー内でのみ使用できます！")
	}

	// Check permissions
	member, err := ctx.Session.GuildMember(ctx.GetGuild(), ctx.GetUser().ID)
	if err != nil {
		return ctx.ReplyEphemeral("❌ 権限の確認に失敗しました！")
	}

	hasPermission := false
	for _, roleID := range member.Roles {
		role, err := ctx.Session.State.Role(ctx.GetGuild(), roleID)
		if err != nil {
			continue
		}
		if role.Permissions&discordgo.PermissionManageGuild != 0 || role.Permissions&discordgo.PermissionAdministrator != 0 {
			hasPermission = true
			break
		}
	}

	if !hasPermission {
		return ctx.ReplyEphemeral("❌ このコマンドを使用するには**サーバー管理**権限が必要です！")
	}

	return c.showMainMenu(ctx)
}

func (c *ConfigCommand) showMainMenu(ctx *Context) error {
	embedBuilder := embed.New().
		SetTitle("⚙️ サーバー設定パネル").
		SetDescription("設定したい機能を選択してください").
		SetColor(embed.M3Colors.Primary).
		AddField("🎫 チケットシステム", "サポートチケット機能", true).
		AddField("🛡️ モデレーション", "自動管理機能", true).
		AddField("👋 ウェルカム", "新メンバー歓迎機能", true).
		AddField("📝 ログ", "サーバーログ機能", true).
		AddField("🔔 Bump通知", "DISBOARD Bump通知", true).
		AddBlankField(true).
		SetFooter("ボタンをクリックして設定を開始", "")

	components := []discordgo.MessageComponent{
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.Button{
					Style:    discordgo.PrimaryButton,
					Label:    "🎫 チケット設定",
					CustomID: "config_main_tickets",
				},
				discordgo.Button{
					Style:    discordgo.SecondaryButton,
					Label:    "🛡️ モデレーション設定",
					CustomID: "config_main_moderation",
				},
				discordgo.Button{
					Style:    discordgo.SecondaryButton,
					Label:    "👋 ウェルカム設定",
					CustomID: "config_main_welcome",
				},
			},
		},
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.Button{
					Style:    discordgo.SecondaryButton,
					Label:    "📝 ログ設定",
					CustomID: "config_main_logging",
				},
				discordgo.Button{
					Style:    discordgo.SecondaryButton,
					Label:    "🔔 Bump設定",
					CustomID: "config_main_bump",
				},
			},
		},
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.Button{
					Style:    discordgo.SuccessButton,
					Label:    "📋 設定確認",
					CustomID: "config_main_view",
				},
				discordgo.Button{
					Style:    discordgo.DangerButton,
					Label:    "🗑️ 設定リセット",
					CustomID: "config_main_reset",
				},
			},
		},
	}

	return ctx.Session.InteractionRespond(ctx.Interaction.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds:     []*discordgo.MessageEmbed{embedBuilder.Build()},
			Components: components,
			Flags:      discordgo.MessageFlagsEphemeral,
		},
	})
}

func (c *ConfigCommand) getFeatureName(feature string) string {
	names := map[string]string{
		"tickets":    "🎫 チケットシステム",
		"moderation": "🛡️ モデレーション",
		"welcome":    "👋 ウェルカムシステム",
		"logging":    "📝 ログシステム",
		"all":        "🔄 全設定",
	}
	
	if name, ok := names[feature]; ok {
		return name
	}
	return feature
}