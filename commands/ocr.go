package commands

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/Sumire-Labs/Luna/ai"
	"github.com/Sumire-Labs/Luna/embed"
)

type OCRCommand struct {
	aiService *ai.GeminiStudioService
}

func NewOCRCommand(aiService *ai.GeminiStudioService) *OCRCommand {
	return &OCRCommand{
		aiService: aiService,
	}
}

func (c *OCRCommand) Name() string {
	return "ocr"
}

func (c *OCRCommand) Description() string {
	return "Luna AIを使って画像からテキストを抽出・分析します"
}

func (c *OCRCommand) Usage() string {
	return "/ocr <画像URL または 添付画像> [mode]"
}

func (c *OCRCommand) Category() string {
	return "AI"
}

func (c *OCRCommand) Aliases() []string {
	return []string{"文字認識", "テキスト抽出"}
}

func (c *OCRCommand) Permission() int64 {
	return discordgo.PermissionSendMessages
}

func (c *OCRCommand) Options() []*discordgo.ApplicationCommandOption {
	return []*discordgo.ApplicationCommandOption{
		{
			Type:        discordgo.ApplicationCommandOptionString,
			Name:        "image_url",
			Description: "テキストを抽出したい画像のURL",
			Required:    false,
		},
		{
			Type:        discordgo.ApplicationCommandOptionString,
			Name:        "mode",
			Description: "処理モード",
			Required:    false,
			Choices: []*discordgo.ApplicationCommandOptionChoice{
				{Name: "📄 テキスト抽出", Value: "text"},
				{Name: "🌐 翻訳", Value: "translate"},
				{Name: "📝 要約", Value: "summarize"},
				{Name: "🔍 詳細分析", Value: "analyze"},
			},
		},
		{
			Type:        discordgo.ApplicationCommandOptionAttachment,
			Name:        "image",
			Description: "テキストを抽出したい画像ファイル",
			Required:    false,
		},
	}
}

