package commands

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/Sumire-Labs/Luna/embed"
)

type EmbedBuilderCommand struct{}

func NewEmbedBuilderCommand() *EmbedBuilderCommand {
	return &EmbedBuilderCommand{}
}

func (c *EmbedBuilderCommand) Name() string {
	return "embed"
}

func (c *EmbedBuilderCommand) Description() string {
	return "カスタム埋め込みメッセージを作成します"
}

func (c *EmbedBuilderCommand) Usage() string {
	return "/embed"
}

func (c *EmbedBuilderCommand) Category() string {
	return "ユーティリティ"
}

func (c *EmbedBuilderCommand) Aliases() []string {
	return []string{"埋め込み", "エンベッド"}
}

func (c *EmbedBuilderCommand) Permission() int64 {
	return discordgo.PermissionSendMessages
}

func (c *EmbedBuilderCommand) Options() []*discordgo.ApplicationCommandOption {
	return []*discordgo.ApplicationCommandOption{}
}

func (c *EmbedBuilderCommand) Execute(ctx *Context) error {
	return c.showMainMenu(ctx)
}

func (c *EmbedBuilderCommand) showMainMenu(ctx *Context) error {
	embedBuilder := embed.New().
		SetTitle("📝 埋め込みビルダー").
		SetDescription("作成したい埋め込みの種類を選択してください").
		SetColor(embed.M3Colors.Primary).
		AddField("🎨 カスタム埋め込み", "自由にデザインできる埋め込み", true).
		AddField("📋 テンプレート", "事前定義されたデザイン", true).
		AddField("✏️ 編集機能", "既存の埋め込みを編集", true).
		SetFooter("ボタンをクリックして開始", "")

	components := []discordgo.MessageComponent{
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.Button{
					Style:    discordgo.PrimaryButton,
					Label:    "🎨 カスタム作成",
					CustomID: "embed_main_custom",
				},
				discordgo.Button{
					Style:    discordgo.SecondaryButton,
					Label:    "📋 テンプレート",
					CustomID: "embed_main_template",
				},
				discordgo.Button{
					Style:    discordgo.SecondaryButton,
					Label:    "✏️ 埋め込み編集",
					CustomID: "embed_main_edit",
				},
			},
		},
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.Button{
					Style:    discordgo.SuccessButton,
					Label:    "📚 ヘルプ",
					CustomID: "embed_main_help",
				},
				discordgo.Button{
					Style:    discordgo.SecondaryButton,
					Label:    "🎨 カラーガイド",
					CustomID: "embed_main_colors",
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


