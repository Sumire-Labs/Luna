package commands

import (
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/luna/luna-bot/database"
	"github.com/luna/luna-bot/embed"
)

type ConfigCommand struct{}

func NewConfigCommand() *ConfigCommand {
	return &ConfigCommand{}
}

func (c *ConfigCommand) Name() string {
	return "config"
}

func (c *ConfigCommand) Description() string {
	return "Configure bot features for your server"
}

func (c *ConfigCommand) Usage() string {
	return "/config <setup|view|reset> [feature]"
}

func (c *ConfigCommand) Category() string {
	return "Administration"
}

func (c *ConfigCommand) Aliases() []string {
	return []string{"setup", "configure", "settings"}
}

func (c *ConfigCommand) Permission() int64 {
	return discordgo.PermissionManageGuild
}

func (c *ConfigCommand) Options() []*discordgo.ApplicationCommandOption {
	return []*discordgo.ApplicationCommandOption{
		{
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Name:        "setup",
			Description: "Setup a specific feature",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "feature",
					Description: "Feature to setup",
					Required:    true,
					Choices: []*discordgo.ApplicationCommandOptionChoice{
						{
							Name:  "🎫 Ticket System",
							Value: "tickets",
						},
						{
							Name:  "🛡️ Moderation",
							Value: "moderation",
						},
						{
							Name:  "👋 Welcome System",
							Value: "welcome",
						},
						{
							Name:  "📝 Logging",
							Value: "logging",
						},
					},
				},
			},
		},
		{
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Name:        "view",
			Description: "View current configuration",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "feature",
					Description: "Specific feature to view (leave empty for all)",
					Required:    false,
					Choices: []*discordgo.ApplicationCommandOptionChoice{
						{
							Name:  "🎫 Ticket System",
							Value: "tickets",
						},
						{
							Name:  "🛡️ Moderation",
							Value: "moderation",
						},
						{
							Name:  "👋 Welcome System",
							Value: "welcome",
						},
						{
							Name:  "📝 Logging",
							Value: "logging",
						},
					},
				},
			},
		},
		{
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Name:        "reset",
			Description: "Reset feature configuration",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "feature",
					Description: "Feature to reset",
					Required:    true,
					Choices: []*discordgo.ApplicationCommandOptionChoice{
						{
							Name:  "🎫 Ticket System",
							Value: "tickets",
						},
						{
							Name:  "🛡️ Moderation",
							Value: "moderation",
						},
						{
							Name:  "👋 Welcome System",
							Value: "welcome",
						},
						{
							Name:  "📝 Logging",
							Value: "logging",
						},
						{
							Name:  "🔄 All Settings",
							Value: "all",
						},
					},
				},
			},
		},
		{
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Name:        "export",
			Description: "Export all server settings",
		},
	}
}

func (c *ConfigCommand) Execute(ctx *Context) error {
	if ctx.GetGuild() == "" {
		return ctx.ReplyEphemeral("This command can only be used in servers!")
	}

	// Check permissions
	member, err := ctx.Session.GuildMember(ctx.GetGuild(), ctx.GetUser().ID)
	if err != nil {
		return ctx.ReplyEphemeral("Failed to check your permissions!")
	}

	hasPermission := false
	for _, roleID := range member.Roles {
		role, err := ctx.Session.State.Role(ctx.GetGuild(), roleID)
		if err != nil {
			continue
		}
		if role.Permissions&discordgo.PermissionManageGuild != 0 || role.Permissions&discordgo.PermissionAdministrator != 0 {
			hasPermission = true
			break
		}
	}

	if !hasPermission {
		return ctx.ReplyEphemeral("❌ You need the **Manage Server** permission to use this command!")
	}

	// Get subcommand
	subcommand := ctx.Interaction.ApplicationCommandData().Options[0].Name
	
	switch subcommand {
	case "setup":
		return c.handleSetup(ctx)
	case "view":
		return c.handleView(ctx)
	case "reset":
		return c.handleReset(ctx)
	case "export":
		return c.handleExport(ctx)
	default:
		return ctx.ReplyEphemeral("Unknown subcommand!")
	}
}

func (c *ConfigCommand) handleSetup(ctx *Context) error {
	feature := ctx.Interaction.ApplicationCommandData().Options[0].Options[0].StringValue()
	
	switch feature {
	case "tickets":
		return c.setupTickets(ctx)
	case "moderation":
		return c.setupModeration(ctx)
	case "welcome":
		return c.setupWelcome(ctx)
	case "logging":
		return c.setupLogging(ctx)
	default:
		return ctx.ReplyEphemeral("Unknown feature!")
	}
}

