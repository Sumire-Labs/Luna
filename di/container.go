package di

import (
	"context"
	"database/sql"

	"github.com/bwmarrin/discordgo"
	"github.com/luna/luna-bot/ai"
	"github.com/luna/luna-bot/bot"
	"github.com/luna/luna-bot/commands"
	"github.com/luna/luna-bot/config"
	"github.com/luna/luna-bot/database"
	"github.com/luna/luna-bot/logging"
)

type Container struct {
	Config           *config.Config
	DB               *sql.DB
	Session          *discordgo.Session
	Bot              *bot.Bot
	CommandRegistry  *commands.Registry
	DatabaseService  *database.Service
	Logger           *logging.Logger
	AIService        *ai.Service
	GeminiStudio     *ai.GeminiStudioService
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
	
	// AI Service の初期化（オプション）
	if cfg.GoogleCloud.UseStudioAPI && cfg.GoogleCloud.StudioAPIKey != "" {
		// Google AI Studio API を優先
		if err := container.initGeminiStudio(); err != nil {
			println("Warning: Gemini Studio service initialization failed:", err.Error())
		}
	} else if cfg.GoogleCloud.ProjectID != "" {
		// Vertex AI を使用
		if err := container.initAIService(); err != nil {
			println("Warning: Vertex AI service initialization failed:", err.Error())
		}
	}
	
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
	c.Logger = logging.NewLogger(c.Session, c.Config, c.DatabaseService)
	
	// ログハンドラーを登録
	c.Logger.RegisterHandlers()
}

func (c *Container) initAIService() error {
	aiService, err := ai.NewService(&c.Config.GoogleCloud)
	if err != nil {
		return err
	}
	c.AIService = aiService
	return nil
}

func (c *Container) initGeminiStudio() error {
	geminiStudio := ai.NewGeminiStudioService(
		c.Config.GoogleCloud.StudioAPIKey,
		c.Config.GoogleCloud.GeminiModel,
	)
	c.GeminiStudio = geminiStudio
	return nil
}

func (c *Container) initCommands() {
	c.CommandRegistry = commands.NewRegistry(c.Session, c.Config, c.DatabaseService)
	
	c.CommandRegistry.Register(commands.NewPingCommand())
	c.CommandRegistry.Register(commands.NewAvatarCommand())
	c.CommandRegistry.Register(commands.NewConfigCommand())
	c.CommandRegistry.Register(commands.NewEmbedBuilderCommand())
	
	// AI コマンドの登録（AIサービスが利用可能な場合のみ）
	if c.AIService != nil {
		c.CommandRegistry.Register(commands.NewAICommand(c.AIService))
		c.CommandRegistry.Register(commands.NewImageCommand(c.AIService))
	}
	
	// OCR コマンドの登録（Gemini Studio APIが利用可能な場合）
	if c.GeminiStudio != nil {
		c.CommandRegistry.Register(commands.NewOCRCommand(c.GeminiStudio))
	}
}

func (c *Container) Cleanup() error {
	if c.Session != nil {
		c.Session.Close()
	}

	if c.DB != nil {
		c.DB.Close()
	}
	
	if c.AIService != nil {
		c.AIService.Close()
	}

	return nil
}