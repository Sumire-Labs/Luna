package database

import (
	"database/sql"
	"fmt"
	"log"
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
		`CREATE TABLE IF NOT EXISTS guild_settings (
			guild_id TEXT PRIMARY KEY,
			
			-- Ticket System Settings
			ticket_enabled BOOLEAN DEFAULT FALSE,
			ticket_category_id TEXT,
			ticket_support_role_id TEXT,
			ticket_admin_role_id TEXT,
			ticket_log_channel_id TEXT,
			ticket_transcript_channel_id TEXT,
			ticket_auto_close_hours INTEGER DEFAULT 24,
			ticket_max_per_user INTEGER DEFAULT 3,
			
			-- Moderation Settings
			moderation_enabled BOOLEAN DEFAULT FALSE,
			moderation_log_channel_id TEXT,
			automod_enabled BOOLEAN DEFAULT FALSE,
			
			-- Welcome System Settings
			welcome_enabled BOOLEAN DEFAULT FALSE,
			welcome_channel_id TEXT,
			welcome_message TEXT,
			welcome_role_id TEXT,
			
			-- Logging Settings
			logging_enabled BOOLEAN DEFAULT FALSE,
			log_channel_id TEXT,
			log_message_edits BOOLEAN DEFAULT TRUE,
			log_message_deletes BOOLEAN DEFAULT TRUE,
			log_member_joins BOOLEAN DEFAULT TRUE,
			log_member_leaves BOOLEAN DEFAULT TRUE,
			log_channel_events BOOLEAN DEFAULT FALSE,
			log_role_events BOOLEAN DEFAULT FALSE,
			log_voice_events BOOLEAN DEFAULT FALSE,
			log_moderation_events BOOLEAN DEFAULT FALSE,
			log_server_events BOOLEAN DEFAULT FALSE,
			log_nickname_changes BOOLEAN DEFAULT FALSE,
			
			-- Bump Settings
			bump_enabled BOOLEAN DEFAULT FALSE,
			bump_channel_id TEXT,
			bump_role_id TEXT,
			bump_last_time DATETIME,
			bump_reminder_sent BOOLEAN DEFAULT FALSE,
			
			-- General Settings
			settings_json TEXT,
			
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (guild_id) REFERENCES guilds(id)
		)`,
		`CREATE TABLE IF NOT EXISTS tickets (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			guild_id TEXT NOT NULL,
			channel_id TEXT NOT NULL UNIQUE,
			creator_id TEXT NOT NULL,
			assigned_id TEXT,
			category TEXT DEFAULT 'general',
			title TEXT NOT NULL,
			description TEXT,
			status TEXT DEFAULT 'open',
			priority INTEGER DEFAULT 1,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			closed_at DATETIME,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (guild_id) REFERENCES guilds(id),
			FOREIGN KEY (creator_id) REFERENCES users(id),
			FOREIGN KEY (assigned_id) REFERENCES users(id)
		)`,
		`CREATE TABLE IF NOT EXISTS ticket_messages (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			ticket_id INTEGER NOT NULL,
			user_id TEXT NOT NULL,
			message_id TEXT NOT NULL,
			content TEXT,
			attachments_json TEXT,
			message_type TEXT DEFAULT 'user',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (ticket_id) REFERENCES tickets(id),
			FOREIGN KEY (user_id) REFERENCES users(id)
		)`,
		`CREATE INDEX IF NOT EXISTS idx_command_usage_guild ON command_usage(guild_id)`,
		`CREATE INDEX IF NOT EXISTS idx_command_usage_user ON command_usage(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_command_usage_command ON command_usage(command)`,
		`CREATE INDEX IF NOT EXISTS idx_command_usage_executed ON command_usage(executed_at)`,
		`CREATE INDEX IF NOT EXISTS idx_tickets_guild ON tickets(guild_id)`,
		`CREATE INDEX IF NOT EXISTS idx_tickets_creator ON tickets(creator_id)`,
		`CREATE INDEX IF NOT EXISTS idx_tickets_status ON tickets(status)`,
		`CREATE INDEX IF NOT EXISTS idx_ticket_messages_ticket ON ticket_messages(ticket_id)`,
		// Migration: Drop and recreate bracket_usage table with new structure
		`DROP TABLE IF EXISTS bracket_usage`,
		`CREATE TABLE IF NOT EXISTS bracket_usage (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			guild_id TEXT NOT NULL,
			user_id TEXT NOT NULL,
			half_width_pairs INTEGER DEFAULT 0,
			full_width_pairs INTEGER DEFAULT 0,
			total_pairs INTEGER DEFAULT 0,
			last_updated DATETIME DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(guild_id, user_id),
			FOREIGN KEY (guild_id) REFERENCES guilds(id),
			FOREIGN KEY (user_id) REFERENCES users(id)
		)`,
		`CREATE INDEX IF NOT EXISTS idx_bracket_usage_guild_new ON bracket_usage(guild_id)`,
		`CREATE INDEX IF NOT EXISTS idx_bracket_usage_total_new ON bracket_usage(total_pairs DESC)`,
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

// Guild Settings Management

type GuildSettings struct {
	GuildID string `json:"guild_id"`
	
	// Ticket System
	TicketEnabled           bool   `json:"ticket_enabled"`
	TicketCategoryID        string `json:"ticket_category_id"`
	TicketSupportRoleID     string `json:"ticket_support_role_id"`
	TicketAdminRoleID       string `json:"ticket_admin_role_id"`
	TicketLogChannelID      string `json:"ticket_log_channel_id"`
	TicketTranscriptChannelID string `json:"ticket_transcript_channel_id"`
	TicketAutoCloseHours    int    `json:"ticket_auto_close_hours"`
	TicketMaxPerUser        int    `json:"ticket_max_per_user"`
	
	// Moderation
	ModerationEnabled    bool   `json:"moderation_enabled"`
	ModerationLogChannelID string `json:"moderation_log_channel_id"`
	AutomodEnabled       bool   `json:"automod_enabled"`
	
	// Welcome System
	WelcomeEnabled   bool   `json:"welcome_enabled"`
	WelcomeChannelID string `json:"welcome_channel_id"`
	WelcomeMessage   string `json:"welcome_message"`
	WelcomeRoleID    string `json:"welcome_role_id"`
	
	// Logging
	LoggingEnabled      bool `json:"logging_enabled"`
	LogChannelID        string `json:"log_channel_id"`
	LogMessageEdits     bool `json:"log_message_edits"`
	LogMessageDeletes   bool `json:"log_message_deletes"`
	LogMemberJoins      bool `json:"log_member_joins"`
	LogMemberLeaves     bool `json:"log_member_leaves"`
	LogChannelEvents    bool `json:"log_channel_events"`
	LogRoleEvents       bool `json:"log_role_events"`
	LogVoiceEvents      bool `json:"log_voice_events"`
	LogModerationEvents bool `json:"log_moderation_events"`
	LogServerEvents     bool `json:"log_server_events"`
	LogNicknameChanges  bool `json:"log_nickname_changes"`
	
	// Bump Settings
	BumpEnabled        bool      `json:"bump_enabled"`
	BumpChannelID      string    `json:"bump_channel_id"`
	BumpRoleID         string    `json:"bump_role_id"`
	BumpLastTime       *time.Time `json:"bump_last_time"`
	BumpReminderSent   bool      `json:"bump_reminder_sent"`
	
	// Metadata
	SettingsJSON string    `json:"settings_json"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

func (s *Service) GetGuildSettings(guildID string) (*GuildSettings, error) {
	settings := &GuildSettings{GuildID: guildID}
	
	query := `
		SELECT 
			ticket_enabled, ticket_category_id, ticket_support_role_id, 
			ticket_admin_role_id, ticket_log_channel_id, ticket_transcript_channel_id,
			ticket_auto_close_hours, ticket_max_per_user,
			moderation_enabled, moderation_log_channel_id, automod_enabled,
			welcome_enabled, welcome_channel_id, welcome_message, welcome_role_id,
			logging_enabled, log_channel_id, log_message_edits, log_message_deletes,
			log_member_joins, log_member_leaves, log_channel_events, log_role_events,
			log_voice_events, log_moderation_events, log_server_events, log_nickname_changes,
			bump_enabled, bump_channel_id, bump_role_id, bump_last_time, bump_reminder_sent,
			settings_json, created_at, updated_at
		FROM guild_settings 
		WHERE guild_id = ?
	`
	
	err := s.db.QueryRow(query, guildID).Scan(
		&settings.TicketEnabled, &settings.TicketCategoryID, &settings.TicketSupportRoleID,
		&settings.TicketAdminRoleID, &settings.TicketLogChannelID, &settings.TicketTranscriptChannelID,
		&settings.TicketAutoCloseHours, &settings.TicketMaxPerUser,
		&settings.ModerationEnabled, &settings.ModerationLogChannelID, &settings.AutomodEnabled,
		&settings.WelcomeEnabled, &settings.WelcomeChannelID, &settings.WelcomeMessage, &settings.WelcomeRoleID,
		&settings.LoggingEnabled, &settings.LogChannelID, &settings.LogMessageEdits, &settings.LogMessageDeletes,
		&settings.LogMemberJoins, &settings.LogMemberLeaves, &settings.LogChannelEvents, &settings.LogRoleEvents,
		&settings.LogVoiceEvents, &settings.LogModerationEvents, &settings.LogServerEvents, &settings.LogNicknameChanges,
		&settings.BumpEnabled, &settings.BumpChannelID, &settings.BumpRoleID, &settings.BumpLastTime, &settings.BumpReminderSent,
		&settings.SettingsJSON, &settings.CreatedAt, &settings.UpdatedAt,
	)
	
	if err == sql.ErrNoRows {
		// Return default settings
		settings.TicketAutoCloseHours = 24
		settings.TicketMaxPerUser = 3
		settings.LogMessageEdits = true
		settings.LogMessageDeletes = true
		settings.LogMemberJoins = true
		settings.LogMemberLeaves = true
		return settings, nil
	}
	
	return settings, err
}

func (s *Service) UpsertGuildSettings(settings *GuildSettings) error {
	query := `
		INSERT INTO guild_settings (
			guild_id, ticket_enabled, ticket_category_id, ticket_support_role_id,
			ticket_admin_role_id, ticket_log_channel_id, ticket_transcript_channel_id,
			ticket_auto_close_hours, ticket_max_per_user,
			moderation_enabled, moderation_log_channel_id, automod_enabled,
			welcome_enabled, welcome_channel_id, welcome_message, welcome_role_id,
			logging_enabled, log_channel_id, log_message_edits, log_message_deletes,
			log_member_joins, log_member_leaves, log_channel_events, log_role_events,
			log_voice_events, log_moderation_events, log_server_events, log_nickname_changes,
			bump_enabled, bump_channel_id, bump_role_id, bump_last_time, bump_reminder_sent,
			settings_json
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(guild_id) DO UPDATE SET
			ticket_enabled = excluded.ticket_enabled,
			ticket_category_id = excluded.ticket_category_id,
			ticket_support_role_id = excluded.ticket_support_role_id,
			ticket_admin_role_id = excluded.ticket_admin_role_id,
			ticket_log_channel_id = excluded.ticket_log_channel_id,
			ticket_transcript_channel_id = excluded.ticket_transcript_channel_id,
			ticket_auto_close_hours = excluded.ticket_auto_close_hours,
			ticket_max_per_user = excluded.ticket_max_per_user,
			moderation_enabled = excluded.moderation_enabled,
			moderation_log_channel_id = excluded.moderation_log_channel_id,
			automod_enabled = excluded.automod_enabled,
			welcome_enabled = excluded.welcome_enabled,
			welcome_channel_id = excluded.welcome_channel_id,
			welcome_message = excluded.welcome_message,
			welcome_role_id = excluded.welcome_role_id,
			logging_enabled = excluded.logging_enabled,
			log_channel_id = excluded.log_channel_id,
			log_message_edits = excluded.log_message_edits,
			log_message_deletes = excluded.log_message_deletes,
			log_member_joins = excluded.log_member_joins,
			log_member_leaves = excluded.log_member_leaves,
			log_channel_events = excluded.log_channel_events,
			log_role_events = excluded.log_role_events,
			log_voice_events = excluded.log_voice_events,
			log_moderation_events = excluded.log_moderation_events,
			log_server_events = excluded.log_server_events,
			log_nickname_changes = excluded.log_nickname_changes,
			bump_enabled = excluded.bump_enabled,
			bump_channel_id = excluded.bump_channel_id,
			bump_role_id = excluded.bump_role_id,
			bump_last_time = excluded.bump_last_time,
			bump_reminder_sent = excluded.bump_reminder_sent,
			settings_json = excluded.settings_json,
			updated_at = CURRENT_TIMESTAMP
	`
	
	// Ensure SettingsJSON is not nil
	settingsJSON := settings.SettingsJSON
	if settingsJSON == "" {
		settingsJSON = "{}"
	}

	_, err := s.db.Exec(query,
		settings.GuildID, settings.TicketEnabled, settings.TicketCategoryID, settings.TicketSupportRoleID,
		settings.TicketAdminRoleID, settings.TicketLogChannelID, settings.TicketTranscriptChannelID,
		settings.TicketAutoCloseHours, settings.TicketMaxPerUser,
		settings.ModerationEnabled, settings.ModerationLogChannelID, settings.AutomodEnabled,
		settings.WelcomeEnabled, settings.WelcomeChannelID, settings.WelcomeMessage, settings.WelcomeRoleID,
		settings.LoggingEnabled, settings.LogChannelID, settings.LogMessageEdits, settings.LogMessageDeletes,
		settings.LogMemberJoins, settings.LogMemberLeaves, settings.LogChannelEvents, settings.LogRoleEvents,
		settings.LogVoiceEvents, settings.LogModerationEvents, settings.LogServerEvents, settings.LogNicknameChanges,
		settings.BumpEnabled, settings.BumpChannelID, settings.BumpRoleID, settings.BumpLastTime, settings.BumpReminderSent,
		settingsJSON,
	)
	
	if err != nil {
		log.Printf("Database error in UpsertGuildSettings: %v", err)
		log.Printf("Guild ID: %s", settings.GuildID)
	}
	
	return err
}

func (s *Service) ResetGuildSettings(guildID, feature string) error {
	var query string
	
	switch feature {
	case "tickets":
		query = `
			UPDATE guild_settings SET
				ticket_enabled = FALSE,
				ticket_category_id = NULL,
				ticket_support_role_id = NULL,
				ticket_admin_role_id = NULL,
				ticket_log_channel_id = NULL,
				ticket_transcript_channel_id = NULL,
				ticket_auto_close_hours = 24,
				ticket_max_per_user = 3,
				updated_at = CURRENT_TIMESTAMP
			WHERE guild_id = ?
		`
	case "moderation":
		query = `
			UPDATE guild_settings SET
				moderation_enabled = FALSE,
				moderation_log_channel_id = NULL,
				automod_enabled = FALSE,
				updated_at = CURRENT_TIMESTAMP
			WHERE guild_id = ?
		`
	case "welcome":
		query = `
			UPDATE guild_settings SET
				welcome_enabled = FALSE,
				welcome_channel_id = NULL,
				welcome_message = NULL,
				welcome_role_id = NULL,
				updated_at = CURRENT_TIMESTAMP
			WHERE guild_id = ?
		`
	case "logging":
		query = `
			UPDATE guild_settings SET
				logging_enabled = FALSE,
				log_channel_id = NULL,
				log_message_edits = TRUE,
				log_message_deletes = TRUE,
				log_member_joins = TRUE,
				log_member_leaves = TRUE,
				updated_at = CURRENT_TIMESTAMP
			WHERE guild_id = ?
		`
	case "all":
		query = `DELETE FROM guild_settings WHERE guild_id = ?`
	default:
		return fmt.Errorf("unknown feature: %s", feature)
	}
	
	_, err := s.db.Exec(query, guildID)
	return err
}

// Bump関連のメソッド
func (s *Service) UpdateBumpTime(guildID string) error {
	query := `
		UPDATE guild_settings SET
			bump_last_time = CURRENT_TIMESTAMP,
			bump_reminder_sent = FALSE,
			updated_at = CURRENT_TIMESTAMP
		WHERE guild_id = ?
	`
	_, err := s.db.Exec(query, guildID)
	return err
}

func (s *Service) MarkBumpReminderSent(guildID string) error {
	query := `
		UPDATE guild_settings SET
			bump_reminder_sent = TRUE,
			updated_at = CURRENT_TIMESTAMP
		WHERE guild_id = ?
	`
	_, err := s.db.Exec(query, guildID)
	return err
}

func (s *Service) GetBumpableGuilds() ([]*GuildSettings, error) {
	query := `
		SELECT guild_id, bump_enabled, bump_channel_id, bump_role_id, 
		       bump_last_time, bump_reminder_sent
		FROM guild_settings
		WHERE bump_enabled = TRUE 
		AND bump_last_time IS NOT NULL 
		AND bump_reminder_sent = FALSE
		AND datetime(bump_last_time, '+2 hours') <= datetime('now')
	`
	
	rows, err := s.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var guilds []*GuildSettings
	for rows.Next() {
		settings := &GuildSettings{}
		err := rows.Scan(
			&settings.GuildID,
			&settings.BumpEnabled,
			&settings.BumpChannelID,
			&settings.BumpRoleID,
			&settings.BumpLastTime,
			&settings.BumpReminderSent,
		)
		if err != nil {
			continue
		}
		guilds = append(guilds, settings)
	}
	
	return guilds, rows.Err()
}

// Bracket usage methods
func (s *Service) UpdateBracketUsage(guildID, userID string, halfWidthPairs, fullWidthPairs int) error {
	totalPairs := halfWidthPairs + fullWidthPairs
	query := `
		INSERT INTO bracket_usage (guild_id, user_id, half_width_pairs, full_width_pairs, total_pairs, last_updated)
		VALUES (?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
		ON CONFLICT(guild_id, user_id) DO UPDATE SET
			half_width_pairs = half_width_pairs + ?,
			full_width_pairs = full_width_pairs + ?,
			total_pairs = total_pairs + ?,
			last_updated = CURRENT_TIMESTAMP
	`
	_, err := s.db.Exec(query, guildID, userID, halfWidthPairs, fullWidthPairs, totalPairs, 
		halfWidthPairs, fullWidthPairs, totalPairs)
	return err
}

func (s *Service) GetBracketRanking(guildID string, limit int) ([]BracketStats, error) {
	query := `
		SELECT user_id, half_width_pairs, full_width_pairs, total_pairs
		FROM bracket_usage
		WHERE guild_id = ?
		ORDER BY total_pairs DESC
		LIMIT ?
	`
	
	rows, err := s.db.Query(query, guildID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var rankings []BracketStats
	for rows.Next() {
		var stats BracketStats
		err := rows.Scan(&stats.UserID, &stats.HalfWidthPairs, &stats.FullWidthPairs, &stats.TotalPairs)
		if err != nil {
			continue
		}
		rankings = append(rankings, stats)
	}
	
	return rankings, rows.Err()
}

func (s *Service) GetUserBracketStats(guildID, userID string) (*BracketStats, error) {
	query := `
		SELECT half_width_pairs, full_width_pairs, total_pairs
		FROM bracket_usage
		WHERE guild_id = ? AND user_id = ?
	`
	
	var stats BracketStats
	stats.UserID = userID
	
	err := s.db.QueryRow(query, guildID, userID).Scan(
		&stats.HalfWidthPairs, &stats.FullWidthPairs, &stats.TotalPairs,
	)
	
	if err == sql.ErrNoRows {
		return &BracketStats{UserID: userID}, nil
	}
	
	return &stats, err
}

type BracketStats struct {
	UserID         string
	HalfWidthPairs int
	FullWidthPairs int
	TotalPairs     int
}