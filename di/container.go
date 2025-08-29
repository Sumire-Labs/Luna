package di

import (
	"context"
	"database/sql"

	"github.com/bwmarrin/discordgo"
	"github.com/Sumire-Labs/Luna/ai"
	"github.com/Sumire-Labs/Luna/bot"
	"github.com/Sumire-Labs/Luna/bump"
	"github.com/Sumire-Labs/Luna/commands"
	"github.com/Sumire-Labs/Luna/config"
	"github.com/Sumire-Labs/Luna/database"
	"github.com/Sumire-Labs/Luna/logging"
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
	VertexGemini     *ai.VertexGeminiService
	AIServices       *ai.Services  // New: holds all AI services
	BumpHandler      *bump.Handler
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
	
	// AI Service の初期化（新しいファクトリーパターンを使用）
	if err := container.initAIServicesWithFactory(); err != nil {
		println("Warning: AI services initialization failed:", err.Error())
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
	c.BumpHandler = bump.NewHandler(c.Session, c.DatabaseService)
	
	// ハンドラーを登録
	c.Logger.RegisterHandlers()
	c.BumpHandler.RegisterHandlers()
	
	// 起動時に保留中のBumpリマインダーをチェック
	go c.BumpHandler.CheckPendingReminders()
}

func (c *Container) initAIServicesWithFactory() error {
	factory := ai.NewFactory(&c.Config.GoogleCloud)
	services, err := factory.CreateServices()
	if err != nil {
		return err
	}
	
	// Store the services container
	c.AIServices = services
	
	// Also store individual services for backward compatibility
	c.GeminiStudio = services.GeminiStudio
	c.VertexGemini = services.VertexGemini
	c.AIService = services.LegacyService
	
	return nil
}

// Legacy methods kept for compatibility
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

func (c *Container) initVertexGemini() error {
	vertexGemini, err := ai.NewVertexGeminiService(&c.Config.GoogleCloud)
	if err != nil {
		return err
	}
	c.VertexGemini = vertexGemini
	return nil
}

func (c *Container) initCommands() {
	c.CommandRegistry = commands.NewRegistry(c.Session, c.Config, c.DatabaseService)
	
	c.CommandRegistry.Register(commands.NewPingCommand())
	c.CommandRegistry.Register(commands.NewAvatarCommand())
	c.CommandRegistry.Register(commands.NewConfigCommand())
	c.CommandRegistry.Register(commands.NewEmbedBuilderCommand())
	c.CommandRegistry.Register(commands.NewActivityCommand(c.DatabaseService))
	c.CommandRegistry.Register(commands.NewLockdownCommand())
	c.CommandRegistry.Register(commands.NewPurgeCommand())
	
	// AI コマンドの登録
	if c.VertexGemini != nil {
		// 新しいVertex AI Gemini API使用時
		c.CommandRegistry.Register(commands.NewAICommandWithVertex(c.VertexGemini))
		// Imagenコマンドは旧APIが必要、GeminiStudioが使える場合は日本語翻訳対応
		if c.AIService != nil {
			if c.GeminiStudio != nil {
				c.CommandRegistry.Register(commands.NewImageCommandWithGemini(c.AIService, c.GeminiStudio))
			} else {
				c.CommandRegistry.Register(commands.NewImageCommand(c.AIService))
			}
		}
	} else if c.AIService != nil {
		// 旧Vertex AI使用時
		c.CommandRegistry.Register(commands.NewAICommand(c.AIService))
		// GeminiStudioが使える場合は日本語翻訳対応のImageCommand
		if c.GeminiStudio != nil {
			c.CommandRegistry.Register(commands.NewImageCommandWithGemini(c.AIService, c.GeminiStudio))
		} else {
			c.CommandRegistry.Register(commands.NewImageCommand(c.AIService))
		}
	} else if c.GeminiStudio != nil {
		// Google AI Studio使用時（askコマンドのみ、imageは非対応）
		c.CommandRegistry.Register(commands.NewAICommandWithStudio(c.GeminiStudio))
	}
	
	// OCR・翻訳コマンドの登録
	if c.VertexGemini != nil && c.GeminiStudio != nil {
		c.CommandRegistry.Register(commands.NewOCRCommandWithBoth(c.GeminiStudio, c.VertexGemini))
		c.CommandRegistry.Register(commands.NewTranslateCommand(c.GeminiStudio))
	} else if c.VertexGemini != nil {
		c.CommandRegistry.Register(commands.NewOCRCommandWithVertex(c.VertexGemini))
	} else if c.GeminiStudio != nil {
		c.CommandRegistry.Register(commands.NewOCRCommand(c.GeminiStudio))
		c.CommandRegistry.Register(commands.NewTranslateCommand(c.GeminiStudio))
	}
	
	// War Thunder コマンドの登録
	c.CommandRegistry.Register(commands.NewWTCommand())
	
	// Brackets コマンドの登録
	c.CommandRegistry.Register(commands.NewBracketsCommand(c.DatabaseService))
}

func (c *Container) Cleanup() error {
	if c.Session != nil {
		c.Session.Close()
	}

	if c.DB != nil {
		c.DB.Close()
	}
	
	// Close all AI services through the new container
	if c.AIServices != nil {
		c.AIServices.Close()
	}

	return nil
}