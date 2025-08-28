package commands

import (
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/Sumire-Labs/Luna/database"
	"github.com/Sumire-Labs/Luna/embed"
)

type BracketsCommand struct {
	db *database.Service
}

func NewBracketsCommand(db *database.Service) *BracketsCommand {
	return &BracketsCommand{
		db: db,
	}
}

func (cmd *BracketsCommand) Name() string {
	return "brackets"
}

func (cmd *BracketsCommand) Description() string {
	return "かっこペア使用量ランキングを表示"
}

func (cmd *BracketsCommand) Usage() string {
	return "/brackets [user]"
}

func (cmd *BracketsCommand) Category() string {
	return "統計"
}

func (cmd *BracketsCommand) Aliases() []string {
	return []string{"kakko", "parentheses"}
}

func (cmd *BracketsCommand) Permission() int64 {
	return 0 // Everyone can use
}

func (cmd *BracketsCommand) Options() []*discordgo.ApplicationCommandOption {
	return []*discordgo.ApplicationCommandOption{
		{
			Type:        discordgo.ApplicationCommandOptionUser,
			Name:        "user",
			Description: "特定ユーザーの統計を表示",
			Required:    false,
		},
	}
}

func (cmd *BracketsCommand) Execute(ctx *Context) error {
	guildID := ctx.GetGuild()
	if guildID == "" {
		return ctx.ReplyEphemeral("❌ このコマンドはサーバー内でのみ使用できます")
	}

	// Check if specific user stats requested
	targetUser := ctx.GetUserArg("user")
	if targetUser != nil {
		return cmd.showUserStats(ctx, guildID, targetUser)
	}

	// Show ranking
	return cmd.showRanking(ctx, guildID)
}

func (cmd *BracketsCommand) showRanking(ctx *Context, guildID string) error {
	// Get top 10 rankings
	rankings, err := cmd.db.GetBracketRanking(guildID, 10)
	if err != nil {
		return ctx.ReplyEphemeral("❌ ランキングの取得に失敗しました")
	}

	if len(rankings) == 0 {
		return ctx.ReplyEmbed(
			embed.New().
				SetTitle("📊 かっこペア使用量ランキング").
				SetColor(0xFF6B6B).
				SetDescription("まだデータがありません。\nメッセージに完全なかっこペア () （） を使ってみましょう！").
				Build(),
		)
	}

	// Build ranking embed
	embedBuilder := embed.New().
		SetTitle("📊 かっこペア使用量ランキング TOP10").
		SetColor(0x4285F4).
		SetThumbnail("https://cdn.discordapp.com/attachments/123/456/brackets_icon.png")

	// Create ranking list
	var description strings.Builder
	medals := []string{"🥇", "🥈", "🥉"}
	
	for i, stats := range rankings {
		// Get user from Discord
		user, err := ctx.Session.User(stats.UserID)
		username := "Unknown User"
		if err == nil {
			username = user.Username
		}

		medal := ""
		if i < len(medals) {
			medal = medals[i]
		} else {
			medal = fmt.Sprintf("%d.", i+1)
		}

		// Format stats - show half-width and full-width pairs separately
		description.WriteString(fmt.Sprintf(
			"%s **%s**\n   () %d回  （） %d回  合計: **%d回**\n\n",
			medal, username,
			stats.HalfWidthPairs, stats.FullWidthPairs,
			stats.TotalPairs,
		))
	}

	embedBuilder.SetDescription(description.String())

	// Get requesting user's rank
	requestingUserID := ctx.GetUser().ID
	userStats, _ := cmd.db.GetUserBracketStats(guildID, requestingUserID)
	
	if userStats != nil && userStats.TotalPairs > 0 {
		// Find user's rank
		userRank := 0
		for i, stats := range rankings {
			if stats.UserID == requestingUserID {
				userRank = i + 1
				break
			}
		}

		if userRank == 0 {
			// User not in top 10, get their actual rank
			allRankings, _ := cmd.db.GetBracketRanking(guildID, 999)
			for i, stats := range allRankings {
				if stats.UserID == requestingUserID {
					userRank = i + 1
					break
				}
			}
		}

		if userRank > 10 {
			embedBuilder.AddField(
				"📍 あなたの順位",
				fmt.Sprintf("**%d位** - 合計 %d回 () %d回 （） %d回",
					userRank, userStats.TotalPairs,
					userStats.HalfWidthPairs, userStats.FullWidthPairs),
				false,
			)
		}
	}

	embedBuilder.SetFooter("💡 /brackets @ユーザー で個別統計を表示", "")

	return ctx.ReplyEmbed(embedBuilder.Build())
}

