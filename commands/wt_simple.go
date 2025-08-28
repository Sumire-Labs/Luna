package commands

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/Sumire-Labs/Luna/embed"
	"github.com/Sumire-Labs/Luna/services"
)

type WTCommand struct {
	wtService *services.WarThunderSimpleService
}

func NewWTCommand() *WTCommand {
	return &WTCommand{
		wtService: services.NewWarThunderSimpleService(),
	}
}

func (cmd *WTCommand) Name() string {
	return "br"
}

func (cmd *WTCommand) Description() string {
	return "War Thunder BR ルーレット"
}

func (cmd *WTCommand) Usage() string {
	return "/br"
}

func (cmd *WTCommand) Category() string {
	return "ゲーム"
}

func (cmd *WTCommand) Aliases() []string {
	return []string{}
}

func (cmd *WTCommand) Permission() int64 {
	return 0 // Everyone can use
}

func (cmd *WTCommand) Options() []*discordgo.ApplicationCommandOption {
	return []*discordgo.ApplicationCommandOption{}
}

func (cmd *WTCommand) Execute(ctx *Context) error {
	// Create initial selection embed
	initialEmbed := embed.New().
		SetTitle("🎮 War Thunder BR ルーレット").
		SetColor(0x4285F4).
		SetDescription("ゲームモードを選択してBRルーレットを回しましょう！\n\n除外したいBRがある場合は、先に「BR除外設定」ボタンで設定してください。").
		AddField("🛩️ 空軍", "BR 1.0 - 14.0", true).
		AddField("🚗 陸軍", "BR 1.0 - 12.0", true).
		AddField("🚢 海軍", "BR 1.0 - 8.7", true).
		SetFooter("モードを選択後、ルーレットが回転します！", "").
		Build()
	
	// Create selection components
	components := []discordgo.MessageComponent{
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.Button{
					CustomID: "br_mode_air",
					Label:    "空軍",
					Style:    discordgo.PrimaryButton,
					Emoji:    &discordgo.ComponentEmoji{Name: "🛩️"},
				},
				discordgo.Button{
					CustomID: "br_mode_ground",
					Label:    "陸軍",
					Style:    discordgo.PrimaryButton,
					Emoji:    &discordgo.ComponentEmoji{Name: "🚗"},
				},
				discordgo.Button{
					CustomID: "br_mode_naval",
					Label:    "海軍",
					Style:    discordgo.PrimaryButton,
					Emoji:    &discordgo.ComponentEmoji{Name: "🚢"},
				},
			},
		},
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.Button{
					CustomID: "br_exclude_settings",
					Label:    "BR除外設定",
					Style:    discordgo.SecondaryButton,
					Emoji:    &discordgo.ComponentEmoji{Name: "⚙️"},
				},
			},
		},
	}
	
	// Send initial embed with components
	return ctx.ReplyWithComponents(initialEmbed, components)
}

func (cmd *WTCommand) createResultEmbed(gameMode services.GameMode, br, minBR, maxBR float64) *discordgo.MessageEmbed {
	color := cmd.getGameModeColor(gameMode)
	
	// Use spinning roulette GIF for all results
	gifURL := "https://media.giphy.com/media/3oEjI67Egb456McTgQ/giphy.gif"
	
	builder := embed.New().
		SetTitle(fmt.Sprintf("%s War Thunder BR ルーレット", gameMode.Emoji())).
		SetColor(color).
		SetDescription(fmt.Sprintf("# **%.1f**", br)).
		SetImage(gifURL)
	
	// Add footer with range info if custom
	defaultMin, defaultMax := cmd.wtService.GetDefaultBRRange(gameMode)
	if minBR != defaultMin || maxBR != defaultMax {
		builder.SetFooter(fmt.Sprintf("BR範囲: %.1f - %.1f", minBR, maxBR), "")
	}
	
	return builder.Build()
}

func (cmd *WTCommand) getGameModeColor(gameMode services.GameMode) int {
	switch gameMode {
	case services.GameModeAir:
		return 0x87CEEB // Sky blue
	case services.GameModeGround:
		return 0x8B4513 // Saddle brown
	case services.GameModeNaval:
		return 0x1E90FF // Dodger blue
	default:
		return 0x4285F4 // Default blue
	}
}



