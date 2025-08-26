package commands

import (
	"fmt"
	"time"

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
	return "wt"
}

func (cmd *WTCommand) Description() string {
	return "War Thunder BR ルーレット"
}

func (cmd *WTCommand) Usage() string {
	return "/wt [mode] [min_br] [max_br]"
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
	return []*discordgo.ApplicationCommandOption{
		{
			Type:        discordgo.ApplicationCommandOptionString,
			Name:        "mode",
			Description: "ゲームモード",
			Required:    false,
			Choices: []*discordgo.ApplicationCommandOptionChoice{
				{Name: "🛩️ 空軍", Value: "air"},
				{Name: "🚗 陸軍", Value: "ground"},
				{Name: "🚢 海軍", Value: "naval"},
			},
		},
		{
			Type:        discordgo.ApplicationCommandOptionNumber,
			Name:        "min_br",
			Description: "最小BR",
			Required:    false,
		},
		{
			Type:        discordgo.ApplicationCommandOptionNumber,
			Name:        "max_br",
			Description: "最大BR",
			Required:    false,
		},
	}
}

func (cmd *WTCommand) Execute(ctx *Context) error {
	// Get arguments
	modeStr := ctx.GetStringArg("mode")
	if modeStr == "" {
		modeStr = "ground" // Default to ground
	}
	
	gameMode := services.GameMode(modeStr)
	
	// Get default BR range for the mode
	defaultMin, defaultMax := cmd.wtService.GetDefaultBRRange(gameMode)
	
	// Override if specified
	minBR := defaultMin
	maxBR := defaultMax
	
	if minArg, ok := ctx.GetArg("min_br"); ok {
		if min, ok := minArg.(float64); ok {
			minBR = min
		}
	}
	
	if maxArg, ok := ctx.GetArg("max_br"); ok {
		if max, ok := maxArg.(float64); ok {
			maxBR = max
		}
	}
	
	// Validate BR range
	if minBR > maxBR {
		return ctx.ReplyEphemeral("❌ 最小BRが最大BRより大きくなっています")
	}
	
	// Defer reply for roulette spin animation
	if err := ctx.DeferReply(false); err != nil {
		return err
	}
	
	// Show spinning roulette animation first
	spinningEmbed := embed.New().
		SetTitle(fmt.Sprintf("%s War Thunder BR ルーレット", gameMode.Emoji())).
		SetColor(cmd.getGameModeColor(gameMode)).
		SetDescription("🎰 **ルーレット回転中...** 🎰").
		SetImage("https://media.giphy.com/media/3oEjI67Egb456McTgQ/giphy.gif"). // Spinning wheel GIF
		Build()
	
	// Update with spinning animation
	_, err := ctx.Session.InteractionResponseEdit(ctx.Interaction.Interaction, &discordgo.WebhookEdit{
		Embeds: &[]*discordgo.MessageEmbed{spinningEmbed},
	})
	if err != nil {
		return err
	}
	
	// Wait for dramatic effect
	time.Sleep(3 * time.Second)
	
	// Get random BR
	selectedBR, err := cmd.wtService.GetRandomBR(gameMode, minBR, maxBR)
	if err != nil {
		return ctx.EditReply(fmt.Sprintf("❌ エラー: %s", err.Error()))
	}
	
	// Create final result embed
	resultEmbed := cmd.createResultEmbed(gameMode, selectedBR, minBR, maxBR)
	
	// Create simple spin again button
	components := []discordgo.MessageComponent{
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.Button{
					CustomID: fmt.Sprintf("wt_spin_%s_%.1f_%.1f", gameMode, minBR, maxBR),
					Label:    "もう一回",
					Style:    discordgo.PrimaryButton,
					Emoji:    &discordgo.ComponentEmoji{Name: "🎲"},
				},
			},
		},
	}
	
	// Update with final result
	_, err = ctx.Session.InteractionResponseEdit(ctx.Interaction.Interaction, &discordgo.WebhookEdit{
		Embeds:     &[]*discordgo.MessageEmbed{resultEmbed},
		Components: &components,
	})
	
	return err
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



