package repositories

import (
	"database/sql"
	"time"
)

// GuildSettings represents all configuration settings for a guild
type GuildSettings struct {
	GuildID            string
	Prefix             string
	LogChannelID       sql.NullString
	WelcomeChannelID   sql.NullString
	WelcomeMessage     sql.NullString
	AutoRole           sql.NullString
	ModRoleID          sql.NullString
	// Ticket system
	TicketCategoryID   sql.NullString
	TicketSupportRole  sql.NullString
	TicketAdminRole    sql.NullString
	TicketLogChannel   sql.NullString
	TicketAutoClose    sql.NullInt64
	// Bump settings
	BumpChannelID      sql.NullString
	BumpRoleID         sql.NullString
	BumpMessage        sql.NullString
	BumpReminderHours  sql.NullInt64
	BumpPingEnabled    sql.NullBool
	LastBumpTime       sql.NullTime
	BumpReminderSent   sql.NullBool
	UpdatedAt          time.Time
}

// GuildRepository handles guild-related database operations
type GuildRepository struct {
	*BaseRepository
}

// NewGuildRepository creates a new guild repository
func NewGuildRepository(db *sql.DB) *GuildRepository {
	return &GuildRepository{
		BaseRepository: NewBaseRepository(db),
	}
}

// UpsertGuild inserts or updates a guild
func (r *GuildRepository) UpsertGuild(id, name, prefix string) error {
	_, err := r.db.Exec(`
		INSERT INTO guilds (id, name, prefix, joined_at, updated_at)
		VALUES (?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		ON CONFLICT(id) DO UPDATE SET
			name = excluded.name,
			prefix = excluded.prefix,
			updated_at = CURRENT_TIMESTAMP
	`, id, name, prefix)
	return err
}

// GetGuildPrefix retrieves the custom prefix for a guild
func (r *GuildRepository) GetGuildPrefix(guildID string) (string, error) {
	var prefix string
	err := r.db.QueryRow(`
		SELECT COALESCE(
			(SELECT prefix FROM guild_settings WHERE guild_id = ?),
			(SELECT prefix FROM guilds WHERE id = ?),
			'/'
		)
	`, guildID, guildID).Scan(&prefix)
	
	if err == sql.ErrNoRows {
		return "/", nil
	}
	return prefix, err
}

// GetGuildSettings retrieves all settings for a guild
func (r *GuildRepository) GetGuildSettings(guildID string) (*GuildSettings, error) {
	settings := &GuildSettings{GuildID: guildID}
	
	err := r.db.QueryRow(`
		SELECT 
			COALESCE(prefix, '/'),
			log_channel_id,
			welcome_channel_id,
			welcome_message,
			auto_role,
			mod_role_id,
			ticket_category_id,
			ticket_support_role,
			ticket_admin_role,
			ticket_log_channel,
			ticket_auto_close,
			bump_channel_id,
			bump_role_id,
			bump_message,
			bump_reminder_hours,
			bump_ping_enabled,
			last_bump_time,
			bump_reminder_sent,
			updated_at
		FROM guild_settings
		WHERE guild_id = ?
	`, guildID).Scan(
		&settings.Prefix,
		&settings.LogChannelID,
		&settings.WelcomeChannelID,
		&settings.WelcomeMessage,
		&settings.AutoRole,
		&settings.ModRoleID,
		&settings.TicketCategoryID,
		&settings.TicketSupportRole,
		&settings.TicketAdminRole,
		&settings.TicketLogChannel,
		&settings.TicketAutoClose,
		&settings.BumpChannelID,
		&settings.BumpRoleID,
		&settings.BumpMessage,
		&settings.BumpReminderHours,
		&settings.BumpPingEnabled,
		&settings.LastBumpTime,
		&settings.BumpReminderSent,
		&settings.UpdatedAt,
	)
	
	if err == sql.ErrNoRows {
		// Return default settings if none exist
		return &GuildSettings{
			GuildID: guildID,
			Prefix:  "/",
		}, nil
	}
	
	return settings, err
}

