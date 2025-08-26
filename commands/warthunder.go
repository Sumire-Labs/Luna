package commands

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/Sumire-Labs/Luna/embed"
	"github.com/Sumire-Labs/Luna/models"
	"github.com/Sumire-Labs/Luna/services"
)

type WarThunderCommand struct {
	wtService *services.WarThunderService
}

func NewWarThunderCommand(wtService *services.WarThunderService) *WarThunderCommand {
	return &WarThunderCommand{
		wtService: wtService,
	}
}

func (cmd *WarThunderCommand) Name() string {
	return "wt"
}

func (cmd *WarThunderCommand) Description() string {
	return "War Thunder BR ルーレット - 陸/空/海のBRルーレットを実行"
}

func (cmd *WarThunderCommand) Usage() string {
	return "/wt [game_mode] [min_br] [max_br]"
}

func (cmd *WarThunderCommand) Category() string {
	return "ゲーム"
}

func (cmd *WarThunderCommand) Aliases() []string {
	return []string{"warthunder", "wtbr"}
}

func (cmd *WarThunderCommand) Permission() int64 {
	return 0 // Everyone can use this command
}

func (cmd *WarThunderCommand) Options() []*discordgo.ApplicationCommandOption {
	return []*discordgo.ApplicationCommandOption{
		{
			Type:        discordgo.ApplicationCommandOptionString,
			Name:        "mode",
			Description: "ゲームモード (air/ground/naval)",
			Required:    false,
			Choices: []*discordgo.ApplicationCommandOptionChoice{
				{Name: "🛩️ 空軍 (Air)", Value: "air"},
				{Name: "🚗 陸軍 (Ground)", Value: "ground"},
				{Name: "🚢 海軍 (Naval)", Value: "naval"},
			},
		},
		{
			Type:        discordgo.ApplicationCommandOptionNumber,
			Name:        "min_br",
			Description: "最小BR (例: 1.0)",
			Required:    false,
		},
		{
			Type:        discordgo.ApplicationCommandOptionNumber,
			Name:        "max_br",
			Description: "最大BR (例: 13.0)",
			Required:    false,
		},
		{
			Type:        discordgo.ApplicationCommandOptionString,
			Name:        "action",
			Description: "アクション",
			Required:    false,
			Choices: []*discordgo.ApplicationCommandOptionChoice{
				{Name: "🎲 ルーレット実行", Value: "spin"},
				{Name: "⚙️ 設定", Value: "config"},
				{Name: "📊 統計", Value: "stats"},
			},
		},
	}
}

func (cmd *WarThunderCommand) Execute(ctx *Context) error {
	// Get arguments
	modeStr := ctx.GetStringArg("mode")
	minBRArg, _ := ctx.GetArg("min_br")
	maxBRArg, _ := ctx.GetArg("max_br")
	action := ctx.GetStringArg("action")

	if action == "" {
		action = "spin"
	}

	user := ctx.GetUser()
	guildID := ctx.GetGuild()

	switch action {
	case "spin":
		return cmd.handleSpin(ctx, user.ID, guildID, modeStr, minBRArg, maxBRArg)
	case "config":
		return cmd.handleConfig(ctx, user.ID, guildID, modeStr)
	case "stats":
		return cmd.handleStats(ctx)
	default:
		return cmd.handleSpin(ctx, user.ID, guildID, modeStr, minBRArg, maxBRArg)
	}
}

