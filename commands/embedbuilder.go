package commands

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/luna/luna-bot/embed"
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
	return []*discordgo.ApplicationCommandOption{
		{
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Name:        "create",
			Description: "新しい埋め込みを作成します",
		},
		{
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Name:        "edit",
			Description: "既存の埋め込みを編集します",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "message_id",
					Description: "編集する埋め込みメッセージのID",
					Required:    true,
				},
			},
		},
		{
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Name:        "template",
			Description: "テンプレートから埋め込みを作成します",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "type",
					Description: "テンプレートの種類",
					Required:    true,
					Choices: []*discordgo.ApplicationCommandOptionChoice{
						{Name: "📢 お知らせ", Value: "announcement"},
						{Name: "📋 ルール", Value: "rules"},
						{Name: "❓ FAQ", Value: "faq"},
						{Name: "🎉 イベント", Value: "event"},
						{Name: "⚠️ 警告", Value: "warning"},
					},
				},
			},
		},
	}
}

func (c *EmbedBuilderCommand) Execute(ctx *Context) error {
	if len(ctx.Options) == 0 {
		return ctx.ReplyEphemeral("サブコマンドを指定してください: `/embed create`, `/embed edit`, `/embed template`")
	}

	subCommand := ctx.Options[0]
	
	switch subCommand.Name {
	case "create":
		return c.handleCreateEmbed(ctx)
	case "edit":
		return c.handleEditEmbed(ctx, subCommand)
	case "template":
		return c.handleTemplateEmbed(ctx, subCommand)
	default:
		return ctx.ReplyEphemeral("❌ 不明なサブコマンドです")
	}
}

func (c *EmbedBuilderCommand) handleCreateEmbed(ctx *Context) error {
	modal := &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseModal,
		Data: &discordgo.InteractionResponseData{
			CustomID: "embed_create_modal",
			Title:    "📝 埋め込み作成",
			Components: []discordgo.MessageComponent{
				&discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						&discordgo.TextInput{
							CustomID:    "embed_title",
							Label:       "タイトル",
							Style:       discordgo.TextInputShort,
							Placeholder: "埋め込みのタイトルを入力...",
							Required:    false,
							MaxLength:   256,
						},
					},
				},
				&discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						&discordgo.TextInput{
							CustomID:    "embed_description",
							Label:       "説明",
							Style:       discordgo.TextInputParagraph,
							Placeholder: "埋め込みの説明を入力...",
							Required:    false,
							MaxLength:   4000,
						},
					},
				},
				&discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						&discordgo.TextInput{
							CustomID:    "embed_color",
							Label:       "カラー (16進数 例: #6750A4 または 0x6750A4)",
							Style:       discordgo.TextInputShort,
							Placeholder: "#6750A4",
							Required:    false,
						},
					},
				},
				&discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						&discordgo.TextInput{
							CustomID:    "embed_image",
							Label:       "画像URL (オプション)",
							Style:       discordgo.TextInputShort,
							Placeholder: "https://example.com/image.png",
							Required:    false,
						},
					},
				},
				&discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						&discordgo.TextInput{
							CustomID:    "embed_footer",
							Label:       "フッター (オプション)",
							Style:       discordgo.TextInputShort,
							Placeholder: "フッターテキスト",
							Required:    false,
							MaxLength:   2048,
						},
					},
				},
			},
		},
	}

	return ctx.Session.InteractionRespond(ctx.Interaction.Interaction, modal)
}

func (c *EmbedBuilderCommand) handleEditEmbed(ctx *Context, subCommand *discordgo.ApplicationCommandInteractionDataOption) error {
	messageID := subCommand.Options[0].StringValue()
	
	// メッセージを取得して編集可能かチェック
	message, err := ctx.Session.ChannelMessage(ctx.GetChannel(), messageID)
	if err != nil {
		return ctx.ReplyEphemeral("❌ 指定されたメッセージが見つかりません")
	}
	
	if message.Author.ID != ctx.Session.State.User.ID {
		return ctx.ReplyEphemeral("❌ このボットが作成したメッセージのみ編集できます")
	}
	
	if len(message.Embeds) == 0 {
		return ctx.ReplyEphemeral("❌ 指定されたメッセージには埋め込みがありません")
	}
	
	currentEmbed := message.Embeds[0]
	
	modal := &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseModal,
		Data: &discordgo.InteractionResponseData{
			CustomID: fmt.Sprintf("embed_edit_modal_%s", messageID),
			Title:    "✏️ 埋め込み編集",
			Components: []discordgo.MessageComponent{
				&discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						&discordgo.TextInput{
							CustomID:    "embed_title",
							Label:       "タイトル",
							Style:       discordgo.TextInputShort,
							Placeholder: "埋め込みのタイトルを入力...",
							Required:    false,
							MaxLength:   256,
							Value:       currentEmbed.Title,
						},
					},
				},
				&discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						&discordgo.TextInput{
							CustomID:    "embed_description",
							Label:       "説明",
							Style:       discordgo.TextInputParagraph,
							Placeholder: "埋め込みの説明を入力...",
							Required:    false,
							MaxLength:   4000,
							Value:       currentEmbed.Description,
						},
					},
				},
				&discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						&discordgo.TextInput{
							CustomID:    "embed_color",
							Label:       "カラー (16進数 例: #6750A4 または 0x6750A4)",
							Style:       discordgo.TextInputShort,
							Placeholder: "#6750A4",
							Required:    false,
							Value:       fmt.Sprintf("#%06X", currentEmbed.Color),
						},
					},
				},
				&discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						&discordgo.TextInput{
							CustomID:    "embed_image",
							Label:       "画像URL (オプション)",
							Style:       discordgo.TextInputShort,
							Placeholder: "https://example.com/image.png",
							Required:    false,
							Value:       getImageURL(currentEmbed),
						},
					},
				},
				&discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						&discordgo.TextInput{
							CustomID:    "embed_footer",
							Label:       "フッター (オプション)",
							Style:       discordgo.TextInputShort,
							Placeholder: "フッターテキスト",
							Required:    false,
							MaxLength:   2048,
							Value:       getFooterText(currentEmbed),
						},
					},
				},
			},
		},
	}

	return ctx.Session.InteractionRespond(ctx.Interaction.Interaction, modal)
}