func (c *ConfigCommand) handleView(ctx *Context) error {
	// TODO: Get database service from context
	// For now, return a placeholder
	
	var feature string
	if len(ctx.Interaction.ApplicationCommandData().Options[0].Options) > 0 {
		feature = ctx.Interaction.ApplicationCommandData().Options[0].Options[0].StringValue()
	}

	embedBuilder := embed.New().
		SetTitle("🔧 Server Configuration").
		SetColor(embed.M3Colors.Primary)

	if feature == "" {
		// Show all settings
		embedBuilder.SetDescription("Current server configuration overview")
		embedBuilder.AddField("🎫 Ticket System", "❌ Not configured", true)
		embedBuilder.AddField("🛡️ Moderation", "❌ Not configured", true)
		embedBuilder.AddField("👋 Welcome System", "❌ Not configured", true)
		embedBuilder.AddField("📝 Logging", "❌ Not configured", true)
		embedBuilder.AddBlankField(true)
		embedBuilder.AddBlankField(true)
		embedBuilder.AddField("💡 Tip", "Use `/config setup <feature>` to configure individual features", false)
	} else {
		// Show specific feature
		embedBuilder.SetDescription(fmt.Sprintf("Configuration for **%s**", c.getFeatureName(feature)))
		embedBuilder.AddField("Status", "❌ Not configured", false)
		embedBuilder.AddField("💡 Setup", fmt.Sprintf("Use `/config setup %s` to configure this feature", feature), false)
	}

	return ctx.ReplyEmbed(embedBuilder.Build())
}

func (c *ConfigCommand) handleReset(ctx *Context) error {
	feature := ctx.Interaction.ApplicationCommandData().Options[0].Options[0].StringValue()
	
	// Create confirmation embed
	embedBuilder := embed.Warning(
		"⚠️ Reset Configuration",
		fmt.Sprintf("Are you sure you want to reset the **%s** configuration?\n\n**This action cannot be undone!**", c.getFeatureName(feature)),
	)

	// Create confirmation buttons
	components := []discordgo.MessageComponent{
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.Button{
					Style:    discordgo.DangerButton,
					Label:    "🗑️ Reset",
					CustomID: fmt.Sprintf("config_reset_confirm_%s", feature),
				},
				discordgo.Button{
					Style:    discordgo.SecondaryButton,
					Label:    "❌ Cancel",
					CustomID: "config_reset_cancel",
				},
			},
		},
	}

	return ctx.Session.InteractionRespond(ctx.Interaction.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds:     []*discordgo.MessageEmbed{embedBuilder},
			Components: components,
			Flags:      discordgo.MessageFlagsEphemeral,
		},
	})
}

func (c *ConfigCommand) handleExport(ctx *Context) error {
	// TODO: Implement settings export
	return ctx.ReplyEphemeral("🚧 Export feature is coming soon!")
}

func (c *ConfigCommand) setupTickets(ctx *Context) error {
	embedBuilder := embed.New().
		SetTitle("🎫 Ticket System Setup").
		SetDescription("Let's configure your ticket system! Click the button below to start the setup wizard.").
		SetColor(embed.M3Colors.Primary).
		AddField("📋 What we'll configure:", strings.Join([]string{
			"• Ticket category (where channels are created)",
			"• Support role (who can view tickets)",
			"• Admin role (who can manage tickets)",  
			"• Log channel (for ticket events)",
			"• Auto-close settings",
		}, "\n"), false)

	components := []discordgo.MessageComponent{
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.Button{
					Style:    discordgo.PrimaryButton,
					Label:    "🚀 Start Setup",
					CustomID: "ticket_setup_start",
				},
				discordgo.Button{
					Style:    discordgo.SecondaryButton,
					Label:    "❌ Cancel",
					CustomID: "setup_cancel",
				},
			},
		},
	}

	return ctx.Session.InteractionRespond(ctx.Interaction.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds:     []*discordgo.MessageEmbed{embedBuilder.Build()},
			Components: components,
			Flags:      discordgo.MessageFlagsEphemeral,
		},
	})
}

func (c *ConfigCommand) setupModeration(ctx *Context) error {
	return ctx.ReplyEphemeral("🚧 Moderation setup is coming soon!")
}

func (c *ConfigCommand) setupWelcome(ctx *Context) error {
	return ctx.ReplyEphemeral("🚧 Welcome system setup is coming soon!")
}

func (c *ConfigCommand) setupLogging(ctx *Context) error {
	return ctx.ReplyEphemeral("🚧 Logging setup is coming soon!")
}

func (c *ConfigCommand) getFeatureName(feature string) string {
	names := map[string]string{
		"tickets":    "🎫 Ticket System",
		"moderation": "🛡️ Moderation",
		"welcome":    "👋 Welcome System",
		"logging":    "📝 Logging",
		"all":        "🔄 All Settings",
	}
	
	if name, ok := names[feature]; ok {
		return name
	}
	return feature
}