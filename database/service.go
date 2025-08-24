package database

import (
	"database/sql"
	"fmt"
	"time"
)

type Service struct {
	db *sql.DB
}

func NewService(db *sql.DB) *Service {
	return &Service{db: db}
}

func (s *Service) Migrate() error {
	migrations := []string{
		`CREATE TABLE IF NOT EXISTS guilds (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			prefix TEXT DEFAULT '/',
			language TEXT DEFAULT 'en',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS users (
			id TEXT PRIMARY KEY,
			username TEXT NOT NULL,
			discriminator TEXT,
			avatar TEXT,
			bot BOOLEAN DEFAULT FALSE,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS command_usage (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			guild_id TEXT,
			user_id TEXT NOT NULL,
			command TEXT NOT NULL,
			args TEXT,
			success BOOLEAN DEFAULT TRUE,
			error_message TEXT,
			executed_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (guild_id) REFERENCES guilds(id),
			FOREIGN KEY (user_id) REFERENCES users(id)
		)`,
		`CREATE TABLE IF NOT EXISTS user_settings (
			user_id TEXT PRIMARY KEY,
			theme TEXT DEFAULT 'material3',
			color_scheme TEXT DEFAULT 'dynamic',
			language TEXT DEFAULT 'en',
			notifications BOOLEAN DEFAULT TRUE,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users(id)
		)`,
		`CREATE INDEX IF NOT EXISTS idx_command_usage_guild ON command_usage(guild_id)`,
		`CREATE INDEX IF NOT EXISTS idx_command_usage_user ON command_usage(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_command_usage_command ON command_usage(command)`,
		`CREATE INDEX IF NOT EXISTS idx_command_usage_executed ON command_usage(executed_at)`,
	}

	for _, migration := range migrations {
		if _, err := s.db.Exec(migration); err != nil {
			return fmt.Errorf("migration failed: %w", err)
		}
	}

	return nil
}

func (s *Service) LogCommand(guildID, userID, command, args string, success bool, errorMsg string) error {
	query := `
		INSERT INTO command_usage (guild_id, user_id, command, args, success, error_message)
		VALUES (?, ?, ?, ?, ?, ?)
	`
	_, err := s.db.Exec(query, guildID, userID, command, args, success, errorMsg)
	return err
}

func (s *Service) GetCommandStats(guildID string, since time.Time) (map[string]int, error) {
	query := `
		SELECT command, COUNT(*) as count
		FROM command_usage
		WHERE guild_id = ? AND executed_at >= ?
		GROUP BY command
		ORDER BY count DESC
	`
	
	rows, err := s.db.Query(query, guildID, since)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	stats := make(map[string]int)
	for rows.Next() {
		var command string
		var count int
		if err := rows.Scan(&command, &count); err != nil {
			return nil, err
		}
		stats[command] = count
	}

	return stats, rows.Err()
}

func (s *Service) UpsertUser(id, username, discriminator, avatar string, isBot bool) error {
	query := `
		INSERT INTO users (id, username, discriminator, avatar, bot)
		VALUES (?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			username = excluded.username,
			discriminator = excluded.discriminator,
			avatar = excluded.avatar,
			bot = excluded.bot,
			updated_at = CURRENT_TIMESTAMP
	`
	_, err := s.db.Exec(query, id, username, discriminator, avatar, isBot)
	return err
}

func (s *Service) UpsertGuild(id, name, prefix string) error {
	query := `
		INSERT INTO guilds (id, name, prefix)
		VALUES (?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			name = excluded.name,
			prefix = excluded.prefix,
			updated_at = CURRENT_TIMESTAMP
	`
	_, err := s.db.Exec(query, id, name, prefix)
	return err
}

func (s *Service) GetGuildPrefix(guildID string) (string, error) {
	var prefix string
	query := `SELECT prefix FROM guilds WHERE id = ?`
	err := s.db.QueryRow(query, guildID).Scan(&prefix)
	if err == sql.ErrNoRows {
		return "/", nil
	}
	return prefix, err
}