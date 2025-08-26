package commands

import (
	"fmt"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/Sumire-Labs/Luna/database"
	"github.com/Sumire-Labs/Luna/embed"
)

type ActivityCommand struct {
	db *database.Service
}

func NewActivityCommand(db *database.Service) *ActivityCommand {
	return &ActivityCommand{
		db: db,
	}
}

func (c *ActivityCommand) Name() string {
	return "activity"
}

func (c *ActivityCommand) Description() string {
	return "サーバーの活動統計を表示します"
}

func (c *ActivityCommand) Usage() string {
	return "/activity [期間]"
}

func (c *ActivityCommand) Category() string {
	return "統計"
}

func (c *ActivityCommand) Aliases() []string {
	return []string{"stats", "統計", "活動"}
}

func (c *ActivityCommand) Permission() int64 {
	return discordgo.PermissionSendMessages
}

func (c *ActivityCommand) Options() []*discordgo.ApplicationCommandOption {
	return []*discordgo.ApplicationCommandOption{
		{
			Type:        discordgo.ApplicationCommandOptionString,
			Name:        "period",
			Description: "統計期間（デフォルト: 7日間）",
			Required:    false,
			Choices: []*discordgo.ApplicationCommandOptionChoice{
				{Name: "📅 今日", Value: "today"},
				{Name: "📊 7日間", Value: "7days"},
				{Name: "📈 30日間", Value: "30days"},
				{Name: "📉 全期間", Value: "all"},
			},
		},
	}
}

func (c *ActivityCommand) Execute(ctx *Context) error {
	if ctx.GetGuild() == "" {
		return ctx.ReplyEphemeral("❌ このコマンドはサーバー内でのみ使用できます！")
	}
	
	// 期間を取得
	period := ctx.GetStringArg("period")
	if period == "" {
		period = "7days"
	}
	
	// 処理中メッセージ
	ctx.DeferReply(false)
	
	// 期間に応じた開始時刻を計算
	since := c.getPeriodStart(period)
	periodName := c.getPeriodName(period)
	
	// サーバー情報を取得
	guild, err := ctx.Session.Guild(ctx.GetGuild())
	if err != nil {
		return ctx.EditReply("❌ サーバー情報の取得に失敗しました")
	}
	
	// コマンド統計を取得
	commandStats, err := c.db.GetCommandStats(ctx.GetGuild(), since)
	if err != nil {
		return ctx.EditReply(fmt.Sprintf("❌ 統計データの取得に失敗しました: %v", err))
	}
	
	// 総使用回数を計算
	totalCommands := 0
	for _, count := range commandStats {
		totalCommands += count
	}
	
	// サーバー統計を取得
	serverStats := c.getServerStats(ctx.Session, guild)
	
	// 埋め込みを作成
	activityEmbed := embed.New().
		SetTitle(fmt.Sprintf("📊 %s の活動統計", guild.Name)).
		SetDescription(fmt.Sprintf("**期間**: %s", periodName)).
		SetColor(embed.M3Colors.Primary).
		SetThumbnail(guild.IconURL("256"))
	
	// サーバー基本情報
	activityEmbed.AddField("🏠 サーバー情報", fmt.Sprintf(
		"👥 **メンバー数**: %d\n"+
		"💬 **チャンネル数**: %d\n"+
		"🎭 **ロール数**: %d\n"+
		"📅 **作成日**: サーバー作成日",
		guild.MemberCount,
		serverStats.ChannelCount,
		len(guild.Roles),
	), true)
	
	// コマンド使用統計
	if totalCommands > 0 {
		topCommands := c.getTopCommands(commandStats, 5)
		commandText := fmt.Sprintf("📈 **総使用回数**: %d\n\n", totalCommands)
		
		for i, cmd := range topCommands {
			percentage := float64(cmd.Count) / float64(totalCommands) * 100
			bar := c.createProgressBar(percentage, 10)
			commandText += fmt.Sprintf("%d. `/%s` - %d回 (%.1f%%)\n%s\n", 
				i+1, cmd.Name, cmd.Count, percentage, bar)
		}
		
		activityEmbed.AddField("🎯 コマンド使用統計", commandText, true)
	} else {
		activityEmbed.AddField("🎯 コマンド使用統計", "📭 この期間にコマンドの使用はありません", true)
	}
	
	// アクティビティレベルを表示
	activityLevel := c.getActivityLevel(totalCommands, period)
	activityEmbed.AddField("⚡ 活動レベル", activityLevel, false)
	
	// フッター
	activityEmbed.SetFooter(fmt.Sprintf("統計取得者: %s • %s", 
		ctx.GetUser().Username, 
		time.Now().Format("2006-01-02 15:04")), 
		ctx.GetUser().AvatarURL("64"))
	
	return ctx.EditReplyEmbed(activityEmbed.Build())
}

