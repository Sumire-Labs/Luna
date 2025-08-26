package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

type Config struct {
	Discord     DiscordConfig     `toml:"discord" mapstructure:"discord"`
	Database    DatabaseConfig    `toml:"database" mapstructure:"database"`
	Bot         BotConfig         `toml:"bot" mapstructure:"bot"`
	GoogleCloud GoogleCloudConfig `toml:"google_cloud" mapstructure:"google_cloud"`
	Logging     LoggingConfig     `toml:"logging" mapstructure:"logging"`
	Features    FeaturesConfig    `toml:"features" mapstructure:"features"`
}

type DiscordConfig struct {
	Token       string `toml:"token" mapstructure:"token"`
	AppID       string `toml:"app_id" mapstructure:"app_id"`
	GuildID     string `toml:"guild_id" mapstructure:"guild_id"`
	Permissions int64  `toml:"permissions" mapstructure:"permissions"`
}

type DatabaseConfig struct {
	Path           string `toml:"path" mapstructure:"path"`
	MaxConnections int    `toml:"max_connections" mapstructure:"max_connections"`
}

type BotConfig struct {
	Prefix        string   `toml:"prefix" mapstructure:"prefix"`
	StatusMessage string   `toml:"status_message" mapstructure:"status_message"`
	ActivityType  int      `toml:"activity_type" mapstructure:"activity_type"`
	Owners        []string `toml:"owners" mapstructure:"owners"`
	Debug         bool     `toml:"debug" mapstructure:"debug"`
}

type GoogleCloudConfig struct {
	ProjectID       string `toml:"project_id" mapstructure:"project_id"`
	Location        string `toml:"location" mapstructure:"location"`
	CredentialsPath string `toml:"credentials_path" mapstructure:"credentials_path"`
	GeminiModel     string `toml:"gemini_model" mapstructure:"gemini_model"`
	ImagenModel     string `toml:"imagen_model" mapstructure:"imagen_model"`
	UseStudioAPI    bool   `toml:"use_studio_api" mapstructure:"use_studio_api"` // Google AI Studio APIを使うか
	StudioAPIKey    string `toml:"studio_api_key" mapstructure:"studio_api_key"` // Google AI Studio APIキー
}

type LoggingConfig struct {
	Level  string `toml:"level" mapstructure:"level"`
	Format string `toml:"format" mapstructure:"format"`
	Output string `toml:"output" mapstructure:"output"`
}

type FeaturesConfig struct {
	EnableAI         bool `toml:"enable_ai" mapstructure:"enable_ai"`
	EnableLogging    bool `toml:"enable_logging" mapstructure:"enable_logging"`
	EnableTickets    bool `toml:"enable_tickets" mapstructure:"enable_tickets"`
	EnableModeration bool `toml:"enable_moderation" mapstructure:"enable_moderation"`
	EnableMusic      bool `toml:"enable_music" mapstructure:"enable_music"`
}

func Load() (*Config, error) {
	// 設定ファイル名と形式を設定
	viper.SetConfigName("config")
	viper.SetConfigType("toml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./config")
	viper.AddConfigPath("$HOME/.luna")

	// デフォルト値設定
	setDefaults()

	// 設定ファイルを読み込み
	if err := viper.ReadInConfig(); err != nil {
		// 設定ファイルが見つからない場合は環境変数フォールバック
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			fmt.Println("Warning: config.toml not found, falling back to environment variables")
			return loadFromEnv()
		}
		return nil, fmt.Errorf("error reading config file: %w", err)
	}

	fmt.Printf("Using config file: %s\n", viper.ConfigFileUsed())

	// 環境変数でのオーバーライドを有効化
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("error unmarshaling config: %w", err)
	}

	// 必須項目の検証
	if err := validateConfig(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func setDefaults() {
	// Discord設定
	viper.SetDefault("discord.permissions", 8)
	
	// データベース設定
	viper.SetDefault("database.path", "./data/luna.db")
	viper.SetDefault("database.max_connections", 10)
	
	// ボット設定
	viper.SetDefault("bot.prefix", "/")
	viper.SetDefault("bot.status_message", "Luna Bot v1.0")
	viper.SetDefault("bot.activity_type", 0)
	viper.SetDefault("bot.debug", false)
	viper.SetDefault("bot.owners", []string{})
	
	// Google Cloud設定
	viper.SetDefault("google_cloud.location", "us-central1")
	viper.SetDefault("google_cloud.gemini_model", "gemini-2.5-flash-lite")
	viper.SetDefault("google_cloud.imagen_model", "imagen-4.0-fast-generate-preview-06-06")
	viper.SetDefault("google_cloud.use_studio_api", false)
	
	// ログ設定
	viper.SetDefault("logging.level", "info")
	viper.SetDefault("logging.format", "text")
	viper.SetDefault("logging.output", "console")
	
	// 機能設定
	viper.SetDefault("features.enable_ai", true)
	viper.SetDefault("features.enable_logging", true)
	viper.SetDefault("features.enable_tickets", true)
	viper.SetDefault("features.enable_moderation", false)
	viper.SetDefault("features.enable_music", false)
}

// 環境変数フォールバック（後方互換性）
func loadFromEnv() (*Config, error) {
	if err := godotenv.Load(); err != nil && !os.IsNotExist(err) {
		fmt.Printf("Warning: Error loading .env file: %v\n", err)
	}

	// 環境変数バインディング
	viper.BindEnv("discord.token", "DISCORD_TOKEN")
	viper.BindEnv("discord.app_id", "DISCORD_APP_ID")
	viper.BindEnv("discord.guild_id", "DISCORD_GUILD_ID")
	viper.BindEnv("bot.debug", "DEBUG")
	viper.BindEnv("google_cloud.project_id", "GOOGLE_CLOUD_PROJECT_ID")
	viper.BindEnv("google_cloud.location", "GOOGLE_CLOUD_LOCATION")
	viper.BindEnv("google_cloud.credentials_path", "GOOGLE_APPLICATION_CREDENTIALS")
	viper.BindEnv("google_cloud.use_studio_api", "USE_GOOGLE_AI_STUDIO")
	viper.BindEnv("google_cloud.studio_api_key", "GOOGLE_AI_STUDIO_API_KEY")

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("error unmarshaling config: %w", err)
	}

	if err := validateConfig(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func validateConfig(cfg *Config) error {
	if cfg.Discord.Token == "" {
		return fmt.Errorf("discord.token is required")
	}
	
	return nil
}