package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

type Config struct {
	Discord     DiscordConfig     `mapstructure:"discord"`
	Database    DatabaseConfig    `mapstructure:"database"`
	Bot         BotConfig         `mapstructure:"bot"`
	GoogleCloud GoogleCloudConfig `mapstructure:"google_cloud"`
}

type DiscordConfig struct {
	Token       string `mapstructure:"token"`
	AppID       string `mapstructure:"app_id"`
	GuildID     string `mapstructure:"guild_id"`
	Permissions int64  `mapstructure:"permissions"`
}

type DatabaseConfig struct {
	Path           string `mapstructure:"path"`
	MaxConnections int    `mapstructure:"max_connections"`
}

type BotConfig struct {
	Prefix        string   `mapstructure:"prefix"`
	StatusMessage string   `mapstructure:"status_message"`
	ActivityType  int      `mapstructure:"activity_type"`
	Owners        []string `mapstructure:"owners"`
	Debug         bool     `mapstructure:"debug"`
}

type GoogleCloudConfig struct {
	ProjectID       string `mapstructure:"project_id"`
	Location        string `mapstructure:"location"`
	CredentialsPath string `mapstructure:"credentials_path"`
	GeminiModel     string `mapstructure:"gemini_model"`
	ImagenModel     string `mapstructure:"imagen_model"`
	UseStudioAPI    bool   `mapstructure:"use_studio_api"`    // Google AI Studio APIを使うか
	StudioAPIKey    string `mapstructure:"studio_api_key"`    // Google AI Studio APIキー
}

func Load() (*Config, error) {
	if err := godotenv.Load(); err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("error loading .env file: %w", err)
	}

	// config.yamlは使用せず、環境変数のみで設定

	viper.SetDefault("discord.permissions", 8)
	viper.SetDefault("database.path", "./data/luna.db")
	viper.SetDefault("database.max_connections", 10)
	viper.SetDefault("bot.prefix", "/")
	viper.SetDefault("bot.status_message", "Luna Bot v1.0")
	viper.SetDefault("bot.activity_type", 0)
	viper.SetDefault("bot.debug", false)

	viper.BindEnv("discord.token", "DISCORD_TOKEN")
	viper.BindEnv("discord.app_id", "DISCORD_APP_ID")
	viper.BindEnv("discord.guild_id", "DISCORD_GUILD_ID")
	viper.BindEnv("bot.debug", "DEBUG")
	viper.BindEnv("google_cloud.project_id", "GOOGLE_CLOUD_PROJECT_ID")
	viper.BindEnv("google_cloud.location", "GOOGLE_CLOUD_LOCATION")
	viper.BindEnv("google_cloud.credentials_path", "GOOGLE_APPLICATION_CREDENTIALS")
	viper.BindEnv("google_cloud.use_studio_api", "USE_GOOGLE_AI_STUDIO")
	viper.BindEnv("google_cloud.studio_api_key", "GOOGLE_AI_STUDIO_API_KEY")
	
	// デフォルト値設定
	viper.SetDefault("google_cloud.location", "us-central1")
	viper.SetDefault("google_cloud.gemini_model", "gemini-2.5-pro-preview-0206")  // Gemini 2.5 Pro 最新版！
	viper.SetDefault("google_cloud.imagen_model", "imagen-4.0-generate-preview-0606")  // Imagen 4 最新版！

	// YAMLファイルは使用せず、環境変数のみで設定するためReadInConfig()は削除

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("error unmarshaling config: %w", err)
	}

	if cfg.Discord.Token == "" {
		return nil, fmt.Errorf("DISCORD_TOKEN is required")
	}

	return &cfg, nil
}