// UpsertGuildSettings inserts or updates guild settings
func (r *GuildRepository) UpsertGuildSettings(settings *GuildSettings) error {
	_, err := r.db.Exec(`
		INSERT INTO guild_settings (
			guild_id, prefix, log_channel_id, welcome_channel_id, welcome_message,
			auto_role, mod_role_id, ticket_category_id, ticket_support_role,
			ticket_admin_role, ticket_log_channel, ticket_auto_close,
			bump_channel_id, bump_role_id, bump_message, bump_reminder_hours,
			bump_ping_enabled, last_bump_time, bump_reminder_sent, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
		ON CONFLICT(guild_id) DO UPDATE SET
			prefix = COALESCE(excluded.prefix, guild_settings.prefix),
			log_channel_id = COALESCE(excluded.log_channel_id, guild_settings.log_channel_id),
			welcome_channel_id = COALESCE(excluded.welcome_channel_id, guild_settings.welcome_channel_id),
			welcome_message = COALESCE(excluded.welcome_message, guild_settings.welcome_message),
			auto_role = COALESCE(excluded.auto_role, guild_settings.auto_role),
			mod_role_id = COALESCE(excluded.mod_role_id, guild_settings.mod_role_id),
			ticket_category_id = COALESCE(excluded.ticket_category_id, guild_settings.ticket_category_id),
			ticket_support_role = COALESCE(excluded.ticket_support_role, guild_settings.ticket_support_role),
			ticket_admin_role = COALESCE(excluded.ticket_admin_role, guild_settings.ticket_admin_role),
			ticket_log_channel = COALESCE(excluded.ticket_log_channel, guild_settings.ticket_log_channel),
			ticket_auto_close = COALESCE(excluded.ticket_auto_close, guild_settings.ticket_auto_close),
			bump_channel_id = COALESCE(excluded.bump_channel_id, guild_settings.bump_channel_id),
			bump_role_id = COALESCE(excluded.bump_role_id, guild_settings.bump_role_id),
			bump_message = COALESCE(excluded.bump_message, guild_settings.bump_message),
			bump_reminder_hours = COALESCE(excluded.bump_reminder_hours, guild_settings.bump_reminder_hours),
			bump_ping_enabled = COALESCE(excluded.bump_ping_enabled, guild_settings.bump_ping_enabled),
			last_bump_time = COALESCE(excluded.last_bump_time, guild_settings.last_bump_time),
			bump_reminder_sent = COALESCE(excluded.bump_reminder_sent, guild_settings.bump_reminder_sent),
			updated_at = CURRENT_TIMESTAMP
	`,
		settings.GuildID, settings.Prefix, settings.LogChannelID, settings.WelcomeChannelID,
		settings.WelcomeMessage, settings.AutoRole, settings.ModRoleID,
		settings.TicketCategoryID, settings.TicketSupportRole, settings.TicketAdminRole,
		settings.TicketLogChannel, settings.TicketAutoClose,
		settings.BumpChannelID, settings.BumpRoleID, settings.BumpMessage,
		settings.BumpReminderHours, settings.BumpPingEnabled,
		settings.LastBumpTime, settings.BumpReminderSent,
	)
	return err
}

// ResetGuildSettings resets specific guild settings to defaults
func (r *GuildRepository) ResetGuildSettings(guildID, feature string) error {
	var query string
	
	switch feature {
	case "moderation":
		query = `UPDATE guild_settings SET mod_role_id = NULL WHERE guild_id = ?`
	case "welcome":
		query = `UPDATE guild_settings SET welcome_channel_id = NULL, welcome_message = NULL, auto_role = NULL WHERE guild_id = ?`
	case "logging":
		query = `UPDATE guild_settings SET log_channel_id = NULL WHERE guild_id = ?`
	case "tickets":
		query = `UPDATE guild_settings SET ticket_category_id = NULL, ticket_support_role = NULL, 
				ticket_admin_role = NULL, ticket_log_channel = NULL, ticket_auto_close = NULL WHERE guild_id = ?`
	case "bump":
		query = `UPDATE guild_settings SET bump_channel_id = NULL, bump_role_id = NULL, 
				bump_message = NULL, bump_reminder_hours = NULL, bump_ping_enabled = NULL,
				last_bump_time = NULL, bump_reminder_sent = NULL WHERE guild_id = ?`
	case "all":
		query = `DELETE FROM guild_settings WHERE guild_id = ?`
	default:
		return nil
	}
	
	_, err := r.db.Exec(query, guildID)
	return err
}