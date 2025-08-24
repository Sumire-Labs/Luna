package commands

import (
	"github.com/bwmarrin/discordgo"
)

type Command interface {
	Name() string
	Description() string
	Usage() string
	Category() string
	Aliases() []string
	Permission() int64
	Options() []*discordgo.ApplicationCommandOption
	Execute(ctx *Context) error
}

type Context struct {
	Session     *discordgo.Session
	Interaction *discordgo.InteractionCreate
	Args        map[string]interface{}
}

func NewContext(s *discordgo.Session, i *discordgo.InteractionCreate) *Context {
	ctx := &Context{
		Session:     s,
		Interaction: i,
		Args:        make(map[string]interface{}),
	}

	if i.ApplicationCommandData().Options != nil {
		for _, opt := range i.ApplicationCommandData().Options {
			ctx.Args[opt.Name] = opt.Value
		}
	}

	return ctx
}

func (c *Context) Reply(content string) error {
	return c.Session.InteractionRespond(c.Interaction.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: content,
		},
	})
}

func (c *Context) ReplyEmbed(embed *discordgo.MessageEmbed) error {
	return c.Session.InteractionRespond(c.Interaction.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{embed},
		},
	})
}

func (c *Context) ReplyEphemeral(content string) error {
	return c.Session.InteractionRespond(c.Interaction.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: content,
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	})
}

func (c *Context) ReplyEmbedEphemeral(embed *discordgo.MessageEmbed) error {
	return c.Session.InteractionRespond(c.Interaction.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{embed},
			Flags:  discordgo.MessageFlagsEphemeral,
		},
	})
}

func (c *Context) DeferReply(ephemeral bool) error {
	flags := uint64(0)
	if ephemeral {
		flags = discordgo.MessageFlagsEphemeral
	}

	return c.Session.InteractionRespond(c.Interaction.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Flags: flags,
		},
	})
}

func (c *Context) EditReply(content string) error {
	_, err := c.Session.InteractionResponseEdit(c.Interaction.Interaction, &discordgo.WebhookEdit{
		Content: &content,
	})
	return err
}

func (c *Context) EditReplyEmbed(embed *discordgo.MessageEmbed) error {
	_, err := c.Session.InteractionResponseEdit(c.Interaction.Interaction, &discordgo.WebhookEdit{
		Embeds: &[]*discordgo.MessageEmbed{embed},
	})
	return err
}

func (c *Context) GetUser() *discordgo.User {
	if c.Interaction.Member != nil {
		return c.Interaction.Member.User
	}
	return c.Interaction.User
}

func (c *Context) GetGuild() string {
	if c.Interaction.GuildID != "" {
		return c.Interaction.GuildID
	}
	return ""
}

func (c *Context) GetChannel() string {
	return c.Interaction.ChannelID
}

func (c *Context) GetArg(name string) (interface{}, bool) {
	val, ok := c.Args[name]
	return val, ok
}

func (c *Context) GetStringArg(name string) string {
	if val, ok := c.Args[name]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}

func (c *Context) GetUserArg(name string) *discordgo.User {
	if val, ok := c.Args[name]; ok {
		if user, ok := val.(*discordgo.User); ok {
			return user
		}
	}
	return nil
}

func (c *Context) GetIntArg(name string) int64 {
	if val, ok := c.Args[name]; ok {
		if num, ok := val.(float64); ok {
			return int64(num)
		}
	}
	return 0
}

func (c *Context) GetBoolArg(name string) bool {
	if val, ok := c.Args[name]; ok {
		if b, ok := val.(bool); ok {
			return b
		}
	}
	return false
}