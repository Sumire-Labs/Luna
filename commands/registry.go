package commands

import (
	"fmt"
	"log"
	"sync"

	"github.com/bwmarrin/discordgo"
	"github.com/luna/luna-bot/config"
	"github.com/luna/luna-bot/database"
)

type Registry struct {
	session           *discordgo.Session
	config            *config.Config
	db                *database.Service
	commands          map[string]Command
	interactionHandler *InteractionHandler
	mutex             sync.RWMutex
}

func NewRegistry(session *discordgo.Session, cfg *config.Config, db *database.Service) *Registry {
	return &Registry{
		session:            session,
		config:             cfg,
		db:                 db,
		commands:           make(map[string]Command),
		interactionHandler: NewInteractionHandler(session, cfg, db),
	}
}

func (r *Registry) Register(cmd Command) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if _, exists := r.commands[cmd.Name()]; exists {
		return fmt.Errorf("command %s already registered", cmd.Name())
	}

	r.commands[cmd.Name()] = cmd

	for _, alias := range cmd.Aliases() {
		if _, exists := r.commands[alias]; exists {
			return fmt.Errorf("alias %s already registered", alias)
		}
		r.commands[alias] = cmd
	}

	return nil
}

func (r *Registry) Get(name string) (Command, bool) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	cmd, ok := r.commands[name]
	return cmd, ok
}

func (r *Registry) GetAll() []Command {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	seen := make(map[string]bool)
	cmds := make([]Command, 0)

	for _, cmd := range r.commands {
		if !seen[cmd.Name()] {
			seen[cmd.Name()] = true
			cmds = append(cmds, cmd)
		}
	}

	return cmds
}

func (r *Registry) RegisterSlashCommands() error {
	applicationCommands := make([]*discordgo.ApplicationCommand, 0)

	for _, cmd := range r.GetAll() {
		appCmd := &discordgo.ApplicationCommand{
			Name:        cmd.Name(),
			Description: cmd.Description(),
			Options:     cmd.Options(),
		}
		applicationCommands = append(applicationCommands, appCmd)
	}

	var guildID string
	if r.config.Discord.GuildID != "" {
		guildID = r.config.Discord.GuildID
	}

	registeredCommands := make([]*discordgo.ApplicationCommand, len(applicationCommands))
	for i, cmd := range applicationCommands {
		registered, err := r.session.ApplicationCommandCreate(
			r.session.State.User.ID,
			guildID,
			cmd,
		)
		if err != nil {
			return fmt.Errorf("failed to register command %s: %w", cmd.Name, err)
		}
		registeredCommands[i] = registered
		log.Printf("Registered slash command: %s", cmd.Name)
	}

	r.session.AddHandler(r.handleInteraction)
	r.session.AddHandler(r.interactionHandler.HandleComponentInteraction)
	r.session.AddHandler(r.interactionHandler.HandleModalSubmit)

	return nil
}

func (r *Registry) handleInteraction(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.Type != discordgo.InteractionApplicationCommand {
		return
	}

	cmdName := i.ApplicationCommandData().Name
	cmd, ok := r.Get(cmdName)
	if !ok {
		log.Printf("Unknown command: %s", cmdName)
		return
	}

	ctx := NewContext(s, i)

	go func() {
		execErr := cmd.Execute(ctx)
		if execErr != nil {
			log.Printf("Error executing command %s: %v", cmdName, execErr)
			
			errorMsg := fmt.Sprintf("An error occurred while executing the command: %v", execErr)
			if ctx.Interaction.Interaction.AppID != "" {
				ctx.EditReply(errorMsg)
			} else {
				ctx.Reply(errorMsg)
			}
		}

		guildID := ""
		if i.GuildID != "" {
			guildID = i.GuildID
		}

		user := ctx.GetUser()
		if user != nil {
			var errorMessage string
			if execErr != nil {
				errorMessage = execErr.Error()
			}
			
			r.db.LogCommand(
				guildID,
				user.ID,
				cmdName,
				fmt.Sprintf("%v", ctx.Args),
				execErr == nil,
				errorMessage,
			)
		}
	}()
}

func (r *Registry) UnregisterSlashCommands() error {
	var guildID string
	if r.config.Discord.GuildID != "" {
		guildID = r.config.Discord.GuildID
	}

	commands, err := r.session.ApplicationCommands(r.session.State.User.ID, guildID)
	if err != nil {
		return fmt.Errorf("failed to get application commands: %w", err)
	}

	for _, cmd := range commands {
		err := r.session.ApplicationCommandDelete(r.session.State.User.ID, guildID, cmd.ID)
		if err != nil {
			return fmt.Errorf("failed to delete command %s: %w", cmd.Name, err)
		}
		log.Printf("Unregistered slash command: %s", cmd.Name)
	}

	return nil
}