func (c *OCRCommand) Execute(ctx *Context) error {
	// AIサービスが利用可能かチェック
	if c.aiService == nil {
		return ctx.ReplyEphemeral("❌ OCR機能は現在利用できません（Google AI Studio設定を確認してください）")
	}
	
	// オプションから情報を取得
	imageURL := ctx.GetStringArg("image_url")
	mode := ctx.GetStringArg("mode")
	attachment := ctx.GetAttachmentArg("image")
	
	// デフォルトモード
	if mode == "" {
		mode = "text"
	}
	
	// 画像の取得方法を決定
	var finalImageURL string
	
	if attachment != nil {
		// 添付ファイルを優先
		finalImageURL = attachment.URL
	} else if imageURL != "" {
		// URLオプション
		finalImageURL = imageURL
	} else {
		// 直前のメッセージから画像を探す
		recentImageURL, err := c.findRecentImage(ctx)
		if err != nil {
			return ctx.ReplyEphemeral("❌ 画像が見つかりません。画像を添付するか、画像URLを指定してください。")
		}
		finalImageURL = recentImageURL
	}
	
	// 処理中メッセージ
	defer ctx.Session.InteractionRespond(ctx.Interaction.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
	})
	
	// 進行状況メッセージ
	progressEmbed := embed.New().
		SetTitle("🔍 画像解析中...").
		SetDescription("Gemini 2.5が画像からテキストを抽出しています...").
		SetColor(embed.M3Colors.Info).
		AddField("📸 画像URL", finalImageURL, false).
		AddField("🎯 処理モード", c.getModeDescription(mode), false).
		SetFooter("解析には10秒〜30秒程度かかる場合があります", "")
	
	ctx.Session.InteractionResponseEdit(ctx.Interaction.Interaction, &discordgo.WebhookEdit{
		Embeds: &[]*discordgo.MessageEmbed{progressEmbed.Build()},
	})
	
	// 画像をダウンロード
	aiCtx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()
	
	imageData, mimeType, err := ai.DownloadImage(aiCtx, finalImageURL)
	if err != nil {
		errorEmbed := embed.New().
			SetTitle("❌ 画像の取得に失敗しました").
			SetDescription(fmt.Sprintf("エラー: %v", err)).
			SetColor(embed.M3Colors.Error).
			AddField("💡 ヒント", "画像URLが正しいか、画像が公開されているか確認してください", false).
			AddField("📋 対応形式", strings.Join(ai.GetSupportedImageTypes(), ", "), false).
			SetFooter("ファイルサイズは20MB以下にしてください", "")
		
		_, _ = ctx.Session.InteractionResponseEdit(ctx.Interaction.Interaction, &discordgo.WebhookEdit{
			Embeds: &[]*discordgo.MessageEmbed{errorEmbed.Build()},
		})
		return nil
	}
	
	// OCR処理実行
	result, err := c.aiService.OCRWithGemini(aiCtx, imageData, mimeType, ctx.GetUser().ID, mode)
	if err != nil {
		errorEmbed := embed.New().
			SetTitle("❌ OCR処理に失敗しました").
			SetDescription(fmt.Sprintf("エラー: %v", err)).
			SetColor(embed.M3Colors.Error).
			AddField("💡 ヒント", "画像が鮮明でない、またはテキストが判読困難な可能性があります", false).
			SetFooter("別の画像で再度お試しください", "")
		
		_, _ = ctx.Session.InteractionResponseEdit(ctx.Interaction.Interaction, &discordgo.WebhookEdit{
			Embeds: &[]*discordgo.MessageEmbed{errorEmbed.Build()},
		})
		return nil
	}
	
	// 結果が長すぎる場合は分割
	if len(result) > 1800 {
		result = result[:1800] + "\n\n...\n*（結果が長すぎるため省略されました）*"
	}
	
	// 成功応答
	successEmbed := embed.New().
		SetTitle("✨ OCR処理完了！").
		SetColor(embed.M3Colors.Success).
		AddField("🎯 処理モード", c.getModeDescription(mode), true).
		AddField("📸 画像形式", mimeType, true).
		AddField("📄 抽出結果", result, false).
		SetFooter(fmt.Sprintf("処理者: %s • Model: Gemini 2.5 Pro", ctx.GetUser().Username), ctx.GetUser().AvatarURL(""))
	
	_, err = ctx.Session.InteractionResponseEdit(ctx.Interaction.Interaction, &discordgo.WebhookEdit{
		Embeds: &[]*discordgo.MessageEmbed{successEmbed.Build()},
	})
	
	return err
}

// findRecentImage は最近のメッセージから画像を探します
func (c *OCRCommand) findRecentImage(ctx *Context) (string, error) {
	// 最新20件のメッセージを取得
	messages, err := ctx.Session.ChannelMessages(ctx.GetChannel(), 20, "", "", "")
	if err != nil {
		return "", fmt.Errorf("メッセージの取得に失敗")
	}
	
	// 画像を含むメッセージを探す
	for _, msg := range messages {
		// 添付ファイルをチェック
		for _, attachment := range msg.Attachments {
			if strings.HasPrefix(attachment.ContentType, "image/") {
				return attachment.URL, nil
			}
		}
		
		// 埋め込みの画像をチェック
		for _, embed := range msg.Embeds {
			if embed.Image != nil {
				return embed.Image.URL, nil
			}
			if embed.Thumbnail != nil {
				return embed.Thumbnail.URL, nil
			}
		}
	}
	
	return "", fmt.Errorf("直近のメッセージに画像が見つかりませんでした")
}

// getModeDescription はモードの説明を返します
func (c *OCRCommand) getModeDescription(mode string) string {
	descriptions := map[string]string{
		"text":      "📄 テキスト抽出",
		"translate": "🌐 翻訳 (日本語)",
		"summarize": "📝 要約",
		"analyze":   "🔍 詳細分析",
	}
	
	if desc, ok := descriptions[mode]; ok {
		return desc
	}
	return "📄 テキスト抽出"
}