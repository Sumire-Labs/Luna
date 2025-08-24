package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

type Config struct {
	Discord  DiscordConfig  `mapstructure:"discord"`
	Database DatabaseConfig `mapstructure:"database"`
	Bot      BotConfig      `mapstructure:"bot"`
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

func Load() (*Config, error) {
	if err := godotenv.Load(); err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("error loading .env file: %w", err)
	}

	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./configs")
	viper.AddConfigPath(".")

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

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("error reading config file: %w", err)
		}
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("error unmarshaling config: %w", err)
	}

	if cfg.Discord.Token == "" {
		return nil, fmt.Errorf("DISCORD_TOKEN is required")
	}

	return &cfg, nil
}