func (cmd *WarThunderCommand) handleSpin(ctx *Context, userID, guildID, modeStr string, minBRArg, maxBRArg interface{}) error {
	// Default to Air if no mode specified
	gameMode := models.GameModeAir
	if modeStr != "" {
		switch strings.ToLower(modeStr) {
		case "air":
			gameMode = models.GameModeAir
		case "ground":
			gameMode = models.GameModeGround
		case "naval":
			gameMode = models.GameModeNaval
		default:
			return ctx.ReplyEphemeral("❌ 無効なゲームモードです。air, ground, naval のいずれかを指定してください。")
		}
	}

	// Get user's roulette configuration
	config, err := cmd.wtService.GetRouletteConfig(userID, guildID, gameMode)
	if err != nil {
		return ctx.ReplyEphemeral("❌ 設定の取得に失敗しました: " + err.Error())
	}

	// Override BR range if specified
	if minBRArg != nil {
		if minBR, ok := minBRArg.(float64); ok {
			config.MinBR = minBR
		}
	}
	if maxBRArg != nil {
		if maxBR, ok := maxBRArg.(float64); ok {
			config.MaxBR = maxBR
		}
	}

	// Validate BR range
	if config.MinBR > config.MaxBR {
		return ctx.ReplyEphemeral("❌ 最小BRが最大BRより大きくなっています。")
	}

	// Defer reply for potentially long operation
	if err := ctx.DeferReply(false); err != nil {
		return err
	}

	// Spin the roulette
	result, err := cmd.wtService.SpinRoulette(config)
	if err != nil {
		return ctx.EditReply("❌ ルーレットの実行に失敗しました: " + err.Error())
	}

	// Create result embed
	embed := cmd.createResultEmbed(result)
	
	// Add spin again button
	components := []discordgo.MessageComponent{
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.Button{
					CustomID: fmt.Sprintf("wt_spin_%s_%.1f_%.1f", gameMode.String(), config.MinBR, config.MaxBR),
					Label:    "🎲 もう一回",
					Style:    discordgo.PrimaryButton,
					Emoji:    &discordgo.ComponentEmoji{Name: "🎲"},
				},
				discordgo.Button{
					CustomID: fmt.Sprintf("wt_config_%s", gameMode.String()),
					Label:    "⚙️ 設定",
					Style:    discordgo.SecondaryButton,
					Emoji:    &discordgo.ComponentEmoji{Name: "⚙️"},
				},
			},
		},
	}

	_, err = ctx.Session.InteractionResponseEdit(ctx.Interaction.Interaction, &discordgo.WebhookEdit{
		Embeds:     &[]*discordgo.MessageEmbed{embed},
		Components: &components,
	})

	return err
}

func (cmd *WarThunderCommand) handleConfig(ctx *Context, userID, guildID, modeStr string) error {
	return ctx.ReplyEphemeral("⚙️ 設定機能は実装中です。現在は基本的なルーレット機能のみ利用可能です。")
}

func (cmd *WarThunderCommand) handleStats(ctx *Context) error {
	airCount, err := cmd.wtService.GetVehicleCount(&[]models.GameMode{models.GameModeAir}[0])
	if err != nil {
		airCount = 0
	}

	groundCount, err := cmd.wtService.GetVehicleCount(&[]models.GameMode{models.GameModeGround}[0])
	if err != nil {
		groundCount = 0
	}

	navalCount, err := cmd.wtService.GetVehicleCount(&[]models.GameMode{models.GameModeNaval}[0])
	if err != nil {
		navalCount = 0
	}

	totalCount, err := cmd.wtService.GetVehicleCount(nil)
	if err != nil {
		totalCount = 0
	}

	statsEmbed := embed.New().
		SetTitle("📊 War Thunder データベース統計").
		SetColor(0x4285F4).
		AddField("🛩️ 航空機", fmt.Sprintf("%d 機", airCount), true).
		AddField("🚗 地上車両", fmt.Sprintf("%d 両", groundCount), true).
		AddField("🚢 艦艇", fmt.Sprintf("%d 隻", navalCount), true).
		AddField("📊 総計", fmt.Sprintf("%d 車両", totalCount), false).
		SetFooter("Luna Bot - War Thunder BR ルーレット", "").
		Build()

	return ctx.ReplyEmbed(statsEmbed)
}

