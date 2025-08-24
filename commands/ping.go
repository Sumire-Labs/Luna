package commands

import (
	"fmt"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/luna/luna-bot/embed"
)

type PingCommand struct{}

func NewPingCommand() *PingCommand {
	return &PingCommand{}
}

func (c *PingCommand) Name() string {
	return "ping"
}

func (c *PingCommand) Description() string {
	return "Check the bot's latency and response time"
}

func (c *PingCommand) Usage() string {
	return "/ping"
}

func (c *PingCommand) Category() string {
	return "Utility"
}

func (c *PingCommand) Aliases() []string {
	return []string{"pong", "latency"}
}

func (c *PingCommand) Permission() int64 {
	return 0
}

func (c *PingCommand) Options() []*discordgo.ApplicationCommandOption {
	return []*discordgo.ApplicationCommandOption{}
}

func (c *PingCommand) Execute(ctx *Context) error {
	start := time.Now()
	
	if err := ctx.DeferReply(false); err != nil {
		return err
	}

	apiLatency := time.Since(start).Milliseconds()
	heartbeat := ctx.Session.HeartbeatLatency().Milliseconds()

	embedBuilder := embed.New().
		SetTitle("🏓 Pong!").
		SetColor(embed.M3Colors.Primary).
		AddField("📡 API Latency", fmt.Sprintf("`%dms`", apiLatency), true).
		AddField("💓 Websocket Heartbeat", fmt.Sprintf("`%dms`", heartbeat), true).
		AddBlankField(true)

	quality := c.getConnectionQuality(apiLatency)
	embedBuilder.AddField("📊 Connection Quality", quality, false)

	statusMessage := c.getStatusMessage(apiLatency)
	embedBuilder.SetFooter(statusMessage, "")

	return ctx.EditReplyEmbed(embedBuilder.Build())
}

func (c *PingCommand) getConnectionQuality(latency int64) string {
	switch {
	case latency < 50:
		return "🟢 **Excellent** - Lightning fast!"
	case latency < 100:
		return "🟢 **Good** - Running smoothly"
	case latency < 200:
		return "🟡 **Fair** - Slight delay"
	case latency < 500:
		return "🟠 **Poor** - Noticeable lag"
	default:
		return "🔴 **Critical** - Severe latency issues"
	}
}

func (c *PingCommand) getStatusMessage(latency int64) string {
	switch {
	case latency < 50:
		return "Bot is performing optimally"
	case latency < 100:
		return "Bot is running normally"
	case latency < 200:
		return "Bot is experiencing minor delays"
	case latency < 500:
		return "Bot performance may be degraded"
	default:
		return "Bot is experiencing connectivity issues"
	}
}