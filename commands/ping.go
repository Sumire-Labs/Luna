package commands

import (
	"fmt"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/Sumire-Labs/Luna/embed"
)

type PingCommand struct{}

func NewPingCommand() *PingCommand {
	return &PingCommand{}
}

func (c *PingCommand) Name() string {
	return "ping"
}

func (c *PingCommand) Description() string {
	return "ボットの応答速度とレイテンシを確認します"
}

func (c *PingCommand) Usage() string {
	return "/ping"
}

func (c *PingCommand) Category() string {
	return "ユーティリティ"
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
		SetTitle("🏓 ポン！").
		SetColor(embed.M3Colors.Primary).
		AddField("📡 API レイテンシ", fmt.Sprintf("`%dms`", apiLatency), true).
		AddField("💓 WebSocket ハートビート", fmt.Sprintf("`%dms`", heartbeat), true).
		AddBlankField(true)

	quality := c.getConnectionQuality(apiLatency)
	embedBuilder.AddField("📊 接続品質", quality, false)

	statusMessage := c.getStatusMessage(apiLatency)
	embedBuilder.SetFooter(statusMessage, "")

	return ctx.EditReplyEmbed(embedBuilder.Build())
}

func (c *PingCommand) getConnectionQuality(latency int64) string {
	switch {
	case latency < 50:
		return "🟢 **優秀** - 超高速！"
	case latency < 100:
		return "🟢 **良好** - スムーズに動作"
	case latency < 200:
		return "🟡 **普通** - 軽微な遅延"
	case latency < 500:
		return "🟠 **悪い** - 目立つ遅延"
	default:
		return "🔴 **危険** - 深刻な遅延問題"
	}
}

func (c *PingCommand) getStatusMessage(latency int64) string {
	switch {
	case latency < 50:
		return "ボットは最適な性能で動作中"
	case latency < 100:
		return "ボットは正常に動作中"
	case latency < 200:
		return "ボットは軽微な遅延が発生中"
	case latency < 500:
		return "ボットの性能が低下している可能性"
	default:
		return "ボットに接続問題が発生中"
	}
}