func (c *ActivityCommand) getPeriodStart(period string) time.Time {
	now := time.Now()
	switch period {
	case "today":
		return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	case "7days":
		return now.AddDate(0, 0, -7)
	case "30days":
		return now.AddDate(0, 0, -30)
	case "all":
		return time.Unix(0, 0) // 1970年から（実質全期間）
	default:
		return now.AddDate(0, 0, -7) // デフォルトは7日間
	}
}

func (c *ActivityCommand) getPeriodName(period string) string {
	switch period {
	case "today":
		return "📅 今日"
	case "7days":
		return "📊 過去7日間"
	case "30days":
		return "📈 過去30日間"
	case "all":
		return "📉 全期間"
	default:
		return "📊 過去7日間"
	}
}

type CommandStat struct {
	Name  string
	Count int
}

func (c *ActivityCommand) getTopCommands(stats map[string]int, limit int) []CommandStat {
	var commands []CommandStat
	for name, count := range stats {
		commands = append(commands, CommandStat{Name: name, Count: count})
	}
	
	// バブルソートで降順にソート
	for i := 0; i < len(commands)-1; i++ {
		for j := 0; j < len(commands)-1-i; j++ {
			if commands[j].Count < commands[j+1].Count {
				commands[j], commands[j+1] = commands[j+1], commands[j]
			}
		}
	}
	
	// 上位N個を返す
	if len(commands) > limit {
		return commands[:limit]
	}
	return commands
}

func (c *ActivityCommand) createProgressBar(percentage float64, length int) string {
	filled := int(percentage / 100 * float64(length))
	if filled > length {
		filled = length
	}
	
	bar := strings.Repeat("█", filled) + strings.Repeat("░", length-filled)
	return fmt.Sprintf("`%s` %.1f%%", bar, percentage)
}

func (c *ActivityCommand) getActivityLevel(totalCommands int, period string) string {
	var level string
	var emoji string
	var description string
	
	// 期間に応じた基準値を設定
	var thresholds map[string][4]int
	switch period {
	case "today":
		thresholds = map[string][4]int{
			"levels": {0, 5, 15, 30}, // 低、中、高、非常に高
		}
	case "7days":
		thresholds = map[string][4]int{
			"levels": {0, 20, 60, 120},
		}
	case "30days":
		thresholds = map[string][4]int{
			"levels": {0, 50, 150, 300},
		}
	default: // all
		thresholds = map[string][4]int{
			"levels": {0, 100, 300, 600},
		}
	}
	
	levels := thresholds["levels"]
	
	if totalCommands <= levels[0] {
		emoji = "😴"
		level = "非常に低い"
		description = "もっとボットを活用してみてください！"
	} else if totalCommands <= levels[1] {
		emoji = "😐"
		level = "低い"
		description = "まだまだ活用の余地があります"
	} else if totalCommands <= levels[2] {
		emoji = "😊"
		level = "中程度"
		description = "良いペースで利用されています"
	} else if totalCommands <= levels[3] {
		emoji = "🔥"
		level = "高い"
		description = "とても活発に利用されています！"
	} else {
		emoji = "🚀"
		level = "非常に高い"
		description = "驚異的な活動レベルです！"
	}
	
	return fmt.Sprintf("%s **%s** (%d回)\n%s", emoji, level, totalCommands, description)
}

type ServerStats struct {
	ChannelCount int
}

func (c *ActivityCommand) getServerStats(session *discordgo.Session, guild *discordgo.Guild) ServerStats {
	channels, err := session.GuildChannels(guild.ID)
	channelCount := 0
	if err == nil {
		channelCount = len(channels)
	}
	
	return ServerStats{
		ChannelCount: channelCount,
	}
}