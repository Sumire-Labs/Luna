package commands

import (
	"context"
	"fmt"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/Sumire-Labs/Luna/ai"
	"github.com/Sumire-Labs/Luna/embed"
)

type TranslateCommand struct {
	geminiStudio *ai.GeminiStudioService
}

func NewTranslateCommand(geminiStudio *ai.GeminiStudioService) *TranslateCommand {
	return &TranslateCommand{
		geminiStudio: geminiStudio,
	}
}

func (c *TranslateCommand) Name() string {
	return "translate"
}

func (c *TranslateCommand) Description() string {
	return "テキストを指定した言語に翻訳します"
}

func (c *TranslateCommand) Usage() string {
	return "/translate <テキスト> [言語]"
}

func (c *TranslateCommand) Category() string {
	return "ユーティリティ"
}

func (c *TranslateCommand) Aliases() []string {
	return []string{"翻訳", "tr"}
}

func (c *TranslateCommand) Permission() int64 {
	return discordgo.PermissionSendMessages
}

func (c *TranslateCommand) Options() []*discordgo.ApplicationCommandOption {
	return []*discordgo.ApplicationCommandOption{
		{
			Type:        discordgo.ApplicationCommandOptionString,
			Name:        "text",
			Description: "翻訳したいテキスト",
			Required:    true,
		},
		{
			Type:        discordgo.ApplicationCommandOptionString,
			Name:        "language",
			Description: "翻訳先の言語（デフォルト: 日本語）",
			Required:    false,
			Choices: []*discordgo.ApplicationCommandOptionChoice{
				{Name: "🇯🇵 日本語", Value: "japanese"},
				{Name: "🇺🇸 英語", Value: "english"},
				{Name: "🇰🇷 韓国語", Value: "korean"},
				{Name: "🇨🇳 中国語", Value: "chinese"},
				{Name: "🇪🇸 スペイン語", Value: "spanish"},
				{Name: "🇫🇷 フランス語", Value: "french"},
				{Name: "🇩🇪 ドイツ語", Value: "german"},
				{Name: "🇮🇹 イタリア語", Value: "italian"},
				{Name: "🇷🇺 ロシア語", Value: "russian"},
				{Name: "🇵🇹 ポルトガル語", Value: "portuguese"},
			},
		},
	}
}

func (c *TranslateCommand) Execute(ctx *Context) error {
	// オプションから情報を取得
	text := ctx.GetStringArg("text")
	language := ctx.GetStringArg("language")
	
	if text == "" {
		return ctx.ReplyEphemeral("❌ 翻訳するテキストを入力してください")
	}
	
	// デフォルト言語
	if language == "" {
		language = "japanese"
	}
	
	// AI サービスが利用可能かチェック
	if c.geminiStudio == nil {
		return ctx.ReplyEphemeral("❌ 翻訳機能は現在利用できません（Google AI Studio設定を確認してください）")
	}
	
	// 処理中メッセージ
	ctx.DeferReply(false)
	
	// 翻訳プロンプトを作成
	prompt := c.createTranslatePrompt(text, language)
	
	// Geminiで翻訳
	aiCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	
	translation, err := c.geminiStudio.AskGemini(aiCtx, prompt, ctx.GetUser().ID)
	if err != nil {
		errorEmbed := embed.New().
			SetTitle("❌ 翻訳エラー").
			SetDescription(fmt.Sprintf("翻訳処理中にエラーが発生しました: %v", err)).
			SetColor(embed.M3Colors.Error).
			SetFooter("Luna Translation", "")
		
		return ctx.EditReplyEmbed(errorEmbed.Build())
	}
	
	// 結果を表示
	resultEmbed := embed.New().
		SetTitle("🌐 翻訳結果").
		SetColor(embed.M3Colors.Primary).
		AddField("📝 原文", c.truncateText(text, 1000), false).
		AddField("🎯 翻訳先", c.getLanguageName(language), true).
		AddField("✨ 翻訳結果", translation, false).
		SetFooter(fmt.Sprintf("翻訳者: %s", ctx.GetUser().Username), ctx.GetUser().AvatarURL("64"))
	
	return ctx.EditReplyEmbed(resultEmbed.Build())
}

func (c *TranslateCommand) createTranslatePrompt(text, language string) string {
	langMap := map[string]string{
		"japanese":   "日本語",
		"english":    "英語",
		"korean":     "韓国語", 
		"chinese":    "中国語",
		"spanish":    "スペイン語",
		"french":     "フランス語",
		"german":     "ドイツ語",
		"italian":    "イタリア語",
		"russian":    "ロシア語",
		"portuguese": "ポルトガル語",
	}
	
	targetLang := langMap[language]
	if targetLang == "" {
		targetLang = "日本語"
	}
	
	return fmt.Sprintf(`以下のテキストを%sに翻訳してください。自然で読みやすい翻訳を心がけ、翻訳結果のみを返答してください。

翻訳するテキスト:
%s`, targetLang, text)
}

func (c *TranslateCommand) getLanguageName(language string) string {
	langMap := map[string]string{
		"japanese":   "🇯🇵 日本語",
		"english":    "🇺🇸 英語",
		"korean":     "🇰🇷 韓国語",
		"chinese":    "🇨🇳 中国語",
		"spanish":    "🇪🇸 スペイン語",
		"french":     "🇫🇷 フランス語",
		"german":     "🇩🇪 ドイツ語",
		"italian":    "🇮🇹 イタリア語",
		"russian":    "🇷🇺 ロシア語",
		"portuguese": "🇵🇹 ポルトガル語",
	}
	
	if name, ok := langMap[language]; ok {
		return name
	}
	return "🇯🇵 日本語"
}

func (c *TranslateCommand) truncateText(text string, maxLen int) string {
	if len(text) <= maxLen {
		return text
	}
	return text[:maxLen-3] + "..."
}