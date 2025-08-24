package di

import (
	"context"
	"database/sql"

	"github.com/bwmarrin/discordgo"
	"github.com/luna/luna-bot/bot"
	"github.com/luna/luna-bot/commands"
	"github.com/luna/luna-bot/config"
	"github.com/luna/luna-bot/database"
)

type Container struct {
	Config           *config.Config
	DB               *sql.DB
	Session          *discordgo.Session
	Bot              *bot.Bot
	CommandRegistry  *commands.Registry
	DatabaseService  *database.Service
}

func NewContainer(ctx context.Context, cfg *config.Config) (*Container, error) {
	container := &Container{
		Config: cfg,
	}

	if err := container.initDatabase(); err != nil {
		return nil, err
	}

	if err := container.initDiscordSession(); err != nil {
		return nil, err
	}

	container.initServices()
	container.initCommands()

	return container, nil
}

func (c *Container) initDatabase() error {
	db, err := database.Connect(c.Config.Database)
	if err != nil {
		return err
	}

	c.DB = db
	c.DatabaseService = database.NewService(db)

	if err := c.DatabaseService.Migrate(); err != nil {
		return err
	}

	return nil
}

func (c *Container) initDiscordSession() error {
	session, err := discordgo.New("Bot " + c.Config.Discord.Token)
	if err != nil {
		return err
	}

	session.Identify.Intents = discordgo.IntentsAll

	c.Session = session
	return nil
}

func (c *Container) initServices() {
	c.Bot = bot.New(c.Session, c.Config, c.DatabaseService)
}

func (c *Container) initCommands() {
	c.CommandRegistry = commands.NewRegistry(c.Session, c.Config, c.DatabaseService)
	
	c.CommandRegistry.Register(commands.NewPingCommand())
	c.CommandRegistry.Register(commands.NewAvatarCommand())
}

func (c *Container) Cleanup() error {
	if c.Session != nil {
		c.Session.Close()
	}

	if c.DB != nil {
		c.DB.Close()
	}

	return nil
}