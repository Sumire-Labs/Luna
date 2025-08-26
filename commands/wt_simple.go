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
	
	// Get random BR
	selectedBR, err := cmd.wtService.GetRandomBR(gameMode, minBR, maxBR)
	if err != nil {
		return ctx.ReplyEphemeral(fmt.Sprintf("❌ エラー: %s", err.Error()))
	}
	
	// Create result embed
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
	
	return ctx.Session.InteractionRespond(ctx.Interaction.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds:     []*discordgo.MessageEmbed{resultEmbed},
			Components: components,
		},
	})
}

func (cmd *WTCommand) createResultEmbed(gameMode services.GameMode, br, minBR, maxBR float64) *discordgo.MessageEmbed {
	color := cmd.getGameModeColor(gameMode)
	
	// Select appropriate GIF based on game mode and BR level
	gifURL := cmd.getResultGIF(gameMode, br)
	
	builder := embed.New().
		SetTitle(fmt.Sprintf("%s War Thunder BR ルーレット", gameMode.Emoji())).
		SetColor(color).
		SetDescription(fmt.Sprintf("# **%.1f**", br))
	
	// Add GIF if available
	if gifURL != "" {
		builder.SetImage(gifURL)
	}
	
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


func (cmd *WTCommand) getResultGIF(gameMode services.GameMode, br float64) string {
	// Return animated GIFs based on game mode and BR
	switch gameMode {
	case services.GameModeAir:
		if br >= 10.0 {
			return "https://media.giphy.com/media/3oEjI1erPMTMBFmNHi/giphy.gif" // Jet fighter
		} else if br >= 5.0 {
			return "https://media.giphy.com/media/l0HlD7sTICn3X5Jf2/giphy.gif" // WW2 fighter
		}
		return "https://media.giphy.com/media/3o7TKUZfJKUKuSWgZG/giphy.gif" // Biplane
		
	case services.GameModeGround:
		if br >= 8.0 {
			return "https://media.giphy.com/media/3o7TKqm1mNujcBPSpy/giphy.gif" // Modern tank
		}
		return "https://media.giphy.com/media/xT9IgLbNugVohGx8Bi/giphy.gif" // WW2 tank
		
	case services.GameModeNaval:
		return "https://media.giphy.com/media/xUOwGi5bbHxbT1XncA/giphy.gif" // Battleship
		
	default:
		return ""
	}
}