func (c *EmbedBuilderCommand) handleTemplateEmbed(ctx *Context, subCommand *discordgo.ApplicationCommandInteractionDataOption) error {
	templateType := subCommand.Options[0].StringValue()
	
	var embedBuilder *embed.Builder
	
	switch templateType {
	case "announcement":
		embedBuilder = embed.New().
			SetTitle("📢 重要なお知らせ").
			SetDescription("ここにお知らせ内容を記入してください。").
			SetColor(embed.M3Colors.Info).
			AddField("📅 日時", "YYYY/MM/DD HH:MM", true).
			AddField("👤 投稿者", "管理者", true).
			AddField("🔗 詳細", "詳細情報がある場合はここに", false)
			
	case "rules":
		embedBuilder = embed.New().
			SetTitle("📋 サーバールール").
			SetDescription("このサーバーを快適に利用するためのルールです。").
			SetColor(embed.M3Colors.Primary).
			AddField("1️⃣ 基本的なマナー", "他の参加者を尊重し、礼儀正しく行動してください。", false).
			AddField("2️⃣ スパム禁止", "不要なメッセージの連投は禁止です。", false).
			AddField("3️⃣ 適切なチャンネル使用", "各チャンネルの目的に沿った投稿をしてください。", false).
			SetFooter("ルール違反には警告・キック・BANの対象となります", "")
			
	case "faq":
		embedBuilder = embed.New().
			SetTitle("❓ よくある質問").
			SetDescription("頻繁にお問い合わせいただく質問をまとめました。").
			SetColor(embed.M3Colors.Info).
			AddField("Q1: ○○はどうすればいいですか？", "A1: ○○の方法について説明...", false).
			AddField("Q2: ○○ができません", "A2: ○○の対処法について...", false).
			AddField("Q3: その他の質問", "A3: サポートチャンネルでお気軽にお尋ねください", false)
			
	case "event":
		embedBuilder = embed.New().
			SetTitle("🎉 イベント開催のお知らせ").
			SetDescription("楽しいイベントを開催します！ぜひご参加ください。").
			SetColor(embed.M3Colors.Success).
			AddField("📅 開催日時", "YYYY/MM/DD HH:MM〜", true).
			AddField("📍 場所", "○○チャンネル", true).
			AddField("🎯 参加条件", "特になし（どなたでも参加可能）", false).
			AddField("🏆 景品", "参加者全員にプレゼント！", false).
			SetFooter("参加表明は下のボタンをクリック", "")
			
	case "warning":
		embedBuilder = embed.New().
			SetTitle("⚠️ 重要な警告").
			SetDescription("緊急かつ重要な情報です。必ずお読みください。").
			SetColor(embed.M3Colors.Warning).
			AddField("🚨 警告内容", "具体的な警告内容をここに記載", false).
			AddField("📋 対処方法", "推奨される対処方法について", false).
			AddField("📞 お問い合わせ", "不明な点があれば管理者までご連絡ください", false).
			SetFooter("この警告を確認したら反応してください", "")
			
	default:
		return ctx.ReplyEphemeral("❌ 不明なテンプレートタイプです")
	}
	
	// テンプレート埋め込みを送信
	response := &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{embedBuilder.Build()},
			Components: []discordgo.MessageComponent{
				&discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						&discordgo.Button{
							Style:    discordgo.SecondaryButton,
							Label:    "✏️ 編集",
							CustomID: fmt.Sprintf("template_edit_%s", templateType),
						},
						&discordgo.Button{
							Style:    discordgo.DangerButton,
							Label:    "🗑️ 削除",
							CustomID: "template_delete",
						},
					},
				},
			},
		},
	}
	
	return ctx.Session.InteractionRespond(ctx.Interaction.Interaction, response)
}

// ヘルパー関数
func getImageURL(embed *discordgo.MessageEmbed) string {
	if embed.Image != nil {
		return embed.Image.URL
	}
	return ""
}

func getFooterText(embed *discordgo.MessageEmbed) string {
	if embed.Footer != nil {
		return embed.Footer.Text
	}
	return ""
}