func (cmd *WarThunderCommand) createResultEmbed(result *models.RouletteResult) *discordgo.MessageEmbed {
	vehicle := result.Vehicle
	
	// Create vehicle type icon
	typeIcon := cmd.getVehicleTypeIcon(vehicle.VehicleType)
	
	// Create special markers
	var markers []string
	if vehicle.Premium {
		markers = append(markers, "⭐ プレミアム")
	}
	if vehicle.Squadron {
		markers = append(markers, "🏆 スコードロン")
	}
	if vehicle.Event {
		markers = append(markers, "🎃 イベント")
	}
	
	markerText := ""
	if len(markers) > 0 {
		markerText = strings.Join(markers, " • ")
	}

	title := fmt.Sprintf("🎲 %s ルーレット結果", result.Config.GameMode.Emoji())
	
	builder := embed.New().
		SetTitle(title).
		SetColor(cmd.getGameModeColor(result.Config.GameMode)).
		AddField("🎯 車両名", fmt.Sprintf("**%s**", vehicle.Name), false).
		AddField("⚡ Battle Rating", fmt.Sprintf("**%.1f**", vehicle.BR), true).
		AddField("🏛️ 国籍", fmt.Sprintf("%s %s", vehicle.Nation.Flag(), string(vehicle.Nation)), true).
		AddField("📝 車両種別", fmt.Sprintf("%s %s", typeIcon, string(vehicle.VehicleType)), true).
		AddField("🔢 ランク", fmt.Sprintf("**%s**", romanNumeral(vehicle.Rank)), true)

	if markerText != "" {
		builder.AddField("🏷️ 特殊", markerText, false)
	}

	// Add BR range info
	builder.AddField("📊 設定BR範囲", fmt.Sprintf("%.1f - %.1f", result.Config.MinBR, result.Config.MaxBR), true)

	builder.SetFooter("Luna Bot - War Thunder BR ルーレット", "")

	// Set image if available
	if vehicle.ImageURL != "" {
		builder.SetImage(vehicle.ImageURL)
	}

	return builder.Build()
}

func (cmd *WarThunderCommand) getVehicleTypeIcon(vType models.VehicleType) string {
	switch vType {
	case models.VehicleTypeFighter:
		return "✈️"
	case models.VehicleTypeAttacker:
		return "🚁"
	case models.VehicleTypeBomber:
		return "🛩️"
	case models.VehicleTypeHelicopter:
		return "🚁"
	case models.VehicleTypeLightTank:
		return "🏎️"
	case models.VehicleTypeMediumTank:
		return "🚗"
	case models.VehicleTypeHeavyTank:
		return "🚛"
	case models.VehicleTypeTankDestroyer:
		return "💥"
	case models.VehicleTypeSPAA:
		return "🎯"
	case models.VehicleTypeSPG:
		return "💣"
	case models.VehicleTypeBoat:
		return "🛥️"
	case models.VehicleTypeFrigate:
		return "⛵"
	case models.VehicleTypeDestroyer:
		return "🚢"
	case models.VehicleTypeCruiser:
		return "🛳️"
	case models.VehicleTypeBattleship:
		return "⚓"
	case models.VehicleTypeCarrier:
		return "🛫"
	case models.VehicleTypeSubmarine:
		return "🚤"
	default:
		return "❓"
	}
}

func (cmd *WarThunderCommand) getGameModeColor(gameMode models.GameMode) int {
	switch gameMode {
	case models.GameModeAir:
		return 0x87CEEB // Sky blue
	case models.GameModeGround:
		return 0x8B4513 // Saddle brown
	case models.GameModeNaval:
		return 0x1E90FF // Dodger blue
	default:
		return 0x4285F4 // Default blue
	}
}

func romanNumeral(num int) string {
	switch num {
	case 1:
		return "I"
	case 2:
		return "II"
	case 3:
		return "III"
	case 4:
		return "IV"
	case 5:
		return "V"
	case 6:
		return "VI"
	case 7:
		return "VII"
	case 8:
		return "VIII"
	default:
		return strconv.Itoa(num)
	}
}

