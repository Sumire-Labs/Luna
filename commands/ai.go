package commands

import (
	"bytes"
	"context"
	"fmt"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/Sumire-Labs/Luna/ai"
	"github.com/Sumire-Labs/Luna/embed"
)

type AICommand struct {
	aiService      *ai.Service
	geminiStudio   *ai.GeminiStudioService
	vertexGemini   *ai.VertexGeminiService
}

func NewAICommand(aiService *ai.Service) *AICommand {
	return &AICommand{
		aiService: aiService,
	}
}

func NewAICommandWithStudio(geminiStudio *ai.GeminiStudioService) *AICommand {
	return &AICommand{
		geminiStudio: geminiStudio,
	}
}

func NewAICommandWithVertex(vertexGemini *ai.VertexGeminiService) *AICommand {
	return &AICommand{
		vertexGemini: vertexGemini,
	}
}

func (c *AICommand) Name() string {
	return "ask"
}

func (c *AICommand) Description() string {
	return "Gemini AIに質問して回答を得ます"
}

func (c *AICommand) Usage() string {
	return "/ask <質問>"
}

func (c *AICommand) Category() string {
	return "AI"
}

func (c *AICommand) Aliases() []string {
	return []string{"質問", "gemini"}
}

func (c *AICommand) Permission() int64 {
	return discordgo.PermissionSendMessages
}

func (c *AICommand) Options() []*discordgo.ApplicationCommandOption {
	return []*discordgo.ApplicationCommandOption{
		{
			Type:        discordgo.ApplicationCommandOptionString,
			Name:        "question",
			Description: "AIに聞きたい質問",
			Required:    true,
		},
	}
}

func (c *AICommand) Execute(ctx *Context) error {
	// オプションから質問を取得
	question := ctx.GetStringArg("question")
	
	if question == "" {
		return ctx.ReplyEphemeral("❌ 質問を入力してください")
	}
	
	// AIサービスが利用可能かチェック
	if c.aiService == nil && c.geminiStudio == nil && c.vertexGemini == nil {
		return ctx.ReplyEphemeral("❌ AI機能は現在利用できません（設定を確認してください）")
	}
	
	// 処理中メッセージ
	ctx.DeferReply(false)
	
	// Geminiに質問
	aiCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	
	var answer string
	var err error
	
	// 利用可能なサービスの優先順位で実行
	if c.vertexGemini != nil {
		// 新しいVertex AI Gemini APIを優先
		answer, err = c.vertexGemini.AskGemini(aiCtx, question, ctx.GetUser().ID)
	} else if c.geminiStudio != nil {
		// Google AI Studio API
		answer, err = c.geminiStudio.AskGemini(aiCtx, question, ctx.GetUser().ID)
	} else {
		// 旧Vertex AI Predict API（非推奨）
		answer, err = c.aiService.AskGemini(aiCtx, question, ctx.GetUser().ID)
	}
	if err != nil {
		errorEmbed := embed.New().
			SetTitle("❌ エラーが発生しました").
			SetDescription(fmt.Sprintf("AIからの応答取得に失敗しました: %v", err)).
			SetColor(embed.M3Colors.Error).
			SetFooter("時間をおいて再度お試しください", "")
		
		return ctx.EditReplyEmbed(errorEmbed.Build())
	}
	
	// 回答が長すぎる場合は切り詰める
	if len(answer) > 1900 {
		answer = answer[:1900] + "...\n\n*（回答が長すぎるため省略されました）*"
	}
	
	// 成功応答
	responseEmbed := embed.New().
		SetTitle("🤖 Gemini AI の回答").
		SetColor(embed.M3Colors.Primary).
		AddField("💬 質問", question, false).
		AddField("📝 回答", answer, false).
		SetFooter(fmt.Sprintf("回答者: %s • Model: Gemini 2.5 Pro", ctx.GetUser().Username), ctx.GetUser().AvatarURL(""))
	
	return ctx.EditReplyEmbed(responseEmbed.Build())
}

// ImageCommand は画像生成コマンドです
type ImageCommand struct {
	aiService *ai.Service
}

func NewImageCommand(aiService *ai.Service) *ImageCommand {
	return &ImageCommand{
		aiService: aiService,
	}
}

func (c *ImageCommand) Name() string {
	return "imagine"
}

func (c *ImageCommand) Description() string {
	return "Imagen AIで画像を生成します"
}

func (c *ImageCommand) Usage() string {
	return "/imagine <プロンプト>"
}

func (c *ImageCommand) Category() string {
	return "AI"
}

func (c *ImageCommand) Aliases() []string {
	return []string{"画像生成", "imagen"}
}

