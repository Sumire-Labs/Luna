package embed

import (
	"fmt"
	"time"

	"github.com/bwmarrin/discordgo"
)

type MaterialColor struct {
	Primary   int
	Secondary int
	Tertiary  int
	Error     int
	Success   int
	Warning   int
	Info      int
	Surface   int
}

var M3Colors = MaterialColor{
	Primary:   0x6750A4,
	Secondary: 0x625B71,
	Tertiary:  0x7D5260,
	Error:     0xBA1A1A,
	Success:   0x4CAF50,
	Warning:   0xFF9800,
	Info:      0x2196F3,
	Surface:   0x1C1B1F,
}

type Builder struct {
	embed *discordgo.MessageEmbed
}

func New() *Builder {
	return &Builder{
		embed: &discordgo.MessageEmbed{
			Color:     M3Colors.Primary,
			Timestamp: time.Now().Format(time.RFC3339),
		},
	}
}

func (b *Builder) SetTitle(title string) *Builder {
	b.embed.Title = title
	return b
}

func (b *Builder) SetDescription(desc string) *Builder {
	b.embed.Description = desc
	return b
}

func (b *Builder) SetColor(color int) *Builder {
	b.embed.Color = color
	return b
}

func (b *Builder) SetAuthor(name, iconURL, url string) *Builder {
	b.embed.Author = &discordgo.MessageEmbedAuthor{
		Name:    name,
		IconURL: iconURL,
		URL:     url,
	}
	return b
}

func (b *Builder) SetFooter(text, iconURL string) *Builder {
	b.embed.Footer = &discordgo.MessageEmbedFooter{
		Text:    text,
		IconURL: iconURL,
	}
	return b
}

func (b *Builder) SetThumbnail(url string) *Builder {
	b.embed.Thumbnail = &discordgo.MessageEmbedThumbnail{
		URL: url,
	}
	return b
}

func (b *Builder) SetImage(url string) *Builder {
	b.embed.Image = &discordgo.MessageEmbedImage{
		URL: url,
	}
	return b
}

func (b *Builder) AddField(name, value string, inline bool) *Builder {
	b.embed.Fields = append(b.embed.Fields, &discordgo.MessageEmbedField{
		Name:   name,
		Value:  value,
		Inline: inline,
	})
	return b
}

func (b *Builder) AddBlankField(inline bool) *Builder {
	return b.AddField("\u200b", "\u200b", inline)
}

func (b *Builder) SetTimestamp() *Builder {
	now := time.Now().Format(time.RFC3339)
	b.embed.Timestamp = now
	return b
}

func (b *Builder) Build() *discordgo.MessageEmbed {
	return b.embed
}

func Success(title, description string) *discordgo.MessageEmbed {
	return New().
		SetTitle(fmt.Sprintf("✅ %s", title)).
		SetDescription(description).
		SetColor(M3Colors.Success).
		Build()
}

func Error(title, description string) *discordgo.MessageEmbed {
	return New().
		SetTitle(fmt.Sprintf("❌ %s", title)).
		SetDescription(description).
		SetColor(M3Colors.Error).
		Build()
}

func Warning(title, description string) *discordgo.MessageEmbed {
	return New().
		SetTitle(fmt.Sprintf("⚠️ %s", title)).
		SetDescription(description).
		SetColor(M3Colors.Warning).
		Build()
}

func Info(title, description string) *discordgo.MessageEmbed {
	return New().
		SetTitle(fmt.Sprintf("ℹ️ %s", title)).
		SetDescription(description).
		SetColor(M3Colors.Info).
		Build()
}

func Loading(description string) *discordgo.MessageEmbed {
	return New().
		SetDescription(fmt.Sprintf("⏳ %s", description)).
		SetColor(M3Colors.Secondary).
		Build()
}