func (cmd *BracketsCommand) showUserStats(ctx *Context, guildID string, user *discordgo.User) error {
	stats, err := cmd.db.GetUserBracketStats(guildID, user.ID)
	if err != nil {
		return ctx.ReplyEphemeral("❌ ユーザー統計の取得に失敗しました")
	}

	if stats.TotalPairs == 0 {
		return ctx.ReplyEmbed(
			embed.New().
				SetTitle(fmt.Sprintf("📊 %s のかっこペア統計", user.Username)).
				SetColor(0xFF6B6B).
				SetDescription("まだデータがありません\n完全なかっこペア () （） を使ってみましょう！").
				SetThumbnail(user.AvatarURL("256")).
				Build(),
		)
	}

	// Since we only count complete pairs, there's no imbalance issue
	balanceColor := 0x4CAF50 // Green - always balanced

	// Get user's rank
	rankings, _ := cmd.db.GetBracketRanking(guildID, 999)
	userRank := 0
	for i, rankStats := range rankings {
		if rankStats.UserID == user.ID {
			userRank = i + 1
			break
		}
	}

	// Build stats embed
	embedBuilder := embed.New().
		SetTitle(fmt.Sprintf("📊 %s のかっこペア統計", user.Username)).
		SetColor(balanceColor).
		SetThumbnail(user.AvatarURL("256"))

	// Add fields
	embedBuilder.
		AddField("🏆 順位", fmt.Sprintf("**%d位** / %d人中", userRank, len(rankings)), true).
		AddField("📈 合計ペア数", fmt.Sprintf("**%d回**", stats.TotalPairs), true).
		AddField("✅ 状態", "完全ペア（バランス完璧！）", false)

	// Add detailed stats
	embedBuilder.AddField(
		"📊 詳細統計",
		fmt.Sprintf("半角かっこペア `()`: **%d回**\n全角かっこペア `（）`: **%d回**",
			stats.HalfWidthPairs, stats.FullWidthPairs),
		false,
	)

	// Add fun facts based on pair usage
	halfWidthPercentage := 0.0
	if stats.TotalPairs > 0 {
		halfWidthPercentage = float64(stats.HalfWidthPairs) / float64(stats.TotalPairs) * 100
	}
	
	funFact := ""
	if stats.TotalPairs > 500 {
		funFact = "🔥 500ペア以上！かっこペアマスター！"
	} else if stats.TotalPairs > 100 {
		funFact = "💪 100ペア以上！かなりのペアユーザー！"
	} else if stats.TotalPairs > 50 {
		funFact = "👍 50ペア達成！順調にペアを使っています！"
	}
	
	// Add preference comment
	if stats.HalfWidthPairs > stats.FullWidthPairs * 2 {
		if funFact != "" {
			funFact += "\n"
		}
		funFact += "📱 半角かっこ派ですね！"
	} else if stats.FullWidthPairs > stats.HalfWidthPairs * 2 {
		if funFact != "" {
			funFact += "\n"
		}
		funFact += "📝 全角かっこ派ですね！"
	} else if stats.HalfWidthPairs > 0 && stats.FullWidthPairs > 0 {
		if funFact != "" {
			funFact += "\n"
		}
		funFact += "🎯 両方バランス良く使っています！"
	}

	if funFact != "" {
		embedBuilder.AddField("💡 コメント", funFact, false)
	}

	embedBuilder.SetFooter(
		fmt.Sprintf("半角かっこ率: %.1f%%", halfWidthPercentage),
		"",
	)

	return ctx.ReplyEmbed(embedBuilder.Build())
}