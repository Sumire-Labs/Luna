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
	return "かっこ使用量ランキングを表示"
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
				SetTitle("📊 かっこ使用量ランキング").
				SetColor(0xFF6B6B).
				SetDescription("まだデータがありません。\nメッセージにかっこを使ってみましょう！").
				Build(),
		)
	}

	// Build ranking embed
	embedBuilder := embed.New().
		SetTitle("📊 かっこ使用量ランキング TOP10").
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

		// Format stats
		balance := stats.OpenBrackets - stats.CloseBrackets
		balanceStr := ""
		if balance > 0 {
			balanceStr = fmt.Sprintf(" ⚠️ (+%d)", balance)
		} else if balance < 0 {
			balanceStr = fmt.Sprintf(" ⚠️ (%d)", balance)
		} else {
			balanceStr = " ✅"
		}

		description.WriteString(fmt.Sprintf(
			"%s **%s**\n   ( %d回  ) %d回  計: **%d回**%s\n\n",
			medal, username,
			stats.OpenBrackets, stats.CloseBrackets,
			stats.TotalBrackets, balanceStr,
		))
	}

	embedBuilder.SetDescription(description.String())

	// Get requesting user's rank
	requestingUserID := ctx.GetUser().ID
	userStats, _ := cmd.db.GetUserBracketStats(guildID, requestingUserID)
	
	if userStats != nil && userStats.TotalBrackets > 0 {
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
				fmt.Sprintf("**%d位** - 合計 %d回 ( %d回 ) %d回",
					userRank, userStats.TotalBrackets,
					userStats.OpenBrackets, userStats.CloseBrackets),
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

	if stats.TotalBrackets == 0 {
		return ctx.ReplyEmbed(
			embed.New().
				SetTitle(fmt.Sprintf("📊 %s のかっこ統計", user.Username)).
				SetColor(0xFF6B6B).
				SetDescription("まだデータがありません").
				SetThumbnail(user.AvatarURL("256")).
				Build(),
		)
	}

	// Calculate balance and ratio
	balance := stats.OpenBrackets - stats.CloseBrackets
	balanceStatus := "✅ 完璧なバランス！"
	balanceColor := 0x4CAF50 // Green
	
	if balance > 0 {
		balanceStatus = fmt.Sprintf("⚠️ 開きかっこが %d 個多い", balance)
		balanceColor = 0xFFC107 // Amber
	} else if balance < 0 {
		balanceStatus = fmt.Sprintf("⚠️ 閉じかっこが %d 個多い", -balance)
		balanceColor = 0xFF9800 // Orange
	}

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
		SetTitle(fmt.Sprintf("📊 %s のかっこ統計", user.Username)).
		SetColor(balanceColor).
		SetThumbnail(user.AvatarURL("256"))

	// Add fields
	embedBuilder.
		AddField("🏆 順位", fmt.Sprintf("**%d位** / %d人中", userRank, len(rankings)), true).
		AddField("📈 合計使用回数", fmt.Sprintf("**%d回**", stats.TotalBrackets), true).
		AddField("⚖️ バランス", balanceStatus, false)

	// Add detailed stats
	embedBuilder.AddField(
		"📊 詳細統計",
		fmt.Sprintf("開きかっこ `(` `（`: **%d回**\n閉じかっこ `)` `）`: **%d回**",
			stats.OpenBrackets, stats.CloseBrackets),
		false,
	)

	// Add fun facts
	avgPercentage := 0.0
	if stats.TotalBrackets > 0 {
		avgPercentage = float64(stats.OpenBrackets) / float64(stats.TotalBrackets) * 100
	}
	
	funFact := ""
	if balance == 0 && stats.TotalBrackets > 100 {
		funFact = "🎯 100回以上使って完璧なバランス！素晴らしい！"
	} else if stats.TotalBrackets > 500 {
		funFact = "🔥 500回以上のかっこ使用！かっこマスター！"
	} else if stats.TotalBrackets > 100 {
		funFact = "💪 100回以上のかっこ使用！かなりのヘビーユーザー！"
	} else if balance > 10 {
		funFact = "😅 閉じ忘れが多いかも？"
	} else if balance < -10 {
		funFact = "🤔 閉じかっこが多すぎるかも？"
	}

	if funFact != "" {
		embedBuilder.AddField("💡 コメント", funFact, false)
	}

	embedBuilder.SetFooter(
		fmt.Sprintf("開きかっこ率: %.1f%%", avgPercentage),
		"",
	)

	return ctx.ReplyEmbed(embedBuilder.Build())
}