func (c *ImageCommand) Permission() int64 {
	return discordgo.PermissionSendMessages
}

func (c *ImageCommand) Options() []*discordgo.ApplicationCommandOption {
	return []*discordgo.ApplicationCommandOption{
		{
			Type:        discordgo.ApplicationCommandOptionString,
			Name:        "prompt",
			Description: "生成したい画像の説明",
			Required:    true,
		},
		{
			Type:        discordgo.ApplicationCommandOptionString,
			Name:        "style",
			Description: "画像のスタイル",
			Required:    false,
			Choices: []*discordgo.ApplicationCommandOptionChoice{
				{Name: "🎨 アート", Value: "artistic"},
				{Name: "📷 写実的", Value: "photorealistic"},
				{Name: "🖼️ アニメ", Value: "anime"},
				{Name: "🎮 ゲーム", Value: "game"},
				{Name: "✏️ スケッチ", Value: "sketch"},
			},
		},
	}
}

func (c *ImageCommand) Execute(ctx *Context) error {
	// オプションからプロンプトとスタイルを取得
	prompt := ctx.GetStringArg("prompt")
	style := ctx.GetStringArg("style")
	
	if prompt == "" {
		return ctx.ReplyEphemeral("❌ 画像の説明を入力してください")
	}
	
	// AIサービスが利用可能かチェック
	if c.aiService == nil {
		return ctx.ReplyEphemeral("❌ AI画像生成機能は現在利用できません（設定を確認してください）")
	}
	
	// スタイルをプロンプトに追加
	fullPrompt := prompt
	switch style {
	case "artistic":
		fullPrompt += ", artistic style, masterpiece"
	case "photorealistic":
		fullPrompt += ", photorealistic, high quality photo, 8k resolution"
	case "anime":
		fullPrompt += ", anime style, manga art, japanese animation"
	case "game":
		fullPrompt += ", game art, concept art, digital painting"
	case "sketch":
		fullPrompt += ", pencil sketch, hand drawn, black and white"
	}
	
	// 処理中メッセージ
	ctx.DeferReply(false)
	
	// 生成開始メッセージ
	startEmbed := embed.New().
		SetTitle("🎨 画像生成中...").
		SetDescription("AIが画像を生成しています。しばらくお待ちください...").
		SetColor(embed.M3Colors.Info).
		AddField("📝 プロンプト", prompt, false).
		SetFooter("生成には30秒〜1分程度かかる場合があります", "")
	
	ctx.EditReplyEmbed(startEmbed.Build())
	
	// Imagenで画像生成
	aiCtx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	
	imageData, err := c.aiService.GenerateImage(aiCtx, fullPrompt, ctx.GetUser().ID)
	if err != nil {
		errorEmbed := embed.New().
			SetTitle("❌ 画像生成に失敗しました").
			SetDescription(fmt.Sprintf("エラー: %v", err)).
			SetColor(embed.M3Colors.Error).
			AddField("💡 ヒント", "プロンプトを変更して再度お試しください", false).
			SetFooter("画像生成は複雑なプロンプトで失敗することがあります", "")
		
		return ctx.EditReplyEmbed(errorEmbed.Build())
	}
	
	// 画像をDiscordにアップロード
	file := &discordgo.File{
		Name:        fmt.Sprintf("imagen_%s.png", time.Now().Format("20060102_150405")),
		ContentType: "image/png",
		Reader:      bytes.NewReader(imageData),
	}
	
	// 成功応答
	successEmbed := embed.New().
		SetTitle("✨ 画像生成完了！").
		SetColor(embed.M3Colors.Success).
		AddField("📝 プロンプト", prompt, false).
		SetImage(fmt.Sprintf("attachment://%s", file.Name)).
		SetFooter(fmt.Sprintf("生成者: %s • Model: Imagen 4", ctx.GetUser().Username), ctx.GetUser().AvatarURL(""))
	
	if style != "" {
		successEmbed.AddField("🎨 スタイル", getStyleName(style), true)
	}
	
	// ファイル付きの応答編集はWebhookEditを使う必要がある
	_, err = ctx.Session.InteractionResponseEdit(ctx.Interaction.Interaction, &discordgo.WebhookEdit{
		Embeds: &[]*discordgo.MessageEmbed{successEmbed.Build()},
		Files:  []*discordgo.File{file},
	})
	
	return err
}

func getStyleName(style string) string {
	styles := map[string]string{
		"artistic":       "🎨 アート",
		"photorealistic": "📷 写実的",
		"anime":          "🖼️ アニメ",
		"game":           "🎮 ゲーム",
		"sketch":         "✏️ スケッチ",
	}
	
	if name, ok := styles[style]; ok {
		return name
	}
	return style
}